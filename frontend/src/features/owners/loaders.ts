import { getOwners, getOwner } from "./api";
import type { Owner } from "@/types/owner";
import type { Pet } from "@/types";

export interface OwnersLoaderData {
  owners: Owner[];
  pets: Pet[];
}

export const ownersLoader = async (): Promise<OwnersLoaderData> => {
  const owners = await getOwners();
  const pets: Pet[] = [];

  owners.forEach((owner) => {
    if (!owner.pets || owner.pets.length === 0) {
      const placeholderPet: Pet = {
        id: `owner-${owner.id}`,
        ownerId: owner.id,
        ownerName: owner.ownerName,
        name: "-",
        species: "-",
        petNumber: "-",
      };
      pets.push(placeholderPet);
    } else {
      owner.pets.forEach((pet) => {
        pets.push({
          ...pet,
          name: pet.name || "-",
          species: pet.species || "-",
        });
      });
    }
  });

  return { owners, pets };
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
