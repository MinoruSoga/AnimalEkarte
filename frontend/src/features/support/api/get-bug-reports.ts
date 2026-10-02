/**
 * get-bug-reports.ts — バグ報告一覧の取得 API（全スタッフ・全医院に公開）
 *
 * GET /v1/support/bug-reports（認証済みスタッフ全員 — 権限ゲート・医院絞りなし）
 * screenshot_url は署名付き URL（TTL 短め）。一覧行ごとに含まれる。
 */
import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

import type { BugReport } from "../types";

interface ListResponse {
  data: BugReport[];
}

async function getBugReports(): Promise<BugReport[]> {
  const { data } = await axios.get<ListResponse>("/v1/support/bug-reports");
  return data.data;
}

export function useGetBugReports(enabled: boolean = true) {
  return useQuery({
    queryKey: queryKeys.supportBugReports.all(),
    queryFn: getBugReports,
    enabled,
    staleTime: QUERY_STALE_TIMES.SHORT,
  });
}
