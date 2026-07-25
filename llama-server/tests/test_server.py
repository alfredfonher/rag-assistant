import threading
import time
from concurrent.futures import ThreadPoolExecutor

from fastapi.testclient import TestClient

import server


class FakeEmbedding:
    def __init__(self) -> None:
        self.documents: list[list[str]] = []
        self.queries: list[str] = []

    def encode(self, texts: list[str]) -> list[list[float]]:
        self.documents.append(texts)
        return [[float(index), 0.5] for index, _ in enumerate(texts)]

    def encode_query(self, text: str) -> list[float]:
        self.queries.append(text)
        return [0.25, 0.75]


class FakeLLM:
    def __init__(self) -> None:
        self.calls = []

    def chat(self, messages, **kwargs) -> str:
        self.calls.append((messages, kwargs))
        return "grounded answer"


def _call_concurrently(function):
    barrier = threading.Barrier(2)

    def call():
        barrier.wait()
        return function()

    with ThreadPoolExecutor(max_workers=2) as executor:
        futures = [executor.submit(call) for _ in range(2)]
        return [future.result() for future in futures]


def test_embedding_model_is_initialized_once_for_concurrent_first_calls(monkeypatch) -> None:
    created = 0
    count_lock = threading.Lock()

    class SlowEmbedding:
        def __init__(self, *args, **kwargs) -> None:
            nonlocal created
            with count_lock:
                created += 1
            time.sleep(0.05)

    monkeypatch.setattr(server, "_embedding", None)
    monkeypatch.setattr(server, "EmbeddingModel", SlowEmbedding)

    models = _call_concurrently(server._get_embedding)

    assert created == 1
    assert models[0] is models[1]


def test_llm_is_initialized_once_for_concurrent_first_calls(monkeypatch) -> None:
    created = 0
    count_lock = threading.Lock()

    class SlowLLM:
        def __init__(self, *args, **kwargs) -> None:
            nonlocal created
            with count_lock:
                created += 1
            time.sleep(0.05)

    monkeypatch.setattr(server, "_llm", None)
    monkeypatch.setattr(server, "LLM", SlowLLM)

    models = _call_concurrently(server._get_llm)

    assert created == 1
    assert models[0] is models[1]


def test_embedding_endpoints_use_json_body_contracts(monkeypatch) -> None:
    embedding = FakeEmbedding()
    monkeypatch.setattr(server, "_embedding", embedding)
    client = TestClient(server.app)

    documents = client.post("/v1/embeddings", json={"input": ["one", "two"]})
    assert documents.status_code == 200
    assert documents.json()["data"] == [
        {"object": "embedding", "embedding": [0.0, 0.5], "index": 0},
        {"object": "embedding", "embedding": [1.0, 0.5], "index": 1},
    ]
    assert embedding.documents == [["one", "two"]]

    query = client.post("/v1/embeddings/query", json={"text": "find this"})
    assert query.status_code == 200
    assert query.json() == {"embedding": [0.25, 0.75], "model": "nomic-embed-text"}
    assert embedding.queries == ["find this"]


def test_query_embedding_rejects_query_parameter_contract(monkeypatch) -> None:
    monkeypatch.setattr(server, "_embedding", FakeEmbedding())
    response = TestClient(server.app).post("/v1/embeddings/query?text=wrong-location")
    assert response.status_code == 422


def test_chat_completion_maps_request_and_response(monkeypatch) -> None:
    llm = FakeLLM()
    monkeypatch.setattr(server, "_llm", llm)
    response = TestClient(server.app).post(
        "/v1/chat/completions",
        json={
            "model": "gemma3",
            "messages": [{"role": "user", "content": "question"}],
            "max_tokens": 42,
            "temperature": 0.2,
            "top_p": 0.8,
            "stop": ["STOP"],
        },
    )

    assert response.status_code == 200
    body = response.json()
    assert body["choices"][0]["message"] == {"role": "assistant", "content": "grounded answer"}
    messages, kwargs = llm.calls[0]
    assert messages[0].role == "user"
    assert messages[0].content == "question"
    assert kwargs == {"max_tokens": 42, "temperature": 0.2, "top_p": 0.8, "stop": ["STOP"]}


def test_health_does_not_load_models(monkeypatch) -> None:
    monkeypatch.setattr(server, "_embedding", None)
    monkeypatch.setattr(server, "_llm", None)
    response = TestClient(server.app).get("/healthz")
    assert response.status_code == 200
    assert response.json()["embedding_loaded"] is False
    assert response.json()["llm_loaded"] is False
