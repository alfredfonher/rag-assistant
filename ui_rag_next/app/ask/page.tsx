"use client";

import { FormEvent, KeyboardEvent, useEffect, useRef, useState } from "react";
import { AlertCircle, Ban, BookOpen, Loader2, Radio, Send, Square } from "lucide-react";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  BackendUnavailableError,
  MalformedQueryStreamError,
  QueryStreamHttpError,
  queryStream,
} from "@/lib/api";
import {
  applyQueryStreamMessage,
  createQueryViewState,
  type QueryPhase,
  type QueryViewState,
} from "@/lib/query-stream.mjs";

type RequestIssue = {
  kind: "unavailable" | "http" | "malformed" | "cancelled";
  message: string;
};

const phaseLabels: Record<QueryPhase | "idle", string> = {
  idle: "Ready for a question",
  starting: "Starting query",
  retrieving: "Retrieving relevant context",
  streaming: "Receiving answer",
  completed: "Query completed",
  error: "Query stopped",
};

export default function AskPage() {
  const [query, setQuery] = useState("");
  const [submittedQuery, setSubmittedQuery] = useState("");
  const [view, setView] = useState<QueryViewState | null>(null);
  const [issue, setIssue] = useState<RequestIssue | null>(null);
  const [validationError, setValidationError] = useState("");
  const conversationId = useRef<string>();
  const abortController = useRef<AbortController>();
  const mounted = useRef(true);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      abortController.current?.abort();
    };
  }, []);

  const phase = view?.phase ?? "idle";
  const active = phase === "starting" || phase === "retrieving" || phase === "streaming";

  async function submitQuery(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const normalizedQuery = query.trim();
    if (!normalizedQuery) {
      setValidationError("Enter a question before submitting.");
      return;
    }
    if (abortController.current) return;

    const controller = new AbortController();
    abortController.current = controller;
    setValidationError("");
    setIssue(null);
    setSubmittedQuery(normalizedQuery);
    setView(createQueryViewState(conversationId.current));

    try {
      const payload = conversationId.current
        ? { query: normalizedQuery, conversation_id: conversationId.current }
        : { query: normalizedQuery };

      for await (const message of queryStream(payload, { signal: controller.signal })) {
        if (!mounted.current) return;
        setView((current) => {
          const next = applyQueryStreamMessage(current ?? createQueryViewState(conversationId.current), message);
          conversationId.current = next.conversationId;
          return next;
        });
      }
    } catch (error) {
      if (!mounted.current) return;

      let nextIssue: RequestIssue;
      if (controller.signal.aborted) {
        nextIssue = { kind: "cancelled", message: "The query was cancelled." };
      } else if (error instanceof BackendUnavailableError) {
        nextIssue = { kind: "unavailable", message: "The backend is unavailable. Check the service and try again." };
      } else if (error instanceof QueryStreamHttpError) {
        nextIssue = error.status >= 500
          ? { kind: "unavailable", message: `The backend is unavailable (HTTP ${error.status}).` }
          : { kind: "http", message: error.body || `The backend rejected the request (HTTP ${error.status}).` };
      } else if (error instanceof MalformedQueryStreamError) {
        nextIssue = { kind: "malformed", message: error.message };
      } else {
        nextIssue = { kind: "unavailable", message: "The query failed before a valid response was received." };
      }

      setIssue(nextIssue);
      setView((current) => current ? { ...current, phase: "error" } : { ...createQueryViewState(conversationId.current), phase: "error" });
    } finally {
      if (abortController.current === controller) abortController.current = undefined;
    }
  }

  function handleQueryKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === "Enter" && (event.ctrlKey || event.metaKey)) {
      event.preventDefault();
      event.currentTarget.form?.requestSubmit();
    }
  }

  const liveStatus = issue?.message
    ?? view?.backendError?.message
    ?? (view?.outcome === "insufficient_context" ? "The backend could not find enough relevant context to answer." : phaseLabels[phase]);

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Query workspace"
        title="Ask your knowledge base"
        description="Send grounded questions to the RAG service and follow retrieval, answer, and citation events as they arrive."
      />

      <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_320px]">
        <Card className="min-h-[540px] overflow-hidden">
          <div className="flex min-h-[540px] flex-col">
            <div className="flex items-center justify-between border-b border-border px-5 py-4 sm:px-7">
              <div className="flex items-center gap-3">
                <span className={`h-2.5 w-2.5 rounded-full ${active ? "animate-pulse bg-primary" : phase === "error" ? "bg-red-400" : "bg-muted-foreground"}`} />
                <p className="text-sm font-medium" aria-live="polite" aria-atomic="true">{liveStatus}</p>
              </div>
              {conversationId.current && <span className="hidden text-xs text-muted-foreground sm:inline">Follow-up context active</span>}
            </div>

            <div className="flex-1 p-5 sm:p-7">
              {!submittedQuery && !issue && (
                <div className="grid h-full min-h-[310px] place-items-center text-center">
                  <div className="max-w-sm">
                    <Radio className="mx-auto h-7 w-7 text-primary" aria-hidden="true" />
                    <h2 className="mt-4 font-semibold">Start a grounded conversation</h2>
                    <p className="mt-2 text-sm leading-6 text-muted-foreground">Ask a specific question. The answer will remain plain text and citations will identify the retrieved source chunks.</p>
                  </div>
                </div>
              )}

              {submittedQuery && (
                <div className="space-y-6">
                  <section aria-labelledby="submitted-question">
                    <p id="submitted-question" className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">Question</p>
                    <p className="mt-2 rounded-xl bg-muted px-4 py-3 text-sm leading-6">{submittedQuery}</p>
                  </section>

                  <section aria-labelledby="answer-heading">
                    <div className="flex items-center gap-2">
                      <h2 id="answer-heading" className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">Answer</h2>
                      {active && <Loader2 className="h-3.5 w-3.5 animate-spin text-primary" aria-hidden="true" />}
                    </div>
                    {view?.answer ? (
                      <p className="mt-3 whitespace-pre-wrap text-[15px] leading-7">{view.answer}</p>
                    ) : view?.outcome === "insufficient_context" ? (
                      <div className="mt-3 flex gap-3 rounded-xl border border-amber-400/30 bg-amber-400/5 p-4 text-sm text-amber-100">
                        <Ban className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
                        <p>The backend could not find enough relevant context to answer this question.</p>
                      </div>
                    ) : active ? (
                      <p className="mt-3 text-sm text-muted-foreground">Waiting for answer content...</p>
                    ) : null}
                  </section>

                  {(issue || view?.backendError) && (
                    <div role="alert" className={`flex gap-3 rounded-xl border p-4 text-sm ${issue?.kind === "cancelled" ? "border-border bg-muted/50 text-muted-foreground" : "border-red-400/30 bg-red-400/5 text-red-100"}`}>
                      <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
                      <div>
                        <p className="font-semibold">
                          {issue?.kind === "unavailable" ? "Backend unavailable" : issue?.kind === "malformed" ? "Malformed event stream" : issue?.kind === "cancelled" ? "Query cancelled" : issue ? "Request rejected" : "Backend error"}
                        </p>
                        <p className="mt-1 leading-6">{issue?.message ?? view?.backendError?.message}</p>
                        {view?.backendError?.code && <p className="mt-1 font-mono text-xs opacity-75">{view.backendError.code}</p>}
                      </div>
                    </div>
                  )}

                  {view && view.citations.length > 0 && (
                    <section aria-labelledby="citations-heading">
                      <h2 id="citations-heading" className="flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">
                        <BookOpen className="h-3.5 w-3.5" aria-hidden="true" /> Citations
                      </h2>
                      <ol className="mt-3 grid gap-3 sm:grid-cols-2">
                        {view.citations.map((citation) => (
                          <li key={`${citation.document_id}:${citation.chunk_id}`} className="rounded-xl border border-border bg-background/40 p-4">
                            <p className="break-all font-mono text-xs text-primary">{citation.document_id}</p>
                            <p className="mt-1 break-all font-mono text-[11px] text-muted-foreground">Chunk {citation.chunk_id}</p>
                            {citation.snippet && <p className="mt-3 whitespace-pre-wrap text-sm leading-6 text-muted-foreground">{citation.snippet}</p>}
                          </li>
                        ))}
                      </ol>
                    </section>
                  )}
                </div>
              )}
            </div>

            <form onSubmit={submitQuery} className="border-t border-border bg-background/35 p-4 sm:p-5">
              <label htmlFor="query" className="text-sm font-semibold">Question</label>
              <div className="mt-2 rounded-xl border border-border bg-background/70 p-3 focus-within:border-primary focus-within:ring-2 focus-within:ring-primary/25">
                <textarea
                  id="query"
                  value={query}
                  onChange={(event) => {
                    setQuery(event.target.value);
                    if (validationError) setValidationError("");
                  }}
                  onKeyDown={handleQueryKeyDown}
                  aria-describedby="query-help query-error"
                  aria-invalid={Boolean(validationError)}
                  placeholder="Ask about the indexed knowledge base..."
                  rows={3}
                  disabled={active}
                  className="w-full resize-y bg-transparent p-1 text-sm leading-6 outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-60"
                />
                <div className="mt-2 flex items-center justify-between gap-3">
                  <p id="query-help" className="text-xs text-muted-foreground">Ctrl+Enter or Command+Enter to submit</p>
                  {active ? (
                    <Button type="button" variant="outline" onClick={() => abortController.current?.abort()}>
                      <Square className="mr-2 h-3.5 w-3.5 fill-current" aria-hidden="true" /> Cancel
                    </Button>
                  ) : (
                    <Button type="submit" disabled={!query.trim()}>
                      <Send className="mr-2 h-4 w-4" aria-hidden="true" /> Ask
                    </Button>
                  )}
                </div>
              </div>
              <p id="query-error" className="mt-2 min-h-5 text-xs text-red-300" role={validationError ? "alert" : undefined}>{validationError}</p>
            </form>
          </div>
        </Card>

        <Card className="h-fit p-5">
          <h2 className="text-sm font-semibold">Live request</h2>
          <dl className="mt-4 space-y-4 text-xs">
            <div><dt className="text-muted-foreground">Endpoint</dt><dd className="mt-1 font-mono">POST /v1/query/stream</dd></div>
            <div><dt className="text-muted-foreground">Transport</dt><dd className="mt-1">SSE over same-origin fetch</dd></div>
            <div><dt className="text-muted-foreground">Phase</dt><dd className="mt-1 capitalize">{phase}</dd></div>
            <div><dt className="text-muted-foreground">Conversation</dt><dd className="mt-1 break-all font-mono">{conversationId.current ?? "Created by backend"}</dd></div>
          </dl>
          <p className="mt-5 border-t border-border pt-4 text-xs leading-5 text-muted-foreground">Follow-up questions reuse the conversation identifier returned during this page session.</p>
        </Card>
      </div>
    </div>
  );
}
