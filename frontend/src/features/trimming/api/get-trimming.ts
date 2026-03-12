import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { TrimmingRecord } from "@/types";
import { transformTrimming } from "./transforms";
import type { BackendTrimming } from "./types";

interface TrimmingListResponse {
  data: BackendTrimming[];
  total: number;
  page: number;
  limit: number;
}

export const getTrimming = async (id: string): Promise<TrimmingRecord> => {
  const { data } = await axios.get<BackendTrimming>(`/v1/trimmings/${id}`);
  return transformTrimming(data);
};

export const useGetTrimming = (id: string) => {
  return useQuery({
    queryKey: ["trimming", id],
    queryFn: () => getTrimming(id),
    enabled: !!id,
  });
};

// Fetch trimmings by pet ID
export const getTrimmingsByPetId = async (
  petId: string
): Promise<TrimmingRecord[]> => {
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings", {
    params: { pet_id: petId },
  });
  return data.data.map(transformTrimming);
};

export const useGetTrimmingsByPetId = (petId: string) => {
  return useQuery({
    queryKey: ["trimmings", "pet", petId],
    queryFn: () => getTrimmingsByPetId(petId),
    enabled: !!petId,
  });
};

// Fetch trimmings by owner ID
export const getTrimmingsByOwnerId = async (
  ownerId: string
): Promise<TrimmingRecord[]> => {
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings", {
    params: { owner_id: ownerId },
  });
  return data.data.map(transformTrimming);
};

export const useGetTrimmingsByOwnerId = (ownerId: string) => {
  return useQuery({
    queryKey: ["trimmings", "owner", ownerId],
    queryFn: () => getTrimmingsByOwnerId(ownerId),
    enabled: !!ownerId,
  });
};

// Fetch trimmings by status
export const getTrimmingsByStatus = async (
  status: string
): Promise<TrimmingRecord[]> => {
  const { data } = await axios.get<TrimmingListResponse>(
    `/v1/trimmings/status/${status}`
  );
  return data.data.map(transformTrimming);
};

export const useGetTrimmingsByStatus = (status: string) => {
  return useQuery({
    queryKey: ["trimmings", "status", status],
    queryFn: () => getTrimmingsByStatus(status),
    enabled: !!status,
  });
};
