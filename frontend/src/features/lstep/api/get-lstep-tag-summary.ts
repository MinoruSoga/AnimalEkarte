import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

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
    queryKey: ["lstep-tag-summary"],
    queryFn: async () => {
      const clinicId = localStorage.getItem("auth_current_clinic:v1");
      const { data } = await axios.get<LstepTagSummaryResponse>(
        `/v1/clinics/${clinicId}/lstep/tag-summary`
      );
      return data;
    },
    staleTime: 5 * 60 * 1000, // 5分キャッシュ
  });
}
