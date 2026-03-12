import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { MedicalRecord } from "@/types";
import { transformMedicalRecord } from "./transforms";
import type { BackendMedicalRecord } from "./types";

interface MedicalRecordListResponse {
  data: BackendMedicalRecord[];
  total: number;
  page: number;
  limit: number;
}

export const getMedicalRecord = async (id: string): Promise<MedicalRecord> => {
  const { data } = await axios.get<BackendMedicalRecord>(
    `/v1/medical-records/${id}`
  );
  return transformMedicalRecord(data);
};

export const useGetMedicalRecord = (id: string) => {
  return useQuery({
    queryKey: ["medical-record", id],
    queryFn: () => getMedicalRecord(id),
    enabled: !!id,
  });
};

// Fetch medical records by pet ID
export const getMedicalRecordsByPetId = async (
  petId: string
): Promise<MedicalRecord[]> => {
  const { data } = await axios.get<MedicalRecordListResponse>(
    "/v1/medical-records",
    { params: { pet_id: petId } }
  );
  return data.data.map(transformMedicalRecord);
};

export const useGetMedicalRecordsByPetId = (petId: string) => {
  return useQuery({
    queryKey: ["medical-records", "pet", petId],
    queryFn: () => getMedicalRecordsByPetId(petId),
    enabled: !!petId,
  });
};

// Fetch medical records by owner ID
export const getMedicalRecordsByOwnerId = async (
  ownerId: string
): Promise<MedicalRecord[]> => {
  const { data } = await axios.get<MedicalRecordListResponse>(
    "/v1/medical-records",
    { params: { owner_id: ownerId } }
  );
  return data.data.map(transformMedicalRecord);
};

export const useGetMedicalRecordsByOwnerId = (ownerId: string) => {
  return useQuery({
    queryKey: ["medical-records", "owner", ownerId],
    queryFn: () => getMedicalRecordsByOwnerId(ownerId),
    enabled: !!ownerId,
  });
};
