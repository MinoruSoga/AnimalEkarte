import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { isPersistedPetId } from "@/lib/pet-id";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { PetResponse } from "@/types/generated/pet-responses";

/** GET /v1/pets list は detail より薄いが、共通フィールドは PetResponse と同形（pet_name_kana）。 */
type BackendPetList = Omit<
  PetResponse,
  "version" | "phone" | "deceased_at" | "created_at" | "updated_at"
>;

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
  /**
   * #86 拠点横断一覧。所属医院が現在拠点のみの既定 scope では省略し、
   * 複数拠点はコンマ直列化して `clinic_ids` として送る（backend が検証済み）。
   */
  clinicIds?: string[];
}

interface GetPetsQueryOptions {
  enabled?: boolean;
  /** ページ切替中に前回データを保持する一覧画面だけが明示的に有効化する。 */
  preservePreviousData?: boolean;
  /** 既定は STATIC(30分)。新規登録直後に出る必要がある検索用途は短い値を渡す。 */
  staleTime?: number;
  refetchOnWindowFocus?: boolean;
}

function transformBackendPetListToFrontend(pet: BackendPetList): Pet {
  // list DTO は version/phone/timestamps を持たないため、detail 変換へ渡すときだけ既定値を補う。
  return transformBackendPetToFrontend({
    ...pet,
    version: 0,
    phone: pet.owner?.phone ?? "",
    created_at: "",
    updated_at: "",
  });
}

/**
 * Shared hook for fetching a single pet by ID.
 * Uses the same query key as features/pets to share React Query cache.
 */
export function useGetPet(petId: string) {
  // BUG-022: ローカル pending の temp-* ID で GET /v1/pets/:id を打たない
  return useQuery({
    queryKey: queryKeys.pets.detail(petId),
    queryFn: async (): Promise<Pet> => {
      const { data } = await axios.get<PetResponse>(`/v1/pets/${petId}`);
      return transformBackendPetToFrontend(data);
    },
    enabled: isPersistedPetId(petId),
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
    // EMR-222: clinic scope でもキャッシュを分離する（search/species と同じ扱い）。
    clinicIds: options.clinicIds?.length ? options.clinicIds : undefined,
  };
  const hasServerListKey = Object.values(serverListKey).some(
    (value) => value !== undefined && value !== "",
  );
  const baseQueryKey = queryKeys.pets.list(ownerId, options);
  const query = useQuery({
    queryKey: hasServerListKey ? [...baseQueryKey, serverListKey] : baseQueryKey,
    queryFn: async () => {
      const params = {
        ...(ownerId ? { owner_id: ownerId } : {}),
        ...(options.includeDeceased ? { include_deceased: "true" } : {}),
        ...(options.page !== undefined ? { page: options.page } : {}),
        ...(options.limit !== undefined ? { limit: options.limit } : {}),
        ...(options.search ? { search: options.search } : {}),
        ...(options.species ? { species: options.species } : {}),
        ...(options.clinicIds?.length ? { clinic_ids: options.clinicIds.join(",") } : {}),
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
    placeholderData: queryOptions.preservePreviousData ? keepPreviousData : undefined,
    staleTime: queryOptions.staleTime ?? QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    refetchOnWindowFocus: queryOptions.refetchOnWindowFocus,
  });

  return {
    ...query,
    data: query.data?.pets,
    total: query.data?.total,
    page: query.data?.page,
    limit: query.data?.limit,
  };
}
