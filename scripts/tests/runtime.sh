#!/usr/bin/env bash
# scripts/tests/runtime.sh — GREEN tests for lifecycle start/stop/status with fake processes
set -Eeuo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
export RAG_RUNTIME_DIR="$ROOT/.rag-assistant/runtime"
export RAG_LOG_DIR="$ROOT/.rag-assistant/logs"
export RAG_SHUTDOWN_TIMEOUT=2

# Use temp files for cross-subshell counting
_COUNT_DIR=$(mktemp -d)
echo 0 > "$_COUNT_DIR/pass"
echo 0 > "$_COUNT_DIR/fail"
trap 'rm -rf "$_COUNT_DIR"' EXIT

pass() { echo $(( $(cat "$_COUNT_DIR/pass") + 1 )) > "$_COUNT_DIR/pass"; printf '  PASS: %s\n' "$1"; }
fail() { echo $(( $(cat "$_COUNT_DIR/fail") + 1 )) > "$_COUNT_DIR/fail"; printf '  FAIL: %s\n' "$1"; }

assert_exit() {
  local desc="$1" expected="$2" actual="$3"
  if [[ "$actual" == "$expected" ]]; then pass "$desc"; else fail "$desc (expected=$expected got=$actual)"; fi
}

assert_file_exists() {
  local desc="$1" path="$2"
  if [[ -f "$path" ]]; then pass "$desc"; else fail "$desc (missing $path)"; fi
}

assert_file_not_exists() {
  local desc="$1" path="$2"
  if [[ ! -f "$path" ]]; then pass "$desc"; else fail "$desc (exists $path)"; fi
}

assert_output_contains() {
  local desc="$1" needle="$2" haystack="$3"
  if [[ "$haystack" == *"$needle"* ]]; then pass "$desc"; else fail "$desc (missing '$needle')"; fi
}

# Create a fake long-running process script
make_fake() {
  local name="${1:-fake}" port="${2:-19999}" pidfile="${3:-/tmp/rag-fake.pid}"
  local script="/tmp/rag-fake-${name}.sh"
  cat > "$script" <<FAKE
#!/usr/bin/env bash
echo \$\$ > "$pidfile"
trap 'exit 0' TERM
while true; do sleep 0.1; done
FAKE
  chmod +x "$script"
  echo "$script"
}

cleanup_all() {
  # Kill any remaining fake processes
  for f in "$RAG_RUNTIME_DIR"/*.pid; do
    [[ -f "$f" ]] || continue
    local pid
    pid=$(<"$f")
    kill -TERM "$pid" 2>/dev/null || true
    sleep 0.2
  done
  rm -rf "$RAG_RUNTIME_DIR" "$RAG_LOG_DIR" /tmp/rag-fake-*.sh
}
trap cleanup_all EXIT

cd "$ROOT"
source scripts/runtime.sh

# --- Test: status on empty state ---
printf '\n--- status on empty state ---\n'
mkdir -p "$RAG_RUNTIME_DIR"
rc=0; cmd_status "non-existent" 2>/dev/null || rc=$?
assert_exit "status on empty returns 1" "1" "$rc"
rm -rf "$RAG_RUNTIME_DIR"

# --- Test: start creates PID file ---
printf '\n--- start creates PID file ---\n'
mkdir -p "$RAG_RUNTIME_DIR" "$RAG_LOG_DIR"
script=$(make_fake "start-test" "19981" "$RAG_RUNTIME_DIR/start-test.pid")
rc=0; cmd_start "test-svc" "bash $script" 2>/dev/null || rc=$?
assert_exit "start returns 0" "0" "$rc"
assert_file_exists "PID file created" "$RAG_RUNTIME_DIR/test-svc.pid"
cmd_stop "test-svc" 2>/dev/null || true
rm -rf "$RAG_RUNTIME_DIR"

# --- Test: idempotent start ---
printf '\n--- idempotent start ---\n'
mkdir -p "$RAG_RUNTIME_DIR" "$RAG_LOG_DIR"
script=$(make_fake "idem-test" "19982" "$RAG_RUNTIME_DIR/idem-test.pid")
cmd_start "test-idem" "bash $script" 2>/dev/null
# Start again — should be idempotent
rc=0; cmd_start "test-idem" "bash $script" 2>/dev/null || rc=$?
assert_exit "idempotent start returns 0" "0" "$rc"
cmd_stop "test-idem" 2>/dev/null || true
rm -rf "$RAG_RUNTIME_DIR"

# --- Test: stop sends SIGTERM and cleans PID file ---
printf '\n--- stop SIGTERM cleanup ---\n'
mkdir -p "$RAG_RUNTIME_DIR" "$RAG_LOG_DIR"
script=$(make_fake "stop-test" "19983" "$RAG_RUNTIME_DIR/stop-test.pid")
cmd_start "test-stop" "bash $script" 2>/dev/null
pidfile="$RAG_RUNTIME_DIR/test-stop.pid"
pid=$(<"$pidfile")
rc=0; cmd_stop "test-stop" 2>/dev/null || rc=$?
assert_exit "stop returns 0" "0" "$rc"
sleep 0.3
assert_file_not_exists "PID file removed after stop" "$pidfile"
if kill -0 "$pid" 2>/dev/null; then
  fail "process $pid still alive after stop"
else
  pass "process terminated after stop"
fi
rm -rf "$RAG_RUNTIME_DIR"

# --- Test: stop non-existent returns error ---
printf '\n--- stop non-existent ---\n'
mkdir -p "$RAG_RUNTIME_DIR"
rc=0; cmd_stop "non-existent-svc" 2>/dev/null || rc=$?
assert_exit "stop non-existent returns 1" "1" "$rc"
rm -rf "$RAG_RUNTIME_DIR"

# --- Test: status returns running ---
printf '\n--- status running ---\n'
mkdir -p "$RAG_RUNTIME_DIR" "$RAG_LOG_DIR"
script=$(make_fake "status-test" "19985" "$RAG_RUNTIME_DIR/status-test.pid")
cmd_start "test-status" "bash $script" 2>/dev/null
rc=0; cmd_status "test-status" 2>/dev/null || rc=$?
assert_exit "status running returns 0" "0" "$rc"
output=$(cmd_status "test-status" 2>/dev/null || true)
assert_output_contains "status shows RUNNING" "RUNNING" "$output"
cmd_stop "test-status" 2>/dev/null || true
rm -rf "$RAG_RUNTIME_DIR"

# --- Summary ---
PASS=$(cat "$_COUNT_DIR/pass")
FAIL=$(cat "$_COUNT_DIR/fail")
printf '\n=== runtime tests: %d passed, %d failed ===\n' "$PASS" "$FAIL"
[[ "$FAIL" -gt 0 ]] && exit 1
exit 0
