import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Pet } from "@/types";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { PetResponse } from "@/types/generated/pet-responses";
import type { CreatePetRequest } from "@/types/pet";

export const createPet = async (req: CreatePetRequest): Promise<Pet> => {
  const { data } = await axios.post<PetResponse>("/v1/pets", req);
  return transformBackendPetToFrontend(data);
};

export const useCreatePet = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createPet,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pets.list() });
    },
    onError: (error) => handleApiError(error, "ペット登録"),
  });
};
