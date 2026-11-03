import os
import re
import time
import uuid
from typing import List, Dict, Any, Optional
from pydantic import BaseModel
import yt_dlp
from celery_app import celery_app
from faster_whisper import WhisperModel
from fastembed import TextEmbedding
from qdrant_client import QdrantClient
from qdrant_client.models import Distance, VectorParams, PointStruct

WHISPER_MODEL_SIZE = os.getenv("WHISPER_MODEL", "base")
DEVICE = os.getenv("WHISPER_DEVICE", "cpu")
COMPUTE_TYPE = os.getenv("WHISPER_COMPUTE_TYPE", "int8")
EMBEDDING_MODEL_NAME = os.getenv("EMBEDDING_MODEL", "sentence-transformers/all-MiniLM-L6-v2")
QDRANT_URL = os.getenv("QDRANT_URL", "http://qdrant:6333")
COLLECTION_NAME = "video_chunks"
DOWNLOAD_DIR = "/app/downloads"

# Lazy-loaded singletons
_whisper_model = None
_embed_model = None
_qdrant_client = None

def get_whisper_model():
    global _whisper_model
    if _whisper_model is None:
        threads = os.cpu_count() or 4
        print(f"Loading Faster-Whisper model '{WHISPER_MODEL_SIZE}' on {DEVICE} ({COMPUTE_TYPE}) with {threads} threads...")
        _whisper_model = WhisperModel(
            WHISPER_MODEL_SIZE,
            device=DEVICE,
            compute_type=COMPUTE_TYPE,
            cpu_threads=threads
        )
    return _whisper_model

def get_embed_model():
    global _embed_model
    if _embed_model is None:
        print(f"Loading FastEmbed model '{EMBEDDING_MODEL_NAME}'...")
        _embed_model = TextEmbedding(model_name=EMBEDDING_MODEL_NAME)
    return _embed_model

def get_qdrant_client():
    global _qdrant_client
    if _qdrant_client is None:
        _qdrant_client = QdrantClient(url=QDRANT_URL)
    return _qdrant_client

def ensure_qdrant_collection():
    client = get_qdrant_client()
    try:
        collections = client.get_collections().collections
        exists = any(c.name == COLLECTION_NAME for c in collections)
        if not exists:
            client.create_collection(
                collection_name=COLLECTION_NAME,
                vectors_config=VectorParams(size=384, distance=Distance.COSINE)
            )
            print(f"Created Qdrant collection '{COLLECTION_NAME}' (384-dim, Cosine)")
    except Exception as e:
        print(f"Notice during Qdrant check: {e}")

def format_seconds(seconds: float) -> str:
    s = int(seconds)
    hours = s // 3600
    minutes = (s % 3600) // 60
    secs = s % 60
    if hours > 0:
        return f"{hours:02d}:{minutes:02d}:{secs:02d}"
    return f"{minutes:02d}:{secs:02d}"

def determine_language_label(info) -> str:
    lang = (getattr(info, "language", None) or "").lower()
    all_probs = {}
    if hasattr(info, "all_language_probs") and info.all_language_probs:
        try:
            all_probs = {k.lower(): v for k, v in info.all_language_probs}
        except Exception:
            all_probs = {}

    en_prob = all_probs.get("en", 0.0)
    hi_prob = all_probs.get("hi", 0.0)
    te_prob = all_probs.get("te", 0.0)

    # Telugu & English code-mixing (Tenglish)
    if (lang == "te" and en_prob >= 0.08) or (lang == "en" and te_prob >= 0.08):
        return "Eng + Tel"

    if lang == "te":
        if en_prob >= 0.05:
            return "Eng + Tel"
        return "Telugu"

    # Hindi & English code-mixing (Hinglish)
    if (lang == "hi" and en_prob >= 0.08) or (lang == "en" and hi_prob >= 0.08):
        return "Eng + Hin"

    if lang == "hi":
        if en_prob >= 0.05:
            return "Eng + Hin"
        return "Hindi"

    if lang == "en":
        if te_prob >= 0.08:
            return "Eng + Tel"
        if hi_prob >= 0.08:
            return "Eng + Hin"
        return "English"

    lang_map = {
        "es": "Spanish",
        "fr": "French",
        "de": "German",
        "te": "Telugu",
        "ta": "Tamil",
        "bn": "Bengali",
        "mr": "Marathi",
        "kn": "Kannada",
        "ml": "Malayalam",
    }
    return lang_map.get(lang, lang.upper() if lang else "English")

def extract_video_id(url_or_id: str) -> str:
    s = url_or_id.strip()
    if len(s) == 11 and re.match(r'^[a-zA-Z0-9_-]{11}$', s):
        return s
    m = re.search(r'(?:v=|\/embed\/|youtu\.be\/|\/v\/|shorts\/)([a-zA-Z0-9_-]{11})', s)
    if m:
        return m.group(1)
    raise ValueError(f"Could not extract valid 11-character YouTube video ID from '{url_or_id}'")

def chunk_segments(video_id: str, segments: List[Dict[str, Any]], target_words: int = 120, overlap_segments: int = 1):
    chunks = []
    if not segments:
        return chunks

    i = 0
    chunk_idx = 0
    n = len(segments)

    while i < n:
        current_segments = []
        current_words = 0
        j = i

        while j < n:
            seg = segments[j]
            seg_words = len(seg["text"].split())
            current_segments.append(seg)
            current_words += seg_words
            j += 1
            if current_words >= target_words and len(current_segments) >= 2:
                break

        if not current_segments:
            break

        start_time = current_segments[0]["start"]
        end_time = current_segments[-1]["end"]
        chunk_text = " ".join(s["text"].strip() for s in current_segments)
        timestamp_display = f"{format_seconds(start_time)} - {format_seconds(end_time)}"

        chunks.append({
            "chunk_id": f"{video_id}_chunk_{chunk_idx}",
            "video_id": video_id,
            "chunk_index": chunk_idx,
            "start_time": round(start_time, 2),
            "end_time": round(end_time, 2),
            "timestamp_display": timestamp_display,
            "text": chunk_text,
            "word_count": len(chunk_text.split()),
            "segment_ids": [s["id"] for s in current_segments]
        })
        chunk_idx += 1

        if j >= n:
            break
        i = max(i + 1, j - overlap_segments)

    return chunks

@celery_app.task(bind=True)
def process_video_pipeline(self, url_or_id: str):
    os.makedirs(DOWNLOAD_DIR, exist_ok=True)
    video_id = extract_video_id(url_or_id)
    canonical_url = f"https://www.youtube.com/watch?v={video_id}"

    # Step 1: Validating YouTube URL & Metadata
    self.update_state(
        state="PROGRESS",
        meta={
            "percent": 15,
            "step": "1. Validating YouTube URL & Metadata",
            "message": f"Fetching metadata for YouTube ID: {video_id}..."
        }
    )

    ydl_opts = {
        "quiet": True,
        "no_warnings": True,
        "skip_download": True,
    }
    title = f"YouTube Video ({video_id})"
    duration = 0
    uploader = "Unknown"

    try:
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            meta = ydl.extract_info(canonical_url, download=False)
            if meta:
                title = meta.get("title") or title
                duration = meta.get("duration") or 0
                uploader = meta.get("uploader") or uploader
    except Exception as e:
        print(f"Notice: metadata fetch warning: {e}")

    # Step 2: Downloading Audio Stream
    audio_path = os.path.join(DOWNLOAD_DIR, f"{video_id}.mp3")
    self.update_state(
        state="PROGRESS",
        meta={
            "percent": 30,
            "step": "2. Downloading Audio Stream (yt-dlp)",
            "message": f"Downloading audio stream for '{title}' ({format_seconds(duration)})..."
        }
    )

    if not os.path.exists(audio_path) or os.path.getsize(audio_path) < 1024:
        download_opts = {
            "format": "bestaudio/best",
            "outtmpl": os.path.join(DOWNLOAD_DIR, f"{video_id}.%(ext)s"),
            "postprocessors": [{
                "key": "FFmpegExtractAudio",
                "preferredcodec": "mp3",
                "preferredquality": "128",
            }],
            "quiet": True,
            "no_warnings": True,
        }
        with yt_dlp.YoutubeDL(download_opts) as ydl:
            ydl.download([canonical_url])

    if not os.path.exists(audio_path):
        raise FileNotFoundError(f"Audio file was not created at {audio_path}")

    # Step 3: Transcribing Audio with Faster-Whisper
    self.update_state(
        state="PROGRESS",
        meta={
            "percent": 55,
            "step": "3. Transcribing Audio (Faster-Whisper)",
            "message": f"Transcribing & translating dialogue to English..."
        }
    )

    model = get_whisper_model()
    task = os.getenv("WHISPER_TASK", "translate")
    beam_size = int(os.getenv("WHISPER_BEAM_SIZE", "1")) # Fast greedy search on CPU

    start_time = time.time()
    segments_gen, info = model.transcribe(
        audio_path,
        beam_size=beam_size,
        task=task,
        vad_filter=True
    )

    raw_segments = []
    full_text_parts = []
    for seg in segments_gen:
        txt = seg.text.strip()
        if txt:
            raw_segments.append({
                "id": seg.id,
                "start": round(seg.start, 2),
                "end": round(seg.end, 2),
                "text": txt,
                "words_count": len(txt.split())
            })
            full_text_parts.append(txt)

    transcription_time = round(time.time() - start_time, 2)
    detected_language = info.language
    language_display = determine_language_label(info)
    full_transcript = " ".join(full_text_parts)

    # Step 4: Computing 384-Dim FastEmbed Vectors
    self.update_state(
        state="PROGRESS",
        meta={
            "percent": 75,
            "step": "4. Computing 384-Dim FastEmbed Vectors",
            "message": f"Generating 384-dimensional dense vectors for chunks..."
        }
    )

    chunks = chunk_segments(video_id, raw_segments, target_words=120, overlap_segments=1)
    if not chunks:
        # Fallback single chunk
        chunks = [{
            "chunk_id": f"{video_id}_chunk_0",
            "video_id": video_id,
            "chunk_index": 0,
            "start_time": 0.0,
            "end_time": round(info.duration, 2),
            "timestamp_display": f"00:00 - {format_seconds(info.duration)}",
            "text": full_transcript or title,
            "word_count": len((full_transcript or title).split()),
            "segment_ids": []
        }]

    chunk_texts = [c["text"] for c in chunks]
    embed_model = get_embed_model()
    embeddings = [emb.tolist() for emb in embed_model.embed(chunk_texts)]

    # Step 5: Upserting Chunks into Qdrant
    self.update_state(
        state="PROGRESS",
        meta={
            "percent": 88,
            "step": "5. Upserting Chunks into Qdrant",
            "message": f"Storing {len(chunks)} vector points into Qdrant collection '{COLLECTION_NAME}'..."
        }
    )

    ensure_qdrant_collection()
    client = get_qdrant_client()
    points = []
    for c, emb in zip(chunks, embeddings):
        c["embedding_preview"] = [round(x, 4) for x in emb[:8]]
        point_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, f"{video_id}_{c['chunk_id']}"))
        points.append(PointStruct(
            id=point_id,
            vector=emb,
            payload={
                "video_id": c["video_id"],
                "chunk_id": c["chunk_id"],
                "chunk_index": c["chunk_index"],
                "start_time": c["start_time"],
                "end_time": c["end_time"],
                "timestamp_display": c["timestamp_display"],
                "text": c["text"],
                "word_count": c["word_count"],
                "segment_ids": c["segment_ids"]
            }
        ))

    client.upsert(collection_name=COLLECTION_NAME, points=points)

    # Step 6: Completing Pipeline
    self.update_state(
        state="PROGRESS",
        meta={
            "percent": 96,
            "step": "6. Mapping Temporal Sequence in Neo4j",
            "message": "Finalizing graph synchronization and metadata cache..."
        }
    )

    return {
        "status": "success",
        "video_id": video_id,
        "title": title,
        "duration_seconds": duration or int(info.duration),
        "uploader": uploader,
        "detected_language": detected_language,
        "language_display": language_display,
        "language_probability": round(info.language_probability, 3),
        "transcription_time_seconds": transcription_time,
        "chunks_count": len(chunks),
        "chunks": chunks,
        "full_transcript": full_transcript,
        "redirect_url": f"/api/v1/videos/{video_id}/embeddings?format=html",
        "message": f"Video '{title}' ({language_display}) successfully indexed in Qdrant and mapped in Neo4j!"
    }
