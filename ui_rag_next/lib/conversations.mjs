const queryStates = new Set([
  "streaming",
  "retrieving",
  "answered",
  "insufficient_context",
  "unsupported",
]);

const maxConversationIdLength = 512;
const maxVisibleCitations = 8;

export function visibleCitationLimit() {
  return maxVisibleCitations;
}

function isRecord(value) {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

export function normalizeConversationId(value) {
  if (typeof value !== "string" || value.length === 0 || value.length > maxConversationIdLength) return null;
  if (/^[\s]|[\s]$|[\u0000-\u001f\u007f]/.test(value)) return null;
  return value;
}

function parseCitation(value) {
  if (!isRecord(value)
    || typeof value.document_id !== "string"
    || typeof value.chunk_id !== "string"
    || (value.snippet !== undefined && typeof value.snippet !== "string")) return null;
  return {
    document_id: value.document_id,
    chunk_id: value.chunk_id,
    ...(value.snippet === undefined ? {} : { snippet: value.snippet }),
  };
}

export function parseConversationSummaries(value) {
  if (!Array.isArray(value)) return null;
  const summaries = [];
  for (const item of value) {
    if (!isRecord(item)
      || !normalizeConversationId(item.id)
      || !Number.isInteger(item.turns_count)
      || item.turns_count < 0) return null;
    summaries.push({ id: item.id, turns_count: item.turns_count });
  }
  return summaries;
}

export function parseConversation(value) {
  if (!isRecord(value) || !normalizeConversationId(value.id) || !Array.isArray(value.turns)) return null;
  const turns = [];
  for (const turn of value.turns) {
    if (!isRecord(turn)
      || typeof turn.query !== "string"
      || typeof turn.state !== "string"
      || !queryStates.has(turn.state)
      || typeof turn.answer !== "string"
      || typeof turn.created_at !== "string"
      || Number.isNaN(Date.parse(turn.created_at))) return null;

    let citations;
    if (turn.citations !== undefined) {
      if (!Array.isArray(turn.citations)) return null;
      citations = turn.citations.map(parseCitation);
      if (citations.some((citation) => citation === null)) return null;
    }
    turns.push({
      query: turn.query,
      state: turn.state,
      answer: turn.answer,
      ...(citations === undefined ? {} : { citations }),
      created_at: turn.created_at,
    });
  }
  return { id: value.id, turns };
}

function boundedText(value, limit) {
  if (value.length <= limit) return value;
  return `${value.slice(0, limit).trimEnd()}...`;
}

export function deriveConversationRows(conversations) {
  return conversations.map((conversation) => {
    const latestTurn = conversation.turns.reduce((latest, turn) => (
      !latest || Date.parse(turn.created_at) > Date.parse(latest.created_at) ? turn : latest
    ), null);
    const previewSource = latestTurn?.query.trim() || latestTurn?.answer.trim() || "No turn content";
    return {
      conversation,
      id: conversation.id,
      turnsCount: conversation.turns.length,
      preview: boundedText(previewSource.replace(/\s+/g, " "), 120),
      updatedAt: latestTurn?.created_at ?? null,
    };
  }).sort((left, right) => {
    const dateDifference = (right.updatedAt ? Date.parse(right.updatedAt) : -Infinity)
      - (left.updatedAt ? Date.parse(left.updatedAt) : -Infinity);
    return dateDifference || left.id.localeCompare(right.id);
  });
}

export function deriveConversationSummaryRows(summaries) {
  return summaries.map((summary) => ({
    conversation: { id: summary.id, turns: [] },
    id: summary.id,
    turnsCount: summary.turns_count,
    preview: "Turn details loading...",
    updatedAt: null,
  })).sort((left, right) => left.id.localeCompare(right.id));
}

export function mergeConversationRow(rows, conversation) {
  const hydrated = deriveConversationRows([conversation])[0];
  return rows.map((row) => row.id === hydrated.id ? hydrated : row).sort((left, right) => {
    const dateDifference = (right.updatedAt ? Date.parse(right.updatedAt) : -Infinity)
      - (left.updatedAt ? Date.parse(left.updatedAt) : -Infinity);
    return dateDifference || left.id.localeCompare(right.id);
  });
}

export function boundedTurn(turn) {
  return {
    query: boundedText(turn.query, 8000),
    answer: boundedText(turn.answer, 12000),
    citations: (turn.citations ?? []).slice(0, maxVisibleCitations).map((citation) => ({
      document_id: boundedText(citation.document_id, 256),
      chunk_id: boundedText(citation.chunk_id, 256),
      ...(citation.snippet ? { snippet: boundedText(citation.snippet, 500) } : {}),
    })),
  };
}

export function conversationPath(id) {
  const normalized = normalizeConversationId(id);
  if (!normalized) throw new TypeError("Invalid conversation ID.");
  return `/v1/conversations/${encodeURIComponent(normalized)}`;
}

export function askConversationHref(id) {
  const normalized = normalizeConversationId(id);
  if (!normalized) throw new TypeError("Invalid conversation ID.");
  return `/ask?${new URLSearchParams({ conversation_id: normalized }).toString()}`;
}

export function conversationIdFromSearch(search) {
  try {
    return normalizeConversationId(new URLSearchParams(search).get("conversation_id"));
  } catch {
    return null;
  }
}

export function describeConversationError(error) {
  if (error?.name === "BackendUnavailableError") return "The backend is unavailable. Check the service and try again.";
  if (error?.name === "APIHttpError") {
    if (error.status === 404) return "The conversation no longer exists. The list has been refreshed.";
    if (error.backendError?.message) return error.backendError.message;
    return error.status >= 500
      ? `The backend is unavailable (HTTP ${error.status}).`
      : `The backend rejected the request (HTTP ${error.status}).`;
  }
  if (error?.name === "MalformedAPIResponseError") return error.message;
  return "The request failed before a valid response was received.";
}
