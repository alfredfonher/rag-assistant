import type { Conversation, ConversationSummary, ConversationTurn } from "./contracts";

export type ConversationRow = {
  conversation: Omit<Conversation, "turns"> & {
    turns: Array<ConversationTurn & { hiddenCitationCount?: number }>;
  };
  id: string;
  turnsCount: number;
  retainedTurnOffset: number;
  preview: string;
  updatedAt: string | null;
  hydrationState: "loading" | "ready" | "error";
};

export type ConversationHydrationResult =
  | { id: string; status: "fulfilled"; conversation: Conversation }
  | { id: string; status: "rejected"; error: unknown };

export function visibleCitationLimit(): number;
export function retainedTurnLimit(): number;
export function detailHydrationConcurrency(): number;
export function normalizeConversationId(value: unknown): string | null;
export function parseConversationSummaries(value: unknown): ConversationSummary[] | null;
export function parseConversation(value: unknown): Conversation | null;
export function deriveConversationRows(conversations: Conversation[]): ConversationRow[];
export function deriveConversationSummaryRows(summaries: ConversationSummary[]): ConversationRow[];
export function mergeConversationRow(rows: ConversationRow[], conversation: Conversation): ConversationRow[];
export function markConversationRowUnavailable(rows: ConversationRow[], id: string): ConversationRow[];
export function boundedTurn(turn: ConversationTurn): Pick<ConversationTurn, "query" | "answer"> & { citations: NonNullable<ConversationTurn["citations"]> };
export function hydrateConversationDetails(
  summaries: ConversationSummary[],
  load: (id: string, options: { signal: AbortSignal }) => Promise<Conversation>,
  options?: {
    signal?: AbortSignal;
    concurrency?: number;
    timeoutMs?: number;
    onResult?: (result: ConversationHydrationResult) => void;
  },
): Promise<ConversationHydrationResult[]>;
export function conversationPath(id: string): string;
export function askConversationHref(id: string): string;
export function conversationIdFromSearch(search: string): string | null;
export function describeConversationError(error: unknown): string;
