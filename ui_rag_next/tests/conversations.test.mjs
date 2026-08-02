import assert from "node:assert/strict";
import test from "node:test";

import {
  askConversationHref,
  boundedTurn,
  conversationIdFromSearch,
  conversationPath,
  deriveConversationRows,
  deriveConversationSummaryRows,
  describeConversationError,
  mergeConversationRow,
  parseConversation,
  parseConversationSummaries,
  visibleCitationLimit,
} from "../lib/conversations.mjs";

const turn = (query, created_at, citations = []) => ({ query, state: "answered", answer: `Answer to ${query}`, citations, created_at });

test("conversation payload parsing accepts exact Go envelopes and strips unknown fields", () => {
  assert.deepEqual(parseConversationSummaries([{ id: "conv-1", turns_count: 2, unexpected: "ignored" }]), [{ id: "conv-1", turns_count: 2 }]);
  assert.deepEqual(parseConversation({ id: "conv-1", turns: [turn("Hello", "2026-08-01T10:00:00Z")], secret: "ignored" }), {
    id: "conv-1",
    turns: [turn("Hello", "2026-08-01T10:00:00Z")],
  });
  assert.equal(parseConversationSummaries([{ id: "conv-1", turns_count: -1 }]), null);
  assert.equal(parseConversation({ id: "conv-1", turns: [{ ...turn("Hello", "not-a-date") }] }), null);
  assert.equal(parseConversation({ id: "conv-1", turns: [{ ...turn("Hello", "2026-08-01T10:00:00Z"), state: "invented" }] }), null);
});

test("conversation rows derive preview and count with deterministic newest-first ordering", () => {
  const rows = deriveConversationRows([
    { id: "conv-b", turns: [turn("Older question", "2026-08-01T09:00:00Z")] },
    { id: "conv-a", turns: [turn("First", "2026-08-01T08:00:00Z"), turn("  Latest   question  ", "2026-08-01T10:00:00Z")] },
    { id: "conv-empty", turns: [] },
  ]);
  assert.deepEqual(rows.map(({ id }) => id), ["conv-a", "conv-b", "conv-empty"]);
  assert.equal(rows[0].preview, "Latest question");
  assert.equal(rows[0].turnsCount, 2);
  assert.equal(rows[2].updatedAt, null);
});

test("conversation summaries render deterministically before details hydrate independently", () => {
  const summaries = deriveConversationSummaryRows([
    { id: "conv-b", turns_count: 2 },
    { id: "conv-a", turns_count: 1 },
  ]);
  assert.deepEqual(summaries.map(({ id }) => id), ["conv-a", "conv-b"]);
  assert.deepEqual(summaries.map(({ turnsCount }) => turnsCount), [1, 2]);

  const hydrated = mergeConversationRow(summaries, { id: "conv-b", turns: [turn("Ready", "2026-08-01T10:00:00Z")] });
  assert.deepEqual(hydrated.map(({ id }) => id), ["conv-b", "conv-a"]);
  assert.equal(hydrated[0].preview, "Ready");
  assert.equal(hydrated[1].turnsCount, 1);
});

test("conversation URLs encode IDs and resume only valid query values", () => {
  const id = "conv/a?b=c";
  assert.equal(conversationPath(id), "/v1/conversations/conv%2Fa%3Fb%3Dc");
  assert.equal(askConversationHref(id), "/ask?conversation_id=conv%2Fa%3Fb%3Dc");
  assert.equal(conversationIdFromSearch("?conversation_id=conv%2Fa%3Fb%3Dc"), id);
  assert.equal(conversationIdFromSearch("?conversation_id=%0Aunsafe"), null);
  assert.equal(conversationIdFromSearch("?other=value"), null);
});

test("rendered turns bound text and citation exposure", () => {
  const citations = Array.from({ length: 12 }, (_, index) => ({ document_id: `doc-${index}`, chunk_id: `chunk-${index}`, snippet: "x".repeat(700) }));
  const visible = boundedTurn({ ...turn("q".repeat(9000), "2026-08-01T10:00:00Z", citations), answer: "a".repeat(13000) });
  assert.equal(visible.citations.length, visibleCitationLimit());
  assert.match(visible.query, /\.\.\.$/);
  assert.match(visible.answer, /\.\.\.$/);
  assert.ok(visible.citations[0].snippet.length <= 503);
});

test("conversation errors distinguish 404 races, backend detail, and availability", () => {
  assert.equal(describeConversationError({ name: "APIHttpError", status: 404 }), "The conversation no longer exists. The list has been refreshed.");
  assert.equal(describeConversationError({ name: "APIHttpError", status: 422, backendError: { message: "invalid conversation" } }), "invalid conversation");
  assert.equal(describeConversationError({ name: "APIHttpError", status: 503 }), "The backend is unavailable (HTTP 503).");
  assert.equal(describeConversationError({ name: "BackendUnavailableError" }), "The backend is unavailable. Check the service and try again.");
});
