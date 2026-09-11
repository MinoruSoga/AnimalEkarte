import { SessionPending } from "@/components/shared/auth/SessionPending";

export function RootHydrateFallback() {
  return <SessionPending message="画面を読み込んでいます" />;
}
