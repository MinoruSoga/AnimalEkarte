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

// Fetch vaccinations by pet ID
export const getVaccinationsByPetId = async (
  petId: string
): Promise<VaccinationRecord[]> => {
  const { data } = await axios.get<BackendVaccination[]>("/v1/vaccinations", {
    params: { pet_id: petId },
  });
  return data.map(transformVaccination);
};

export const useGetVaccinationsByPetId = (petId: string) => {
  return useQuery({
    queryKey: ["vaccinations", "pet", petId],
    queryFn: () => getVaccinationsByPetId(petId),
    enabled: !!petId,
  });
};

// Fetch vaccinations by owner ID
export const getVaccinationsByOwnerId = async (
  ownerId: string
): Promise<VaccinationRecord[]> => {
  const { data } = await axios.get<BackendVaccination[]>("/v1/vaccinations", {
    params: { owner_id: ownerId },
  });
  return data.map(transformVaccination);
};

export const useGetVaccinationsByOwnerId = (ownerId: string) => {
  return useQuery({
    queryKey: ["vaccinations", "owner", ownerId],
    queryFn: () => getVaccinationsByOwnerId(ownerId),
    enabled: !!ownerId,
  });
};
