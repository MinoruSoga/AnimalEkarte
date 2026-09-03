import { useState, useTransition } from "react";

import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { extractApiErrorMessage } from "@/lib/handle-api-error";
import type { OwnerSearchItem, PetSearchItem } from "@/types/generated/identitylink-responses";

import {
  ownerMemberKey,
  petMemberKey,
  useOwnerGroupLookup,
  usePetGroupLookup,
} from "./use-identity-group-lookup";
import { useOwnerSearchQuery, usePetSearchQuery } from "./use-identity-link-search";
import {
  useCreateOwnerLink,
  useCreatePetLink,
  useLinkedTreatmentHistory,
  useUnlinkOwnerMember,
  useUnlinkPetMember,
} from "./use-identity-link-mutations";

const SEARCH_DEBOUNCE_MS = 300;

function sameOwner(a: OwnerSearchItem, b: { clinic_id: number; owner_id: number }): boolean {
  return a.clinic_id === b.clinic_id && a.owner_id === b.owner_id;
}

function samePet(a: PetSearchItem, b: { clinic_id: number; pet_id: number }): boolean {
  return a.clinic_id === b.clinic_id && a.pet_id === b.pet_id;
}

/**
 * Phase 1 manual identity-link ワークベンチ。検索・逆引き・作成/解除を
 * react-query（useQuery/useQueries/useMutation）に委譲し、view/edit の gate は
 * 呼び出し元（IdentityLinksPage）が担う。
 */
export function useIdentityLinksWorkbench(canEdit: boolean) {
  const [ownerQuery, setOwnerQuery] = useState("");
  const [petQuery, setPetQuery] = useState("");
  const debouncedOwnerQuery = useDebouncedValue(ownerQuery, SEARCH_DEBOUNCE_MS).trim();
  const debouncedPetQuery = useDebouncedValue(petQuery, SEARCH_DEBOUNCE_MS).trim();

  const ownerSearch = useOwnerSearchQuery(debouncedOwnerQuery);
  const petSearch = usePetSearchQuery(debouncedPetQuery);

  const [selectedOwners, setSelectedOwners] = useState<OwnerSearchItem[]>([]);
  const [selectedPets, setSelectedPets] = useState<PetSearchItem[]>([]);

  const ownerGroupLookup = useOwnerGroupLookup(selectedOwners);
  const petGroupLookup = usePetGroupLookup(selectedPets);

  // ペットリンクの親飼主グループが未確定なら、ペット逆引きで判明した owner_group_id を補完に使う。
  // 飼主側セッションを優先し、異なる id で上書きしない。
  const effectiveOwnerGroupId = ownerGroupLookup.sessionGroupId ?? petGroupLookup.sessionOwnerGroupId;

  const [historyText, setHistoryText] = useState("");
  const [actionError, setActionError] = useState<string | null>(null);
  const [pending, startTransition] = useTransition();

  const createOwnerLink = useCreateOwnerLink();
  const unlinkOwnerMember = useUnlinkOwnerMember();
  const createPetLink = useCreatePetLink();
  const unlinkPetMember = useUnlinkPetMember();
  const linkedHistory = useLinkedTreatmentHistory();

  const errorMessage = actionError ?? ownerGroupLookup.errorMessage ?? petGroupLookup.errorMessage;

  const resolveOwnerGroupId = (item: OwnerSearchItem): number | null =>
    ownerGroupLookup.groupIdsByMember[ownerMemberKey(item.clinic_id, item.owner_id)] ?? null;

  const resolvePetGroupId = (item: PetSearchItem): number | null =>
    petGroupLookup.groupIdsByMember[petMemberKey(item.clinic_id, item.pet_id)] ?? null;

  const toggleOwner = (item: OwnerSearchItem) => {
    setSelectedOwners((prev) =>
      prev.some((p) => sameOwner(p, item))
        ? prev.filter((p) => !sameOwner(p, item))
        : [...prev, item],
    );
  };

  const togglePet = (item: PetSearchItem) => {
    setSelectedPets((prev) =>
      prev.some((p) => samePet(p, item)) ? prev.filter((p) => !samePet(p, item)) : [...prev, item],
    );
  };

  const onLinkOwners = () => {
    if (!canEdit || selectedOwners.length < 2) return;
    setActionError(null);
    startTransition(async () => {
      try {
        await createOwnerLink.mutateAsync(
          selectedOwners.map((o) => ({ clinic_id: o.clinic_id, owner_id: o.owner_id })),
        );
      } catch (e: unknown) {
        setActionError(extractApiErrorMessage(e, "飼主リンク"));
      }
    });
  };

  const onUnlinkOwner = (item: OwnerSearchItem) => {
    const groupId = resolveOwnerGroupId(item);
    if (!canEdit || groupId == null) return;
    setActionError(null);
    startTransition(async () => {
      try {
        await unlinkOwnerMember.mutateAsync({
          groupId,
          member: { clinic_id: item.clinic_id, owner_id: item.owner_id },
        });
        setSelectedOwners((prev) => prev.filter((p) => !sameOwner(p, item)));
      } catch (e: unknown) {
        setActionError(extractApiErrorMessage(e, "飼主の連携解除"));
      }
    });
  };

  const onLinkPets = () => {
    if (!canEdit || effectiveOwnerGroupId == null || selectedPets.length < 2) return;
    setActionError(null);
    startTransition(async () => {
      try {
        await createPetLink.mutateAsync({
          ownerGroupId: effectiveOwnerGroupId,
          members: selectedPets.map((p) => ({ clinic_id: p.clinic_id, pet_id: p.pet_id })),
        });
      } catch (e: unknown) {
        setActionError(extractApiErrorMessage(e, "ペットリンク"));
      }
    });
  };

  const onUnlinkPet = (item: PetSearchItem) => {
    const groupId = resolvePetGroupId(item);
    if (!canEdit || groupId == null) return;
    setActionError(null);
    startTransition(async () => {
      try {
        await unlinkPetMember.mutateAsync({
          groupId,
          member: { clinic_id: item.clinic_id, pet_id: item.pet_id },
        });
        setSelectedPets((prev) => prev.filter((p) => !samePet(p, item)));
      } catch (e: unknown) {
        setActionError(extractApiErrorMessage(e, "ペットの連携解除"));
      }
    });
  };

  const onLoadHistory = (item: PetSearchItem) => {
    setActionError(null);
    startTransition(async () => {
      try {
        const hist = await linkedHistory.mutateAsync({
          clinicId: item.clinic_id,
          petId: item.pet_id,
        });
        setHistoryText(
          hist.items
            .map(
              (h) =>
                `[医院 ${h.clinic_id}/ペット ${h.pet_id}] ${h.record_date} ${h.record_no}: ${h.content}`,
            )
            .join("\n") || "（履歴なし）",
        );
      } catch (e: unknown) {
        setActionError(extractApiErrorMessage(e, "履歴取得"));
      }
    });
  };

  return {
    ownerQuery,
    setOwnerQuery,
    petQuery,
    setPetQuery,
    ownerHits: ownerSearch.data ?? [],
    petHits: petSearch.data ?? [],
    selectedOwners,
    selectedPets,
    // 飼主リンク欄の表示は「解決済みグループ」を飼主/ペットどちらの逆引き経由でも同じ値で見せる
    // （元実装は単一 state でこの合流を素で持っていたため、UI 上の見え方を変えない）。
    ownerGroupId: effectiveOwnerGroupId,
    petGroupId: petGroupLookup.sessionGroupId,
    canLinkPets: effectiveOwnerGroupId != null,
    historyText,
    errorMessage,
    pending,
    toggleOwner,
    togglePet,
    resolveOwnerGroupId,
    resolvePetGroupId,
    onLinkOwners,
    onUnlinkOwner,
    onLinkPets,
    onUnlinkPet,
    onLoadHistory,
  };
}
