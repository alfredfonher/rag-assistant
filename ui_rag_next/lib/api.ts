import type {
  IngestDocumentRequest,
  QueryRequest,
  QueryResponse,
  ResourceName,
  ResourceRecord,
  ServiceStatus,
} from "@/lib/contracts";

export const API_BASE_URL = "/backend";

export const endpoints = {
  health: "/healthz",
  readiness: "/readyz",
  query: "/v1/query",
  queryStream: "/v1/query/stream",
  ingest: "/v1/documents/ingest",
  resource: (name: ResourceName) => `/v1/${name}`,
} as const;

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`API request failed with status ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
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
  ingest: (payload: IngestDocumentRequest) =>
    request<ResourceRecord>(endpoints.ingest, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
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

export async function queryStream(payload: QueryRequest): Promise<ReadableStream<Uint8Array>> {
  const response = await fetch(`${API_BASE_URL}${endpoints.queryStream}`, {
    method: "POST",
    headers: {
      Accept: "text/event-stream",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok || !response.body) {
    throw new Error(`Streaming query failed with status ${response.status}`);
  }

  return response.body;
}
