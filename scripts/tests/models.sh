#!/usr/bin/env bash
set -Eeuo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
[[ -x "$ROOT/scripts/models.sh" ]] || fail "models script missing"
[[ -f "$ROOT/models/manifest.tsv" ]] || fail "manifest missing"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/bin" "$work/models" "$work/state"
cat > "$work/bin/curl" <<'EOF'
#!/usr/bin/env bash
set -Eeuo pipefail
out=''
while (($#)); do case "$1" in -o|--output) out=$2; shift 2;; -C|--continue-at|--retry) shift 2;; --fail|--location) shift;; *) shift;; esac; done
[[ -n "$out" ]] || exit 2
[[ ${MOCK_FAIL:-0} == 1 ]] && exit 22
printf '%s' '-bytes' >> "$out"
EOF
chmod +x "$work/bin/curl"
hash=$(printf fixture-model-bytes | sha256sum | cut -d' ' -f1)
printf 'fixture\tfixture.gguf\thttps://example.invalid/fixture\t%s\n' "$hash" > "$work/manifest.tsv"
printf 'fixture-model' > "$work/models/fixture.gguf.partial"

env PATH="$work/bin:$PATH" RAG_MODELS_ROOT="$work/models" RAG_MODEL_STATE="$work/state" \
  RAG_MODEL_MANIFEST="$work/manifest.tsv" bash "$ROOT/scripts/models.sh" fetch >/dev/null \
  || fail "fixture fetch failed"
[[ -f "$work/state/verified-models.tsv" ]] || fail "receipt missing"
grep -q $'fixture\t' "$work/state/verified-models.tsv" || fail "receipt identity missing"
mtime=$(stat -c %Y "$work/models/fixture.gguf")
env PATH="$work/bin:$PATH" RAG_MODELS_ROOT="$work/models" RAG_MODEL_STATE="$work/state" \
  RAG_MODEL_MANIFEST="$work/manifest.tsv" bash "$ROOT/scripts/models.sh" fetch >/dev/null \
  || fail "idempotent fetch failed"
[[ $(stat -c %Y "$work/models/fixture.gguf") == "$mtime" ]] || fail "verified artifact rewritten"

printf bad > "$work/models/fixture.gguf"
set +e
env PATH="$work/bin:$PATH" RAG_MODELS_ROOT="$work/models" RAG_MODEL_STATE="$work/state" \
  RAG_MODEL_MANIFEST="$work/manifest.tsv" bash "$ROOT/scripts/models.sh" verify >/dev/null 2>&1
status=$?
set -e
(( status != 0 )) || fail "mismatch was accepted"
[[ ! -f "$work/state/verified-models.tsv" ]] || fail "stale receipt survived mismatch"

printf 'fixture\tfixture.gguf\thttps://example.invalid/fixture\n' > "$work/manifest.tsv"
set +e
env PATH="$work/bin:$PATH" RAG_MODELS_ROOT="$work/models" RAG_MODEL_STATE="$work/state" \
  RAG_MODEL_MANIFEST="$work/manifest.tsv" bash "$ROOT/scripts/models.sh" fetch >/dev/null 2>&1
status=$?
set -e
(( status != 0 )) || fail "missing checksum was accepted"

printf 'fixture\tfixture.gguf\thttps://example.invalid/fixture\t%s\n' "$hash" > "$work/manifest.tsv"
rm -f "$work/models/fixture.gguf" "$work/models/fixture.gguf.partial" "$work/state/verified-models.tsv"
set +e
env MOCK_FAIL=1 PATH="$work/bin:$PATH" RAG_MODELS_ROOT="$work/models" RAG_MODEL_STATE="$work/state" \
  RAG_MODEL_MANIFEST="$work/manifest.tsv" bash "$ROOT/scripts/models.sh" fetch >/dev/null 2>&1
status=$?
set -e
(( status != 0 )) || fail "unavailable URL was accepted"
[[ ! -f "$work/state/verified-models.tsv" ]] || fail "unavailable URL left receipt"

printf 'models tests: PASS\n'
