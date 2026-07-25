from embedding import EmbeddingModel


class FakeLlama:
    def __init__(self) -> None:
        self.inputs: list[list[str]] = []

    def embed(self, texts: list[str]) -> list[list[float]]:
        self.inputs.append(texts)
        return [[float(index)] for index, _ in enumerate(texts)]


def embedding_model_with(runtime: FakeLlama) -> EmbeddingModel:
    model = EmbeddingModel.__new__(EmbeddingModel)
    model._llama = runtime
    return model


def test_documents_use_nomic_search_document_prefix() -> None:
    runtime = FakeLlama()
    model = embedding_model_with(runtime)

    assert model.encode(["first", "second"]) == [[0.0], [1.0]]
    assert runtime.inputs == [["search_document: first", "search_document: second"]]


def test_queries_use_nomic_search_query_prefix() -> None:
    runtime = FakeLlama()
    model = embedding_model_with(runtime)

    assert model.encode_query("question") == [0.0]
    assert runtime.inputs == [["search_query: question"]]
