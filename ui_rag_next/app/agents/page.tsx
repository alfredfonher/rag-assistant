import { ResourcePage } from "@/components/resource-page";

export default function AgentsPage() {
  return <ResourcePage resource="Agents" description="Configure retrieval and answer behavior through backend-managed agents." capability="Agent CRUD is available at /v1/agents. This shell does not speculate about model, prompt, or policy fields." />;
}
