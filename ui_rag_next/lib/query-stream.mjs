export class SseParser {
  constructor() {
    this.buffer = "";
    this.dataLines = [];
    this.event = undefined;
    this.id = undefined;
    this.finished = false;
  }

  push(chunk) {
    if (this.finished) throw new Error("Cannot push to a finished SSE parser.");
    this.buffer += chunk;
    return this.drain(false);
  }

  finish(chunk = "") {
    if (this.finished) throw new Error("Cannot finish an SSE parser twice.");
    this.buffer += chunk;
    this.finished = true;
    return this.drain(true);
  }

  drain(final) {
    const messages = [];

    while (this.buffer.length > 0) {
      const newline = this.buffer.search(/[\r\n]/);
      if (newline === -1) {
        if (!final) break;
        this.processLine(this.buffer, messages);
        this.buffer = "";
        break;
      }

      if (!final && this.buffer[newline] === "\r" && newline === this.buffer.length - 1) break;

      const line = this.buffer.slice(0, newline);
      const separatorLength = this.buffer[newline] === "\r" && this.buffer[newline + 1] === "\n" ? 2 : 1;
      this.buffer = this.buffer.slice(newline + separatorLength);
      this.processLine(line, messages);
    }

    if (final) this.dispatch(messages);
    return messages;
  }

  processLine(line, messages) {
    if (line === "") {
      this.dispatch(messages);
      return;
    }
    if (line.startsWith(":")) return;

    const colon = line.indexOf(":");
    const field = colon === -1 ? line : line.slice(0, colon);
    let value = colon === -1 ? "" : line.slice(colon + 1);
    if (value.startsWith(" ")) value = value.slice(1);

    if (field === "data") this.dataLines.push(value);
    if (field === "event") this.event = value;
    if (field === "id" && !value.includes("\0")) this.id = value;
  }

  dispatch(messages) {
    if (this.dataLines.length > 0) {
      messages.push({ data: this.dataLines.join("\n"), event: this.event, id: this.id });
    }
    this.dataLines = [];
    this.event = undefined;
  }
}

export function createQueryViewState(conversationId) {
  return {
    phase: "starting",
    outcome: null,
    answer: "",
    citations: [],
    conversationId,
    backendError: null,
  };
}

export function applyQueryStreamMessage(current, message) {
  const { frame } = message;
  const event = message.event ?? frame.event;
  const citations = [...current.citations];
  const citationIndexes = new Map(citations.map((citation, index) => [`${citation.document_id}\0${citation.chunk_id}`, index]));

  for (const citation of frame.citations ?? []) {
    const key = `${citation.document_id}\0${citation.chunk_id}`;
    const existingIndex = citationIndexes.get(key);
    if (existingIndex === undefined) {
      citationIndexes.set(key, citations.length);
      citations.push(citation);
    } else if (!citations[existingIndex].snippet && citation.snippet) {
      citations[existingIndex] = citation;
    }
  }

  let phase = current.phase;
  let outcome = current.outcome;
  if (frame.state === "insufficient_context") {
    phase = "completed";
    outcome = "insufficient_context";
  } else if (frame.error) {
    phase = "error";
    outcome = "backend_error";
  } else if (frame.state === "unsupported") {
    phase = "error";
    outcome = "backend_error";
  } else if (frame.state === "answered") {
    phase = "completed";
    outcome = "answered";
  } else if (frame.state === "retrieving" || event === "retrieval") {
    phase = "retrieving";
  } else if (event === "start") {
    phase = "starting";
  } else {
    phase = "streaming";
  }

  return {
    phase,
    outcome,
    answer: frame.answer === undefined ? current.answer : frame.answer,
    citations,
    conversationId: frame.conversation_id ?? current.conversationId,
    backendError: frame.error ?? current.backendError,
  };
}
