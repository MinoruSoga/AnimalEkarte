import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformCheckupGlobal } from "./transforms";
import type { BackendCheckupGlobal, CheckupFilters } from "./types";
import type { CheckupRecord } from "./transforms";

interface CheckupsListResponse {
  data: BackendCheckupGlobal[];
  total: number;
  page: number;
  limit: number;
}

// X-16②: BE が実サーバページング封筒 {data,total,page,limit} を返すようになったため、
// FE も total を捨てず消費する（先例: medical-records の get-medical-records.ts）。
export interface CheckupsResult {
  data: CheckupRecord[];
  total: number;
  page: number;
  limit: number;
}

const DEFAULT_PAGE = 1;
const DEFAULT_LIMIT = 20;

export const getCheckups = async (filters?: CheckupFilters): Promise<CheckupsResult> => {
  const params: Record<string, string | number> = {
    page: filters?.page ?? DEFAULT_PAGE,
    limit: filters?.limit ?? DEFAULT_LIMIT,
  };
  if (filters?.petId) params.pet_id = filters.petId;
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.nextStartDate) params.next_start_date = filters.nextStartDate;
  if (filters?.nextEndDate) params.next_end_date = filters.nextEndDate;

  const { data } = await axios.get<CheckupsListResponse>("/v1/checkups", { params });
  return {
    data: (data.data ?? []).map(transformCheckupGlobal),
    total: data.total,
    page: data.page,
    limit: data.limit,
  };
};

export const useGetCheckups = (filters?: CheckupFilters) => {
  return useQuery({
    queryKey: queryKeys.checkups.list(filters),
    queryFn: () => getCheckups(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
