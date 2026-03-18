import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { VaccinationRecord } from "@/types";
import { transformVaccination } from "./transforms";
import type { BackendVaccination } from "./types";

export const getVaccination = async (id: string): Promise<VaccinationRecord> => {
  const { data } = await axios.get<BackendVaccination>(`/v1/vaccinations/${id}`);
  return transformVaccination(data);
};

export const useGetVaccination = (id: string) => {
  return useQuery({
    queryKey: ["vaccination", id],
    queryFn: () => getVaccination(id),
    enabled: !!id,
  });
};

