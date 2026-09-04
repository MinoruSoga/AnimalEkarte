import { useCallback, useState, type RefObject } from "react";
import { toast } from "sonner";

import { handleApiError } from "@/lib/handle-api-error";
import type { PetMutations } from "@/types/pet";
import type { PetFormData } from "../types";

interface PendingOwnerChange {
  id: string;
  name: string;
}

interface UseOwnerPetChangeConfirmArgs {
  editingPetRef: RefObject<PetFormData | null>;
  canEditRef: RefObject<boolean>;
  petMutations: PetMutations | undefined;
  ownerDiscountRate: number | null | undefined;
  ownerMembershipType: string | undefined;
  setPetModalOpen: (open: boolean) => void;
}

// BUG-373: 飼主変更 — discount_rate/membership_type が異なる時のみ確認モーダル
export function useOwnerPetChangeConfirm({
  editingPetRef,
  canEditRef,
  petMutations,
  ownerDiscountRate,
  ownerMembershipType,
  setPetModalOpen,
}: UseOwnerPetChangeConfirmArgs) {
  const [pendingOwnerChange, setPendingOwnerChange] = useState<PendingOwnerChange | null>(null);

  const handlePetChangeOwner = useCallback(
    (newOwner: { id: string; name: string; discountRate: number; membershipType: string }) => {
      const currentEditingPet = editingPetRef.current;
      if (
        canEditRef.current !== true ||
        !currentEditingPet?.id ||
        currentEditingPet.status === "死亡" ||
        !petMutations
      ) {
        return;
      }
      const needsConfirm =
        (ownerDiscountRate ?? 0) !== newOwner.discountRate ||
        ownerMembershipType !== newOwner.membershipType;
      if (needsConfirm) {
        setPendingOwnerChange({ id: newOwner.id, name: newOwner.name });
      } else {
        if (canEditRef.current !== true) return;
        petMutations.updatePetMutate(
          { id: currentEditingPet.id, req: { owner_id: Number(newOwner.id) } },
          {
            onSuccess: () => {
              toast.success(`飼主を ${newOwner.name} に変更しました`);
              setPetModalOpen(false);
            },
            onError: (error) => {
              handleApiError(error, "飼主変更");
            },
          },
        );
      }
    },
    [
      petMutations,
      ownerDiscountRate,
      ownerMembershipType,
      setPetModalOpen,
      editingPetRef,
      canEditRef,
    ],
  );

  const handleConfirmOwnerChange = useCallback(() => {
    const currentEditingPet = editingPetRef.current;
    if (
      canEditRef.current !== true ||
      !pendingOwnerChange ||
      !currentEditingPet?.id ||
      currentEditingPet.status === "死亡" ||
      !petMutations
    ) {
      return;
    }
    const newOwner = pendingOwnerChange;
    if (canEditRef.current !== true) return;
    petMutations.updatePetMutate(
      { id: currentEditingPet.id, req: { owner_id: Number(newOwner.id) } },
      {
        onSuccess: () => {
          toast.success(`飼主を ${newOwner.name} に変更しました`);
          setPendingOwnerChange(null);
          setPetModalOpen(false);
        },
        onError: (error) => {
          handleApiError(error, "飼主変更");
          setPendingOwnerChange(null);
        },
      },
    );
  }, [pendingOwnerChange, petMutations, setPetModalOpen, editingPetRef, canEditRef]);

  return {
    pendingOwnerChange,
    setPendingOwnerChange,
    handlePetChangeOwner,
    handleConfirmOwnerChange,
  };
}
