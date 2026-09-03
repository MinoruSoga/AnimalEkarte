import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import {
  HOSPITALIZATION_LIST_DEFAULT_LIMIT,
  HOSPITALIZATION_LIST_DEFAULT_PAGE,
  type HospitalizationFilterStatus,
  type HospitalizationWireStatus,
  toHospitalizationWireStatus,
} from "../constants";
import { transformHospitalization } from "./transforms";
import type { Hospitalization } from "./transforms";
import type { BackendHospitalization } from "./types";

interface HospitalizationPaginatedResponse {
  data: BackendHospitalization[];
  total: number;
  page: number;
  limit: number;
}

export interface HospitalizationFilters {
  petId?: string;
  startDate?: string; // YYYY-MM-DD（入院開始日の範囲）
  endDate?: string; // YYYY-MM-DD
  /** UI タブ値（active/reserved/discharged/all）。wire 変換は getHospitalizations 内で行う。 */
  statusFilter?: HospitalizationFilterStatus;
  page?: number;
  limit?: number;
}

/** サーバページング封筒を捨てない list 結果（BUG-009）。 */
export interface HospitalizationsResult {
  data: Hospitalization[];
  total: number;
  page: number;
  limit: number;
}

const getHospitalizations = async (
  filters?: HospitalizationFilters,
): Promise<HospitalizationsResult> => {
  const page = filters?.page ?? HOSPITALIZATION_LIST_DEFAULT_PAGE;
  const limit = filters?.limit ?? HOSPITALIZATION_LIST_DEFAULT_LIMIT;
  const params: Record<string, string | number> = { page, limit };

  if (filters?.petId) params.pet_id = filters.petId;
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;

  if (filters?.statusFilter !== undefined) {
    const wireStatus: HospitalizationWireStatus | undefined = toHospitalizationWireStatus(
      filters.statusFilter,
    );
    if (wireStatus !== undefined) {
      params.status = wireStatus;
    }
  }

  const { data } = await axios.get<HospitalizationPaginatedResponse>("/v1/hospitalizations", {
    params,
  });
  return {
    data: (data.data ?? []).map(transformHospitalization),
    total: data.total,
    page: data.page,
    limit: data.limit,
  };
};

export const useGetHospitalizations = (filters?: HospitalizationFilters) => {
  return useQuery({
    queryKey: queryKeys.hospitalizations.list(filters),
    queryFn: () => getHospitalizations(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

/** テスト用に export（request params 契約の固定）。 */
export { getHospitalizations };
