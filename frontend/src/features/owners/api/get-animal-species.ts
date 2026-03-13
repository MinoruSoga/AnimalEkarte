import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { BackendAnimalSpecies } from "./types";

export const getAnimalSpecies = async (): Promise<BackendAnimalSpecies[]> => {
  const { data } = await axios.get<BackendAnimalSpecies[]>("/v1/masters/animal-species");
  return data;
};

export const useGetAnimalSpecies = () => {
  return useQuery({
    queryKey: ["masters", "animal-species"],
    queryFn: getAnimalSpecies,
    staleTime: 30 * 60 * 1000, // 静的マスタデータは30分キャッシュ
  });
};
