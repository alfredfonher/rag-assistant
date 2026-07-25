#!/usr/bin/env bash
set -Eeuo pipefail
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
MANIFEST=${RAG_MODEL_MANIFEST:-$ROOT/models/manifest.tsv}
MODELS=${RAG_MODELS_ROOT:-$ROOT/models}
STATE=${RAG_MODEL_STATE:-$ROOT/.rag-assistant}
RECEIPT=$STATE/verified-models.tsv
mkdir -p "$MODELS" "$STATE"
die() { rm -f "$RECEIPT"; printf 'models: %s\n' "$*" >&2; exit 1; }
manifest_digest=$(sha256sum "$MANIFEST" | cut -d' ' -f1) || die 'manifest is unreadable'
read_manifest() {
  local id name url sha extra
  while IFS=$'\t' read -r id name url sha extra || [[ -n $id ]]; do
    [[ -z $id || $id == \#* ]] && continue
    [[ -n $id && -n $name && -n $url && $sha =~ ^[0-9a-fA-F]{64}$ && -z ${extra:-} ]] || die 'malformed manifest or missing checksum'
    printf '%s\t%s\t%s\t%s\n' "$id" "$name" "$url" "${sha,,}"
  done < "$MANIFEST"
}
fingerprint() { stat -c '%d\t%i\t%s\t%Y\t%Z' "$1"; }
receipt_row() { awk -F '\t' -v id="$1" '$1 == id {print; exit}' "$RECEIPT" 2>/dev/null || true; }
verify_one() {
  local id=$1 name=$2 expected=$3 target row actual fp
  target=$MODELS/$name
  [[ -f $target ]] || return 1
  actual=$(sha256sum "$target" | cut -d' ' -f1)
  [[ $actual == "$expected" ]] || return 1
  fp=$(fingerprint "$target")
  row=$(receipt_row "$id")
  [[ $row == "$id"$'\t'"$manifest_digest"$'\t'"$target"$'\t'"$expected"$'\t'"$fp" ]]
}
write_receipt() {
  local tmp=$STATE/.verified-models.tsv.tmp id name expected target fp
  : > "$tmp"
  while IFS=$'\t' read -r id name _ expected; do
    target=$MODELS/$name; fp=$(fingerprint "$target")
    printf '%s\t%s\t%s\t%s\t%s\n' "$id" "$manifest_digest" "$target" "$expected" "$fp" >> "$tmp"
  done < <(read_manifest)
  mv -f "$tmp" "$RECEIPT"
}
verify_all() {
  local id name _ expected
  [[ -s $RECEIPT ]] || return 1
  while IFS=$'\t' read -r id name _ expected; do verify_one "$id" "$name" "$expected" || return 1; done < <(read_manifest)
}
fetch_all() {
  local id name url expected target partial
  while IFS=$'\t' read -r id name url expected; do
    target=$MODELS/$name; partial=$target.partial
    if verify_one "$id" "$name" "$expected"; then continue; fi
    rm -f "$RECEIPT"
    curl --fail --location --retry 2 --continue-at - --output "$partial" "$url" || { rm -f "$partial"; die "download failed for $id"; }
    [[ $(sha256sum "$partial" | cut -d' ' -f1) == "$expected" ]] || { rm -f "$partial"; die "checksum mismatch for $id"; }
    mv -f "$partial" "$target"
  done < <(read_manifest)
  write_receipt
  verify_all || die 'verification receipt is invalid'
  printf 'Models verified: %s\n' "$RECEIPT"
}
case ${1:-} in
  fetch) fetch_all ;;
  verify) verify_all || die 'models are missing, mutated, or unverified'; printf 'Models verified: %s\n' "$RECEIPT" ;;
  *) printf 'usage: %s fetch|verify\n' "$0" >&2; exit 2 ;;
esac
