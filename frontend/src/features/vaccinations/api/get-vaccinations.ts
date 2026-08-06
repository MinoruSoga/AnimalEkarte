import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { VaccinationRecord } from "@/types";
import { transformVaccination } from "./transforms";
import type { BackendVaccination } from "./types";

export interface VaccinationFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string; // YYYY-MM-DD
  /** Server-side pet filter (snake_case pet_id). Required for pet history panels. */
  petId?: string;
  page?: number;
  limit?: number;
}

const getVaccinations = async (
  filters?: VaccinationFilters,
): Promise<VaccinationRecord[]> => {
  const params: Record<string, string | number> = {
    page: filters?.page ?? 1,
    // Default limit was BE 20, which hid 2026 rows behind 2029 seed page-window (BUG-007).
    limit: filters?.limit ?? HISTORY_FETCH_LIMIT,
  };
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.petId) params.pet_id = filters.petId;

  const { data } = await axios.get<{ data: BackendVaccination[] }>(
    "/v1/vaccinations",
    { params },
  );
  return (data.data ?? []).map(transformVaccination);
};

export const useGetVaccinations = (filters?: VaccinationFilters) => {
  // petId: undefined/omitted → list mode (enabled). petId: "" → wait for selection.
  const petScoped = filters != null && "petId" in filters;
  const enabled = !petScoped || Boolean(filters?.petId);

  return useQuery({
    queryKey: queryKeys.vaccinations.list(filters),
    queryFn: () => getVaccinations(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
    enabled,
  });
};
