# rag-assistant local runtime

This repository runs the Go RAG service beside the Python llama.cpp service without Docker. PR1 provides reproducible prerequisites, immutable model acquisition, and receipt-backed Python health telemetry. PR2 adds lifecycle management with start/stop/status, PID identity safety, and graceful signal shutdown.

## Clean-clone quick path

```bash
cp .env.example .env
bash scripts/bootstrap.sh
bash scripts/models.sh fetch
bash scripts/tests/bootstrap.sh
bash scripts/tests/models.sh
llama-server/.venv/bin/python -m pytest -q llama-server/tests/test_server.py
```

`scripts/bootstrap.sh` requires Linux, Bash 4.4+, Python **3.12**, `uv`, `curl`, GNU `sha256sum`, `setsid`, and Go 1.22+. It reports every missing prerequisite and a remediation before changing environment state. `models.sh fetch` resumes `.partial` files, verifies the exact pinned SHA-256, atomically installs the GGUF, and writes `.rag-assistant/verified-models.tsv`. Repeating fetch is safe and does not rewrite verified artifacts.

## Configuration

`.env` is optional. Shell values supplied on the command line take precedence over file values; parsing accepts literal assignments only and never evaluates shell code. Model binaries, partial downloads, `.env`, and verification state are ignored by Git.

## Lifecycle management

```bash
bash scripts/runtime.sh start <service-name> <command...>
bash scripts/runtime.sh status <service-name>
bash scripts/runtime.sh stop <service-name>
```

PR2 provides deterministic start/stop/status with:
- **PID identity safety**: rejects reused PID/PGID/UID/cmdline/start-time
- **Idempotent start**: skip if already running with matching identity
- **Graceful shutdown**: SIGTERM with configurable timeout, then SIGKILL fallback
- **Atomic PID files**: written via temp+rename, stale detection via `kill -0`
- **Go graceful shutdown**: `main.go` handles SIGTERM/SIGINT with 10s drain timeout

Environment variables: `RAG_DATA_DIR`, `RAG_LOG_DIR`, `RAG_SHUTDOWN_TIMEOUT` (default 5s).

## Verification and health

```bash
bash scripts/models.sh verify
```

`GET /healthz` remains HTTP 200 with its existing `status`, model paths, and lazy `*_loaded` booleans. It adds `embedding_verified` and `llm_verified`, which are true only when the manifest digest, immutable tuple, receipt, and current file fingerprint match. Health checks never load models; lazy-loaded state is telemetry, not readiness.

## Readiness

`GET /readyz` returns HTTP 200 when the llama-server dependency is reachable and both embedding and LLM models are loaded. Returns HTTP 503 with dependency reasons when the server is unreachable or models are not loaded. Readiness never loads models; it checks the existing health endpoint.

## End-to-end smoke test

```bash
bash scripts/smoke.sh
```

Requires a running llama-server on port 8090 (configurable via `RAG_LLAMA_PORT`). The smoke test ingests `scripts/fixtures/smoke.md`, queries for the answer, verifies the readiness endpoint, and proves no GGUF binaries or temp files leak into the data directory. On failure, it cleans up all managed processes and temp data.

## Rollback

To roll back PR1, revert the PR1 setup, manifest, model-script, receipt-health, test, and documentation changes. Do not delete ignored `models/*.gguf`; a later bootstrap can reuse them after verification. Do not commit model binaries or `.env`.

To roll back PR2, revert `scripts/runtime.sh`, `scripts/lib/{process,wait}.sh`, `scripts/tests/runtime.sh`, and the Go signal-shutdown changes in `service/cmd/server/`. PR1 remains intact.

PR2 intentionally does not add readiness probing, Go dependency-backed health, or the real-model smoke test; those belong to PR3.
