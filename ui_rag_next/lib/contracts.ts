export interface ServiceStatus {
  ok: boolean;
  status: number;
}

export interface QueryRequest {
  query: string;
  agent_id?: string;
  collection_ids?: string[];
  conversation_id?: string;
}

export interface QueryResponse {
  answer: string;
  citations?: unknown[];
  conversation_id?: string;
}

export interface IngestDocumentRequest {
  path: string;
  collection_id?: string;
}

export interface ResourceRecord {
  id: string;
  [key: string]: unknown;
}

export type Agent = ResourceRecord;
export type Collection = ResourceRecord;
export type Document = ResourceRecord;
export type Conversation = ResourceRecord;

export type ResourceName =
  | "agents"
  | "collections"
  | "documents"
  | "conversations";
