const supportedExtensions = [".txt", ".md", ".markdown"];

export function validateDocumentPath(path) {
  const normalized = path.trim();
  if (!normalized) return "Enter a path relative to the configured ingest root.";
  if (/^(?:[a-zA-Z]:[\\/]|[\\/]{1,2})/.test(normalized)) {
    return "Use a relative path, not an absolute path.";
  }
  if (normalized.split(/[\\/]+/).includes("..") || normalized === ".") {
    return "Path traversal is not allowed.";
  }
  const lower = normalized.toLowerCase();
  if (!supportedExtensions.some((extension) => lower.endsWith(extension))) {
    return "Use a .txt, .md, or .markdown document.";
  }
  return null;
}

export function createIngestResult(response) {
  const citationCount = response.citations?.length ?? 0;

  if (response.state === "indexed") {
    return {
      tone: "success",
      title: "Document indexed",
      message: "The retrieval index accepted the document.",
      citationCount,
    };
  }
  if (response.state === "unindexed") {
    return {
      tone: "error",
      title: "Document not indexed",
      message: response.error?.message ?? "The backend could not add the document to the retrieval index.",
      citationCount,
    };
  }
  return {
    tone: "warning",
    title: "Document unsupported",
    message: response.error?.message ?? "The backend rejected the document.",
    citationCount,
  };
}

export function describeDocumentRequestError(error) {
  if (error?.name === "BackendUnavailableError") {
    return "The backend is unavailable. Check the service and try again.";
  }
  if (error?.name === "APIHttpError") {
    if (error.backendError?.message) return error.backendError.message;
    return error.status >= 500
      ? `The backend is unavailable (HTTP ${error.status}).`
      : `The backend rejected the request (HTTP ${error.status}).`;
  }
  if (error?.name === "MalformedAPIResponseError") return error.message;
  return "The request failed before a valid response was received.";
}
