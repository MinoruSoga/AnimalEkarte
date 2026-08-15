import { useQuery } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_GC_TIMES, QUERY_STALE_TIMES } from "@/lib/react-query";

export interface OwnerSharedPetApiResponse {
  id: number;
  pet_number: string;
  name: string;
  status: string;
  gender: string;
  animal_species: {
    name: string;
  };
  birth_date: string | null;
  color: string;
  weight: number | null;
  environment: string;
  remarks: string;
  relationship: string;
}

export interface OwnerSharedPetsResponse {
  shared_pets: OwnerSharedPetApiResponse[];
}

export async function getOwnerSharedPets(
  ownerId: string,
): Promise<OwnerSharedPetsResponse> {
  const { data } = await axios.get<OwnerSharedPetsResponse>(
    `/v1/owners/${ownerId}/shared-pets`,
  );
  return data;
}

export function useGetOwnerSharedPets(ownerId: string | undefined) {
  const normalizedOwnerId = ownerId ?? "";

  return useQuery({
    queryKey: queryKeys.ownerSharedPets.detail(normalizedOwnerId),
    queryFn: () => getOwnerSharedPets(normalizedOwnerId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: normalizedOwnerId !== "",
  });
}
