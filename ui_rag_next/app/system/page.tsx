import { PageHeader } from "@/components/page-header";
import { SystemStatus } from "@/components/system-status";
import { Card } from "@/components/ui/card";
import { API_BASE_URL, endpoints } from "@/lib/api";

export default function SystemPage() {
  return (
    <div className="space-y-8">
      <PageHeader eyebrow="Runtime visibility" title="System" description="Inspect browser-visible backend availability and the API contract configured for this frontend." />
      <div className="grid gap-5 lg:grid-cols-2">
        <SystemStatus />
        <Card className="p-5 sm:p-6">
          <h2 className="text-base font-semibold">Configured contract</h2>
          <p className="mt-1 break-all font-mono text-xs text-muted-foreground">{API_BASE_URL}</p>
          <ul className="mt-5 grid gap-2 font-mono text-xs text-muted-foreground">
            <li>GET {endpoints.health}</li>
            <li>GET {endpoints.readiness}</li>
            <li>POST {endpoints.query}</li>
            <li>POST {endpoints.queryStream}</li>
            <li>POST {endpoints.ingest}</li>
            <li>CRUD /v1/agents</li>
            <li>CRUD /v1/collections</li>
            <li>CRUD /v1/documents</li>
            <li>CRUD /v1/conversations</li>
          </ul>
        </Card>
      </div>
    </div>
  );
}
