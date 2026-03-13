import { getOwner } from "./api";
import { axios } from "@/lib/axios";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { Owner } from "@/types/owner";
import type { Pet as BackendPet } from "@/types/generated/models";

interface PetsResponse {
  data: BackendPet[];
  total: number;
  page: number;
  limit: number;
}

export interface OwnersLoaderData {
  pets: Pet[];
}

export const ownersLoader = async (): Promise<OwnersLoaderData> => {
  const { data } = await axios.get<PetsResponse>("/v1/pets", {
    params: { limit: 500 },
  });
  const pets = data.data.map(transformBackendPetToFrontend);
  return { pets };
};

export interface OwnerLoaderData {
  owner: Owner;
}

export const ownerLoader = async ({ params }: { params: Record<string, string | undefined> }): Promise<OwnerLoaderData> => {
  const id = params.id;
  if (!id) {
    throw new Response("Owner ID is required", { status: 400 });
  }
  const owner = await getOwner(id);
  return { owner };
};
