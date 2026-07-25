# rag-llama-server

FastAPI microservice that exposes local llama.cpp embedding and chat models to the Go service.

## Setup

Python 3.12 or 3.13 is required.

```bash
python3.12 -m venv .venv
.venv/bin/pip install -e '.[dev]'
```

Place the GGUF files in the repository `models/` directory, or configure explicit paths.

## Run

```bash
.venv/bin/python server.py
```

The server listens on `127.0.0.1:8090` by default. Models load lazily on their first embedding or chat request, so `GET /healthz` does not load GGUF files.

## Environment

- `RAG_EMBEDDING_MODEL`: nomic embedding GGUF path. Default: `models/nomic-embed-text-v1.5.f32.gguf`.
- `RAG_LLM_MODEL`: chat GGUF path. Default: `models/gemma-3-1b-it-Q4_K_M.gguf`.
- `RAG_LLAMA_HOST`: bind host. Default: `127.0.0.1`.
- `RAG_LLAMA_PORT`: bind port. Default: `8090`.
- `RAG_LLAMA_N_CTX`: llama.cpp context size. Default: `2048`.
- `RAG_LLAMA_N_THREADS`: worker thread count. Default: llama.cpp selection for chat and `1` for embeddings.

## Test

The endpoint tests replace both model wrappers and never load GGUF files.

```bash
.venv/bin/python -m pytest
```
