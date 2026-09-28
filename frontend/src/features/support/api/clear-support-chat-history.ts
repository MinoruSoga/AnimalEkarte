/**
 * clear-support-chat-history.ts — ヘルプチャット会話履歴の削除
 *
 * DELETE /v1/support/chat/history
 * 認証中の clinic × staff の履歴をサーバー側で soft delete する。
 * 「会話をリセット」ボタンから呼ばれ、成功後に履歴キャッシュを破棄する。
 */
import { useMutation, useQueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";

async function clearSupportChatHistory(): Promise<void> {
  await axios.delete("/v1/support/chat/history");
}

export function useClearSupportChatHistory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: clearSupportChatHistory,
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: queryKeys.supportChat.history() });
    },
  });
}
