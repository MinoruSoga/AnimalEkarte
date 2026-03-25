import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { TrimmingUI } from "@/types";
import { transformTrimming } from "./transforms";
import type { BackendTrimming, TrimmingListResponse } from "@/types/trimming";

export const getTrimming = async (id: string): Promise<TrimmingUI> => {
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
): Promise<TrimmingUI[]> => {
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

