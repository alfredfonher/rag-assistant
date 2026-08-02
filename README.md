# rag-assistant

A full-stack RAG system that ingests documents, performs semantic search, and streams LLM responses — all running locally with zero cloud dependencies.

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go (stdlib net/http, hexagonal architecture, zero external deps) |
| Inference | Python FastAPI + llama-cpp-python |
| Frontend | Next.js 16, React 19, shadcn/ui, Prisma |
| TUI Client | Go (Bubbletea) |
| Models | nomic-embed-text v1.5 (embeddings), Gemma 3 1B / Qwen 3.5 0.8B (chat) |
| Orchestration | Docker Compose |
| CI | GitHub Actions |

## Quick Start

**Option A — Docker (recommended):**

```bash
cp .env.example .env
docker compose up --build
```

**Option B — Manual:**

```bash
cp .env.example .env
bash scripts/bootstrap.sh
bash scripts/models.sh fetch
```

The bootstrap script requires Linux, Bash 4.4+, Python 3.12, `uv`, `curl`, `sha256sum`, and Go 1.22+. It reports missing prerequisites before changing environment state. The model fetch script resumes partial downloads, verifies SHA-256 checksums, and writes a verification receipt.

## Architecture

```
┌─────────────────┐     ┌──────────────────┐
│   Next.js UI    │     │   Go TUI Client  │
│  (port 3000)    │     │  (Bubbletea)     │
└────────┬────────┘     └────────┬─────────┘
         │                       │
         ▼                       ▼
┌─────────────────────────────────────────┐
│          Go Backend (port 8080)         │
│     Hexagonal architecture, stdlib      │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│     Python Inference (port 8090)        │
│   FastAPI + llama-cpp-python server     │
└────────────────┬────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────┐
│        llama.cpp (local GGUF)           │
│  Embeddings + Chat models on disk       │
└─────────────────────────────────────────┘
```

## Project Structure

```
rag-assistant/
├── service/           # Go backend (hexagonal ports + adapters)
├── llama-server/      # Python inference server
├── web/               # Next.js frontend
├── cmd/               # Go TUI client
├── scripts/           # Bootstrap, models, lifecycle, smoke tests
├── docker-compose.yml
└── .env.example
```

## Features

- **Document ingestion** — chunk, embed, and index into vector storage
- **Semantic search** — query documents by meaning, not keywords
- **SSE streaming** — real-time token-by-token LLM responses
- **Health probes** — `/healthz` for status, `/readyz` for dependency checks
- **Model verification** — SHA-256 integrity checks on all GGUF files
- **Dual chat models** — Gemma 3 1B or Qwen 3.5 0.8B via config

## Document Ingestion

`POST /v1/documents/ingest` accepts a `.txt`, `.md`, or `.markdown` path relative to `RAG_INGEST_ROOT`; absolute paths and traversal are rejected. The local default is `docs`, resolved relative to the backend process's current working directory; set `RAG_INGEST_ROOT` to override it. Docker Compose sets the root to `/docs`, backed by the read-only `./docs:/docs:ro` mount, so a request for the repository file `docs/docker.md` uses `{"path":"docker.md"}`.

Documents are limited to regular files of at most 10 MiB. The response reports indexing state, document ID, citations, and stable errors; normalized document content is never returned. Index ingestion and the CRUD document registry remain separate flows.

## Testing

```bash
# Go unit tests
go test ./...

# Python inference tests
llama-server/.venv/bin/python -m pytest -q llama-server/tests/

# Full smoke test (requires running llama-server on :8090)
bash scripts/smoke.sh
```

The smoke test ingests a fixture document, queries the answer, verifies the readiness endpoint, and confirms no GGUF binaries leak into the data directory.

## License

MIT
