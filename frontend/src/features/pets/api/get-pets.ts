import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Pet } from "@/types";
import { transformBackendPetToFrontend } from "./transforms";
import type { BackendPet } from "./types";

export const getPets = async (): Promise<Pet[]> => {
  const { data } = await axios.get<BackendPet[]>("/v1/pets");
  return data.map(transformBackendPetToFrontend);
};

export const useGetPets = (ownerId?: string) => {
  return useQuery({
    queryKey: ["pets"],
    queryFn: getPets,
    select: (pets) => {
      if (!ownerId) return pets;
      return pets.filter((pet) => pet.ownerId === ownerId);
    },
  });
};
