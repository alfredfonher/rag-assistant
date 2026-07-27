import type { APIError, Citation, QueryStreamMessage } from "./contracts";

export interface SseMessage {
  data: string;
  event?: string;
  id?: string;
}

export class SseParser {
  push(chunk: string): SseMessage[];
  finish(chunk?: string): SseMessage[];
}

export type QueryPhase = "starting" | "retrieving" | "streaming" | "completed" | "error";
export type QueryOutcome = "answered" | "insufficient_context" | "backend_error" | null;

export interface QueryViewState {
  phase: QueryPhase;
  outcome: QueryOutcome;
  answer: string;
  citations: Citation[];
  conversationId?: string;
  backendError: APIError | null;
}

export function createQueryViewState(conversationId?: string): QueryViewState;
export function applyQueryStreamMessage(current: QueryViewState, message: QueryStreamMessage): QueryViewState;
