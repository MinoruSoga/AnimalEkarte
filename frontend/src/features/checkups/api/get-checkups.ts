import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformCheckupGlobal } from "./transforms";
import type { BackendCheckupGlobal, CheckupFilters } from "./types";
import type { CheckupRecord } from "./transforms";

const getCheckups = async (filters?: CheckupFilters): Promise<CheckupRecord[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.nextStartDate) params.next_start_date = filters.nextStartDate;
  if (filters?.nextEndDate) params.next_end_date = filters.nextEndDate;

  const { data } = await axios.get<{ data: BackendCheckupGlobal[] }>("/v1/checkups", { params });
  return (data.data ?? []).map(transformCheckupGlobal);
};

export const useGetCheckups = (filters?: CheckupFilters) => {
  return useQuery({
    queryKey: queryKeys.checkups.list(filters),
    queryFn: () => getCheckups(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
