import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Pet } from "@/types";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { PetResponse } from "@/types/generated/pet-responses";
import type { UpdatePetRequest } from "@/types/pet";

export const updatePet = async (id: string, req: UpdatePetRequest): Promise<Pet> => {
  const { data } = await axios.patch<PetResponse>(`/v1/pets/${id}`, req);
  return transformBackendPetToFrontend(data);
};

export const useUpdatePet = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdatePetRequest }) => updatePet(id, req),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pets.list() });
      queryClient.invalidateQueries({ queryKey: queryKeys.pets.detail(id) });
    },
    onError: (error) => handleApiError(error, "ペット更新"),
  });
};
