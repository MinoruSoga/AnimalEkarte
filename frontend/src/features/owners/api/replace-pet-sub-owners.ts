import { useMutation, useQueryClient } from "@tanstack/react-query";

import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";

interface ReplacePetSubOwnerItem {
  owner_id: number;
  relationship: string;
}

export interface ReplacePetSubOwnersRequest {
  version: number;
  sub_owners: ReplacePetSubOwnerItem[];
}

export interface ReplacePetSubOwnersVariables {
  petId: string;
  request: ReplacePetSubOwnersRequest;
}

export async function replacePetSubOwners(
  petId: string,
  request: ReplacePetSubOwnersRequest,
): Promise<void> {
  await axios.put<void>(`/v1/pets/${petId}/sub-owners`, request);
}

export function useReplacePetSubOwners() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ petId, request }: ReplacePetSubOwnersVariables) =>
      replacePetSubOwners(petId, request),
    onSuccess: async (_data, { petId }) => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: queryKeys.pets.detail(petId),
        }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.petSubOwners.detail(petId),
        }),
        queryClient.invalidateQueries({
          queryKey: queryKeys.petSubOwners.metadata(petId),
        }),
      ]);
    },
  });
}
