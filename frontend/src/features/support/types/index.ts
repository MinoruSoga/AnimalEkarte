/**
 * support feature の型定義。
 * バックエンド `internal/support` のレスポンスと対応する（snake_case のまま受け取る）。
 */

export type BugReportStatus = "open" | "resolved";

export interface BugReport {
  id: number;
  title: string;
  detail: string;
  page_url: string;
  route_path: string;
  user_agent: string;
  viewport: string;
  app_version: string;
  status: BugReportStatus;
  reporter_staff_id: number;
  reporter_name: string;
  screenshot_url?: string;
  created_at: string;
  updated_at: string;
}

// ── ヘルプチャット（POST /v1/support/chat） ──────────────────────────

export type SupportChatRole = "user" | "assistant";

export interface SupportChatHistoryMessage {
  role: SupportChatRole;
  content: string;
}

/** 検索で取得したマニュアル抜粋を LLM の根拠コンテキストとして送る */
export interface SupportChatContextItem {
  title: string;
  category: string;
  slug: string;
  text: string;
}

export interface SupportChatSource {
  title: string;
  category: string;
  slug: string;
}

export interface SupportChatResponse {
  reply: string;
  sources: SupportChatSource[];
}

export interface SupportChatStatus {
  enabled: boolean;
}
