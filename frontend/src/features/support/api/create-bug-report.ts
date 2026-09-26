/**
 * create-bug-report.ts — バグ報告の作成 API（multipart/form-data）
 *
 * POST /v1/support/bug-reports
 * 認証済み全スタッフが送信できる（バックエンド側で clinic_id を文脈から付与）。
 */
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

import type { BugReport } from "../types";

export interface CreateBugReportParams {
  title: string;
  detail: string;
  pageUrl: string;
  routePath: string;
  userAgent: string;
  viewport: string;
  appVersion: string;
  screenshot: File | null;
}

async function createBugReport(params: CreateBugReportParams): Promise<BugReport> {
  const formData = new FormData();
  formData.append("title", params.title);
  formData.append("detail", params.detail);
  formData.append("page_url", params.pageUrl);
  formData.append("route_path", params.routePath);
  formData.append("user_agent", params.userAgent);
  formData.append("viewport", params.viewport);
  formData.append("app_version", params.appVersion);
  if (params.screenshot !== null) {
    formData.append("screenshot", params.screenshot, params.screenshot.name);
  }
  const { data } = await axios.post<BugReport>("/v1/support/bug-reports", formData, {
    headers: { "Content-Type": "multipart/form-data" },
  });
  return data;
}

export function useCreateBugReport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createBugReport,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.supportBugReports.all() });
      toast.success("バグ報告を送信しました");
    },
    onError: (error) => {
      handleApiError(error, "バグ報告の送信");
    },
  });
}
