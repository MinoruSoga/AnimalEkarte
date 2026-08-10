import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { isPersistedPetId } from "@/lib/pet-id";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_GC_TIMES, QUERY_STALE_TIMES } from "@/lib/react-query";

export interface PetSubOwnerApiResponse {
  owner_id: number;
  name: string;
  name_kana: string;
  relationship: string;
}

export interface PetSubOwnersResponse {
  sub_owners: PetSubOwnerApiResponse[];
}

export interface PetSubOwnerMetadataResponse {
  owner_id: number;
  version: number;
}

export interface SubOwnerCandidate {
  ownerId: number;
  name: string;
  nameKana: string;
}

interface OwnerCandidateApiResponse {
  id: number;
  owner_name: string;
  owner_name_kana: string;
}

interface OwnerCandidatesPageResponse {
  data: OwnerCandidateApiResponse[];
}

export async function getPetSubOwners(
  petId: string,
): Promise<PetSubOwnersResponse> {
  const { data } = await axios.get<PetSubOwnersResponse>(
    `/v1/pets/${petId}/sub-owners`,
  );
  return data;
}

export async function getPetSubOwnerMetadata(
  petId: string,
): Promise<PetSubOwnerMetadataResponse> {
  const { data } = await axios.get<PetSubOwnerMetadataResponse>(
    `/v1/pets/${petId}`,
  );
  return {
    owner_id: data.owner_id,
    version: data.version,
  };
}

export async function getSubOwnerCandidates(
  search: string,
): Promise<SubOwnerCandidate[]> {
  const normalizedSearch = search.trim();
  if (normalizedSearch === "") {
    return [];
  }

  const { data } = await axios.get<OwnerCandidatesPageResponse>("/v1/owners", {
    params: { search: normalizedSearch },
  });
  return data.data.map((owner) => ({
    ownerId: owner.id,
    name: owner.owner_name,
    nameKana: owner.owner_name_kana,
  }));
}

export function useGetPetSubOwners(petId: string) {
  // BUG-022: pending ペットの temp-* など非永続 ID では API を発行しない
  const canFetch = isPersistedPetId(petId);
  return useQuery({
    queryKey: queryKeys.petSubOwners.detail(petId),
    queryFn: () => getPetSubOwners(petId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: canFetch,
  });
}

export function useGetPetSubOwnerMetadata(petId: string) {
  // BUG-022: pending ペットの temp-* など非永続 ID では API を発行しない
  const canFetch = isPersistedPetId(petId);
  return useQuery({
    queryKey: queryKeys.petSubOwners.metadata(petId),
    queryFn: () => getPetSubOwnerMetadata(petId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: canFetch,
  });
}

export function useGetSubOwnerCandidates(search: string, enabled = true) {
  const normalizedSearch = search.trim();

  return useQuery({
    queryKey: queryKeys.owners.subOwnerOptions(normalizedSearch),
    queryFn: () => getSubOwnerCandidates(normalizedSearch),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: enabled && normalizedSearch !== "",
  });
}
