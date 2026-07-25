"""Lightweight llama.cpp HTTP server for rag-assistant.

Exposes OpenAI-compatible endpoints:
  POST /v1/embeddings       - text to vector (nomic-embed-text)
  POST /v1/embeddings/query - single query embedding shorthand
  POST /v1/chat/completions - chat generation (gemma3)
  GET  /healthz             - model load status
"""

from __future__ import annotations

import gc
import logging
import os
import threading
import time
from pathlib import Path
from typing import Any

import uvicorn
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from embedding import EmbeddingModel
from llm import LLM
from models import ChatMessage

LOG = logging.getLogger("rag-llama-server")

_BASE_DIR = Path(__file__).resolve().parent.parent / "models"

EMBEDDING_MODEL_PATH = Path(
    os.environ.get("RAG_EMBEDDING_MODEL", str(_BASE_DIR / "nomic-embed-text-v1.5.f32.gguf"))
)
LLM_MODEL_PATH = Path(
    os.environ.get("RAG_LLM_MODEL", str(_BASE_DIR / "gemma-3-1b-it-Q4_K_M.gguf"))
)

HOST = os.environ.get("RAG_LLAMA_HOST", "127.0.0.1")
PORT = int(os.environ.get("RAG_LLAMA_PORT", "8090"))
N_CTX = int(os.environ.get("RAG_LLAMA_N_CTX", "2048"))
N_THREADS = int(os.environ.get("RAG_LLAMA_N_THREADS", "0"))


# ---------------------------------------------------------------------------
# Request / Response schemas (OpenAI-compatible subset)
# ---------------------------------------------------------------------------

class EmbeddingRequest(BaseModel):
    input: str | list[str]
    model: str = "nomic-embed-text"


class QueryEmbeddingRequest(BaseModel):
    text: str


class EmbeddingObject(BaseModel):
    object: str = "embedding"
    embedding: list[float]
    index: int


class EmbeddingUsage(BaseModel):
    prompt_tokens: int = 0
    total_tokens: int = 0


class EmbeddingResponse(BaseModel):
    object: str = "list"
    data: list[EmbeddingObject]
    model: str
    usage: EmbeddingUsage


class APIChatMessage(BaseModel):
    role: str
    content: str


class ChatCompletionRequest(BaseModel):
    model: str = "gemma3"
    messages: list[APIChatMessage]
    max_tokens: int = Field(default=512, ge=1)
    temperature: float = Field(default=0.3, ge=0.0, le=2.0)
    top_p: float = Field(default=0.9, ge=0.0, le=1.0)
    stream: bool = False
    stop: list[str] | None = None


class ChatChoice(BaseModel):
    index: int = 0
    message: APIChatMessage
    finish_reason: str = "stop"


class ChatUsage(BaseModel):
    prompt_tokens: int = 0
    completion_tokens: int = 0
    total_tokens: int = 0


class ChatCompletionResponse(BaseModel):
    id: str
    object: str = "chat.completion"
    created: int
    model: str
    choices: list[ChatChoice]
    usage: ChatUsage


class HealthResponse(BaseModel):
    status: str
    embedding_model: str
    embedding_loaded: bool
    llm_model: str
    llm_loaded: bool


# ---------------------------------------------------------------------------
# Global model state (lazy-loaded)
# ---------------------------------------------------------------------------

_embedding: EmbeddingModel | None = None
_llm: LLM | None = None
_embedding_lock = threading.Lock()
_llm_lock = threading.Lock()


def _get_embedding() -> EmbeddingModel:
    global _embedding
    if _embedding is not None:
        return _embedding
    with _embedding_lock:
        if _embedding is None:
            LOG.info("Loading embedding model: %s", EMBEDDING_MODEL_PATH)
            _embedding = EmbeddingModel(
                EMBEDDING_MODEL_PATH, n_ctx=N_CTX, n_threads=N_THREADS or 1
            )
            LOG.info("Embedding model loaded")
    return _embedding


def _get_llm() -> LLM:
    global _llm
    if _llm is not None:
        return _llm
    with _llm_lock:
        if _llm is None:
            LOG.info("Loading LLM model: %s", LLM_MODEL_PATH)
            _llm = LLM(LLM_MODEL_PATH, n_ctx=N_CTX, n_threads=N_THREADS or None)
            LOG.info("LLM model loaded")
    return _llm


# ---------------------------------------------------------------------------
# FastAPI app
# ---------------------------------------------------------------------------

app = FastAPI(title="rag-llama-server", version="0.1.0")


@app.get("/healthz", response_model=HealthResponse)
def healthz() -> HealthResponse:
    return HealthResponse(
        status="ok",
        embedding_model=str(EMBEDDING_MODEL_PATH),
        embedding_loaded=_embedding is not None,
        llm_model=str(LLM_MODEL_PATH),
        llm_loaded=_llm is not None,
    )


@app.post("/v1/embeddings", response_model=EmbeddingResponse)
def create_embeddings(req: EmbeddingRequest) -> EmbeddingResponse:
    texts = [req.input] if isinstance(req.input, str) else req.input
    if not texts:
        raise HTTPException(status_code=400, detail="input must not be empty")
    model = _get_embedding()
    try:
        vectors = model.encode(texts)
    except Exception as exc:
        raise HTTPException(status_code=500, detail=str(exc))
    finally:
        gc.collect()
    data = [EmbeddingObject(embedding=v, index=i) for i, v in enumerate(vectors)]
    tokens = sum(len(t.split()) for t in texts)
    return EmbeddingResponse(data=data, model="nomic-embed-text", usage=EmbeddingUsage(prompt_tokens=tokens, total_tokens=tokens))


@app.post("/v1/embeddings/query")
def create_query_embedding(req: QueryEmbeddingRequest) -> dict[str, Any]:
    if not req.text.strip():
        raise HTTPException(status_code=400, detail="text must not be empty")
    model = _get_embedding()
    try:
        vec = model.encode_query(req.text)
    except Exception as exc:
        raise HTTPException(status_code=500, detail=str(exc))
    finally:
        gc.collect()
    return {"embedding": vec, "model": "nomic-embed-text"}


@app.post("/v1/chat/completions", response_model=ChatCompletionResponse)
def create_chat_completion(req: ChatCompletionRequest) -> ChatCompletionResponse:
    if not req.messages:
        raise HTTPException(status_code=400, detail="messages must not be empty")
    llm = _get_llm()
    messages = [ChatMessage(role=m.role, content=m.content) for m in req.messages]
    try:
        content = llm.chat(
            messages,
            max_tokens=req.max_tokens,
            temperature=req.temperature,
            top_p=req.top_p,
            stop=req.stop,
        )
    except Exception as exc:
        raise HTTPException(status_code=500, detail=str(exc))
    now = int(time.time())
    return ChatCompletionResponse(
        id="chatcmpl-rag",
        created=now,
        model=req.model,
        choices=[ChatChoice(message=APIChatMessage(role="assistant", content=content))],
        usage=ChatUsage(prompt_tokens=0, completion_tokens=len(content.split()), total_tokens=0),
    )


def main() -> None:
    logging.basicConfig(level=logging.INFO)
    LOG.info("Starting rag-llama-server on %s:%s", HOST, PORT)
    uvicorn.run(app, host=HOST, port=PORT, log_level="info")


if __name__ == "__main__":
    main()
