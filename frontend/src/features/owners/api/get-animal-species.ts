import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { AnimalSpecies as BackendAnimalSpecies } from "@/types/generated/models";

export const getAnimalSpecies = async (): Promise<BackendAnimalSpecies[]> => {
  const { data } = await axios.get<BackendAnimalSpecies[]>("/v1/masters/animal-species");
  return data;
};

export const useGetAnimalSpecies = (opts?: { includeInactive?: boolean }) => {
  return useQuery({
    queryKey: ["masters", "animal-species", opts?.includeInactive ? "all" : "active"],
    queryFn: getAnimalSpecies,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
