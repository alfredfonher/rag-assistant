# RAG Assistant UI

Standalone Next.js frontend shell for the RAG Assistant backend. It provides responsive navigation, honest resource empty states, typed API boundaries, and browser-side health/readiness checks.

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
pnpm build
pnpm install --offline --frozen-lockfile
```

## Current Scope

- App Router routes for overview, queries, documents, collections, agents, conversations, and system status
- Typed contracts for health, readiness, query, streaming query, path-based ingestion, and CRUD resources
- SSE streaming represented as a fetch `ReadableStream`; `EventSource` is not used
- No fabricated metrics, records, conversations, or backend response fields
- Query and CRUD controls are intentionally not wired until concrete backend payload schemas are validated

This app is independent from Staffflow. Staffflow supplied only known-compatible dependency versions and general configuration patterns; none of its business pages, services, authentication, data, assets, or environment files are included.
