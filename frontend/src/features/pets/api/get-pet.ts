import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Pet } from "@/types";
import { transformBackendPetToFrontend } from "./transforms";
import type { BackendPet } from "./types";

export const getPet = async (id: string): Promise<Pet> => {
  const { data } = await axios.get<BackendPet>(`/v1/pets/${id}`);
  return transformBackendPetToFrontend(data);
};

export const useGetPet = (petId: string) => {
  return useQuery({
    queryKey: ["pet", petId],
    queryFn: () => getPet(petId),
    enabled: !!petId,
  });
};
