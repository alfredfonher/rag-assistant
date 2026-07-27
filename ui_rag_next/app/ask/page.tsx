import { Radio, Send } from "lucide-react";
import { PageHeader } from "@/components/page-header";
import { Card } from "@/components/ui/card";

export default function AskPage() {
  return (
    <div className="space-y-8">
      <PageHeader eyebrow="Query workspace" title="Ask your knowledge base" description="The backend supports standard responses and SSE streaming over fetch. Submission remains disabled until the query response and citation schemas are confirmed end to end." />
      <div className="grid gap-5 lg:grid-cols-[1fr_300px]">
        <Card className="min-h-[480px] p-5 sm:p-7">
          <div className="flex h-full min-h-[420px] flex-col justify-between">
            <div className="grid flex-1 place-items-center text-center">
              <div className="max-w-sm">
                <Radio className="mx-auto h-7 w-7 text-primary" />
                <h2 className="mt-4 font-semibold">No active conversation</h2>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">Answers and citations will appear here after query execution is connected to validated backend payloads.</p>
              </div>
            </div>
            <div className="flex gap-3 rounded-xl border border-border bg-background/60 p-3">
              <label htmlFor="query" className="sr-only">Query</label>
              <textarea id="query" disabled placeholder="Query submission is not connected yet" className="min-h-12 flex-1 resize-none bg-transparent p-2 text-sm text-muted-foreground outline-none disabled:cursor-not-allowed" />
              <button disabled aria-label="Submit query" className="grid h-11 w-11 place-items-center self-end rounded-lg bg-primary text-primary-foreground disabled:opacity-35"><Send className="h-4 w-4" /></button>
            </div>
          </div>
        </Card>
        <Card className="h-fit p-5">
          <h2 className="text-sm font-semibold">Transport contract</h2>
          <dl className="mt-4 space-y-4 text-xs">
            <div><dt className="text-muted-foreground">Standard</dt><dd className="mt-1 font-mono">POST /v1/query</dd></div>
            <div><dt className="text-muted-foreground">Streaming</dt><dd className="mt-1 font-mono">POST /v1/query/stream</dd></div>
            <div><dt className="text-muted-foreground">Protocol</dt><dd className="mt-1">SSE via fetch ReadableStream</dd></div>
          </dl>
        </Card>
      </div>
    </div>
  );
}
