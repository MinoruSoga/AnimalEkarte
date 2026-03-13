import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Hospitalization } from "@/types";
import { transformHospitalization } from "./transforms";
import type { BackendHospitalization } from "./types";

interface HospitalizationListResponse {
  data: BackendHospitalization[];
  total: number;
  page: number;
  limit: number;
}

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

export const getHospitalizationRaw = async (
  id: string
): Promise<BackendHospitalization> => {
  const { data } = await axios.get<BackendHospitalization>(
    `/v1/hospitalizations/${id}`
  );
  return data;
};

export const useGetHospitalizationRaw = (id: string | undefined) => {
  return useQuery({
    queryKey: ["hospitalization", "raw", id],
    queryFn: () => getHospitalizationRaw(id!),
    enabled: !!id,
  });
};

// Fetch hospitalizations by pet ID
export const getHospitalizationsByPetId = async (
  petId: string
): Promise<Hospitalization[]> => {
  const { data } = await axios.get<HospitalizationListResponse>(
    "/v1/hospitalizations",
    { params: { pet_id: petId } }
  );
  return data.data.map(transformHospitalization);
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
  const { data } = await axios.get<HospitalizationListResponse>(
    "/v1/hospitalizations",
    { params: { owner_id: ownerId } }
  );
  return data.data.map(transformHospitalization);
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
  const { data } = await axios.get<HospitalizationListResponse>(
    `/v1/hospitalizations/status/${status}`
  );
  return data.data.map(transformHospitalization);
};

export const useGetHospitalizationsByStatus = (status: string) => {
  return useQuery({
    queryKey: ["hospitalizations", "status", status],
    queryFn: () => getHospitalizationsByStatus(status),
    enabled: !!status,
  });
};
