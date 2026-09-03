import { useMutation, useQueryClient } from "@tanstack/react-query";

import { queryKeys } from "@/lib/query-keys";
import type { OwnerGroupResponse, PetGroupResponse } from "@/types/generated/identitylink-responses";

import {
  createOwnerIdentityGroup,
  createPetIdentityGroup,
  getLinkedTreatmentHistory,
  unlinkOwnerIdentityMember,
  unlinkPetIdentityMember,
} from "../api/identity-links-api";

type OwnerMember = { clinic_id: number; owner_id: number };
type PetMember = { clinic_id: number; pet_id: number };

/** 作成直後に各メンバーの逆引きキャッシュへ結果を書き込み、refetch を待たず UI に反映する。 */
export function useCreateOwnerLink() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (members: OwnerMember[]) => createOwnerIdentityGroup(members),
    onSuccess: (group: OwnerGroupResponse, members) => {
      for (const member of members) {
        queryClient.setQueryData(
          queryKeys.identityLinks.ownerGroup(member.clinic_id, member.owner_id),
          group,
        );
      }
    },
  });
}

export function useUnlinkOwnerMember() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ groupId, member }: { groupId: number; member: OwnerMember }) =>
      unlinkOwnerIdentityMember(groupId, member),
    onSuccess: (_result, { member }) => {
      queryClient.setQueryData(
        queryKeys.identityLinks.ownerGroup(member.clinic_id, member.owner_id),
        null,
      );
    },
  });
}

export function useCreatePetLink() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ ownerGroupId, members }: { ownerGroupId: number; members: PetMember[] }) =>
      createPetIdentityGroup(ownerGroupId, members),
    onSuccess: (group: PetGroupResponse, { members }) => {
      for (const member of members) {
        queryClient.setQueryData(
          queryKeys.identityLinks.petGroup(member.clinic_id, member.pet_id),
          group,
        );
      }
    },
  });
}

export function useUnlinkPetMember() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ groupId, member }: { groupId: number; member: PetMember }) =>
      unlinkPetIdentityMember(groupId, member),
    onSuccess: (_result, { member }) => {
      queryClient.setQueryData(
        queryKeys.identityLinks.petGroup(member.clinic_id, member.pet_id),
        null,
      );
    },
  });
}

/** クリック起点の一回性フェッチ。キャッシュ不要のため useQuery ではなく useMutation で表現する。 */
export function useLinkedTreatmentHistory() {
  return useMutation({
    mutationFn: ({ clinicId, petId }: { clinicId: number; petId: number }) =>
      getLinkedTreatmentHistory(clinicId, petId, true, 1, 20),
  });
}
