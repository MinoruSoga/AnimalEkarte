import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export interface LstepTagOwner {
  owner_id: string;
  owner_name: string;
  line_user_id: string | null;
  last_visit_date: string | null;
  reason?: string; // BE は omitempty で nil/空時に省略
}

export interface LstepTagOwnersParams {
  tag: string;
  page?: number;
  per_page?: number;
  search?: string;
}

export interface LstepTagOwnersResponse {
  owners: LstepTagOwner[];
  total: number;
}

// GET /api/clinics/:clinic_id/lstep/owners
export function useGetLstepTagOwners(params: LstepTagOwnersParams) {
  return useQuery({
    queryKey: ["lstep-tag-owners", params],
    queryFn: async () => {
      const clinicId = localStorage.getItem("auth_current_clinic:v1");
      const { data } = await axios.get<LstepTagOwnersResponse>(
        `/v1/clinics/${clinicId}/lstep/owners`,
        { params }
      );
      return data;
    },
    enabled: !!params.tag,
  });
}
