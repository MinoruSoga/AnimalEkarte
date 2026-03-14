import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { VaccinationRecord } from "@/types";
import { transformVaccination } from "./transforms";
import type { BackendVaccination } from "./types";

export const getVaccinations = async (): Promise<VaccinationRecord[]> => {
  const { data } = await axios.get<{ data: BackendVaccination[] }>("/v1/vaccinations");
  return (data.data ?? []).map(transformVaccination);
};

export const useGetVaccinations = () => {
  return useQuery({
    queryKey: ["vaccinations"],
    queryFn: getVaccinations,
  });
};
