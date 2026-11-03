# Video-RAG Studio 🎬🧠

> **Containerized Multimodal Video Retrieval-Augmented Generation (RAG) System**  
> Powered by Celery & Redis, Faster-Whisper (Multi-Language Auto-Translation), FastEmbed (384-dim), Qdrant Vector DB, Neo4j Knowledge Graph, and Ollama Local LLMs.

---

## Introduction

### Overview
**Video-RAG Studio** is a full-stack, multimodal Retrieval-Augmented Generation (RAG) platform that transforms long-form video content (such as YouTube lectures, tutorials, and podcasts) into an interactive, context-aware AI assistant. 

Unlike traditional RAG systems that treat transcripts as flat, disconnected text, Video-RAG combines **dense vector similarity search**, **graph-based relational expansion**, **asynchronous task execution with Celery & Redis**, and **grounded local LLM synthesis**.

```
                           ┌────────────────────────┐
                           │   YouTube Video URL    │
                           └───────────┬────────────┘
                                       │ 
                                       ▼
                       ┌────────────────────────────────┐
                       │  Go API Gateway (Port 8080)    │
                       └───────────────┬────────────────┘
                                       │ (Enqueues Job in 10ms)
                                       ▼
                       ┌────────────────────────────────┐
                       │   Redis Task Broker (Port 6379)│
                       └───────────────┬────────────────┘
                                       │ (Background Task)
                                       ▼
                       ┌────────────────────────────────┐
                       │   Celery Distributed Worker    │
                       │  ────────────────────────────  │
                       │ 1. yt-dlp Audio Extraction     │
                       │ 2. Faster-Whisper ASR:         │
                       │    - Multi-Language Detection  │
                       │      (English, Hindi, Telugu,  │
                       │       Eng+Hin, Eng+Tel)        │
                       │    - Speech-to-English Trans.  │
                       │ 3. FastEmbed (384-dim vectors) │
                       └───────┬────────────────┬───────┘
             384-Dim Vectors   │                │ Chunks & Entities
                               ▼                ▼
                     ┌──────────────────┐  ┌──────────────────┐
                     │ Qdrant Vector DB │  │   Neo4j Graph    │
                     │  (Dense Search)  │  │ (Temporal Chain) │
                     └─────────┬────────┘  └────────┬─────────┘
                               │                    │
                               └─────────┬──────────┘
                                         ▼
                        ┌─────────────────────────────────┐
                        │   Grounded Multi-Chunk Context   │
                        └────────────────┬────────────────┘
                                         │
                                         ▼
                        ┌─────────────────────────────────┐
                        │    Ollama Local LLM (llama3.2)  │
                        │    Strict Grounded Synthesis    │
                        └────────────────┬────────────────┘
                                         │
                                         ▼
                        ┌─────────────────────────────────┐
                        │   Video-RAG Interactive Studio  │
                        │ (Chat, Visual Graph, MP3 Audio, │
                        │  Live Celery Progress Polling)  │
                        └─────────────────────────────────┘
```

### Key Capabilities & Architecture Highlights

1. **Asynchronous Processing via Celery & Redis (Zero Timeouts)**:
   * Video downloading, transcription, embedding generation, and graph construction are decoupled from the HTTP gateway.
   * Tasks are enqueued into **Redis** and processed by a dedicated **Celery Worker** using multithreaded CPU optimization and greedy beam search (`beam_size=1`).
   * Handles long-form videos (20 min, 1 hour, or more) without hitting HTTP gateway or browser connection timeouts.
   * Real-time UI polling (`GET /api/v1/tasks/:id`) smoothly drives a 6-stage modal progress bar (10% to 100%).

2. **First-Class Multi-Language Detection & Translation**:
   * **Supported Languages**: **English**, **Hindi**, **Telugu**, and code-mixed spoken formats (**Eng + Hin / Hinglish**, **Eng + Tel / Tenglish**).
   * **Faster-Whisper Translation (`task="translate"`)**: Automatically transcribes foreign speech and translates dialogue directly into unified English chunks while retaining millisecond-accurate timestamps.
   * **Unified 384-Dim Vector Space**: Translating into English ensures 100% semantic retrieval accuracy against `sentence-transformers/all-MiniLM-L6-v2` and eliminates cross-lingual vector degradation.
   * **Dynamic Language Badges**: Displays the exact language badge (`English`, `Hindi`, `Telugu`, `Eng + Hin`, `Eng + Tel`) in the video dashboard header.

3. **Preserve Temporal & Semantic Flow with Neo4j Knowledge Graph**:
   * Long-form videos have chronological progression. By pairing sequential relationships (`[:NEXT_CHUNK]`) and domain concepts (`[:DISCUSSES]`) in Neo4j with dense vectors in Qdrant, queries retrieve not only the best-matching moment, but also preceding context and subsequent outcomes.

4. **Grounded Ollama Synthesis (Zero Hallucination)**:
   * Combines Qdrant vector retrieval and Neo4j graph context into local Ollama models (`llama3.2` or `gemma4:12b`).
   * Strict prompt grounding forces the LLM to answer using exclusively facts stated in the video, accompanied by clickable timestamps.

5. **Interactive All-in-One Studio**:
   * Force-directed HTML5 Canvas Knowledge Graph Visualizer with drag, pan, zoom, and node inspector.
   * Live in-app Neo4j Cypher query runner console.
   * Direct one-click MP3 audio downloader and streaming inline audio player with seek support.
   * Multi-video persistent storage across Qdrant and Neo4j disk volumes.

---

## Getting Started

### Software Dependencies
Ensure the following tools are installed on your host system:
* **Docker Engine** (v24.0+) & **Docker Compose** (v2.20+)
* **Ollama** ([Download Ollama](https://ollama.ai)) running on host machine (`http://localhost:11434`)
* **Git** (for source control)
* Optional for local native development: Go 1.22+, Python 3.11+, ffmpeg

### Installation Process

#### 1. Clone the Repository
```bash
git clone https://github.com/utkarshraj-gif/Video-Script.git
cd Video-Script
```

#### 2. Pull the Recommended Local LLM
Ensure Ollama is running, then pull the recommended `llama3.2` model (or `gemma4:12b`):
```bash
ollama pull llama3.2
```

#### 3. Launch the Stack with Docker Compose
Start the complete 6-container microservice stack:
```bash
docker compose up -d
```

Docker Compose spins up:
| Service | Container Name | Port | Description |
|---|---|---|---|
| `backend` | `video-rag-backend` | `8080` | Go (Gin) API Gateway, Task Dispatcher & Interactive Studio UI |
| `celery-worker` | `video-rag-celery-worker` | Internal | Celery Distributed Worker for Whisper, FastEmbed, & Qdrant |
| `redis` | `video-rag-redis` | `6379` | Redis In-Memory Message Broker & Result Backend |
| `whisper-service` | `video-rag-whisper` | `5000` | Python FastAPI Microservice & Task Dispatch Proxy |
| `qdrant` | `video-rag-qdrant` | `6335` (mapped to `6333`) | Qdrant Vector Database (384-dim Cosine Metric) |
| `neo4j` | `video-rag-neo4j` | `7474`, `7687` | Neo4j Graph Database (Passwordless) |

#### 4. Open in Browser
* **Video-RAG Studio (Main UI)**: [http://localhost:8080](http://localhost:8080)
* **Neo4j Graph Browser**: [http://localhost:7474/browser/](http://localhost:7474/browser/)
* **Qdrant Vector Dashboard**: [http://localhost:6335/dashboard](http://localhost:6335/dashboard)

---

## API References

#### 1. Asynchronous Video Ingestion (Celery Dispatch)
Enqueues a YouTube URL into Redis for asynchronous background extraction:
* **Endpoint**: `POST /api/v1/videos/process`
* **Request Body**:
```json
{
  "url": "https://www.youtube.com/watch?v=3WI9RZODuag"
}
```
* **Response (Immediate - < 20ms)**:
```json
{
  "status": "queued",
  "task_id": "df0f75ea-3391-4d85-bf78-b17e0817357c",
  "message": "Video processing queued in Celery worker. Processing in background..."
}
```

#### 2. Celery Task Progress & Status
Polls the real-time execution progress of an active ingestion task:
* **Endpoint**: `GET /api/v1/tasks/:id`
* **In-Progress Response**:
```json
{
  "status": "PROGRESS",
  "task_id": "df0f75ea-3391-4d85-bf78-b17e0817357c",
  "percent": 55,
  "step": "3. Transcribing Audio (Faster-Whisper)",
  "message": "Transcribing & translating dialogue to English..."
}
```
* **Completed Response**:
```json
{
  "status": "SUCCESS",
  "task_id": "df0f75ea-3391-4d85-bf78-b17e0817357c",
  "percent": 100,
  "step": "Complete",
  "result": {
    "status": "success",
    "video_id": "3WI9RZODuag",
    "title": "Share Market Crash Course for Beginners | Learn Stock Market #FREE",
    "language": "Hindi",
    "detected_language": "hi",
    "chunks_count": 35,
    "redirect_url": "/api/v1/videos/3WI9RZODuag/embeddings?format=html"
  }
}
```

#### 3. Grounded RAG Chat & Summarization
Queries video chunks via Qdrant semantic search, expands context with Neo4j, and generates grounded Ollama answers:
* **Endpoint**: `POST /api/v1/videos/:id/chat`
* **Request Body**:
```json
{
  "query": "Summarize this video and its core lessons"
}
```
* **Response**:
```json
{
  "status": "success",
  "query": "Summarize this video and its core lessons",
  "video_id": "3WI9RZODuag",
  "answer": "🎯 **Overview**\nThis video provides a structured guide to investing...",
  "highlights": ["Understanding market fundamentals", "Managing financial risk"],
  "concepts": ["StockMarket", "Investment", "RiskManagement"],
  "timestamp": "Full Video",
  "youtube_url": "https://youtu.be/3WI9RZODuag",
  "ollama_model": "llama3.2",
  "rag_source": "Ollama Grounded Summary (llama3.2)"
}
```

#### 4. Interactive Web Studio Dashboard
Renders the complete UI with chatbot, vector chunk explorer, and interactive graph visualizer:
* **Endpoint**: `GET /api/v1/videos/:id/embeddings?format=html`

#### 5. Neo4j Knowledge Graph Data
Returns graph nodes (Video, Chunks, Concepts) and edges (`[:NEXT_CHUNK]`, `[:DISCUSSES]`):
* **Endpoint**: `GET /api/v1/videos/:id/graph`

#### 6. Live Cypher Execution
Executes custom Cypher queries against the local Neo4j database:
* **Endpoint**: `POST /api/v1/neo4j/query`
* **Request Body**:
```json
{
  "query": "MATCH (v:Video)-[:HAS_CHUNK]->(c:Chunk) RETURN v.title, count(c) AS chunk_count"
}
```

#### 7. Audio Download & Streaming
* **Download MP3**: `GET /api/v1/videos/:id/download-audio?download=true`
* **Stream Audio**: `GET /api/v1/videos/:id/stream-audio` (Supports HTTP `Range: bytes=` seeking)

#### 8. System Health
* **Endpoint**: `GET /health`
* **Response**: `{"status":"success","message":"Video RAG API is running!"}`

---

## Build and Test

### Building Containers Locally
To rebuild the containers after making changes:
```bash
# Rebuild and start entire stack in detached mode
docker compose up --build -d

# Check live logs of all services
docker compose logs -f

# Check Celery background worker logs
docker compose logs -f celery-worker

# Check Go backend logs
docker compose logs -f backend
```

### Running Verification Tests

#### 1. Verify Service Health
```bash
curl -s http://localhost:8080/health
```

#### 2. Test Celery Video Queue
```bash
curl -X POST "http://localhost:8080/api/v1/videos/process" \
     -H "Content-Type: application/json" \
     -d '{"url": "https://www.youtube.com/watch?v=Rr4CW2quazA"}'
```

#### 3. Test Grounded RAG Chat
```bash
curl -X POST "http://localhost:8080/api/v1/videos/Rr4CW2quazA/chat" \
     -H "Content-Type: application/json" \
     -d '{"query": "What is Celery used for?"}'
```

#### 4. Safe Shutdown
To stop all services while preserving all Qdrant vectors and Neo4j graphs:
```bash
docker compose down
```

---

## Contribute

We welcome contributions from the open-source community! To contribute:

1. **Fork the Repository** on GitHub.
2. **Create a Feature Branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make Your Changes**:
   * Follow standard Go formatting conventions (`gofmt -s -w .`).
   * Adhere to Python PEP 8 style guidelines.
   * Ensure Docker builds compile cleanly without warnings.
4. **Commit Your Code**:
   ```bash
   git commit -m "feat: Add support for custom embedding models"
   ```
5. **Open a Pull Request**: Submit your pull request to the target branch with a clear description of the enhancements, screenshots (if UI changes are made), and test verification.

---

## License
Distributed under the MIT License. See `LICENSE` for more information.
