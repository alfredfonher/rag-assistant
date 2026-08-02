"use client";

import { FormEvent, useEffect, useRef, useState } from "react";
import { AlertCircle, Database, FileText, Loader2, RefreshCw, Send, Square } from "lucide-react";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { ingestDocument, listDocuments } from "@/lib/api";
import type { Document, IngestDocumentResponse } from "@/lib/contracts";
import {
  createIngestResult,
  describeDocumentRequestError,
  validateDocumentPath,
} from "@/lib/documents.mjs";

const statusStyles: Record<Document["status"], string> = {
  pending: "border-amber-400/30 bg-amber-400/10 text-amber-200",
  indexing: "border-blue-400/30 bg-blue-400/10 text-blue-200",
  ready: "border-primary/30 bg-primary/10 text-primary",
  error: "border-red-400/30 bg-red-400/10 text-red-200",
  outdated: "border-orange-400/30 bg-orange-400/10 text-orange-200",
};

function formatTimestamp(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

export default function DocumentsPage() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [listLoading, setListLoading] = useState(true);
  const [listError, setListError] = useState("");
  const [path, setPath] = useState("");
  const [pathError, setPathError] = useState("");
  const [ingestError, setIngestError] = useState("");
  const [ingestResponse, setIngestResponse] = useState<IngestDocumentResponse | null>(null);
  const [ingesting, setIngesting] = useState(false);
  const listController = useRef<AbortController>();
  const ingestController = useRef<AbortController>();
  const mounted = useRef(false);

  useEffect(() => {
    mounted.current = true;
    const controller = new AbortController();
    listController.current = controller;

    listDocuments({ signal: controller.signal })
      .then((records) => {
        if (!mounted.current) return;
        setDocuments(records);
        setListError("");
      })
      .catch((error: unknown) => {
        if (!mounted.current || controller.signal.aborted) return;
        setListError(describeDocumentRequestError(error));
      })
      .finally(() => {
        if (mounted.current && listController.current === controller) setListLoading(false);
      });

    return () => {
      mounted.current = false;
      controller.abort();
      listController.current?.abort();
      ingestController.current?.abort();
    };
  }, []);

  async function refreshDocuments() {
    listController.current?.abort();
    const controller = new AbortController();
    listController.current = controller;
    setListLoading(true);
    setListError("");

    try {
      const records = await listDocuments({ signal: controller.signal });
      if (mounted.current) setDocuments(records);
    } catch (error) {
      if (mounted.current && !controller.signal.aborted) setListError(describeDocumentRequestError(error));
    } finally {
      if (mounted.current && listController.current === controller) setListLoading(false);
    }
  }

  async function submitIngest(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (ingestController.current) return;

    const validationError = validateDocumentPath(path);
    if (validationError) {
      setPathError(validationError);
      return;
    }

    const controller = new AbortController();
    ingestController.current = controller;
    setIngesting(true);
    setPathError("");
    setIngestError("");
    setIngestResponse(null);

    try {
      const response = await ingestDocument({ path: path.trim() }, { signal: controller.signal });
      if (!mounted.current) return;
      setIngestResponse(response);
      if (response.state === "indexed") await refreshDocuments();
    } catch (error) {
      if (!mounted.current) return;
      setIngestError(controller.signal.aborted ? "Ingestion was cancelled." : describeDocumentRequestError(error));
    } finally {
      if (ingestController.current === controller) ingestController.current = undefined;
      if (mounted.current) setIngesting(false);
    }
  }

  const ingestResult = ingestResponse ? createIngestResult(ingestResponse) : null;

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Knowledge registry"
        title="Documents"
        description="Inspect document registry records and index text files from the configured ingest root."
        action={(
          <Button type="button" variant="outline" onClick={refreshDocuments} disabled={listLoading}>
            <RefreshCw className={`mr-2 h-4 w-4 ${listLoading ? "animate-spin" : ""}`} aria-hidden="true" />
            Refresh
          </Button>
        )}
      />

      <div className="grid items-start gap-5 xl:grid-cols-[minmax(0,1fr)_380px]">
        <section aria-labelledby="registry-heading" className="min-w-0 space-y-4">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h2 id="registry-heading" className="font-semibold">Document registry</h2>
              <p className="mt-1 text-sm text-muted-foreground">GET /backend/v1/documents</p>
            </div>
            {!listLoading && !listError && <span className="text-sm text-muted-foreground">{documents.length} {documents.length === 1 ? "record" : "records"}</span>}
          </div>

          {listLoading && documents.length === 0 && (
            <Card className="grid min-h-72 place-items-center p-8 text-center" aria-live="polite">
              <div><Loader2 className="mx-auto h-6 w-6 animate-spin text-primary" aria-hidden="true" /><p className="mt-3 text-sm text-muted-foreground">Loading document registry...</p></div>
            </Card>
          )}

          {listError && (
            <Card className="border-red-400/30 p-6" role="alert">
              <div className="flex gap-3 text-red-100"><AlertCircle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" /><div><h3 className="font-semibold">Could not load documents</h3><p className="mt-1 text-sm leading-6">{listError}</p></div></div>
              <Button type="button" variant="outline" className="mt-5" onClick={refreshDocuments}>Try again</Button>
            </Card>
          )}

          {!listLoading && !listError && documents.length === 0 && (
            <Card className="grid min-h-72 place-items-center p-8 text-center">
              <div className="max-w-md"><Database className="mx-auto h-7 w-7 text-primary" aria-hidden="true" /><h3 className="mt-4 font-semibold">No registry records</h3><p className="mt-2 text-sm leading-6 text-muted-foreground">The DocumentRepo returned an empty list. Indexing a path below does not automatically create a CRUD registry record.</p></div>
            </Card>
          )}

          {!listError && documents.length > 0 && (
            <ul className="grid gap-4 2xl:grid-cols-2">
              {documents.map((document) => (
                <li key={document.id}>
                  <Card className="h-full overflow-hidden p-5 sm:p-6">
                    <div className="flex min-w-0 items-start justify-between gap-4">
                      <div className="min-w-0"><p className="break-words font-semibold">{document.filename}</p><p className="mt-1 break-all font-mono text-[11px] text-muted-foreground">{document.id}</p></div>
                      <span className={`shrink-0 rounded-full border px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wide ${statusStyles[document.status]}`}>{document.status}</span>
                    </div>
                    <dl className="mt-5 grid gap-x-5 gap-y-4 text-sm sm:grid-cols-2">
                      <div className="sm:col-span-2"><dt className="text-xs text-muted-foreground">Path</dt><dd className="mt-1 break-all font-mono text-xs">{document.path}</dd></div>
                      <div><dt className="text-xs text-muted-foreground">Collection ID</dt><dd className="mt-1 break-all font-mono text-xs">{document.collection_id}</dd></div>
                      <div><dt className="text-xs text-muted-foreground">Chunks</dt><dd className="mt-1">{document.chunks_count}</dd></div>
                      <div><dt className="text-xs text-muted-foreground">Created</dt><dd className="mt-1 text-xs leading-5">{formatTimestamp(document.created_at)}</dd></div>
                      <div><dt className="text-xs text-muted-foreground">Updated</dt><dd className="mt-1 text-xs leading-5">{formatTimestamp(document.updated_at)}</dd></div>
                      <div className="sm:col-span-2"><dt className="text-xs text-muted-foreground">Error message</dt><dd className={`mt-1 break-words text-xs leading-5 ${document.error_message ? "text-red-200" : "text-muted-foreground"}`}>{document.error_message || "None"}</dd></div>
                    </dl>
                  </Card>
                </li>
              ))}
            </ul>
          )}
        </section>

        <aside className="space-y-4 xl:sticky xl:top-6">
          <Card className="overflow-hidden">
            <div className="border-b border-border p-5 sm:p-6"><div className="flex items-center gap-3"><FileText className="h-5 w-5 text-primary" aria-hidden="true" /><h2 className="font-semibold">Ingest document</h2></div><p className="mt-3 text-sm leading-6 text-muted-foreground">Enter a path relative to the backend&apos;s configured ingest root. Absolute paths and traversal are rejected. This is not a browser file upload.</p></div>
            <form onSubmit={submitIngest} className="p-5 sm:p-6">
              <label htmlFor="document-path" className="text-sm font-semibold">Relative document path</label>
              <input
                id="document-path"
                value={path}
                onChange={(event) => { setPath(event.target.value); if (pathError) setPathError(""); }}
                placeholder="guides/guide.md"
                disabled={ingesting}
                aria-invalid={Boolean(pathError)}
                aria-describedby="document-path-help document-path-error"
                className="mt-2 h-11 w-full rounded-lg border border-border bg-background/70 px-3 font-mono text-sm outline-none placeholder:text-muted-foreground focus:border-primary focus:ring-2 focus:ring-primary/25 disabled:cursor-not-allowed disabled:opacity-60"
              />
              <p id="document-path-help" className="mt-2 text-xs leading-5 text-muted-foreground">Relative to RAG_INGEST_ROOT. Supported extensions: .txt, .md, .markdown</p>
              <p id="document-path-error" className="mt-1 min-h-5 text-xs text-red-300" role={pathError ? "alert" : undefined}>{pathError}</p>
              <div className="mt-3 flex justify-end gap-3">
                {ingesting && <Button type="button" variant="outline" onClick={() => ingestController.current?.abort()}><Square className="mr-2 h-3.5 w-3.5 fill-current" aria-hidden="true" />Cancel</Button>}
                <Button type="submit" disabled={ingesting || Boolean(validateDocumentPath(path))}>{ingesting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" /> : <Send className="mr-2 h-4 w-4" aria-hidden="true" />}Ingest</Button>
              </div>
            </form>
          </Card>

          {ingestError && <Card className="border-red-400/30 p-5 text-sm text-red-100" role="alert"><p className="font-semibold">Ingest request failed</p><p className="mt-1 leading-6">{ingestError}</p></Card>}

          {ingestResponse && ingestResult && (
            <Card className={`p-5 ${ingestResult.tone === "success" ? "border-primary/30" : ingestResult.tone === "warning" ? "border-amber-400/30" : "border-red-400/30"}`} aria-live="polite">
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">{ingestResponse.state}</p>
              <h3 className="mt-2 font-semibold">{ingestResult.title}</h3>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">{ingestResult.message}</p>
              <dl className="mt-4 grid grid-cols-2 gap-4 border-t border-border pt-4 text-xs">
                <div><dt className="text-muted-foreground">Document ID</dt><dd className="mt-1 break-all font-mono">{ingestResponse.document_id ?? "Not returned"}</dd></div>
                <div><dt className="text-muted-foreground">Citations</dt><dd className="mt-1">{ingestResult.citationCount}</dd></div>
              </dl>
            </Card>
          )}

          <p className="px-1 text-xs leading-5 text-muted-foreground">Ingestion indexing and the CRUD document registry are separate backend flows. The ingest service does not persist through DocumentRepo, so this page never auto-creates a registry record.</p>
        </aside>
      </div>
    </div>
  );
}
