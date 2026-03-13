import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Pet } from "@/types";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet as BackendPet } from "@/types/generated/models";

interface PetsListResponse {
  data: BackendPet[];
  total: number;
  page: number;
  limit: number;
}

export const getPets = async (ownerId?: string): Promise<Pet[]> => {
  const params = ownerId ? { owner_id: ownerId } : {};
  const { data } = await axios.get<PetsListResponse>("/v1/pets", { params });
  return data.data.map(transformBackendPetToFrontend);
};

export const useGetPets = (ownerId?: string) => {
  return useQuery({
    queryKey: ownerId ? ["pets", { ownerId }] : ["pets"],
    queryFn: () => getPets(ownerId),
  });
};
