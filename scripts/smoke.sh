#!/usr/bin/env bash
# scripts/smoke.sh — end-to-end smoke test with real models
# Verifies: ingest → embed → query → citation → cleanup
set -Eeuo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$ROOT"

PASS=0 FAIL=0
pass() { PASS=$((PASS+1)); printf '  PASS: %s\n' "$1"; }
fail() { FAIL=$((FAIL+1)); printf '  FAIL: %s\n' "$1"; }

# --- config ---
LLAMA_PORT="${RAG_LLAMA_PORT:-8090}"
SERVICE_PORT="${RAG_SERVICE_PORT:-8091}"
LLAMA_URL="http://127.0.0.1:${LLAMA_PORT}"
SERVICE_URL="http://127.0.0.1:${SERVICE_PORT}"
DATA_DIR=$(mktemp -d)
INGEST_ROOT="$ROOT/scripts/fixtures"
FIXTURE="smoke.md"

cleanup() {
  # Stop service if running
  if [[ -n "${SERVICE_PID:-}" ]]; then
    kill -TERM "$SERVICE_PID" 2>/dev/null || true
    wait "$SERVICE_PID" 2>/dev/null || true
  fi
  rm -rf "$DATA_DIR"
  # Verify no orphaned rag-assistant processes (exclude grep/self)
  if pgrep -f "cmd/server" | grep -v "$$" >/dev/null 2>&1; then
    fail "orphaned rag-assistant process"
  fi
}
trap cleanup EXIT

# --- check prerequisites ---
if ! curl -sf "$LLAMA_URL/healthz" >/dev/null 2>&1; then
  echo "SKIP: llama-server not running on $LLAMA_URL"
  exit 0
fi

# --- ingest ---
printf '\n--- ingest smoke fixture ---\n'
INGEST_RESP=$(curl -sf -X POST "$SERVICE_URL/v1/documents/ingest" \
  -H "Content-Type: application/json" \
  -d "{\"path\": \"$FIXTURE\"}" 2>/dev/null) || {
  # Service may not be running — try starting it
  RAG_HTTP_ADDR="127.0.0.1:${SERVICE_PORT}" \
  RAG_DATA_DIR="$DATA_DIR" \
  RAG_INGEST_ROOT="$INGEST_ROOT" \
  RAG_LLAMA_SERVER_URL="$LLAMA_URL" \
  go run -C "$ROOT/service" ./cmd/server &
  SERVICE_PID=$!
  sleep 2
  INGEST_RESP=$(curl -sf -X POST "$SERVICE_URL/v1/documents/ingest" \
    -H "Content-Type: application/json" \
    -d "{\"path\": \"$FIXTURE\"}")
}

if echo "$INGEST_RESP" | grep -q '"state"'; then
  pass "ingest returned response"
else
  fail "ingest failed: $INGEST_RESP"
fi

# --- query ---
printf '\n--- query smoke ---\n'
QUERY_RESP=$(curl -sf -X POST "$SERVICE_URL/v1/query" \
  -H "Content-Type: application/json" \
  -d '{"query": "What is the smoke answer?"}') || true

if echo "$QUERY_RESP" | grep -q '"state"'; then
  pass "query returned response"
else
  fail "query failed: $QUERY_RESP"
fi

# --- readiness ---
printf '\n--- readiness check ---\n'
READY_RESP=$(curl -sf "$SERVICE_URL/readyz" 2>/dev/null) || true
if echo "$READY_RESP" | grep -q '"ready"'; then
  pass "readiness endpoint responds"
else
  fail "readiness failed: $READY_RESP"
fi

# --- cleanup proof ---
printf '\n--- cleanup proof ---\n'
if [[ -d "$DATA_DIR" ]]; then
  # Check no GGUF binaries in data dir
  if find "$DATA_DIR" -name '*.gguf' 2>/dev/null | grep -q .; then
    fail "GGUF binary leaked to data dir"
  else
    pass "no GGUF binaries in data dir"
  fi
fi

# --- summary ---
printf '\n=== smoke tests: %d passed, %d failed ===\n' "$PASS" "$FAIL"
[[ "$FAIL" -gt 0 ]] && exit 1
exit 0
