import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { requireStoredClinicId } from "@/lib/current-clinic";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { queryKeys } from "@/lib/query-keys";

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
    queryKey: queryKeys.lstepTagOwners.list(params),
    queryFn: async () => {
      const clinicId = requireStoredClinicId();
      const { data } = await axios.get<LstepTagOwnersResponse>(
        `/v1/clinics/${clinicId}/lstep/owners`,
        { params }
      );
      return data;
    },
    enabled: !!params.tag,
  });
}

// GET /api/clinics/:clinic_id/lstep/owners（全ページを逐次取得。CSV出力用）
export async function fetchAllLstepTagOwners(
  tagName: string,
  ownerCount: number
): Promise<LstepTagOwner[]> {
  const clinicId = requireStoredClinicId();
  const perPage = HISTORY_FETCH_LIMIT;
  const totalPages = Math.max(1, Math.ceil(ownerCount / perPage));
  const owners: LstepTagOwner[] = [];

  for (let page = 1; page <= totalPages; page += 1) {
    const { data } = await axios.get<LstepTagOwnersResponse>(
      `/v1/clinics/${clinicId}/lstep/owners`,
      { params: { tag: tagName, page, per_page: perPage } }
    );
    owners.push(...data.owners);
    if (owners.length >= data.total || data.owners.length === 0) {
      break;
    }
  }

  return owners;
}
