import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { TrimmingUI } from "@/types";
import { transformTrimming } from "./transforms";
import type { TrimmingListResponse } from "@/types/trimming";

export interface TrimmingFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string;   // YYYY-MM-DD
}

export const getTrimmings = async (filters?: TrimmingFilters): Promise<TrimmingUI[]> => {
  const params: Record<string, string | number> = { page: 1, limit: 100 };
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings", { params });
  return data.data.reduce<ReturnType<typeof transformTrimming>[]>((acc, d) => {
    if (d.pet?.id != null) acc.push(transformTrimming(d));
    return acc;
  }, []);
};

export const useGetTrimmings = (filters?: TrimmingFilters) => {
  return useQuery({
    queryKey: ["trimmings", filters],
    queryFn: () => getTrimmings(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
