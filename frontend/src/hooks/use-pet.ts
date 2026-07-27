import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { Pet as BackendPet } from "@/types/generated/models";

type BackendPetList = Omit<BackendPet, "name_kana"> & {
  /** GET /v1/pets の list DTO は detail DTO と異なる wire name を使う。 */
  pet_name_kana: string;
};

interface PetListResponse {
  data: BackendPetList[];
  total?: number;
  page?: number;
  limit?: number;
}

interface GetPetsOptions {
  includeDeceased?: boolean;
  page?: number;
  limit?: number;
  search?: string;
  /** animal_species_id の10進文字列。 */
  species?: string;
}

interface GetPetsQueryOptions {
  enabled?: boolean;
  /** ページ切替中に前回データを保持する一覧画面だけが明示的に有効化する。 */
  preservePreviousData?: boolean;
}

function transformBackendPetListToFrontend(pet: BackendPetList): Pet {
  return transformBackendPetToFrontend({
    ...pet,
    name_kana: pet.pet_name_kana,
  });
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
  const serverListKey = {
    page: options.page,
    limit: options.limit,
    search: options.search,
    species: options.species,
  };
  const hasServerListKey = Object.values(serverListKey).some(
    (value) => value !== undefined && value !== "",
  );
  const baseQueryKey = queryKeys.pets.list(ownerId, options);
  const query = useQuery({
    queryKey: hasServerListKey
      ? [...baseQueryKey, serverListKey]
      : baseQueryKey,
    queryFn: async () => {
      const params = {
        ...(ownerId ? { owner_id: ownerId } : {}),
        ...(options.includeDeceased ? { include_deceased: "true" } : {}),
        ...(options.page !== undefined ? { page: options.page } : {}),
        ...(options.limit !== undefined ? { limit: options.limit } : {}),
        ...(options.search ? { search: options.search } : {}),
        ...(options.species ? { species: options.species } : {}),
      };
      const { data } = await axios.get<PetListResponse>("/v1/pets", { params });
      return {
        pets: data.data.map(transformBackendPetListToFrontend),
        total: data.total,
        page: data.page,
        limit: data.limit,
      };
    },
    enabled: queryOptions.enabled ?? true,
    placeholderData: queryOptions.preservePreviousData
      ? keepPreviousData
      : undefined,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

  return {
    ...query,
    data: query.data?.pets,
    total: query.data?.total,
    page: query.data?.page,
    limit: query.data?.limit,
  };
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
