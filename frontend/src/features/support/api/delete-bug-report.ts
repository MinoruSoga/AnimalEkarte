/**
 * delete-bug-report.ts — バグ報告削除 API（全スタッフ公開）
 *
 * DELETE /v1/support/bug-reports/:id（認証済みスタッフ全員 — 他医院の報告も対象）
 * 論理削除 + 添付スクショのオブジェクト削除（best-effort）。
 * Plane 側の既起票チケットは残る。
 */
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

async function deleteBugReport(id: number): Promise<void> {
  await axios.delete(`/v1/support/bug-reports/${id}`);
}

export function useDeleteBugReport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteBugReport,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.supportBugReports.all() });
      toast.success("バグ報告を削除しました");
    },
    onError: (error) => {
      handleApiError(error, "バグ報告の削除");
    },
  });
}
