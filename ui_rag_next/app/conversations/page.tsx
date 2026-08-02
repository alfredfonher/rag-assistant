"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import * as Dialog from "@radix-ui/react-dialog";
import { AlertCircle, ArrowRight, BookOpen, Loader2, MessageSquare, RefreshCw, Trash2, X } from "lucide-react";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { APIHttpError, deleteConversation, getConversation, listConversationSummaries } from "@/lib/api";
import {
  askConversationHref,
  boundedTurn,
  deriveConversationSummaryRows,
  describeConversationError,
  mergeConversationRow,
  visibleCitationLimit,
  type ConversationRow,
} from "@/lib/conversations.mjs";

function formatTimestamp(value: string | null) {
  if (!value) return "No dated turns";
  return new Date(value).toLocaleString();
}

export default function ConversationsPage() {
  const [rows, setRows] = useState<ConversationRow[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState("");
  const controller = useRef<AbortController>();
  const mounted = useRef(false);

  const selected = rows.find((row) => row.id === selectedId) ?? null;

  async function loadConversations(preferredId?: string | null) {
    controller.current?.abort();
    const nextController = new AbortController();
    controller.current = nextController;
    setLoading(true);
    setError("");

    try {
      const summaries = await listConversationSummaries({ signal: nextController.signal });
      const nextRows = deriveConversationSummaryRows(summaries);
      let missingCount = 0;
      if (!mounted.current || controller.current !== nextController) return;
      setRows(nextRows);
      setSelectedId((current) => {
        const requested = preferredId === undefined ? current : preferredId;
        return requested && nextRows.some((row) => row.id === requested) ? requested : nextRows[0]?.id ?? null;
      });
      setLoading(false);

      for (const summary of summaries) {
        void getConversation(summary.id, { signal: nextController.signal }).then((conversation) => {
          if (mounted.current && controller.current === nextController) {
            setRows((current) => mergeConversationRow(current, conversation));
          }
        }).catch((detailError) => {
          if (!mounted.current || nextController.signal.aborted || controller.current !== nextController) return;
          if (detailError instanceof APIHttpError && detailError.status === 404) {
            missingCount += 1;
            setRows((current) => {
              const remaining = current.filter((row) => row.id !== summary.id);
              setSelectedId((selected) => selected === summary.id ? remaining[0]?.id ?? null : selected);
              return remaining;
            });
            setNotice(`${missingCount} conversation ${missingCount === 1 ? "was" : "were"} removed while the list loaded.`);
          }
        });
      }
    } catch (loadError) {
      if (mounted.current && !nextController.signal.aborted) setError(describeConversationError(loadError));
    } finally {
      if (mounted.current && controller.current === nextController) setLoading(false);
    }
  }

  useEffect(() => {
    mounted.current = true;
    void loadConversations(null);
    return () => {
      mounted.current = false;
      controller.current?.abort();
    };
  }, []);

  async function handleDelete() {
    if (!selected || deleting) return;
    const deletedId = selected.id;
    const deleteController = new AbortController();
    setDeleting(true);
    setError("");
    setNotice("");
    setDeleteError("");
    try {
      await deleteConversation(deletedId, { signal: deleteController.signal });
      if (!mounted.current) return;
      setConfirmOpen(false);
      setNotice("Conversation deleted.");
      await loadConversations(null);
    } catch (deleteError) {
      if (!mounted.current) return;
      if (deleteError instanceof APIHttpError && deleteError.status === 404) {
        setConfirmOpen(false);
        setNotice("The conversation had already been removed. The list was refreshed.");
        await loadConversations(null);
      } else {
        setDeleteError(describeConversationError(deleteError));
      }
    } finally {
      if (mounted.current) setDeleting(false);
    }
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Query history"
        title="Conversations"
        description="Review backend-persisted query turns, inspect their citations, or resume a thread in Ask."
        action={(
          <Button type="button" variant="outline" onClick={() => void loadConversations()} disabled={loading || deleting}>
            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? "animate-spin" : ""}`} aria-hidden="true" /> Refresh
          </Button>
        )}
      />

      <p className="sr-only" aria-live="polite" aria-atomic="true">{notice || (loading ? "Loading conversations." : `${rows.length} conversations loaded.`)}</p>
      {notice && <Card className="border-primary/30 px-5 py-4 text-sm" role="status">{notice}</Card>}
      {error && (
        <Card className="border-red-400/30 p-5" role="alert">
          <div className="flex gap-3 text-red-100"><AlertCircle className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" /><div><h2 className="font-semibold">Conversation request failed</h2><p className="mt-1 text-sm leading-6">{error}</p></div></div>
          {!confirmOpen && <Button type="button" variant="outline" className="mt-4" onClick={() => void loadConversations()}>Try again</Button>}
        </Card>
      )}

      {loading && rows.length === 0 && !error && (
        <Card className="grid min-h-80 place-items-center p-8 text-center" aria-live="polite">
          <div><Loader2 className="mx-auto h-6 w-6 animate-spin text-primary" aria-hidden="true" /><p className="mt-3 text-sm text-muted-foreground">Loading conversations and turn details...</p></div>
        </Card>
      )}

      {!loading && !error && rows.length === 0 && (
        <Card className="grid min-h-80 place-items-center p-8 text-center">
          <div className="max-w-md"><MessageSquare className="mx-auto h-7 w-7 text-primary" aria-hidden="true" /><h2 className="mt-4 font-semibold">No conversations yet</h2><p className="mt-2 text-sm leading-6 text-muted-foreground">Completed Ask requests will appear here after the backend persists their turns.</p><Button asChild className="mt-5"><Link href="/ask">Start in Ask</Link></Button></div>
        </Card>
      )}

      {rows.length > 0 && (
        <div className="grid items-start gap-5 lg:grid-cols-[340px_minmax(0,1fr)]">
          <section aria-labelledby="conversation-list-heading" className="min-w-0">
            <div className="mb-3 flex items-center justify-between"><h2 id="conversation-list-heading" className="font-semibold">Saved threads</h2><span className="text-sm text-muted-foreground">{rows.length}</span></div>
            <ul className="space-y-2">
              {rows.map((row) => (
                <li key={row.id}>
                  <button type="button" onClick={() => setSelectedId(row.id)} aria-pressed={selectedId === row.id} className={`w-full rounded-xl border p-4 text-left transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary ${selectedId === row.id ? "border-primary/50 bg-primary/10" : "border-border bg-card hover:bg-muted/60"}`}>
                    <p className="line-clamp-2 text-sm font-semibold leading-5">{row.preview}</p>
                    <div className="mt-3 flex items-center justify-between gap-3 text-xs text-muted-foreground"><span>{row.turnsCount} {row.turnsCount === 1 ? "turn" : "turns"}</span><time dateTime={row.updatedAt ?? undefined}>{formatTimestamp(row.updatedAt)}</time></div>
                    <p className="mt-2 truncate font-mono text-[10px] text-muted-foreground">{row.id}</p>
                  </button>
                </li>
              ))}
            </ul>
          </section>

          {selected && (
            <section aria-labelledby="conversation-detail-heading" className="min-w-0">
              <Card className="overflow-hidden">
                <div className="flex flex-col gap-4 border-b border-border p-5 sm:flex-row sm:items-start sm:justify-between sm:p-6">
                  <div className="min-w-0"><p className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">Conversation detail</p><h2 id="conversation-detail-heading" className="mt-2 break-all font-mono text-sm">{selected.id}</h2><p className="mt-2 text-sm text-muted-foreground">{selected.turnsCount} {selected.turnsCount === 1 ? "turn" : "turns"} · Updated {formatTimestamp(selected.updatedAt)}</p></div>
                  <div className="flex shrink-0 flex-wrap gap-2"><Button asChild><Link href={askConversationHref(selected.id)}>Continue in Ask <ArrowRight className="ml-2 h-4 w-4" aria-hidden="true" /></Link></Button><Button type="button" variant="outline" onClick={() => { setError(""); setDeleteError(""); setConfirmOpen(true); }}><Trash2 className="mr-2 h-4 w-4" aria-hidden="true" /> Delete</Button></div>
                </div>
                <ol className="divide-y divide-border">
                  {selected.conversation.turns.map((turn, index) => {
                    const visible = boundedTurn(turn);
                    const hiddenCitations = (turn.citations?.length ?? 0) - visible.citations.length;
                    return (
                      <li key={`${turn.created_at}:${index}`} className="space-y-5 p-5 sm:p-6">
                        <div className="flex items-center justify-between gap-3"><h3 className="font-semibold">Turn {index + 1}</h3><time dateTime={turn.created_at} className="text-xs text-muted-foreground">{formatTimestamp(turn.created_at)}</time></div>
                        <section><h4 className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">Question</h4><p className="mt-2 whitespace-pre-wrap break-words rounded-xl bg-muted px-4 py-3 text-sm leading-6">{visible.query || "No question text returned."}</p></section>
                        <section><h4 className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">Answer</h4><p className="mt-2 whitespace-pre-wrap break-words text-sm leading-7">{visible.answer || (turn.state === "insufficient_context" ? "Insufficient context to answer." : "No answer text returned.")}</p><p className="mt-2 text-xs capitalize text-muted-foreground">State: {turn.state.replace("_", " ")}</p></section>
                        {visible.citations.length > 0 && <section><h4 className="flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground"><BookOpen className="h-3.5 w-3.5" aria-hidden="true" /> Citations</h4><ol className="mt-3 grid gap-3 xl:grid-cols-2">{visible.citations.map((citation, citationIndex) => <li key={`${citation.document_id}:${citation.chunk_id}:${citationIndex}`} className="rounded-xl border border-border bg-background/40 p-4"><p className="break-all font-mono text-xs text-primary">{citation.document_id}</p><p className="mt-1 break-all font-mono text-[11px] text-muted-foreground">Chunk {citation.chunk_id}</p>{citation.snippet && <p className="mt-3 whitespace-pre-wrap break-words text-sm leading-6 text-muted-foreground">{citation.snippet}</p>}</li>)}</ol>{hiddenCitations > 0 && <p className="mt-3 text-xs text-muted-foreground">Showing {visibleCitationLimit()} citations. {hiddenCitations} additional {hiddenCitations === 1 ? "citation is" : "citations are"} hidden.</p>}</section>}
                      </li>
                    );
                  })}
                </ol>
              </Card>
            </section>
          )}
        </div>
      )}

      <Dialog.Root open={confirmOpen} onOpenChange={(open) => { if (!deleting) setConfirmOpen(open); }}>
        <Dialog.Portal>
          <Dialog.Overlay className="fixed inset-0 z-50 bg-black/70" />
          <Dialog.Content className="fixed left-1/2 top-1/2 z-50 w-[calc(100%-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-2xl border border-border bg-card p-6 shadow-2xl focus:outline-none" onOpenAutoFocus={(event) => { event.preventDefault(); document.getElementById("confirm-delete")?.focus(); }}>
            <div className="flex items-start justify-between gap-4"><div><Dialog.Title className="text-lg font-semibold">Delete conversation?</Dialog.Title><Dialog.Description className="mt-2 text-sm leading-6 text-muted-foreground">This permanently removes the selected backend conversation and all of its persisted turns. This action cannot be undone.</Dialog.Description></div><Dialog.Close asChild><Button type="button" variant="ghost" size="icon" aria-label="Close confirmation"><X className="h-4 w-4" aria-hidden="true" /></Button></Dialog.Close></div>
            <p className="mt-4 break-all rounded-lg bg-muted p-3 font-mono text-xs">{selected?.id}</p>
            {deleteError && <div className="mt-4 flex gap-3 rounded-xl border border-red-400/30 bg-red-400/5 p-4 text-sm text-red-100" role="alert"><AlertCircle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" /><p>{deleteError}</p></div>}
            <div className="mt-6 flex justify-end gap-3"><Dialog.Close asChild><Button type="button" variant="outline" disabled={deleting}>Cancel</Button></Dialog.Close><Button id="confirm-delete" type="button" onClick={() => void handleDelete()} disabled={deleting} className="bg-red-500 text-white hover:bg-red-500/90">{deleting && <Loader2 className="mr-2 h-4 w-4 animate-spin" aria-hidden="true" />}Delete conversation</Button></div>
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
    </div>
  );
}
