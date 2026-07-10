import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Pet } from "@/types";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { PetListResponse } from "@/types/pet";

const getPets = async (ownerId?: string): Promise<Pet[]> => {
  const params = ownerId ? { owner_id: ownerId } : {};
  const { data } = await axios.get<PetListResponse>("/v1/pets", { params });
  return data.data.map(transformBackendPetToFrontend);
};

export const useGetPets = (ownerId?: string) => {
  return useQuery({
    queryKey: ownerId ? ["pets", { ownerId }] : ["pets"],
    queryFn: () => getPets(ownerId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
