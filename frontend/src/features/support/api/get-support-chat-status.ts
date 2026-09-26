/**
 * get-support-chat-status.ts — ヘルプチャットの有効/無効取得
 *
 * GET /v1/support/chat/status
 * LLM が未設定（API キーなし）の環境では enabled=false が返り、
 * ヘルプタブは検索型 UI にフォールバックする。
 */
import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";

import type { SupportChatStatus } from "../types";

async function getSupportChatStatus(): Promise<SupportChatStatus> {
  const { data } = await axios.get<SupportChatStatus>("/v1/support/chat/status");
  return data;
}

export function useGetSupportChatStatus() {
  return useQuery({
    queryKey: queryKeys.supportChat.status(),
    queryFn: getSupportChatStatus,
    // 設定はプロセス起動時の env 依存のため、セッション中は変わらない前提でキャッシュ
    staleTime: Number.POSITIVE_INFINITY,
    gcTime: Number.POSITIVE_INFINITY,
  });
}
