import { useQueries } from "@tanstack/react-query";

import { queryKeys } from "@/lib/query-keys";
import type { OwnerSearchItem, PetSearchItem } from "@/types/generated/identitylink-responses";

import {
  findOwnerIdentityGroupByMember,
  findPetIdentityGroupByMember,
} from "../api/identity-links-api";

export function ownerMemberKey(clinicId: number, ownerId: number): string {
  return `${clinicId}:${ownerId}`;
}

export function petMemberKey(clinicId: number, petId: number): string {
  return `${clinicId}:${petId}`;
}

interface OwnerGroupLookupResult {
  /** member key → 既存グループ id（選択中メンバーのみ。未所属/未解決は含まない） */
  groupIdsByMember: Record<string, number>;
  /** 選択中メンバーの中で最初に解決した既存グループ id。新規リンクの anchor に使う */
  sessionGroupId: number | null;
  /** 反映漏れを表示に出すための、いずれかの選択中メンバーの直近エラー */
  errorMessage: string | null;
}

interface PetGroupLookupResult extends OwnerGroupLookupResult {
  /** 最初に解決したペットグループが属する飼主グループ id（親グループ未確定時の補完用） */
  sessionOwnerGroupId: number | null;
}

/**
 * 選択中の飼主ごとに既存グループを逆引きする。react-query の queryKey がメンバー単位のため、
 * 選択解除で自動的に該当クエリが破棄され、旧来の「stillSelected」手動ガードが不要になる。
 */
export function useOwnerGroupLookup(selectedOwners: OwnerSearchItem[]): OwnerGroupLookupResult {
  const results = useQueries({
    queries: selectedOwners.map((owner) => ({
      queryKey: queryKeys.identityLinks.ownerGroup(owner.clinic_id, owner.owner_id),
      queryFn: () => findOwnerIdentityGroupByMember(owner.clinic_id, owner.owner_id),
    })),
  });

  const groupIdsByMember: Record<string, number> = {};
  let sessionGroupId: number | null = null;
  let errorMessage: string | null = null;

  selectedOwners.forEach((owner, index) => {
    const query = results[index];
    if (query?.data != null) {
      groupIdsByMember[ownerMemberKey(owner.clinic_id, owner.owner_id)] = query.data.id;
      if (sessionGroupId == null) sessionGroupId = query.data.id;
    }
    if (query?.isError) {
      errorMessage = "飼主グループ取得に失敗しました";
    }
  });

  return { groupIdsByMember, sessionGroupId, errorMessage };
}

export function usePetGroupLookup(selectedPets: PetSearchItem[]): PetGroupLookupResult {
  const results = useQueries({
    queries: selectedPets.map((pet) => ({
      queryKey: queryKeys.identityLinks.petGroup(pet.clinic_id, pet.pet_id),
      queryFn: () => findPetIdentityGroupByMember(pet.clinic_id, pet.pet_id),
    })),
  });

  const groupIdsByMember: Record<string, number> = {};
  let sessionGroupId: number | null = null;
  let sessionOwnerGroupId: number | null = null;
  let errorMessage: string | null = null;

  selectedPets.forEach((pet, index) => {
    const query = results[index];
    if (query?.data != null) {
      groupIdsByMember[petMemberKey(pet.clinic_id, pet.pet_id)] = query.data.id;
      if (sessionGroupId == null) sessionGroupId = query.data.id;
      if (sessionOwnerGroupId == null) sessionOwnerGroupId = query.data.owner_group_id;
    }
    if (query?.isError) {
      errorMessage = "ペットグループ取得に失敗しました";
    }
  });

  return { groupIdsByMember, sessionGroupId, sessionOwnerGroupId, errorMessage };
}
