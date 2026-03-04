import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Hospitalization } from "@/types";
import { transformHospitalization } from "./transforms";
import type { BackendHospitalization } from "./types";

export const getHospitalization = async (
  id: string
): Promise<Hospitalization> => {
  const { data } = await axios.get<BackendHospitalization>(
    `/v1/hospitalizations/${id}`
  );
  return transformHospitalization(data);
};

export const useGetHospitalization = (id: string) => {
  return useQuery({
    queryKey: ["hospitalization", id],
    queryFn: () => getHospitalization(id),
    enabled: !!id,
  });
};

// Fetch hospitalizations by pet ID
export const getHospitalizationsByPetId = async (
  petId: string
): Promise<Hospitalization[]> => {
  const { data } = await axios.get<BackendHospitalization[]>(
    `/v1/pets/${petId}/hospitalizations`
  );
  return data.map(transformHospitalization);
};

export const useGetHospitalizationsByPetId = (petId: string) => {
  return useQuery({
    queryKey: ["hospitalizations", "pet", petId],
    queryFn: () => getHospitalizationsByPetId(petId),
    enabled: !!petId,
  });
};

// Fetch hospitalizations by owner ID
export const getHospitalizationsByOwnerId = async (
  ownerId: string
): Promise<Hospitalization[]> => {
  const { data } = await axios.get<BackendHospitalization[]>(
    `/v1/owners/${ownerId}/hospitalizations`
  );
  return data.map(transformHospitalization);
};

export const useGetHospitalizationsByOwnerId = (ownerId: string) => {
  return useQuery({
    queryKey: ["hospitalizations", "owner", ownerId],
    queryFn: () => getHospitalizationsByOwnerId(ownerId),
    enabled: !!ownerId,
  });
};

// Fetch hospitalizations by status
export const getHospitalizationsByStatus = async (
  status: string
): Promise<Hospitalization[]> => {
  const { data } = await axios.get<BackendHospitalization[]>(
    `/v1/hospitalizations/status/${status}`
  );
  return data.map(transformHospitalization);
};

export const useGetHospitalizationsByStatus = (status: string) => {
  return useQuery({
    queryKey: ["hospitalizations", "status", status],
    queryFn: () => getHospitalizationsByStatus(status),
    enabled: !!status,
  });
};
