import { ResourcePage } from "@/components/resource-page";

export default function ConversationsPage() {
  return <ResourcePage resource="Conversations" description="Return to backend-persisted query threads and their grounded context." capability="Conversation CRUD is represented without fabricated history, participants, timestamps, or message counts." />;
}
