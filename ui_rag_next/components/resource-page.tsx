import { Database, Plus } from "lucide-react";
import { PageHeader } from "@/components/page-header";
import { Card } from "@/components/ui/card";

export function ResourcePage({
  resource,
  description,
  capability,
}: {
  resource: string;
  description: string;
  capability: string;
}) {
  return (
    <div className="space-y-8">
      <PageHeader eyebrow="Resource workspace" title={resource} description={description} />
      <Card className="grid min-h-[360px] place-items-center p-8 text-center">
        <div className="max-w-md">
          <span className="mx-auto grid h-12 w-12 place-items-center rounded-2xl border border-border bg-muted text-primary">
            <Database aria-hidden="true" className="h-5 w-5" />
          </span>
          <h2 className="mt-5 text-lg font-semibold">No {resource.toLowerCase()} loaded</h2>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">{capability}</p>
          <div className="mt-5 inline-flex items-center gap-2 rounded-lg border border-dashed border-border px-3 py-2 text-xs text-muted-foreground">
            <Plus aria-hidden="true" className="h-3.5 w-3.5" />
            Creation controls will follow confirmed payload schemas
          </div>
        </div>
      </Card>
    </div>
  );
}
