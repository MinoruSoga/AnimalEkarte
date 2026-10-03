/**
 * send-support-chat.ts — ヘルプチャット送信
 *
 * POST /v1/support/chat
 * マニュアル抜粋（context）を根拠に LLM が回答し、参照記事が sources として返る。
 */
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";

import type { SupportChatContextItem, SupportChatResponse } from "../types";

// history は送信しない — LLM コンテキストはサーバー保存済み履歴から構築される。
export interface SendSupportChatParams {
  message: string;
  context: SupportChatContextItem[];
}

async function sendSupportChat(params: SendSupportChatParams): Promise<SupportChatResponse> {
  const { data } = await axios.post<SupportChatResponse>("/v1/support/chat", params);
  return data;
}

export function useSendSupportChat() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: sendSupportChat,
    // 送信分はサーバーに永続化済み（best-effort）。staleTime=Infinity の履歴キャッシュは
    // 送信を反映しないため invalidate しておき、再マウント時の hydrate が
    // 今回の送信分を取りこぼさないようにする。
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.supportChat.history() });
    },
  });
}
