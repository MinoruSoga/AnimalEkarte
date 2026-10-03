/**
 * update-bug-report-status.ts — バグ報告ステータス更新 API（報告元医院スコープ）
 *
 * PATCH /v1/support/bug-reports/:id/status
 * 閲覧は全医院共有だが変更は報告元医院のスタッフのみ — 他医院の報告への
 * PATCH はサーバー側で 404 を返す。UI 側も canMutate ゲートでボタンを非表示にする。
 */
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

import type { BugReport, BugReportStatus } from "../types";

interface UpdateBugReportStatusParams {
  id: number;
  status: BugReportStatus;
}

async function updateBugReportStatus(params: UpdateBugReportStatusParams): Promise<BugReport> {
  const { data } = await axios.patch<BugReport>(`/v1/support/bug-reports/${params.id}/status`, {
    status: params.status,
  });
  return data;
}

export function useUpdateBugReportStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateBugReportStatus,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.supportBugReports.all() });
      toast.success("ステータスを更新しました");
    },
    onError: (error) => {
      handleApiError(error, "ステータス更新");
    },
  });
}
