import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

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
}

export interface LtvOwnersParams {
  sort?: "total_fee" | "annual_visit_count" | "last_visit_date";
  order?: "asc" | "desc";
  cpm_stage?: string;
  has_line?: boolean;
  min_total_fee?: number;
  max_total_fee?: number;
  search?: string;
  page?: number;
  per_page?: number;
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
