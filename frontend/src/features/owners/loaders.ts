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
      pets.push({
        id: `owner-${owner.id}`,
        ownerId: owner.id,
        ownerNumber: owner.ownerNumber,
        ownerName: owner.name,
        name: "-",
        species: "-",
        petNumber: "-",
        status: undefined,
      } as Pet);
    } else {
      owner.pets.forEach((pet) => {
        pets.push({
          ...pet,
          ownerId: owner.id,
          ownerNumber: owner.ownerNumber,
          ownerName: owner.name,
          name: pet.name || "-",
          species: pet.species || "-",
        } as unknown as Pet);
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
