import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

// Type definitions for aggregation features
export type LtvSortField =
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

export interface LtvOwner {
  owner_id: string;
  owner_name: string;
  has_line: boolean;
  cpm_stage: "encounter" | "growing" | "core" | "noah" | "spot" | "dormant" | null;
  total_fee: number;
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
}

export interface LtvOwnersParams {
  sort?: LtvSortField;
  order?: "asc" | "desc";
  cpm_stage?: string;
  has_line?: boolean;
  min_total_fee?: number;
  max_total_fee?: number;
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
  max_visit_count?: number;
  // Last visit aggregation parameters
  last_visit_bucket?: string;
  include_no_visit?: boolean;
}

export interface LtvOwnersResponse {
  owners: LtvOwner[];
  total: number;
  page: number;
  per_page: number;
}

// GET /api/clinics/:clinic_id/owners/ltv
export function useGetLtvOwners(params: LtvOwnersParams) {
  return useQuery({
    queryKey: ["ltv-owners", params],
    queryFn: async () => {
      const clinicId = localStorage.getItem("auth_current_clinic:v1");
      const { data } = await axios.get<LtvOwnersResponse>(
        `/v1/clinics/${clinicId}/owners/ltv`,
        { params }
      );
      return data;
    },
    staleTime: 5 * 60 * 1000,
  });
}
