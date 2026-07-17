import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { VaccinationRecord } from "@/types";
import { transformVaccination } from "./transforms";
import type { BackendVaccination } from "./types";

export interface VaccinationFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string; // YYYY-MM-DD
}

const getVaccinations = async (
  filters?: VaccinationFilters,
): Promise<VaccinationRecord[]> => {
  const params: Record<string, string> = {};
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  const { data } = await axios.get<{ data: BackendVaccination[] }>(
    "/v1/vaccinations",
    { params },
  );
  return (data.data ?? []).map(transformVaccination);
};

export const useGetVaccinations = (filters?: VaccinationFilters) => {
  return useQuery({
    queryKey: queryKeys.vaccinations.list(filters),
    queryFn: () => getVaccinations(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
