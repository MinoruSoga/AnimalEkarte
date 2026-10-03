/**
 * get-support-chat-exchanges.ts — チャット履歴の共有一覧取得
 *
 * GET /v1/support/chat/exchanges（認証済みスタッフ全員・全医院共有 — 権限ゲート・
 * 医院絞りなし。バグ報告ボードと同じ製品判断）。質問+回答ペアを新しい順で返す。
 */
import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

import type { SupportChatExchange } from "../types";

interface ListResponse {
  data: SupportChatExchange[];
}

async function getSupportChatExchanges(): Promise<SupportChatExchange[]> {
  const { data } = await axios.get<ListResponse>("/v1/support/chat/exchanges");
  return data.data;
}

export function useGetSupportChatExchanges(enabled: boolean = true) {
  return useQuery({
    queryKey: queryKeys.supportChat.exchanges(),
    queryFn: getSupportChatExchanges,
    enabled,
    staleTime: QUERY_STALE_TIMES.SHORT,
  });
}
