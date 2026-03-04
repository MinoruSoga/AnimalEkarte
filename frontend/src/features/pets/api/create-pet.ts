import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Pet } from "@/types";
import { transformBackendPetToFrontend } from "./transforms";
import type { BackendPet, CreatePetRequest } from "./types";

export const createPet = async (req: CreatePetRequest): Promise<Pet> => {
  const { data } = await axios.post<BackendPet>("/v1/pets", req);
  return transformBackendPetToFrontend(data);
};

export const useCreatePet = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createPet,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["pets"] });
    },
  });
};
