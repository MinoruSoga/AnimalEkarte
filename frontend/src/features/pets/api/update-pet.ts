import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Pet } from "@/types";
import { handleApiError } from "@/lib/handle-api-error";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet as BackendPet } from "@/types/generated/models";
import type { UpdatePetRequest } from "@/types/pet";

export const updatePet = async (
  id: string,
  req: UpdatePetRequest
): Promise<Pet> => {
  const { data } = await axios.patch<BackendPet>(`/v1/pets/${id}`, req);
  return transformBackendPetToFrontend(data);
};

export const useUpdatePet = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdatePetRequest }) =>
      updatePet(id, req),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ["pets"] });
      queryClient.invalidateQueries({ queryKey: ["pet", id] });
    },
    onError: (error) => handleApiError(error, "ペット更新"),
  });
};
