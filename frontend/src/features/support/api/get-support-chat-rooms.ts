/**
 * get-support-chat-rooms.ts — チャット履歴の会話ルーム一覧取得
 *
 * GET /v1/support/chat/rooms（認証済みスタッフ全員・全医院共有 — 権限ゲート・
 * 医院絞りなし。バグ報告ボードと同じ製品判断）。同一 clinic×staff の発言
 * ストリームをサーバー側で時間ギャップ区切りの疑似ルームにまとめて返す。
 * ルームは最終発言の新しい順、ルーム内発言は古い順。
 */
import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

import type { SupportChatRoom } from "../types";

interface ListResponse {
  data: SupportChatRoom[];
}

async function getSupportChatRooms(): Promise<SupportChatRoom[]> {
  const { data } = await axios.get<ListResponse>("/v1/support/chat/rooms");
  return data.data;
}

export function useGetSupportChatRooms(enabled: boolean = true) {
  return useQuery({
    queryKey: queryKeys.supportChat.rooms(),
    queryFn: getSupportChatRooms,
    enabled,
    staleTime: QUERY_STALE_TIMES.SHORT,
  });
}
