import { getOwner } from "./api/get-owner";
import { axios } from "@/lib/axios";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { Owner } from "@/types/owner";
import type { Owner as BackendOwner } from "@/types/generated/models";

interface OwnersResponse {
  data: BackendOwner[];
  total: number;
  page: number;
  limit: number;
}

const PER_PAGE = 100; // backend の parsePagination 上限

export interface OwnersLoaderData {
  pets: Pet[];
}

/**
 * 飼主一覧ローダー — /v1/owners（ペットネスト）から全飼主を取得し
 * ペット行にフラット化する。ペット0件の飼主も1行（空ペット）として含める。
 */
export const ownersLoader = async (): Promise<OwnersLoaderData> => {
  try {
    // page 1 で総件数を確認し、残りのページを並列フェッチ
    const { data: firstPage } = await axios.get<OwnersResponse>("/v1/owners", {
      params: { page: 1, limit: PER_PAGE },
    });

    const totalPages = Math.ceil(firstPage.total / PER_PAGE);

    const remainingPages = await Promise.all(
      Array.from({ length: totalPages - 1 }, (_, i) =>
        axios.get<OwnersResponse>("/v1/owners", {
          params: { page: i + 2, limit: PER_PAGE },
        }).then(r => r.data)
      )
    );

    const allOwners = [firstPage, ...remainingPages].flatMap(page => page.data);

    // 各飼主のペットをフラット化。ペットなし飼主は owner 情報のみの空ペット行を1件追加。
    const allPets: Pet[] = allOwners.flatMap(owner => {
      const pets = owner.pets ?? [];
      if (pets.length > 0) {
        return pets.map(pet => transformBackendPetToFrontend({ ...pet, owner }));
      }
      // ペットなし飼主: owner 情報のみを持つ空ペット行を直接構築
      const ownerAddress = [owner.address1, owner.address2].filter(Boolean).join(" ") || undefined;
      const emptyPet: Pet = {
        id: `owner-${owner.id}`,
        ownerId: String(owner.id),
        ownerNumber: owner.id,
        ownerName: owner.owner_name,
        ownerNameKana: owner.owner_name_kana || undefined,
        address: ownerAddress,
        phone: owner.phone,
        petNumber: undefined,
        name: "",
        petNameKana: undefined,
        species: "",
        animalSpeciesId: undefined,
        breed: undefined,
        color: undefined,
        gender: undefined,
        status: undefined,
        birthDate: undefined,
        neuteredDate: undefined,
        weight: undefined,
        food: undefined,
        environment: undefined,
        acquisitionType: undefined,
        dangerLevel: undefined,
        lastVisit: undefined,
        insuranceId: undefined,
        insuranceName: undefined,
        insuranceDetails: undefined,
        remarks: undefined,
      };
      return [emptyPet];
    });

    return { pets: allPets };
  } catch {
    throw new Response("飼主一覧の取得に失敗しました", { status: 500 });
  }
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
