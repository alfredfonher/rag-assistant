import { ResourcePage } from "@/components/resource-page";

export default function DocumentsPage() {
  return <ResourcePage resource="Documents" description="Review source documents made available to retrieval workflows." capability="The API supports document CRUD and JSON path-based ingestion through POST /v1/documents/ingest. No browser upload contract is assumed." />;
}
