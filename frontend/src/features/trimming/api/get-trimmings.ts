import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { queryKeys } from "@/lib/query-keys";
import type { TrimmingUI } from "@/types";
import { transformTrimming } from "./transforms";
import type { TrimmingListResponse } from "@/types/trimming";

export interface TrimmingFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string;   // YYYY-MM-DD
  petId?: string;
  enabled?: boolean;
}

const getTrimmings = async (filters?: TrimmingFilters): Promise<TrimmingUI[]> => {
  const params: Record<string, string | number> = { page: 1, limit: HISTORY_FETCH_LIMIT };
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.petId) params.pet_id = filters.petId;
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings", { params });
  return data.data.reduce<ReturnType<typeof transformTrimming>[]>((acc, d) => {
    if (d.pet?.id != null) acc.push(transformTrimming(d));
    return acc;
  }, []);
};

export const useGetTrimmings = (filters?: TrimmingFilters) => {
  return useQuery({
    queryKey: queryKeys.trimmings.list(filters),
    queryFn: () => getTrimmings(filters),
    enabled: filters?.enabled ?? true,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
