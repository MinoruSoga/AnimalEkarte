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

const PER_PAGE = 100; // backend の parsePagination 上限

export interface OwnersLoaderData {
  pets: Pet[];
}

export const ownersLoader = async (): Promise<OwnersLoaderData> => {
  // page 1 で総件数を確認し、残りのページを並列フェッチ
  const { data: firstPage } = await axios.get<PetsResponse>("/v1/pets", {
    params: { page: 1, limit: PER_PAGE },
  });

  const totalPages = Math.ceil(firstPage.total / PER_PAGE);

  const remainingPages = await Promise.all(
    Array.from({ length: totalPages - 1 }, (_, i) =>
      axios.get<PetsResponse>("/v1/pets", {
        params: { page: i + 2, limit: PER_PAGE },
      }).then(r => r.data)
    )
  );

  const allPets: Pet[] = [firstPage, ...remainingPages]
    .flatMap(page => page.data.map(transformBackendPetToFrontend));

  return { pets: allPets };
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
