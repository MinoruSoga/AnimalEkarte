import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { queryKeys } from "@/lib/query-keys";
import type { TrimmingUI } from "@/types";
import { transformTrimming } from "./transforms";
import type { BackendTrimming, TrimmingListResponse } from "@/types/trimming";

const getTrimming = async (id: string): Promise<TrimmingUI> => {
  const { data } = await axios.get<BackendTrimming>(`/v1/trimmings/${id}`);
  return transformTrimming(data);
};

export const useGetTrimming = (id: string) => {
  return useQuery({
    queryKey: queryKeys.trimmings.detail(id),
    queryFn: () => getTrimming(id),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

// Fetch trimmings by pet ID
const getTrimmingsByPetId = async (petId: string): Promise<TrimmingUI[]> => {
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings", {
    params: { pet_id: petId, page: 1, limit: HISTORY_FETCH_LIMIT },
  });
  return data.data.map(transformTrimming);
};

export const useGetTrimmingsByPetId = (petId: string) => {
  return useQuery({
    queryKey: queryKeys.trimmings.byPet(petId),
    queryFn: () => getTrimmingsByPetId(petId),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
