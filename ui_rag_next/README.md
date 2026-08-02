# RAG Assistant UI

Standalone Next.js frontend for the RAG Assistant backend. It provides responsive navigation, streamed Ask, document, and conversation workflows, honest resource empty states, typed API boundaries, and browser-side health/readiness checks.

## Requirements

- Node.js compatible with Next.js 14
- pnpm 10.14.0
- A pnpm store already containing the locked packages for offline installation

## Offline Setup

```bash
pnpm install --offline --frozen-lockfile
cp .env.example .env.local
pnpm dev
```

Browser requests use the same-origin `/backend` prefix, which Next.js proxies to
`RAG_API_URL` (default `http://localhost:8080`). This server-only variable is read
when Next.js starts, so restart the UI after changing it.

## Validation

```bash
pnpm lint
pnpm typecheck
pnpm test
pnpm build
pnpm install --offline --frozen-lockfile
```

## Current Scope

- App Router routes for overview, queries, documents, collections, agents, conversations, and system status
- A streamed Ask workflow with cancellation, page-session follow-ups, explicit query phases, plain-text answers, and deduplicated citations
- A live Documents page with typed registry listing, refresh/error/empty states, and cancellable ingestion of `.txt`, `.md`, and `.markdown` paths relative to the backend's configured `RAG_INGEST_ROOT`
- A live Conversations page with deterministic history, bounded turn/citation rendering, confirmed deletion, and safe continuation into Ask through `conversation_id`
- Ingestion reports indexed, unindexed, and unsupported outcomes, but does not create CRUD registry records: the backend ingest service writes the retrieval index independently from `DocumentRepo`
- A typed fetch/`ReadableStream` SSE client with incremental parsing, terminal-frame validation, and distinct backend, HTTP, malformed-stream, and cancellation failures; `EventSource` is not used
- Exact runtime-validated contracts for conversation summaries/details, document registry records, and relative-path ingestion responses without normalized document content, plus health, readiness, query frames, and remaining CRUD resources
- No fabricated metrics, records, conversations, or backend response fields
- Rename remains intentionally unavailable because the backend does not persist conversation names

This app is independent from Staffflow. Staffflow supplied only known-compatible dependency versions and general configuration patterns; none of its business pages, services, authentication, data, assets, or environment files are included.
