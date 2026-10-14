import { useQuery, useQueryClient, type QueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";
import type { AggregationCPMStage } from "@/lib/cpm-stage";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

// Type definitions for aggregation features.
// 仕様 §4.1〜4.3 / §5.2 の sort 値を canonical とする。
// 旧名 (`total_fee` / `annual_visit_count` / `visit_count`) は AGG-BE-005 の互換エイリアスとして残置するが、
// 新規 UI からは canonical のみ使用すること。
export type AggregationSortField =
  | "annual_amount"
  | "period_visit_count"
  | "last_visit_date"
  | "days_since_last_visit"
  | "owner_name"
  // 互換エイリアス (BE が受理する限りは UI 側からも送られうる)
  | "total_fee"
  | "annual_visit_count"
  | "visit_count";

export type AmountBasis = "gross_total_amount" | "paid_amount" | "net_paid_amount";

export type PeriodPreset = "last_3_months" | "last_6_months" | "last_12_months" | "calendar_year";

export type LastVisitBucket = "within_3m" | "over_3m" | "over_6m" | "over_1y" | "no_visit";

export interface AggregationOwner {
  owner_id: string;
  owner_name: string;
  annual_visit_count: number;
  total_visit_count: number;
  last_visit_date: string | null;
  first_visit_date: string | null;
  // Aggregation fields (optional, returned by BE based on query)
  annual_amount?: number;
  billing_count?: number;
  period_visit_count?: number;
  period_last_visit_date?: string | null;
  days_since_last_visit?: number | null;
  last_visit_bucket?: LastVisitBucket | null;
  // 累計診療費。BE は canonical の `total_amount` を返す。
  // `total_fee` は AGG-BE-005 の互換エイリアス（旧FE/CSV/外部連携向け）。
  total_amount?: number;
  total_fee?: number;
  // CPM セグメント（ISSUE-180）。BE が各飼主の CPM V1 判定結果を返す（`cpm_stage,omitempty`）。
  cpm_stage?: AggregationCPMStage;
}

export interface AggregationParams {
  sort?: AggregationSortField;
  order?: "asc" | "desc";
  search?: string;
  page?: number;
  per_page?: number;
  // Revenue aggregation parameters
  year?: number;
  amount_basis?: AmountBasis;
  min_amount?: number;
  max_amount?: number;
  include_zero?: boolean;
  // Visit count aggregation parameters
  period_preset?: PeriodPreset;
  from?: string; // YYYY-MM-DD
  to?: string; // YYYY-MM-DD
  min_visit_count?: number;
  max_visit_count?: number;
  // Last visit aggregation parameters
  last_visit_bucket?: string;
  include_no_visit?: boolean;
  // CPM セグメント絞り込み（ISSUE-180）。BE は "cpm_xxx" / "xxx" 双方を受理する。
  cpm_stage?: AggregationCPMStage;
}

export interface AggregationResponse {
  owners: AggregationOwner[];
  total: number;
  page: number;
  per_page: number;
  // EMR-334: cpm_stage 絞り込み・ページネーション前の母集団の CPM 区分人数。
  // 全 7 区分 (cpm_unclassified 含む) のキーを BE が必ず返す。
  cpm_stage_counts: Record<AggregationCPMStage, number>;
}

// EMR-334: BE 集計クエリの deadline 超過 (503 + code=aggregation_timeout) だけを
// 専用エラーとして扱う。axios timeout (ECONNABORTED) や通常の 4xx/5xx は含めない。
const AGGREGATION_TIMEOUT_CODE = "aggregation_timeout";

function isAggregationTimeoutError(error: unknown): boolean {
  const response = (
    error as { response?: { status?: number; data?: { code?: string } } } | null | undefined
  )?.response;
  return response?.status === 503 && response.data?.code === AGGREGATION_TIMEOUT_CODE;
}

// EMR-334: タイムアウト以外のエラーでは QueryClient の有効な既定 retry をそのまま
// 尊重する。query-core retryer の shouldRetry と同じ評価
// (true / failureCount < retry / function / 未設定時はブラウザ既定の 3)。
function defaultShouldRetry(
  queryClient: QueryClient,
  failureCount: number,
  error: unknown,
): boolean {
  const configured = queryClient.getDefaultOptions().queries?.retry;
  if (configured === true) return true;
  if (configured === false) return false;
  if (typeof configured === "function") {
    return configured(failureCount, error as Error);
  }
  return failureCount < (configured ?? 3);
}

// GET /api/clinics/:clinic_id/owners/aggregations
export function useGetOwnerAggregations(params: AggregationParams) {
  const queryClient = useQueryClient();
  const queryKey = queryKeys.ownerAggregations.list(params);
  return useQuery({
    queryKey,
    queryFn: async () => {
      const clinicId = requireStoredClinicId();
      // BUG-012: client timeout so dashboard leaves infinite loading on BE hang.
      const { data } = await axios.get<AggregationResponse>(
        `/v1/clinics/${clinicId}/owners/aggregations`,
        { params, timeout: 25_000 },
      );
      return data;
    },
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    // EMR-334: 集計タイムアウト (503 + aggregation_timeout) のみ自動リトライを止め、
    // 画面内の専用メッセージに委ねる。その他のエラーは QueryClient 既定の
    // リトライに従う (本番 retry:1 / テスト既定 retry:false / 未設定は 3)。
    retry: (failureCount, error) =>
      !isAggregationTimeoutError(error) && defaultShouldRetry(queryClient, failureCount, error),
    // EMR-334: QueryCache の共通エラーハンドラ (lib/react-query.ts) は
    // dispatch(error) で state が更新された後に silentError を読むため、
    // getter で「現在の」エラーを分類する。タイムアウトのみ true になり、
    // その他のエラーや直前エラーの名残で通知を抑止することはない。
    meta: {
      get silentError() {
        return isAggregationTimeoutError(queryClient.getQueryState(queryKey)?.error);
      },
    },
  });
}
