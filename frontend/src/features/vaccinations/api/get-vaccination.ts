import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { VaccinationRecord } from "@/types";
import { transformVaccination } from "./transforms";
import type { BackendVaccination } from "./types";

const getVaccination = async (id: string): Promise<VaccinationRecord> => {
  const { data } = await axios.get<BackendVaccination>(`/v1/vaccinations/${id}`);
  return transformVaccination(data);
};

export const useGetVaccination = (id: string) => {
  return useQuery({
    queryKey: queryKeys.vaccinations.detail(id),
    queryFn: () => getVaccination(id),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
