#!/usr/bin/env bash
# scripts/runtime.sh — lifecycle manager: start, stop, status
# PID identity safety, SIGTERM→SIGKILL, idempotent start, atomic PID files.
set -Eeuo pipefail

ROOT=${RAG_RUNTIME_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}
RAG_RUNTIME_DIR="${RAG_DATA_DIR:-$ROOT/.rag-assistant}/runtime"
RAG_RUNTIME_LOG_DIR="${RAG_LOG_DIR:-$ROOT/.rag-assistant/logs}"
RAG_SHUTDOWN_TIMEOUT="${RAG_SHUTDOWN_TIMEOUT:-5}"

# --- process helpers ---
process_ensure_dir() { mkdir -p "$RAG_RUNTIME_DIR"; }

process_acquire() {
  local service="$1" pidfile="$2" pid="$3"
  process_ensure_dir
  if [[ -f "$pidfile" ]]; then
    local old_pid; old_pid=$(<"$pidfile")
    if kill -0 "$old_pid" 2>/dev/null; then
      [[ "$old_pid" == "$pid" ]] && return 0
      return 1  # different process holds our PID file
    fi
  fi
  local tmp="${pidfile}.tmp.$$"
  printf '%s\n' "$pid" > "$tmp"
  mv -f "$tmp" "$pidfile"
}

process_release() {
  local pidfile="$1"
  [[ -f "$pidfile" ]] || return 1
  local pid; pid=$(<"$pidfile")
  kill -0 "$pid" 2>/dev/null && return 1  # still running
  rm -f "$pidfile"
}

process_read_pid() {
  local pidfile="$1"
  [[ -f "$pidfile" ]] || { echo ""; return; }
  local pid; pid=$(<"$pidfile")
  kill -0 "$pid" 2>/dev/null && echo "$pid" || echo ""
}

# --- wait helper ---
wait_terminate() {
  local pid="$1" timeout="${2:-5}"
  [[ -z "$pid" ]] && return 0
  kill -0 "$pid" 2>/dev/null || return 0
  kill -TERM "$pid" 2>/dev/null
  local elapsed=0
  while kill -0 "$pid" 2>/dev/null && (( elapsed < timeout )); do
    sleep 0.2; elapsed=$((elapsed + 1))
  done
  if kill -0 "$pid" 2>/dev/null; then
    kill -KILL "$pid" 2>/dev/null || true; sleep 0.1
  fi
}

# --- usage ---
usage() {
  cat <<EOF
Usage: $(basename "$0") <start|stop|status> <service-name> [command...]

Commands:
  start <name> <command...>   Start a managed service
  stop  <name>                Stop a managed service
  status <name>               Check service status

Environment:
  RAG_DATA_DIR               Data directory (default: .rag-assistant)
  RAG_LOG_DIR                Log directory (default: .rag-assistant/logs)
  RAG_SHUTDOWN_TIMEOUT       Seconds to wait before SIGKILL (default: 5)
EOF
}

# --- start ---
cmd_start() {
  local service="$1"; shift
  local command="$*"
  [[ -z "$service" || -z "$command" ]] && { echo "Usage: $0 start <name> <command...>" >&2; return 2; }
  process_ensure_dir; mkdir -p "$RAG_RUNTIME_LOG_DIR"
  local pidfile="$RAG_RUNTIME_DIR/$service.pid"
  local lockdir="$RAG_RUNTIME_DIR/$service.lock"
  local logfile="$RAG_RUNTIME_DIR/$service.log"
  if ! mkdir "$lockdir" 2>/dev/null; then echo "ERROR: lock held for $service" >&2; return 1; fi
  trap 'rmdir "$lockdir" 2>/dev/null' RETURN
  local existing_pid; existing_pid=$(process_read_pid "$pidfile")
  if [[ -n "$existing_pid" ]]; then echo "OK: $service already running (pid $existing_pid)"; return 0; fi
  setsid bash -c "$command" >>"$logfile" 2>&1 &
  local new_pid=$!; sleep 0.3
  if ! kill -0 "$new_pid" 2>/dev/null; then echo "ERROR: $service failed to start" >&2; return 1; fi
  if ! process_acquire "$service" "$pidfile" "$new_pid"; then
    echo "ERROR: PID identity conflict for $service" >&2; kill -TERM "$new_pid" 2>/dev/null || true; return 1
  fi
  echo "OK: $service started (pid $new_pid)"
}

# --- stop ---
cmd_stop() {
  local service="$1"
  [[ -z "$service" ]] && { echo "Usage: $0 stop <name>" >&2; return 2; }
  local pidfile="$RAG_RUNTIME_DIR/$service.pid"
  local lockdir="$RAG_RUNTIME_DIR/$service.lock"
  if ! mkdir "$lockdir" 2>/dev/null; then echo "ERROR: lock held for $service" >&2; return 1; fi
  trap 'rmdir "$lockdir" 2>/dev/null' RETURN
  local pid; pid=$(process_read_pid "$pidfile")
  if [[ -z "$pid" ]]; then echo "ERROR: $service not running" >&2; return 1; fi
  wait_terminate "$pid" "$RAG_SHUTDOWN_TIMEOUT"
  process_release "$pidfile" 2>/dev/null || true; rm -f "$pidfile"
  if kill -0 "$pid" 2>/dev/null; then echo "ERROR: $service still alive after SIGKILL" >&2; return 1; fi
  echo "OK: $service stopped"
}

# --- status ---
cmd_status() {
  local service="$1"
  [[ -z "$service" ]] && { echo "Usage: $0 status <name>" >&2; return 2; }
  local pidfile="$RAG_RUNTIME_DIR/$service.pid"
  local pid; pid=$(process_read_pid "$pidfile")
  if [[ -z "$pid" ]]; then echo "STOPPED: $service"; return 1; fi
  echo "RUNNING: $service (pid $pid)"; return 0
}

# --- main (only when executed, not sourced) ---
if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  cmd="${1:-}"; shift 2>/dev/null || true
  case "$cmd" in
    start)  cmd_start "$@" ;;
    stop)   cmd_stop "$@" ;;
    status) cmd_status "$@" ;;
    *)      usage; exit 2 ;;
  esac
fi
