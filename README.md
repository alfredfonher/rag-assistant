# rag-assistant local runtime

This repository runs the Go RAG service beside the Python llama.cpp service without Docker. PR1 provides reproducible prerequisites, immutable model acquisition, and receipt-backed Python health telemetry.

## Clean-clone quick path

```bash
cp .env.example .env
bash scripts/bootstrap.sh
bash scripts/models.sh fetch
bash scripts/tests/bootstrap.sh
bash scripts/tests/models.sh
llama-server/.venv/bin/python -m pytest -q llama-server/tests/test_server.py
```

`scripts/bootstrap.sh` requires Linux, Bash 4.4+, Python **3.12**, `uv`, `curl`, GNU `sha256sum`, `setsid`, `flock`, and Go 1.22+. It reports every missing prerequisite and a remediation before changing environment state. `models.sh fetch` resumes `.partial` files, verifies the exact pinned SHA-256, atomically installs the GGUF, and writes `.rag-assistant/verified-models.tsv`. Repeating fetch is safe and does not rewrite verified artifacts.

## Configuration

`.env` is optional. Shell values supplied on the command line take precedence over file values; parsing accepts literal assignments only and never evaluates shell code. Model binaries, partial downloads, `.env`, and verification state are ignored by Git.

## Verification and health

```bash
bash scripts/models.sh verify
```

`GET /healthz` remains HTTP 200 with its existing `status`, model paths, and lazy `*_loaded` booleans. It adds `embedding_verified` and `llm_verified`, which are true only when the manifest digest, immutable tuple, receipt, and current file fingerprint match. Health checks never load models; lazy-loaded state is telemetry, not readiness.

## Rollback

To roll back PR1, revert the PR1 setup, manifest, model-script, receipt-health, test, and documentation changes. Do not delete ignored `models/*.gguf`; a later bootstrap can reuse them after verification. Do not commit model binaries or `.env`.

PR1 intentionally does not add lifecycle orchestration, Go shutdown/readiness changes, or the real-model smoke test; those belong to PR2 and PR3.
