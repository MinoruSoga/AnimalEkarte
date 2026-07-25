import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { Pet as BackendPet } from "@/types/generated/models";

interface PetListResponse {
  data: BackendPet[];
}

interface GetPetsOptions {
  includeDeceased?: boolean;
}

interface GetPetsQueryOptions {
  enabled?: boolean;
}

/**
 * Shared hook for fetching a single pet by ID.
 * Uses the same query key as features/pets to share React Query cache.
 */
export function useGetPet(petId: string) {
  return useQuery({
    queryKey: queryKeys.pets.detail(petId),
    queryFn: async (): Promise<Pet> => {
      const { data } = await axios.get<BackendPet>(`/v1/pets/${petId}`);
      return transformBackendPetToFrontend(data);
    },
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

/**
 * Shared hook for fetching a list of pets, optionally filtered by ownerId.
 * Uses the same query key as features/pets to share React Query cache.
 */
export function useGetPets(
  ownerId?: string,
  options: GetPetsOptions = {},
  queryOptions: GetPetsQueryOptions = {},
) {
  return useQuery({
    queryKey: queryKeys.pets.list(ownerId, options),
    queryFn: async (): Promise<Pet[]> => {
      const params = {
        ...(ownerId ? { owner_id: ownerId } : {}),
        ...(options.includeDeceased ? { include_deceased: "true" } : {}),
      };
      const { data } = await axios.get<PetListResponse>("/v1/pets", { params });
      return data.data.map(transformBackendPetToFrontend);
    },
    enabled: queryOptions.enabled ?? true,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

/**
 * Shared hook for searching/listing pets, optionally filtered by ownerId.
 */
export function useSearchPets(ownerId?: string) {
  const { data: pets, isLoading, error, isPending } = useGetPets(ownerId);
  return {
    pets: pets ?? [],
    isLoading,
    error,
    isPending,
  };
}
