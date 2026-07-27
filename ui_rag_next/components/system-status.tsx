"use client";

import { useEffect, useState } from "react";
import { AlertTriangle, CheckCircle2, Loader2 } from "lucide-react";
import { API_BASE_URL, api } from "@/lib/api";
import { Card } from "@/components/ui/card";

type CheckState = "checking" | "available" | "unavailable";

function StatusRow({ label, state }: { label: string; state: CheckState }) {
  const Icon = state === "checking" ? Loader2 : state === "available" ? CheckCircle2 : AlertTriangle;
  const text = state === "checking" ? "Checking" : state === "available" ? "Available" : "Unavailable";

  return (
    <div className="flex items-center justify-between gap-4 border-b border-border py-4 last:border-0">
      <span className="text-sm font-medium">{label}</span>
      <span className="flex items-center gap-2 text-xs text-muted-foreground">
        <Icon aria-hidden="true" className={`h-4 w-4 ${state === "checking" ? "animate-spin" : state === "available" ? "text-emerald-400" : "text-amber-400"}`} />
        {text}
      </span>
    </div>
  );
}

export function SystemStatus() {
  const [health, setHealth] = useState<CheckState>("checking");
  const [readiness, setReadiness] = useState<CheckState>("checking");

  useEffect(() => {
    let active = true;
    const check = async () => {
      const [healthResult, readinessResult] = await Promise.allSettled([
        api.health(),
        api.readiness(),
      ]);
      if (!active) return;
      setHealth(healthResult.status === "fulfilled" ? "available" : "unavailable");
      setReadiness(readinessResult.status === "fulfilled" ? "available" : "unavailable");
    };
    void check();
    return () => { active = false; };
  }, []);

  return (
    <Card className="p-5 sm:p-6">
      <div className="mb-2">
        <h2 className="text-base font-semibold">Backend checks</h2>
        <p className="mt-1 break-all font-mono text-xs text-muted-foreground">{API_BASE_URL}</p>
      </div>
      <StatusRow label="Health /healthz" state={health} />
      <StatusRow label="Readiness /readyz" state={readiness} />
      {(health === "unavailable" || readiness === "unavailable") && (
        <p role="status" className="mt-4 rounded-lg border border-amber-400/20 bg-amber-400/5 p-3 text-xs leading-5 text-amber-200">
          The backend could not be reached from this browser. The interface remains available; verify the API URL, service state, and CORS policy.
        </p>
      )}
    </Card>
  );
}
