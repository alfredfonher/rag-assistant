# RAG Assistant UI

Standalone Next.js frontend for the RAG Assistant backend. It provides responsive navigation, a streamed Ask workflow, honest resource empty states, typed API boundaries, and browser-side health/readiness checks.

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
- A typed fetch/`ReadableStream` SSE client with incremental parsing, terminal-frame validation, and distinct backend, HTTP, malformed-stream, and cancellation failures; `EventSource` is not used
- Typed contracts for health, readiness, query frames, path-based ingestion, and CRUD resources
- No fabricated metrics, records, conversations, or backend response fields
- CRUD controls remain intentionally unwired until their concrete backend payload schemas are validated

This app is independent from Staffflow. Staffflow supplied only known-compatible dependency versions and general configuration patterns; none of its business pages, services, authentication, data, assets, or environment files are included.
