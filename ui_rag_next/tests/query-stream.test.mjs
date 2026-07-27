import assert from "node:assert/strict";
import test from "node:test";

import { applyQueryStreamMessage, createQueryViewState, SseParser } from "../lib/query-stream.mjs";

test("SSE parser handles arbitrary chunks, CRLF, fields, and multiple frames", () => {
  const parser = new SseParser();
  const messages = [
    ...parser.push(": keep-alive\r\nid: 1\r\nevent: sta"),
    ...parser.push("rt\r\ndata: {\"state\":\"streaming\"}\r"),
    ...parser.push("\n\r\nid: 2\nevent: retrieval\ndata: {\"state\":"),
    ...parser.push("\"retrieving\"}\n\n"),
    ...parser.finish(),
  ];

  assert.deepEqual(messages, [
    { id: "1", event: "start", data: '{"state":"streaming"}' },
    { id: "2", event: "retrieval", data: '{"state":"retrieving"}' },
  ]);
});

test("SSE parser joins data fields and dispatches a final unseparated frame", () => {
  const parser = new SseParser();
  const messages = parser.finish("event: done\ndata: {\"state\":\ndata: \"answered\"}");

  assert.deepEqual(messages, [
    { event: "done", id: undefined, data: '{"state":\n"answered"}' },
  ]);
});

test("query state assigns repeated complete answers and deduplicates citations", () => {
  let state = createQueryViewState();
  state = applyQueryStreamMessage(state, {
    event: "content",
    frame: {
      state: "streaming",
      answer: "Complete answer",
      conversation_id: "conversation-1",
      citations: [{ document_id: "doc-1", chunk_id: "chunk-1" }],
    },
  });
  state = applyQueryStreamMessage(state, {
    event: "done",
    frame: {
      state: "answered",
      answer: "Complete answer",
      citations: [
        { document_id: "doc-1", chunk_id: "chunk-1", snippet: "Relevant passage" },
        { document_id: "doc-2", chunk_id: "chunk-3" },
      ],
    },
  });

  assert.equal(state.answer, "Complete answer");
  assert.equal(state.phase, "completed");
  assert.equal(state.outcome, "answered");
  assert.equal(state.conversationId, "conversation-1");
  assert.deepEqual(state.citations, [
    { document_id: "doc-1", chunk_id: "chunk-1", snippet: "Relevant passage" },
    { document_id: "doc-2", chunk_id: "chunk-3" },
  ]);
});

test("query state separates insufficient context from backend errors", () => {
  const insufficient = applyQueryStreamMessage(createQueryViewState(), {
    event: "done",
    frame: {
      state: "insufficient_context",
      event: "done",
      kind: "completion",
      conversation_id: "conversation-1",
      error: { code: "insufficient_context", message: "no relevant context found" },
    },
  });
  const unsupported = applyQueryStreamMessage(createQueryViewState(), {
    event: "done",
    frame: {
      state: "unsupported",
      error: { code: "query_not_supported", message: "Query is not supported." },
    },
  });

  assert.equal(insufficient.phase, "completed");
  assert.equal(insufficient.outcome, "insufficient_context");
  assert.equal(insufficient.backendError.message, "no relevant context found");
  assert.equal(unsupported.phase, "error");
  assert.equal(unsupported.outcome, "backend_error");
});
