import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

export interface LstepTagSummaryItem {
  tag_name: string;
  owner_count: number;
  category: "auto" | "manual";
}

export interface LstepTagSummaryResponse {
  tags: LstepTagSummaryItem[];
  total_owners_with_lstep: number;
  as_of: string; // ISO datetime
}

// GET /api/clinics/:clinic_id/lstep/tag-summary
export function useGetLstepTagSummary() {
  return useQuery({
    queryKey: queryKeys.lstepTagSummary(),
    queryFn: async () => {
      const clinicId = requireStoredClinicId();
      const { data } = await axios.get<LstepTagSummaryResponse>(
        `/v1/clinics/${clinicId}/lstep/tag-summary`,
      );
      return data;
    },
    staleTime: QUERY_STALE_TIMES.MEDIUM, // 5分キャッシュ
  });
}
