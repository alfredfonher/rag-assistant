import {
  Bot,
  Boxes,
  FileText,
  Gauge,
  MessageSquareText,
  MessagesSquare,
  ServerCog,
} from "lucide-react";

export const navigation = [
  { href: "/", label: "Overview", icon: Gauge },
  { href: "/ask", label: "Ask", icon: MessageSquareText },
  { href: "/documents", label: "Documents", icon: FileText },
  { href: "/collections", label: "Collections", icon: Boxes },
  { href: "/agents", label: "Agents", icon: Bot },
  { href: "/conversations", label: "Conversations", icon: MessagesSquare },
  { href: "/system", label: "System", icon: ServerCog },
] as const;
