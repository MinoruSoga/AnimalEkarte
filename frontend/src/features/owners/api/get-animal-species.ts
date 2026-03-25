import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES } from "@/lib/react-query";
import type { AnimalSpecies as BackendAnimalSpecies } from "@/types/generated/models";

export const getAnimalSpecies = async (): Promise<BackendAnimalSpecies[]> => {
  const { data } = await axios.get<BackendAnimalSpecies[]>("/v1/masters/animal-species");
  return data;
};

export const useGetAnimalSpecies = () => {
  return useQuery({
    queryKey: ["masters", "animal-species"],
    queryFn: getAnimalSpecies,
    staleTime: QUERY_STALE_TIMES.STATIC, // 静的マスタデータは30分キャッシュ
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
