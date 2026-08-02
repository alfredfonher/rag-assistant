import type { Conversation, ConversationSummary, ConversationTurn } from "./contracts";

export type ConversationRow = {
  conversation: Conversation;
  id: string;
  turnsCount: number;
  preview: string;
  updatedAt: string | null;
};

export function visibleCitationLimit(): number;
export function normalizeConversationId(value: unknown): string | null;
export function parseConversationSummaries(value: unknown): ConversationSummary[] | null;
export function parseConversation(value: unknown): Conversation | null;
export function deriveConversationRows(conversations: Conversation[]): ConversationRow[];
export function deriveConversationSummaryRows(summaries: ConversationSummary[]): ConversationRow[];
export function mergeConversationRow(rows: ConversationRow[], conversation: Conversation): ConversationRow[];
export function boundedTurn(turn: ConversationTurn): Pick<ConversationTurn, "query" | "answer"> & { citations: NonNullable<ConversationTurn["citations"]> };
export function conversationPath(id: string): string;
export function askConversationHref(id: string): string;
export function conversationIdFromSearch(search: string): string | null;
export function describeConversationError(error: unknown): string;
