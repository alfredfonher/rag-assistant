"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { Menu, Search, Sparkles, X } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import { navigation } from "@/lib/navigation";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

function Brand() {
  return (
    <Link
      href="/"
      className="flex items-center gap-3 rounded-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
    >
      <span className="grid h-9 w-9 place-items-center rounded-xl border border-primary/30 bg-primary/10 text-primary">
        <Sparkles aria-hidden="true" className="h-4 w-4" />
      </span>
      <span>
        <span className="block text-sm font-semibold tracking-wide text-foreground">RAG Assistant</span>
        <span className="block text-[10px] uppercase tracking-[0.22em] text-muted-foreground">Knowledge workspace</span>
      </span>
    </Link>
  );
}

function Navigation({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname();

  return (
    <nav aria-label="Primary navigation" className="space-y-1">
      {navigation.map(({ href, label, icon: Icon }) => {
        const active = pathname === href;
        return (
          <Link
            key={href}
            href={href}
            onClick={onNavigate}
            aria-current={active ? "page" : undefined}
            className={cn(
              "group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary",
              active
                ? "bg-primary/12 text-foreground"
                : "text-muted-foreground hover:bg-muted hover:text-foreground",
            )}
          >
            <Icon aria-hidden="true" className={cn("h-4 w-4", active && "text-primary")} />
            {label}
            {active && <span aria-hidden="true" className="ml-auto h-1.5 w-1.5 rounded-full bg-primary" />}
          </Link>
        );
      })}
    </nav>
  );
}

export function AppShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const pageName = navigation.find((item) => item.href === pathname)?.label ?? "Workspace";

  return (
    <div className="min-h-screen md:grid md:grid-cols-[248px_1fr]">
      <aside className="hidden border-r border-border bg-card/55 px-4 py-6 backdrop-blur-xl md:flex md:flex-col">
        <Brand />
        <div className="mt-9 flex-1"><Navigation /></div>
        <p className="rounded-xl border border-border bg-background/40 p-3 text-xs leading-relaxed text-muted-foreground">
          Interface shell only. Resource actions remain intentionally unwired until response schemas are confirmed.
        </p>
      </aside>

      <div className="min-w-0">
        <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-border bg-background/80 px-4 backdrop-blur-xl sm:px-6">
          <div className="flex items-center gap-3">
            <Dialog.Root>
              <Dialog.Trigger asChild>
                <Button variant="ghost" size="icon" className="md:hidden" aria-label="Open navigation">
                  <Menu aria-hidden="true" className="h-5 w-5" />
                </Button>
              </Dialog.Trigger>
              <Dialog.Portal>
                <Dialog.Overlay className="fixed inset-0 z-40 bg-black/70 data-[state=open]:animate-in" />
                <Dialog.Content className="fixed inset-y-0 left-0 z-50 w-[min(86vw,320px)] border-r border-border bg-background p-5 shadow-2xl focus:outline-none">
                  <Dialog.Title className="sr-only">Navigation</Dialog.Title>
                  <div className="flex items-center justify-between">
                    <Brand />
                    <Dialog.Close asChild>
                      <Button variant="ghost" size="icon" aria-label="Close navigation">
                        <X aria-hidden="true" className="h-5 w-5" />
                      </Button>
                    </Dialog.Close>
                  </div>
                  <div className="mt-8"><Navigation /></div>
                </Dialog.Content>
              </Dialog.Portal>
            </Dialog.Root>
            <span className="text-sm font-medium text-foreground">{pageName}</span>
          </div>
          <Link
            href="/ask"
            className="flex items-center gap-2 rounded-lg border border-border bg-card px-3 py-2 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
          >
            <Search aria-hidden="true" className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">Open query workspace</span>
            <span className="sm:hidden">Ask</span>
          </Link>
        </header>
        <main className="mx-auto w-full max-w-[1400px] px-4 py-8 sm:px-6 lg:px-10 lg:py-10">{children}</main>
      </div>
    </div>
  );
}
