#!/usr/bin/env bash
set -Eeuo pipefail
ROOT=${RAG_BOOTSTRAP_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}
missing=()
need() { command -v "$1" >/dev/null 2>&1 || missing+=("$1: $2"); }
need python3.12 'install distro Python 3.12'
need uv 'install uv from the official installer'
need bash 'install Bash >= 4.4'
need curl 'install curl'
need sha256sum 'install coreutils'
need setsid 'install util-linux'
need go 'install Go >= 1.22'
if ((${#missing[@]})); then
  printf 'Missing prerequisites:\n'; printf ' - %s\n' "${missing[@]}"; exit 2
fi
"$(command -v python3.12)" -c 'import sys; assert sys.version_info[:2] == (3, 12)'
cd "$ROOT"
uv sync --python 3.12 --project llama-server --extra dev --locked
printf 'Bootstrap complete: Python 3.12 locked environment is ready.\n'
