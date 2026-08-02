import assert from "node:assert/strict";
import test from "node:test";

import {
  askConversationHref,
  boundedTurn,
  conversationIdFromSearch,
  conversationPath,
  detailHydrationConcurrency,
  deriveConversationRows,
  deriveConversationSummaryRows,
  describeConversationError,
  hydrateConversationDetails,
  markConversationRowUnavailable,
  mergeConversationRow,
  parseConversation,
  parseConversationSummaries,
  visibleCitationLimit,
  retainedTurnLimit,
} from "../lib/conversations.mjs";

const turn = (query, created_at, citations = []) => ({ query, state: "answered", answer: `Answer to ${query}`, citations, created_at });

test("conversation payload parsing accepts exact Go envelopes and strips unknown fields", () => {
  assert.deepEqual(parseConversationSummaries([{ id: "conv-1", turns_count: 2, unexpected: "ignored" }]), [{ id: "conv-1", turns_count: 2 }]);
  assert.deepEqual(parseConversation({ id: "conv-1", turns: [turn("Hello", "2026-08-01T10:00:00Z")], secret: "ignored" }), {
    id: "conv-1",
    turns: [turn("Hello", "2026-08-01T10:00:00Z")],
  });
  assert.equal(parseConversationSummaries([{ id: "conv-1", turns_count: -1 }]), null);
  assert.equal(parseConversationSummaries([{ id: "conv/child", turns_count: 1 }]), null);
  assert.equal(parseConversation({ id: "conv-1", turns: [{ ...turn("Hello", "not-a-date") }] }), null);
  assert.equal(parseConversation({ id: "conv/child", turns: [] }), null);
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
  const id = "conv?a=b #1";
  assert.equal(conversationPath(id), "/v1/conversations/conv%3Fa%3Db%20%231");
  assert.equal(askConversationHref(id), "/ask?conversation_id=conv%3Fa%3Db+%231");
  assert.equal(conversationIdFromSearch("?conversation_id=conv%3Fa%3Db+%231"), id);
  assert.throws(() => conversationPath("conv/child"), /Invalid conversation ID/);
  assert.throws(() => askConversationHref("conv/child"), /Invalid conversation ID/);
  assert.equal(conversationIdFromSearch("?conversation_id=conv%2Fchild"), null);
  assert.equal(conversationIdFromSearch("?conversation_id=%0Aunsafe"), null);
  assert.equal(conversationIdFromSearch("?other=value"), null);
});

test("rendered turns bound text and citation exposure", () => {
  const citations = Array.from({ length: 12 }, (_, index) => ({ document_id: `doc-${index}`, chunk_id: `chunk-${index}`, snippet: "x".repeat(700) }));
  const visible = boundedTurn({ ...turn("q".repeat(9000), "2026-08-01T10:00:00Z", citations), answer: "a".repeat(13000) });
  assert.equal(visible.citations.length, visibleCitationLimit());
  assert.match(visible.query, /\.\.\.$/);
  assert.match(visible.answer, /\.\.\.$/);
  assert.ok(visible.citations[0].snippet.length <= 500);
});

test("hydrated rows retain bounded recent detail while preserving truthful totals", () => {
  const summaries = deriveConversationSummaryRows([{ id: "conv-1", turns_count: 75 }]);
  const turns = Array.from({ length: 75 }, (_, index) => turn(`Question ${index} ${"q".repeat(9000)}`, `2026-08-01T10:${String(index % 60).padStart(2, "0")}:00Z`, Array.from({ length: 12 }, (__, citationIndex) => ({
    document_id: `doc-${citationIndex}-${"d".repeat(300)}`,
    chunk_id: `chunk-${citationIndex}-${"c".repeat(300)}`,
    snippet: "s".repeat(700),
  }))));
  const [row] = mergeConversationRow(summaries, { id: "conv-1", turns });

  assert.equal(row.turnsCount, 75);
  assert.equal(row.retainedTurnOffset, 75 - retainedTurnLimit());
  assert.equal(row.conversation.turns.length, retainedTurnLimit());
  assert.match(row.conversation.turns[0].query, /^Question 25 /);
  assert.ok(row.conversation.turns[0].query.length <= 8000);
  assert.equal(row.conversation.turns[0].citations.length, visibleCitationLimit());
  assert.equal(row.conversation.turns[0].hiddenCitationCount, 4);
  assert.ok(row.conversation.turns[0].citations[0].document_id.length <= 256);
});

test("detail hydration bounds concurrency and isolates partial failures in input order", async () => {
  const summaries = Array.from({ length: 9 }, (_, index) => ({ id: `conv-${index}`, turns_count: 1 }));
  let active = 0;
  let maximum = 0;
  const callbacks = [];
  const results = await hydrateConversationDetails(summaries, async (id) => {
    active += 1;
    maximum = Math.max(maximum, active);
    await new Promise((resolve) => setTimeout(resolve, id === "conv-0" ? 15 : 1));
    active -= 1;
    if (id === "conv-3") throw new Error("detail failed");
    return { id, turns: [] };
  }, { onResult: (result) => callbacks.push(result), timeoutMs: 1000 });

  assert.equal(maximum, detailHydrationConcurrency());
  assert.deepEqual(results.map(({ id }) => id), summaries.map(({ id }) => id));
  assert.equal(results.find(({ id }) => id === "conv-3").status, "rejected");
  assert.equal(callbacks.length, summaries.length);

  const unavailable = markConversationRowUnavailable(deriveConversationSummaryRows(summaries), "conv-3");
  assert.equal(unavailable.find(({ id }) => id === "conv-3").hydrationState, "error");
  assert.equal(unavailable.find(({ id }) => id === "conv-4").hydrationState, "loading");
});

test("detail hydration times out requests and suppresses stale callbacks after refresh abort", async () => {
  const timeoutResults = await hydrateConversationDetails([{ id: "slow", turns_count: 1 }], (_id, { signal }) => new Promise((resolve, reject) => {
    signal.addEventListener("abort", () => reject(signal.reason), { once: true });
  }), { timeoutMs: 5 });
  assert.equal(timeoutResults[0].status, "rejected");
  assert.equal(timeoutResults[0].error.name, "TimeoutError");

  const controller = new AbortController();
  const callbacks = [];
  const hydration = hydrateConversationDetails(
    [{ id: "stale-1", turns_count: 1 }, { id: "stale-2", turns_count: 1 }],
    (_id, { signal }) => new Promise((resolve, reject) => signal.addEventListener("abort", () => reject(signal.reason), { once: true })),
    { signal: controller.signal, onResult: (result) => callbacks.push(result), timeoutMs: 1000 },
  );
  controller.abort();
  await hydration;
  assert.deepEqual(callbacks, []);
});

test("conversation errors distinguish 404 races, backend detail, and availability", () => {
  assert.equal(describeConversationError({ name: "APIHttpError", status: 404 }), "The conversation no longer exists. The list has been refreshed.");
  assert.equal(describeConversationError({ name: "APIHttpError", status: 422, backendError: { message: "invalid conversation" } }), "invalid conversation");
  assert.equal(describeConversationError({ name: "APIHttpError", status: 503 }), "The backend is unavailable (HTTP 503).");
  assert.equal(describeConversationError({ name: "BackendUnavailableError" }), "The backend is unavailable. Check the service and try again.");
});
