#!/usr/bin/env bash

load_rag_env() {
  local file=${1:-.env} line key value
  [[ -f "$file" ]] || return 1
  while IFS= read -r line || [[ -n "$line" ]]; do
    [[ $line =~ ^[[:space:]]*# || $line =~ ^[[:space:]]*$ ]] && continue
    [[ $line =~ ^[[:space:]]*([A-Za-z_][A-Za-z0-9_]*)[[:space:]]*=[[:space:]]*(.*)[[:space:]]*$ ]] || return 1
    key=${BASH_REMATCH[1]}; value=${BASH_REMATCH[2]}
    if [[ $value == \"*\" && $value == *\" ]]; then value=${value:1:${#value}-2}; fi
    if [[ $value == \'*\' && $value == *\' ]]; then value=${value:1:${#value}-2}; fi
    [[ -v "$key" ]] || export "$key=$value"
  done < "$file"
}

export_rag_defaults() {
  : "${RAG_LLAMA_HOST:=127.0.0.1}"
  : "${RAG_LLAMA_PORT:=8090}"
  : "${RAG_SERVICE_PORT:=8080}"
  : "${RAG_LLAMA_N_CTX:=2048}"
  : "${RAG_LLAMA_N_THREADS:=0}"
  export RAG_LLAMA_HOST RAG_LLAMA_PORT RAG_SERVICE_PORT RAG_LLAMA_N_CTX RAG_LLAMA_N_THREADS
}
