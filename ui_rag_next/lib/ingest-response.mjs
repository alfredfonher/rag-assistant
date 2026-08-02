function isRecord(value) {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isAPIError(value) {
  return isRecord(value) && typeof value.code === "string" && typeof value.message === "string";
}

function isCitation(value) {
  return isRecord(value)
    && typeof value.document_id === "string"
    && typeof value.chunk_id === "string"
    && (value.snippet === undefined || typeof value.snippet === "string");
}

function isIngestResponse(value) {
  return isRecord(value)
    && ["indexed", "unindexed", "unsupported"].includes(value.state)
    && (value.document_id === undefined || typeof value.document_id === "string")
    && (value.citations === undefined || (Array.isArray(value.citations) && value.citations.every(isCitation)))
    && (value.error === undefined || isAPIError(value.error));
}

export class APIHttpError extends Error {
  constructor(status, backendError) {
    super(backendError?.message ?? `The backend returned HTTP ${status}.`);
    this.name = "APIHttpError";
    this.status = status;
    this.backendError = backendError;
  }
}

export function parseIngestResponse(status, ok, value) {
  if (!ok) {
    const backendError = isIngestResponse(value) ? value.error : isAPIError(value) ? value : undefined;
    throw new APIHttpError(status, backendError);
  }
  if (isIngestResponse(value)) return value;
  return undefined;
}
