import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ExaminationRecord } from "@/types";
import { transformExamination } from "./transforms";
import type { BackendExamination } from "./types";

interface ExaminationListResponse {
  data: BackendExamination[];
  total: number;
  page: number;
  limit: number;
}

export const getExamination = async (id: string): Promise<ExaminationRecord> => {
  const { data } = await axios.get<BackendExamination>(`/v1/examinations/${id}`);
  return transformExamination(data);
};

export const useGetExamination = (id: string) => {
  return useQuery({
    queryKey: ["examination", id],
    queryFn: () => getExamination(id),
    enabled: !!id,
  });
};

// Fetch examinations by pet ID
export const getExaminationsByPetId = async (
  petId: string
): Promise<ExaminationRecord[]> => {
  const { data } = await axios.get<ExaminationListResponse>("/v1/examinations", {
    params: { pet_id: petId },
  });
  return data.data.map(transformExamination);
};

export const useGetExaminationsByPetId = (petId: string) => {
  return useQuery({
    queryKey: ["examinations", "pet", petId],
    queryFn: () => getExaminationsByPetId(petId),
    enabled: !!petId,
  });
};

// Fetch examinations by owner ID
export const getExaminationsByOwnerId = async (
  ownerId: string
): Promise<ExaminationRecord[]> => {
  const { data } = await axios.get<ExaminationListResponse>("/v1/examinations", {
    params: { owner_id: ownerId },
  });
  return data.data.map(transformExamination);
};

export const useGetExaminationsByOwnerId = (ownerId: string) => {
  return useQuery({
    queryKey: ["examinations", "owner", ownerId],
    queryFn: () => getExaminationsByOwnerId(ownerId),
    enabled: !!ownerId,
  });
};

// Fetch examinations by status
export const getExaminationsByStatus = async (
  status: string
): Promise<ExaminationRecord[]> => {
  const { data } = await axios.get<ExaminationListResponse>("/v1/examinations", {
    params: { status },
  });
  return data.data.map(transformExamination);
};

export const useGetExaminationsByStatus = (status: string) => {
  return useQuery({
    queryKey: ["examinations", "status", status],
    queryFn: () => getExaminationsByStatus(status),
    enabled: !!status,
  });
};
