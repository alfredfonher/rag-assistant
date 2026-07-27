import { ArrowUpRight, Braces, Cable, Files, MessagesSquare } from "lucide-react";
import Link from "next/link";
import { PageHeader } from "@/components/page-header";
import { Card } from "@/components/ui/card";
import { SystemStatus } from "@/components/system-status";

const capabilities = [
  { title: "Grounded queries", text: "Submit standard or streaming queries against configured knowledge sources.", href: "/ask", icon: MessagesSquare },
  { title: "Knowledge organization", text: "Manage path-ingested documents and their collection boundaries.", href: "/documents", icon: Files },
  { title: "Explicit contracts", text: "Connect to versioned REST resources without invented client-side data.", href: "/system", icon: Braces },
];

export default function OverviewPage() {
  return (
    <div className="space-y-10">
      <PageHeader
        eyebrow="Retrieval operations"
        title="A clear surface for grounded answers."
        description="This workspace exposes the backend's documented RAG capabilities without manufacturing activity, metrics, or resource records."
        action={<span className="inline-flex items-center gap-2 rounded-full border border-primary/25 bg-primary/10 px-3 py-1.5 text-xs text-primary"><Cable className="h-3.5 w-3.5" /> API-aware shell</span>}
      />
      <section aria-labelledby="capabilities-title">
        <h2 id="capabilities-title" className="mb-4 text-sm font-semibold">Available workspaces</h2>
        <div className="grid gap-4 lg:grid-cols-3">
          {capabilities.map(({ title, text, href, icon: Icon }) => (
            <Link key={href} href={href} className="group rounded-2xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary">
              <Card className="h-full p-5 transition-colors group-hover:border-primary/35">
                <div className="flex items-start justify-between">
                  <span className="grid h-10 w-10 place-items-center rounded-xl bg-muted text-primary"><Icon className="h-4 w-4" /></span>
                  <ArrowUpRight className="h-4 w-4 text-muted-foreground transition-transform group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
                </div>
                <h3 className="mt-7 font-semibold">{title}</h3>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">{text}</p>
              </Card>
            </Link>
          ))}
        </div>
      </section>
      <section className="grid gap-5 lg:grid-cols-[1.15fr_0.85fr]">
        <SystemStatus />
        <Card className="p-5 sm:p-6">
          <h2 className="text-base font-semibold">Current scope</h2>
          <ul className="mt-4 space-y-3 text-sm leading-6 text-muted-foreground">
            <li>Responsive navigation and route structure</li>
            <li>Typed API boundaries for documented endpoints</li>
            <li>Browser-side health and readiness checks</li>
            <li>No persisted UI state or speculative schemas</li>
          </ul>
        </Card>
      </section>
    </div>
  );
}
