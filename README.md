# Video-RAG Studio 🎬🧠

> **Containerized Multimodal Video Retrieval-Augmented Generation (RAG) System**  
> Powered by Faster-Whisper, FastEmbed (384-dim), Qdrant Vector DB, Neo4j Knowledge Graph, and Ollama Local LLMs.

---

## Introduction

### Overview
**Video-RAG Studio** is a full-stack, multimodal Retrieval-Augmented Generation (RAG) platform that transforms long-form video content (such as YouTube lectures, tutorials, and podcasts) into an interactive, context-aware AI assistant. 

Unlike traditional RAG systems that treat transcripts as flat, disconnected text, Video-RAG combines **dense vector similarity search** with **graph-based relational expansion** and **grounded local LLM synthesis**.

```
                           ┌────────────────────────┐
                           │   YouTube Video URL    │
                           └───────────┬────────────┘
                                       │ (yt-dlp)
                                       ▼
                           ┌────────────────────────┐
                           │ Audio Extraction (.mp3)│
                           └───────────┬────────────┘
                                       │
                                       ▼
                       ┌────────────────────────────────┐
                       │ Faster-Whisper ASR (int8 CPU)  │
                       └───────────────┬────────────────┘
                                       │ Timestamps & Text
                                       ▼
                       ┌────────────────────────────────┐
                       │ FastEmbed (all-MiniLM-L6-v2)   │
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
                        │ (Chat, Visual Graph, MP3 Audio) │
                        └─────────────────────────────────┘
```

### Motivation & Objectives
1. **Preserve Temporal & Semantic Flow**: Long-form videos have chronological progression. By pairing sequential relationships (`[:NEXT_CHUNK]`) and domain entities (`[:DISCUSSES]`) in Neo4j with 384-dimensional dense vectors in Qdrant, queries retrieve not only the best-matching moment, but also preceding causes, subsequent outcomes, and related topics.
2. **Eliminate Raw Chunk Dumping & Hallucinations**: Standard vector search dumps raw speech snippets filled with transcription artifacts. Video-RAG uses local Ollama models (`llama3.2` or `gemma4:12b`) with strict prompt grounding to synthesize clean, accurate answers based solely on what was said in the video.
3. **Frictionless Local Execution**: Zero cloud API dependencies. Runs entirely on your local machine using Docker Compose and local Ollama, ensuring data privacy and fast offline capability.
4. **Interactive All-in-One Studio**: Includes an embedded force-directed Canvas knowledge graph, a live Cypher console, an inline audio player/downloader, and full Qdrant vector exploration.

---

## Getting Started

### Software Dependencies
Ensure the following tools are installed on your host system:
* **Docker Engine** (v24.0+) & **Docker Compose** (v2.20+)
* **Ollama** ([Download Ollama](https://ollama.ai)) running on the host machine (`http://localhost:11434`)
* **Git** (for source control)
* Optional for local native development: Go 1.22+, Python 3.11+, ffmpeg

### Installation Process

#### 1. Clone the Repository
```bash
git clone https://github.com/utkarshraj-gif/Video-Script.git
cd Video-Script
```

#### 2. Pull the Recommended Local LLM
Ensure Ollama is running, then pull the lightweight `llama3.2` model (or `gemma4:12b`):
```bash
ollama pull llama3.2
```

#### 3. Launch the Stack with Docker Compose
From the repository root, start all 4 containerized microservices:
```bash
docker compose up -d
```

Docker Compose spins up:
| Service | Container Name | Port | Description |
|---|---|---|---|
| `backend` | `video-rag-backend` | `8080` | Go (Gin) API & Interactive Web Studio UI |
| `whisper-service` | `video-rag-whisper` | `5000` | Python Faster-Whisper & FastEmbed microservice |
| `qdrant` | `video-rag-qdrant` | `6335` (mapped to `6333`) | Qdrant Vector Database |
| `neo4j` | `video-rag-neo4j` | `7474`, `7687` | Neo4j Graph Database (Passwordless) |

#### 4. Open in Browser
* **Video-RAG Studio (Main UI)**: [http://localhost:8080](http://localhost:8080)
* **Neo4j Graph Browser**: [http://localhost:7474/browser/](http://localhost:7474/browser/)
* **Qdrant Vector Dashboard**: [http://localhost:6335/dashboard](http://localhost:6335/dashboard)

---

### Latest Releases
* **v1.0.0 (Current)**:
  * Full end-to-end containerized pipeline (`yt-dlp` ➔ Faster-Whisper ➔ FastEmbed ➔ Qdrant ➔ Neo4j ➔ Ollama).
  * Interactive HTML5 Canvas Force-Directed Knowledge Graph with drag, pan, zoom, and node inspector.
  * Live in-app Neo4j Cypher query runner console.
  * Direct one-click MP3 audio downloader and streaming inline audio player with seek support.
  * Grounded Ollama RAG integration with executive video summarization and zero-hallucination Q&A.

---

### API References

#### 1. End-to-End Video Ingestion
Processes a YouTube URL through the entire pipeline: audio download, transcription, embedding generation, Qdrant upsert, and Neo4j graph synchronization.
* **Endpoint**: `POST /api/v1/videos/process`
* **Request Body**:
```json
{
  "url": "https://www.youtube.com/watch?v=hsOJhs3_UCM"
}
```
* **Response**:
```json
{
  "status": "success",
  "message": "Video processed and indexed into Qdrant & Neo4j successfully.",
  "video_id": "hsOJhs3_UCM",
  "title": "Complete Session On Knowledge Graph and GraphDb With Langchain",
  "total_chunks": 155,
  "redirect_url": "/api/v1/videos/hsOJhs3_UCM/embeddings?format=html"
}
```

#### 2. Grounded RAG Chat & Summarization
Queries the video using Qdrant vector retrieval, Neo4j relational expansion, and Ollama synthesis.
* **Endpoint**: `POST /api/v1/videos/:id/chat`
* **Request Body**:
```json
{
  "query": "Summarize this video"
}
```
* **Response**:
```json
{
  "status": "success",
  "query": "Summarize this video",
  "video_id": "hsOJhs3_UCM",
  "answer": "🎯 **Overview**\nThis video explains how knowledge graphs and graph databases...",
  "highlights": ["Key point 1...", "Key point 2..."],
  "concepts": ["GraphDatabase", "LangChain", "Neo4j"],
  "timestamp": "Full Video",
  "youtube_url": "https://youtu.be/hsOJhs3_UCM",
  "ollama_model": "llama3.2",
  "rag_source": "Ollama Grounded Summary (llama3.2)"
}
```

#### 3. Interactive Web Studio Dashboard
Renders the complete UI with chatbot, vector chunk explorer, and interactive graph visualizer.
* **Endpoint**: `GET /api/v1/videos/:id/embeddings?format=html`

#### 4. Neo4j Knowledge Graph Data
Returns the graph nodes (Video, Chunks, Concepts) and edges (`[:NEXT_CHUNK]`, `[:DISCUSSES]`) for visualization.
* **Endpoint**: `GET /api/v1/videos/:id/graph`

#### 5. Live Cypher Execution
Executes ad-hoc Cypher queries against the local Neo4j database.
* **Endpoint**: `POST /api/v1/neo4j/query`
* **Request Body**:
```json
{
  "query": "MATCH (c:Chunk)-[:DISCUSSES]->(cp:Concept) RETURN c.timestamp_display, cp.name LIMIT 10"
}
```

#### 6. Audio Download & Streaming
* **Download MP3**: `GET /api/v1/videos/:id/download-audio?download=true`
* **Stream Audio**: `GET /api/v1/videos/:id/stream-audio` (Supports HTTP `Range: bytes=` seeking)

#### 7. System Health
* **Endpoint**: `GET /health`
* **Response**: `{"status":"success","message":"Video RAG API is running!"}`

---

## Build and Test

### Building Containers Locally
To rebuild the containers after making changes:
```bash
# Rebuild and start in detached mode
docker compose up --build -d

# Check live logs of all services
docker compose logs -f

# Check logs of backend service only
docker compose logs -f backend
```

### Running Tests

#### 1. Verify Service Health
```bash
curl -s http://localhost:8080/health
```

#### 2. Test Grounded RAG Chat
```bash
curl -X POST "http://localhost:8080/api/v1/videos/hsOJhs3_UCM/chat" \
     -H "Content-Type: application/json" \
     -d '{"query": "What is a knowledge graph?"}'
```

#### 3. Test Neo4j Cypher Execution
```bash
curl -X POST "http://localhost:8080/api/v1/neo4j/query" \
     -H "Content-Type: application/json" \
     -d '{"query": "MATCH (n) RETURN count(n) AS total_nodes"}'
```

#### 4. Safe Shutdown
To stop all services while safely preserving all indexed Qdrant vectors and Neo4j graphs:
```bash
docker compose down
```

---

## Contribute

We welcome contributions from the open-source community! To contribute to Video-RAG Studio:

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
5. **Push to Your Branch**:
   ```bash
   git push origin feature/your-feature-name
   ```
6. **Open a Pull Request**: Submit your pull request to the `main` or develop branch with a clear description of the enhancements, screenshots (if UI changes are made), and test verification.

---

## License
Distributed under the MIT License. See `LICENSE` for more information.
