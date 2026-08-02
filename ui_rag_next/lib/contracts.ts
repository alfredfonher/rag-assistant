export interface ServiceStatus {
  ok: boolean;
  status: number;
}

export interface QueryRequest {
  query: string;
  conversation_id?: string;
}

export type QueryState =
  | "streaming"
  | "retrieving"
  | "answered"
  | "insufficient_context"
  | "unsupported";

export type QueryStreamEventName =
  | "start"
  | "retrieval"
  | "content"
  | "citation"
  | "done";

export interface Citation {
  document_id: string;
  chunk_id: string;
  snippet?: string;
}

export interface APIError {
  code: string;
  message: string;
}

export interface QueryResponse {
  state: QueryState;
  event?: QueryStreamEventName;
  kind?: string;
  answer?: string;
  citations?: Citation[];
  conversation_id?: string;
  error?: APIError;
}

export interface QueryStreamMessage {
  event?: QueryStreamEventName;
  id?: string;
  frame: QueryResponse;
}

export interface IngestDocumentRequest {
  path: string;
}

export type DocumentStatus = "pending" | "indexing" | "ready" | "error" | "outdated";

export interface Document {
  id: string;
  collection_id: string;
  filename: string;
  path: string;
  status: DocumentStatus;
  chunks_count: number;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export type IngestDocumentState = "indexed" | "unindexed" | "unsupported";

export interface IngestDocumentResponse {
  state: IngestDocumentState;
  document_id?: string;
  citations?: Citation[];
  error?: APIError;
}

export interface ResourceRecord {
  id: string;
  [key: string]: unknown;
}

export type Agent = ResourceRecord;
export type Collection = ResourceRecord;
export type Conversation = ResourceRecord;

export type ResourceName =
  | "agents"
  | "collections"
  | "documents"
  | "conversations";
