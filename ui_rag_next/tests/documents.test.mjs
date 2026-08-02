import assert from "node:assert/strict";
import test from "node:test";

import {
  createIngestResult,
  describeDocumentRequestError,
  validateDocumentPath,
} from "../lib/documents.mjs";
import { APIHttpError, parseIngestResponse } from "../lib/ingest-response.mjs";

test("path validation requires a supported relative document path", () => {
  assert.equal(validateDocumentPath("   "), "Enter a path relative to the configured ingest root.");
  assert.equal(validateDocumentPath("/data/guide.md"), "Use a relative path, not an absolute path.");
  assert.equal(validateDocumentPath("C:\\data\\guide.md"), "Use a relative path, not an absolute path.");
  assert.equal(validateDocumentPath("../guide.md"), "Path traversal is not allowed.");
  assert.equal(validateDocumentPath("guides/../guide.md"), "Path traversal is not allowed.");
  assert.equal(validateDocumentPath("guide.pdf"), "Use a .txt, .md, or .markdown document.");
  assert.equal(validateDocumentPath(" guides/GUIDE.MARKDOWN "), null);
  assert.equal(validateDocumentPath("notes.txt"), null);
});

test("ingest state keeps indexed, unindexed, and unsupported outcomes distinct", () => {
  assert.deepEqual(createIngestResult({ state: "indexed", citations: [{ document_id: "d", chunk_id: "c" }] }), {
    tone: "success",
    title: "Document indexed",
    message: "The retrieval index accepted the document.",
    citationCount: 1,
  });
  assert.equal(createIngestResult({ state: "unindexed", error: { code: "embedding_unavailable", message: "embedding failed" } }).message, "embedding failed");
  assert.equal(createIngestResult({ state: "unsupported" }).title, "Document unsupported");
});

test("request errors prefer backend detail and distinguish transport failures", () => {
  assert.equal(describeDocumentRequestError({ name: "BackendUnavailableError" }), "The backend is unavailable. Check the service and try again.");
  assert.equal(describeDocumentRequestError({ name: "APIHttpError", status: 422, backendError: { message: "empty document" } }), "empty document");
  assert.equal(describeDocumentRequestError({ name: "APIHttpError", status: 503 }), "The backend is unavailable (HTTP 503).");
  assert.equal(describeDocumentRequestError({ name: "MalformedAPIResponseError", message: "Invalid list." }), "Invalid list.");
});

test("non-success ingest envelopes retain backend status and error", () => {
  const backendError = { code: "ingest_root_unavailable", message: "configured ingest root is unavailable" };
  assert.throws(
    () => parseIngestResponse(503, false, { state: "unsupported", error: backendError }),
    (error) => error instanceof APIHttpError
      && error.status === 503
      && error.backendError === backendError,
  );
});

test("successful ingest envelopes remain valid responses", () => {
  const response = { state: "indexed", document_id: "guide", citations: [] };
  assert.equal(parseIngestResponse(201, true, response), response);
});

test("stopping an ingest wait does not claim server-side cancellation", () => {
  assert.equal(
    describeDocumentRequestError({ name: "AbortError" }),
    "Stopped waiting for the ingestion response. The server may still complete the operation; check the document index before trying again.",
  );
});
