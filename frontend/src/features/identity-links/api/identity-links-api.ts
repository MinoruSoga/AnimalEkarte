import { isAxiosError } from "axios";
import { axios } from "@/lib/axios";
import type {
  LinkedTreatmentHistoryResponse,
  OwnerGroupResponse,
  OwnerSearchItem,
  PetGroupResponse,
  PetSearchItem,
} from "@/types/generated/identitylink-responses";

export async function searchOwnersForLink(q: string, limit = 20): Promise<OwnerSearchItem[]> {
  const { data } = await axios.get<{ items: OwnerSearchItem[] }>(
    "/v1/identity-links/owners/search",
    { params: { q, limit } },
  );
  return data.items ?? [];
}

export async function searchPetsForLink(q: string, limit = 20): Promise<PetSearchItem[]> {
  const { data } = await axios.get<{ items: PetSearchItem[] }>("/v1/identity-links/pets/search", {
    params: { q, limit },
  });
  return data.items ?? [];
}

/** Reverse lookup: 404 (no visible group) → null; 403/other errors rethrow. */
export async function findOwnerIdentityGroupByMember(
  clinicId: number,
  ownerId: number,
): Promise<OwnerGroupResponse | null> {
  try {
    const { data } = await axios.get<OwnerGroupResponse>(
      `/v1/identity-links/owners/${clinicId}/${ownerId}/group`,
    );
    return data;
  } catch (err: unknown) {
    if (isAxiosError(err) && err.response?.status === 404) {
      return null;
    }
    throw err;
  }
}

/** Reverse lookup: 404 (no visible group) → null; 403/other errors rethrow. */
export async function findPetIdentityGroupByMember(
  clinicId: number,
  petId: number,
): Promise<PetGroupResponse | null> {
  try {
    const { data } = await axios.get<PetGroupResponse>(
      `/v1/identity-links/pets/${clinicId}/${petId}/group`,
    );
    return data;
  } catch (err: unknown) {
    if (isAxiosError(err) && err.response?.status === 404) {
      return null;
    }
    throw err;
  }
}

export async function createOwnerIdentityGroup(
  members: { clinic_id: number; owner_id: number }[],
): Promise<OwnerGroupResponse> {
  const { data } = await axios.post<OwnerGroupResponse>("/v1/identity-links/owner-groups", {
    members,
  });
  return data;
}

export async function unlinkOwnerIdentityMember(
  groupId: number,
  member: { clinic_id: number; owner_id: number },
): Promise<void> {
  await axios.delete(`/v1/identity-links/owner-groups/${groupId}/members`, { data: member });
}

export async function createPetIdentityGroup(
  ownerGroupId: number,
  members: { clinic_id: number; pet_id: number }[],
): Promise<PetGroupResponse> {
  const { data } = await axios.post<PetGroupResponse>("/v1/identity-links/pet-groups", {
    owner_group_id: ownerGroupId,
    members,
  });
  return data;
}

export async function unlinkPetIdentityMember(
  groupId: number,
  member: { clinic_id: number; pet_id: number },
): Promise<void> {
  await axios.delete(`/v1/identity-links/pet-groups/${groupId}/members`, { data: member });
}

export async function getLinkedTreatmentHistory(
  clinicId: number,
  petId: number,
  includeLinked: boolean,
  page = 1,
  limit = 20,
): Promise<LinkedTreatmentHistoryResponse> {
  const { data } = await axios.get<LinkedTreatmentHistoryResponse>(
    `/v1/identity-links/pets/${clinicId}/${petId}/treatment-history`,
    { params: { include_linked: includeLinked ? "true" : "false", page, limit } },
  );
  return data;
}
