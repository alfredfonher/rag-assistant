#!/usr/bin/env bash
set -Eeuo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
[[ -x "$ROOT/scripts/bootstrap.sh" ]] || fail "bootstrap script missing"
[[ -f "$ROOT/scripts/lib/env.sh" ]] || fail "env library missing"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/bin" "$work/project"
for tool in bash curl sha256sum setsid flock go; do
  ln -s "$(command -v "$tool")" "$work/bin/$tool"
done
cat > "$work/bin/python3.12" <<'EOF'
#!/usr/bin/env bash
if [[ ${1:-} == -c ]]; then exit 0; fi
exit 0
EOF
cat > "$work/bin/uv" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" > "${UV_LOG:?}"
exit 0
EOF
chmod +x "$work/bin/python3.12" "$work/bin/uv"

set +e
PATH="$work/bin" RAG_BOOTSTRAP_ROOT="$work/project" UV_LOG="$work/uv.log" \
  bash "$ROOT/scripts/bootstrap.sh" >/dev/null 2>&1
status=$?
set -e
(( status == 0 )) || fail "supported prerequisites did not bootstrap"

cat > "$work/project/.env" <<'EOF'
RAG_LLAMA_PORT=8088
RAG_LLAMA_HOST="127.0.0.9"
EOF
RAG_BOOTSTRAP_ROOT="$work/project" RAG_LLAMA_PORT=9191 \
  bash -c 'source "$1/scripts/lib/env.sh"; load_rag_env "$2"; [[ $RAG_LLAMA_PORT == 9191 && $RAG_LLAMA_HOST == 127.0.0.9 ]]' _ "$ROOT" "$work/project/.env" \
  || fail "invocation environment did not win over .env"

printf 'bootstrap tests: PASS\n'
