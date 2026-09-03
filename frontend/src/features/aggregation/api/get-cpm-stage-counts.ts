import { useQueries } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";
import { AGGREGATION_CPM_STAGES, type AggregationCPMStage } from "@/lib/cpm-stage";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";
import type { AggregationParams, AggregationResponse } from "./get-aggregations";

export type CPMStageCounts = Record<AggregationCPMStage, number>;

export interface CPMStageCountsResult {
  counts: CPMStageCounts;
  total: number;
  isLoading: boolean;
  isError: boolean;
}

// 人数集計の母集団フィルタだけを残し、人数（total）に影響しないパラメータを除外する。
// page/per_page/sort/order は total を変えないため落とす。
// cpm_stage 自体は各リクエストで個別指定するため除外する。
export function toCPMCountBaseParams(
  params: AggregationParams,
): Omit<AggregationParams, "page" | "per_page" | "sort" | "order" | "cpm_stage"> {
  const {
    page: _page,
    per_page: _perPage,
    sort: _sort,
    order: _order,
    cpm_stage: _cpmStage,
    ...rest
  } = params;
  return rest;
}

// ISSUE-180: CPM セグメント別の人数を取得する。
//
// レスポンスはページネーションされるため、取得済み1ページの件数を数えると母集団全体に
// ならない。BE はフィルタ後の全件数を `total` に返すので、各セグメントに `per_page=1` を
// 付けて `total` を人数として読む（転送は1件分のみ）。6 セグメントを並列取得する。
// BE 変更は不要。6 リクエストを1回に束ねる専用集計エンドポイントは将来の効率化に留める。
export function useGetCPMStageCounts(params: AggregationParams): CPMStageCountsResult {
  const baseParams = toCPMCountBaseParams(params);

  const queries = useQueries({
    queries: AGGREGATION_CPM_STAGES.map((stage) => ({
      queryKey: queryKeys.ownerAggregations.cpmStageCounts(stage, baseParams),
      queryFn: async (): Promise<number> => {
        const clinicId = requireStoredClinicId();
        // BUG-012: bound hang so chips leave "loading" with error instead of forever-pending.
        const { data } = await axios.get<AggregationResponse>(
          `/v1/clinics/${clinicId}/owners/aggregations`,
          {
            params: { ...baseParams, cpm_stage: stage, page: 1, per_page: 1 },
            timeout: 25_000,
          },
        );
        return data.total;
      },
      staleTime: QUERY_STALE_TIMES.MEDIUM,
    })),
  });

  const counts = AGGREGATION_CPM_STAGES.reduce((acc, stage, index) => {
    acc[stage] = queries[index].data ?? 0;
    return acc;
  }, {} as CPMStageCounts);

  const total = AGGREGATION_CPM_STAGES.reduce((sum, stage) => sum + counts[stage], 0);
  const isLoading = queries.some((query) => query.isLoading);
  const isError = queries.some((query) => query.isError);

  return { counts, total, isLoading, isError };
}
