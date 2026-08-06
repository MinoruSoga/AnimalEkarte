import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";
import type { CPMStage } from "@/lib/cpm-stage";
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

export type AmountBasis =
  | "gross_total_amount"
  | "paid_amount"
  | "net_paid_amount";

export type PeriodPreset =
  | "last_3_months"
  | "last_6_months"
  | "last_12_months"
  | "calendar_year";

export type LastVisitBucket =
  | "within_3m"
  | "over_3m"
  | "over_6m"
  | "over_1y"
  | "no_visit";

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
  cpm_stage?: CPMStage;
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
  cpm_stage?: CPMStage;
}

export interface AggregationResponse {
  owners: AggregationOwner[];
  total: number;
  page: number;
  per_page: number;
}

// GET /api/clinics/:clinic_id/owners/aggregations
export function useGetOwnerAggregations(params: AggregationParams) {
  return useQuery({
    queryKey: queryKeys.ownerAggregations.list(params),
    queryFn: async () => {
      const clinicId = requireStoredClinicId();
      // BUG-012: client timeout so dashboard leaves infinite loading on BE hang.
      const { data } = await axios.get<AggregationResponse>(
        `/v1/clinics/${clinicId}/owners/aggregations`,
        { params, timeout: 25_000 }
      );
      return data;
    },
    staleTime: QUERY_STALE_TIMES.MEDIUM,
  });
}
