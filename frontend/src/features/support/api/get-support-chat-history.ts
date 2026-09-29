/**
 * get-support-chat-history.ts — ヘルプチャットの保存済み会話履歴取得
 *
 * GET /v1/support/chat/history
 * 認証中の clinic × staff にスコープされた履歴を古い順で返す。
 * HelpChat マウント時に1度だけローカル履歴へ反映する（再マウント時の
 * 上書き防止は hydrated フラグ側で担保）。
 */
import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";

import type { SupportChatHistoryRecord } from "../types";

interface ListResponse {
  data: SupportChatHistoryRecord[];
}

async function getSupportChatHistory(): Promise<SupportChatHistoryRecord[]> {
  const { data } = await axios.get<ListResponse>("/v1/support/chat/history");
  return data.data;
}

export function useGetSupportChatHistory() {
  return useQuery({
    queryKey: queryKeys.supportChat.history(),
    queryFn: getSupportChatHistory,
    // 履歴はセッション中に自分の送信/リセットでのみ変わるため、再マウントごとの
    // 再取得は不要（リセット時は removeQueries で明示的に破棄する）
    staleTime: Number.POSITIVE_INFINITY,
  });
}
