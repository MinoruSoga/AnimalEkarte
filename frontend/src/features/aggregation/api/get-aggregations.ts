import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

// Type definitions for aggregation features
export type AggregationSortField =
  | "total_fee"
  | "annual_visit_count"
  | "last_visit_date"
  | "annual_amount"
  | "period_visit_count"
  | "visit_count"
  | "owner_name";

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
  total_fee?: number;
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
    queryKey: ["owner-aggregations", params],
    queryFn: async () => {
      const clinicId = localStorage.getItem("auth_current_clinic:v1");
      const { data } = await axios.get<AggregationResponse>(
        `/v1/clinics/${clinicId}/owners/aggregations`,
        { params }
      );
      return data;
    },
    staleTime: 5 * 60 * 1000,
  });
}
