import type {
  APIError,
  Citation,
  Document,
  IngestDocumentRequest,
  IngestDocumentResponse,
  QueryRequest,
  QueryResponse,
  QueryState,
  QueryStreamEventName,
  QueryStreamMessage,
  ResourceName,
  ResourceRecord,
  ServiceStatus,
} from "@/lib/contracts";
import { SseParser } from "@/lib/query-stream.mjs";

export const API_BASE_URL = "/backend";

export const endpoints = {
  health: "/healthz",
  readiness: "/readyz",
  query: "/v1/query",
  queryStream: "/v1/query/stream",
  ingest: "/v1/documents/ingest",
  resource: (name: ResourceName) => `/v1/${name}`,
} as const;

export class APIHttpError extends Error {
  constructor(
    public readonly status: number,
    public readonly backendError?: APIError,
  ) {
    super(backendError?.message ?? `The backend returned HTTP ${status}.`);
    this.name = "APIHttpError";
  }
}

export class MalformedAPIResponseError extends Error {
  constructor(message: string, public readonly cause?: unknown) {
    super(message);
    this.name = "MalformedAPIResponseError";
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...init?.headers,
      },
    });
  } catch (error) {
    if (isAbortError(error, init?.signal ?? undefined)) throw error;
    throw new BackendUnavailableError(error);
  }

  if (!response.ok) {
    throw new APIHttpError(response.status, await readAPIError(response));
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return readJSON(response, init?.signal ?? undefined) as Promise<T>;
}

const documentStatuses = new Set(["pending", "indexing", "ready", "error", "outdated"]);
const ingestStates = new Set(["indexed", "unindexed", "unsupported"]);

function isAPIError(value: unknown): value is APIError {
  return isRecord(value) && typeof value.code === "string" && typeof value.message === "string";
}

function isCitation(value: unknown): value is Citation {
  return isRecord(value)
    && typeof value.document_id === "string"
    && typeof value.chunk_id === "string"
    && (value.snippet === undefined || typeof value.snippet === "string");
}

function isDocument(value: unknown): value is Document {
  return isRecord(value)
    && typeof value.id === "string"
    && typeof value.collection_id === "string"
    && typeof value.filename === "string"
    && typeof value.path === "string"
    && typeof value.status === "string"
    && documentStatuses.has(value.status)
    && typeof value.chunks_count === "number"
    && (value.error_message === undefined || typeof value.error_message === "string")
    && typeof value.created_at === "string"
    && typeof value.updated_at === "string";
}

function isIngestResponse(value: unknown): value is IngestDocumentResponse {
  return isRecord(value)
    && typeof value.state === "string"
    && ingestStates.has(value.state)
    && (value.document_id === undefined || typeof value.document_id === "string")
    && (value.citations === undefined || (Array.isArray(value.citations) && value.citations.every(isCitation)))
    && (value.error === undefined || isAPIError(value.error));
}

async function readJSON(response: Response, signal?: AbortSignal): Promise<unknown> {
  try {
    return await response.json();
  } catch (error) {
    if (isAbortError(error, signal)) throw error;
    throw new MalformedAPIResponseError("The backend returned invalid JSON.", error);
  }
}

async function readAPIError(response: Response): Promise<APIError | undefined> {
  try {
    const value: unknown = await response.json();
    return isAPIError(value) ? value : undefined;
  } catch {
    return undefined;
  }
}

export async function listDocuments(options: { signal?: AbortSignal } = {}): Promise<Document[]> {
  const value = await request<unknown>(endpoints.resource("documents"), {
    cache: "no-store",
    signal: options.signal,
  });
  if (!Array.isArray(value) || !value.every(isDocument)) {
    throw new MalformedAPIResponseError("The backend returned an invalid document list.");
  }
  return value;
}

export async function ingestDocument(
  payload: IngestDocumentRequest,
  options: { signal?: AbortSignal } = {},
): Promise<IngestDocumentResponse> {
  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${endpoints.ingest}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      cache: "no-store",
      signal: options.signal,
    });
  } catch (error) {
    if (isAbortError(error, options.signal)) throw error;
    throw new BackendUnavailableError(error);
  }

  const value = await readJSON(response, options.signal);
  if (isIngestResponse(value)) return value;
  if (!response.ok && isAPIError(value)) throw new APIHttpError(response.status, value);
  throw new MalformedAPIResponseError("The backend returned an invalid ingest response.");
}

async function check(path: string): Promise<ServiceStatus> {
  const response = await fetch(`${API_BASE_URL}${path}`, { cache: "no-store" });
  if (!response.ok) {
    throw new Error(`API check failed with status ${response.status}`);
  }
  return { ok: true, status: response.status };
}

export const api = {
  health: () => check(endpoints.health),
  readiness: () => check(endpoints.readiness),
  query: (payload: QueryRequest) =>
    request<QueryResponse>(endpoints.query, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  documents: {
    list: listDocuments,
    ingest: ingestDocument,
  },
  list: <T extends ResourceRecord>(resource: ResourceName) =>
    request<T[]>(endpoints.resource(resource)),
  create: <T extends ResourceRecord>(resource: ResourceName, payload: unknown) =>
    request<T>(endpoints.resource(resource), {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  update: <T extends ResourceRecord>(
    resource: ResourceName,
    id: string,
    payload: unknown,
  ) =>
    request<T>(`${endpoints.resource(resource)}/${id}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    }),
  remove: (resource: ResourceName, id: string) =>
    request<void>(`${endpoints.resource(resource)}/${id}`, {
      method: "DELETE",
    }),
};

const queryStates = new Set<QueryState>([
  "streaming",
  "retrieving",
  "answered",
  "insufficient_context",
  "unsupported",
]);

const queryEvents = new Set<QueryStreamEventName>([
  "start",
  "retrieval",
  "content",
  "citation",
  "done",
]);

export class BackendUnavailableError extends Error {
  constructor(public readonly cause: unknown) {
    super("The backend could not be reached.");
    this.name = "BackendUnavailableError";
  }
}

export class QueryStreamHttpError extends Error {
  constructor(
    public readonly status: number,
    public readonly body: string,
  ) {
    super(`The backend returned HTTP ${status}${body ? `: ${body}` : "."}`);
    this.name = "QueryStreamHttpError";
  }
}

export class MalformedQueryStreamError extends Error {
  constructor(message: string, public readonly cause?: unknown) {
    super(message);
    this.name = "MalformedQueryStreamError";
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function parseFrame(data: string): QueryResponse {
  let value: unknown;
  try {
    value = JSON.parse(data);
  } catch (error) {
    throw new MalformedQueryStreamError("The backend sent invalid JSON in the event stream.", error);
  }

  if (!isRecord(value) || typeof value.state !== "string" || !queryStates.has(value.state as QueryState)) {
    throw new MalformedQueryStreamError("The backend sent a stream frame with an invalid state.");
  }

  for (const field of ["event", "kind", "answer", "conversation_id"] as const) {
    if (value[field] !== undefined && typeof value[field] !== "string") {
      throw new MalformedQueryStreamError(`The backend sent an invalid ${field} field.`);
    }
  }

  if (value.event !== undefined && !queryEvents.has(value.event as QueryStreamEventName)) {
    throw new MalformedQueryStreamError("The backend sent an unknown stream event.");
  }

  if (value.error !== undefined) {
    if (!isRecord(value.error) || typeof value.error.code !== "string" || typeof value.error.message !== "string") {
      throw new MalformedQueryStreamError("The backend sent an invalid error field.");
    }
  }

  if (value.citations !== undefined) {
    if (!Array.isArray(value.citations)) {
      throw new MalformedQueryStreamError("The backend sent an invalid citations field.");
    }
    for (const citation of value.citations) {
      if (
        !isRecord(citation) ||
        typeof citation.document_id !== "string" ||
        typeof citation.chunk_id !== "string" ||
        (citation.snippet !== undefined && typeof citation.snippet !== "string")
      ) {
        throw new MalformedQueryStreamError("The backend sent an invalid citation.");
      }
    }
  }

  return value as unknown as QueryResponse;
}

function isAbortError(error: unknown, signal?: AbortSignal): boolean {
  return signal?.aborted === true || (error instanceof DOMException && error.name === "AbortError");
}

export async function* queryStream(
  payload: QueryRequest,
  options: { signal?: AbortSignal } = {},
): AsyncGenerator<QueryStreamMessage, void, void> {
  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${endpoints.queryStream}`, {
      method: "POST",
      headers: {
        Accept: "text/event-stream",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
      cache: "no-store",
      signal: options.signal,
    });
  } catch (error) {
    if (isAbortError(error, options.signal)) throw error;
    throw new BackendUnavailableError(error);
  }

  if (!response.ok) {
    let body = "";
    try {
      body = (await response.text()).trim();
    } catch (error) {
      if (isAbortError(error, options.signal)) throw error;
    }
    throw new QueryStreamHttpError(response.status, body);
  }

  if (!response.body) {
    throw new MalformedQueryStreamError("The backend response did not contain an event stream body.");
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  const parser = new SseParser();
  let sawDone = false;

  const decodeMessages = (chunk: string, final = false): QueryStreamMessage[] => {
    const events = final ? parser.finish(chunk) : parser.push(chunk);
    return events.map(({ data, event, id }) => {
      const frame = parseFrame(data);
      if (event !== undefined && !queryEvents.has(event as QueryStreamEventName)) {
        throw new MalformedQueryStreamError(`The backend sent an unknown SSE event: ${event}.`);
      }
      const typedEvent = event as QueryStreamEventName | undefined;
      if (typedEvent === "done" || frame.event === "done" || ["answered", "insufficient_context", "unsupported"].includes(frame.state)) {
        sawDone = true;
      }
      return { event: typedEvent, id, frame };
    });
  };

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      for (const message of decodeMessages(decoder.decode(value, { stream: true }))) {
        yield message;
      }
    }

    for (const message of decodeMessages(decoder.decode(), true)) {
      yield message;
    }
  } catch (error) {
    if (isAbortError(error, options.signal) || error instanceof MalformedQueryStreamError) throw error;
    throw new BackendUnavailableError(error);
  } finally {
    reader.releaseLock();
  }

  if (!sawDone) {
    throw new MalformedQueryStreamError("The event stream ended before a terminal frame was received.");
  }
}
