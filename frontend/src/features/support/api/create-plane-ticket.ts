/**
 * create-plane-ticket.ts — バグ報告の Plane 起票（手動再送）API（管理者向け）
 *
 * POST /v1/support/bug-reports/:id/plane-ticket（hospital-settings:edit 権限）
 * 自動起票に失敗した報告の再送経路。起票済みなら現在の状態を返す冪等操作。
 */
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

import type { BugReport } from "../types";

async function createPlaneTicket(id: number): Promise<BugReport> {
  const { data } = await axios.post<BugReport>(`/v1/support/bug-reports/${id}/plane-ticket`);
  return data;
}

export function useCreatePlaneTicket() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createPlaneTicket,
    onSuccess: (report) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.supportBugReports.all() });
      if (report.plane_issue_url) {
        toast.success("Plane チケットを作成しました");
      }
    },
    onError: (error) => {
      handleApiError(error, "Plane チケット作成");
    },
  });
}
