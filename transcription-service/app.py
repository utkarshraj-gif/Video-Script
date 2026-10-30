import os
import time
import uuid
from typing import Optional, List, Dict, Any
from fastapi import FastAPI, HTTPException, status
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from faster_whisper import WhisperModel
from fastembed import TextEmbedding
from qdrant_client import QdrantClient
from qdrant_client.models import Distance, VectorParams, PointStruct, Filter, FieldCondition, MatchValue

app = FastAPI(
    title="Video RAG - Transcription & Embedding Service",
    description="Microservice providing faster-whisper speech-to-text, 384-dim embeddings, and Qdrant vector indexing",
    version="2.3.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Configuration from environment variables
WHISPER_MODEL_SIZE = os.getenv("WHISPER_MODEL", "base")
DEVICE = os.getenv("WHISPER_DEVICE", "cpu")
COMPUTE_TYPE = os.getenv("WHISPER_COMPUTE_TYPE", "int8")
EMBEDDING_MODEL_NAME = os.getenv("EMBEDDING_MODEL", "sentence-transformers/all-MiniLM-L6-v2")
QDRANT_URL = os.getenv("QDRANT_URL", "http://qdrant:6333")
COLLECTION_NAME = "video_chunks"

whisper_model: Optional[WhisperModel] = None
embedding_model: Optional[TextEmbedding] = None
_qdrant_client_instance: Optional[QdrantClient] = None

def get_whisper_model() -> WhisperModel:
    global whisper_model
    if whisper_model is None:
        print(f"Loading faster-whisper model: {WHISPER_MODEL_SIZE} on {DEVICE} ({COMPUTE_TYPE})...")
        whisper_model = WhisperModel(WHISPER_MODEL_SIZE, device=DEVICE, compute_type=COMPUTE_TYPE)
        print("Faster-whisper model loaded successfully!")
    return whisper_model

def get_embedding_model() -> TextEmbedding:
    global embedding_model
    if embedding_model is None:
        print(f"Loading embedding model: {EMBEDDING_MODEL_NAME} (384 dimensions)...")
        embedding_model = TextEmbedding(model_name=EMBEDDING_MODEL_NAME)
        print("Embedding model loaded successfully!")
    return embedding_model

def get_qdrant_client() -> QdrantClient:
    global _qdrant_client_instance
    if _qdrant_client_instance is not None:
        return _qdrant_client_instance

    # Connect directly to Qdrant server (http://qdrant:6333)
    for attempt in range(5):
        try:
            client = QdrantClient(url=QDRANT_URL, timeout=10)
            client.get_collections()
            _qdrant_client_instance = client
            print(f"Successfully connected to Qdrant server at {QDRANT_URL}")
            return _qdrant_client_instance
        except Exception as e:
            if attempt < 4:
                print(f"Waiting for Qdrant server at {QDRANT_URL} (attempt {attempt+1}/5): {e}")
                time.sleep(1)

    # Fallback to direct client creation
    client = QdrantClient(url=QDRANT_URL, timeout=15)
    _qdrant_client_instance = client
    return _qdrant_client_instance

def ensure_qdrant_collection():
    try:
        client = get_qdrant_client()
        collections = [c.name for c in client.get_collections().collections]
        if COLLECTION_NAME not in collections:
            client.create_collection(
                collection_name=COLLECTION_NAME,
                vectors_config=VectorParams(size=384, distance=Distance.COSINE)
            )
            print(f"Created Qdrant collection '{COLLECTION_NAME}' with 384 dimensions (Cosine)")
        return True
    except Exception as e:
        print(f"Warning ensuring Qdrant collection at {QDRANT_URL}: {e}")
        return False

def execute_qdrant_search(client: QdrantClient, query_vector: list, query_filter, limit: int):
    """Executes vector search compatible across all qdrant-client versions."""
    if hasattr(client, "query_points"):
        resp = client.query_points(
            collection_name=COLLECTION_NAME,
            query=query_vector,
            query_filter=query_filter,
            limit=limit
        )
        return resp.points if hasattr(resp, "points") else resp

    if hasattr(client, "search"):
        return client.search(
            collection_name=COLLECTION_NAME,
            query_vector=query_vector,
            query_filter=query_filter,
            limit=limit
        )

    if hasattr(client, "search_points"):
        return client.search_points(
            collection_name=COLLECTION_NAME,
            vector=query_vector,
            filter=query_filter,
            limit=limit
        )

    raise RuntimeError("No compatible search method found on QdrantClient instance")

@app.on_event("startup")
async def startup_event():
    try:
        get_whisper_model()
    except Exception as e:
        print(f"Warning: Whisper model deferred: {e}")
    try:
        get_embedding_model()
    except Exception as e:
        print(f"Warning: Embedding model deferred: {e}")
    try:
        ensure_qdrant_collection()
    except Exception as e:
        print(f"Warning: Qdrant collection deferred: {e}")

# Helpers
def format_seconds(seconds: float) -> str:
    mins = int(seconds // 60)
    secs = int(seconds % 60)
    return f"{mins:02d}:{secs:02d}"

# Pydantic Models
class SegmentItem(BaseModel):
    id: int
    start: float
    end: float
    text: str

class TranscribeRequest(BaseModel):
    audio_path: Optional[str] = None
    video_id: Optional[str] = None
    language: Optional[str] = None

class TranscribeResponse(BaseModel):
    status: str
    video_id: str
    model: str
    detected_language: str
    language_probability: float
    duration_seconds: float
    transcription_time_seconds: float
    full_transcript: str
    segments_count: int
    segments: List[SegmentItem]

class ChunkItem(BaseModel):
    chunk_id: str
    video_id: str
    chunk_index: int
    start_time: float
    end_time: float
    timestamp_display: str
    text: str
    word_count: int
    segment_ids: List[int]

class ChunkWithEmbedding(BaseModel):
    chunk_id: str
    video_id: str
    chunk_index: int
    start_time: float
    end_time: float
    timestamp_display: str
    text: str
    word_count: int
    segment_ids: List[int]
    embedding_dimensions: int
    embedding_preview: List[float]
    embedding_vector: Optional[List[float]] = None
    stored_in_qdrant: bool

class EmbedChunksRequest(BaseModel):
    video_id: str
    segments: Optional[List[SegmentItem]] = None
    full_transcript: Optional[str] = None
    target_words: Optional[int] = 120
    overlap_segments: Optional[int] = 1
    store_in_qdrant: Optional[bool] = True
    include_full_vector: Optional[bool] = True

class EmbedChunksResponse(BaseModel):
    status: str
    video_id: str
    embedding_model: str
    embedding_dimensions: int
    qdrant_server: str
    collection_name: str
    qdrant_stored: bool
    total_chunks: int
    chunks: List[ChunkWithEmbedding]
    message: str

class SearchRequest(BaseModel):
    query: str
    video_id: Optional[str] = None
    limit: Optional[int] = 3

class SearchMatch(BaseModel):
    score: float
    chunk_id: str
    video_id: str
    start_time: float
    end_time: float
    timestamp_display: str
    text: str
    youtube_timestamp_url: Optional[str] = None

class SearchResponse(BaseModel):
    status: str
    query: str
    matches_count: int
    matches: List[SearchMatch]

# Core Chunking Function
def chunk_segments(
    video_id: str,
    segments: List[SegmentItem],
    target_words: int = 120,
    overlap_segments: int = 1
) -> List[ChunkItem]:
    chunks: List[ChunkItem] = []
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
            seg_words = len(seg.text.split())
            current_segments.append(seg)
            current_words += seg_words
            j += 1

            if current_words >= target_words and len(current_segments) >= 2:
                break

        if not current_segments:
            break

        start_time = current_segments[0].start
        end_time = current_segments[-1].end
        chunk_text = " ".join(s.text.strip() for s in current_segments)
        timestamp_display = f"{format_seconds(start_time)} - {format_seconds(end_time)}"

        chunks.append(ChunkItem(
            chunk_id=f"{video_id}_chunk_{chunk_idx}",
            video_id=video_id,
            chunk_index=chunk_idx,
            start_time=round(start_time, 2),
            end_time=round(end_time, 2),
            timestamp_display=timestamp_display,
            text=chunk_text,
            word_count=len(chunk_text.split()),
            segment_ids=[s.id for s in current_segments]
        ))
        chunk_idx += 1

        if j >= n:
            break
        i = max(i + 1, j - overlap_segments)

    return chunks

# Endpoints
@app.get("/health")
def health_check():
    client = get_qdrant_client()
    qdrant_ok = False
    try:
        client.get_collections()
        qdrant_ok = True
    except Exception:
        qdrant_ok = False

    return {
        "status": "success",
        "service": "whisper-and-embedding-service",
        "whisper_model": WHISPER_MODEL_SIZE,
        "embedding_model": EMBEDDING_MODEL_NAME,
        "embedding_dimensions": 384,
        "qdrant_connected": qdrant_ok,
        "qdrant_url": QDRANT_URL,
        "collection": COLLECTION_NAME
    }

@app.post("/transcribe", response_model=TranscribeResponse)
def transcribe_audio(req: TranscribeRequest):
    audio_path = req.audio_path
    video_id = req.video_id or "unknown"

    if not audio_path and video_id:
        downloads_dir = "/app/downloads"
        candidates = [
            os.path.join(downloads_dir, f"{video_id}.mp3"),
            os.path.join(downloads_dir, f"{video_id}.m4a"),
            os.path.join(downloads_dir, f"{video_id}.wav"),
            os.path.join(downloads_dir, f"{video_id}.opus")
        ]
        for c in candidates:
            if os.path.isfile(c):
                audio_path = c
                break

    if not audio_path:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Either audio_path or a valid video_id with downloaded audio is required."
        )

    if not os.path.isabs(audio_path):
        audio_path = os.path.join("/app", audio_path)

    if not os.path.isfile(audio_path):
        alt_path = os.path.join("/app/downloads", os.path.basename(audio_path))
        if os.path.isfile(alt_path):
            audio_path = alt_path
        else:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Audio file not found at '{audio_path}'"
            )

    try:
        model = get_whisper_model()
        start_time = time.time()

        segments_generator, info = model.transcribe(
            audio_path,
            beam_size=5,
            language=req.language,
            vad_filter=True
        )

        segments = []
        full_text_parts = []
        for i, segment in enumerate(segments_generator):
            cleaned_text = segment.text.strip()
            if cleaned_text:
                segments.append(SegmentItem(
                    id=i,
                    start=round(segment.start, 2),
                    end=round(segment.end, 2),
                    text=cleaned_text
                ))
                full_text_parts.append(cleaned_text)

        transcription_time = round(time.time() - start_time, 2)
        full_transcript = " ".join(full_text_parts)

        return TranscribeResponse(
            status="success",
            video_id=video_id,
            model=WHISPER_MODEL_SIZE,
            detected_language=info.language,
            language_probability=round(info.language_probability, 3),
            duration_seconds=round(info.duration, 2),
            transcription_time_seconds=transcription_time,
            full_transcript=full_transcript,
            segments_count=len(segments),
            segments=segments
        )
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Transcription failed: {str(e)}"
        )

@app.post("/embed-chunks", response_model=EmbedChunksResponse)
def embed_chunks_endpoint(req: EmbedChunksRequest):
    video_id = req.video_id
    if not video_id:
        raise HTTPException(status_code=400, detail="video_id is required")

    segments = req.segments or []
    if not segments and req.full_transcript:
        segments = [SegmentItem(id=0, start=0.0, end=0.0, text=req.full_transcript)]

    if not segments:
        raise HTTPException(status_code=400, detail="segments or full_transcript is required.")

    # 1. Chunk segments
    target_words = req.target_words or 120
    overlap_segs = req.overlap_segments if req.overlap_segments is not None else 1
    chunks = chunk_segments(video_id, segments, target_words, overlap_segs)

    if not chunks:
        raise HTTPException(status_code=400, detail="Could not produce chunks.")

    # 2. Compute 384-dimensional embeddings
    chunk_texts = [c.text for c in chunks]
    embed_model = get_embedding_model()
    embeddings = [emb.tolist() for emb in embed_model.embed(chunk_texts)]

    # 3. Store in Qdrant server
    store_in_qdrant = req.store_in_qdrant if req.store_in_qdrant is not None else True
    qdrant_stored = False

    if store_in_qdrant:
        ensure_qdrant_collection()
        client = get_qdrant_client()
        points = []
        for chunk, emb in zip(chunks, embeddings):
            point_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, f"{video_id}_{chunk.chunk_id}"))
            points.append(PointStruct(
                id=point_id,
                vector=emb,
                payload={
                    "video_id": chunk.video_id,
                    "chunk_id": chunk.chunk_id,
                    "chunk_index": chunk.chunk_index,
                    "start_time": chunk.start_time,
                    "end_time": chunk.end_time,
                    "timestamp_display": chunk.timestamp_display,
                    "text": chunk.text,
                    "word_count": chunk.word_count,
                    "segment_ids": chunk.segment_ids
                }
            ))

        last_upsert_err = None
        for attempt in range(3):
            try:
                client.upsert(collection_name=COLLECTION_NAME, points=points)
                qdrant_stored = True
                print(f"Successfully upserted {len(points)} points to Qdrant collection '{COLLECTION_NAME}'")
                break
            except Exception as e:
                last_upsert_err = e
                print(f"Retrying upsert to Qdrant (attempt {attempt+1}/3): {e}")
                time.sleep(1)

        if not qdrant_stored:
            raise HTTPException(
                status_code=500,
                detail=f"Failed to upsert points to Qdrant at {QDRANT_URL}: {str(last_upsert_err)}"
            )

    # 4. Construct response with embeddings
    include_full = req.include_full_vector if req.include_full_vector is not None else True
    result_chunks = []
    for chunk, emb in zip(chunks, embeddings):
        preview = [round(v, 4) for v in emb[:5]]
        full_vec = [round(v, 4) for v in emb] if include_full else None

        result_chunks.append(ChunkWithEmbedding(
            chunk_id=chunk.chunk_id,
            video_id=chunk.video_id,
            chunk_index=chunk.chunk_index,
            start_time=chunk.start_time,
            end_time=chunk.end_time,
            timestamp_display=chunk.timestamp_display,
            text=chunk.text,
            word_count=chunk.word_count,
            segment_ids=chunk.segment_ids,
            embedding_dimensions=384,
            embedding_preview=preview,
            embedding_vector=full_vec,
            stored_in_qdrant=qdrant_stored
        ))

    return EmbedChunksResponse(
        status="success",
        video_id=video_id,
        embedding_model=EMBEDDING_MODEL_NAME,
        embedding_dimensions=384,
        qdrant_server=QDRANT_URL,
        collection_name=COLLECTION_NAME,
        qdrant_stored=qdrant_stored,
        total_chunks=len(result_chunks),
        chunks=result_chunks,
        message=f"Created {len(result_chunks)} chunks, generated 384-dim embeddings, and stored in Qdrant server ({QDRANT_URL})."
    )

# Alias for backwards compatibility
@app.post("/chunk-and-embed")
def chunk_and_embed_alias(req: EmbedChunksRequest):
    return embed_chunks_endpoint(req)

@app.post("/search", response_model=SearchResponse)
def search_chunks(req: SearchRequest):
    if not req.query.strip():
        raise HTTPException(status_code=400, detail="Search query cannot be empty.")

    client = get_qdrant_client()
    embed_model = get_embedding_model()

    query_vector = list(embed_model.embed([req.query]))[0].tolist()

    query_filter = None
    if req.video_id:
        query_filter = Filter(
            must=[FieldCondition(key="video_id", match=MatchValue(value=req.video_id))]
        )

    limit = req.limit or 3

    try:
        results = execute_qdrant_search(client, query_vector, query_filter, limit)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Qdrant vector search failed: {str(e)}")

    matches = []
    for hit in results:
        payload = getattr(hit, "payload", None) or {}
        score = getattr(hit, "score", 0.0)
        vid = payload.get("video_id", "")
        start_t = payload.get("start_time", 0.0)
        yt_url = f"https://youtu.be/{vid}?t={int(start_t)}" if vid else None

        matches.append(SearchMatch(
            score=round(score, 4),
            chunk_id=payload.get("chunk_id", ""),
            video_id=vid,
            start_time=start_t,
            end_time=payload.get("end_time", 0.0),
            timestamp_display=payload.get("timestamp_display", ""),
            text=payload.get("text", ""),
            youtube_timestamp_url=yt_url
        ))

    return SearchResponse(
        status="success",
        query=req.query,
        matches_count=len(matches),
        matches=matches
    )

@app.get("/qdrant-points")
def list_qdrant_points(video_id: Optional[str] = None, limit: int = 100):
    client = get_qdrant_client()
    ensure_qdrant_collection()
    query_filter = None
    if video_id:
        query_filter = Filter(
            must=[FieldCondition(key="video_id", match=MatchValue(value=video_id))]
        )
    try:
        points, _ = client.scroll(
            collection_name=COLLECTION_NAME,
            scroll_filter=query_filter,
            limit=limit,
            with_payload=True,
            with_vectors=True
        )
        formatted = []
        for p in points:
            vec = p.vector if p.vector is not None else []
            formatted.append({
                "id": str(p.id),
                "payload": p.payload,
                "vector_dimensions": len(vec),
                "vector_preview": [round(float(v), 4) for v in vec[:5]],
                "vector": [round(float(v), 4) for v in vec]
            })
        return {
            "status": "success",
            "collection_name": COLLECTION_NAME,
            "total_points": len(formatted),
            "video_id_filter": video_id,
            "points": formatted
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to scroll points from Qdrant: {str(e)}")
