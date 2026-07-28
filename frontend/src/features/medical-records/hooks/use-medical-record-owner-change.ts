import { useCallback, useLayoutEffect, useRef, useState } from "react";
import type { TransitionStartFunction } from "react";
import { toast } from "sonner";

import { handleApiError } from "@/lib/handle-api-error";
import { usePermission } from "@/hooks/use-permission";

import type { MedicalRecord } from "../api/transforms";
import type { UpdateMedicalRecordRequest } from "../api/types";

interface OwnerSnapshot {
  discountRate?: number;
  membershipType?: string;
}

interface OwnerChangeRequest {
  id: string;
  name: string;
  discountRate: number;
  membershipType: string;
}

interface PendingOwnerChange {
  id: string;
  name: string;
}

interface UseMedicalRecordOwnerChangeArgs {
  owner?: OwnerSnapshot | null;
  recordId?: string;
  existingRecord?: MedicalRecord;
  updateMutation: {
    mutateAsync: (variables: { id: string; req: UpdateMedicalRecordRequest }) => Promise<unknown>;
  };
  startSaveTransition: TransitionStartFunction;
  canEdit?: boolean;
  isSelectedPetDeceased: boolean;
}

export function useMedicalRecordOwnerChange({
  owner,
  recordId,
  existingRecord,
  updateMutation,
  startSaveTransition,
  canEdit: canEditOverride,
  isSelectedPetDeceased,
}: UseMedicalRecordOwnerChangeArgs) {
  const { canEdit: permissionCanEdit } = usePermission("medical-records");
  const canEdit = canEditOverride ?? permissionCanEdit;
  const canEditRef = useRef(canEdit);
  const isSelectedPetDeceasedRef = useRef(isSelectedPetDeceased);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);
  useLayoutEffect(() => {
    isSelectedPetDeceasedRef.current = isSelectedPetDeceased;
  }, [isSelectedPetDeceased]);
  const isMutationAllowed = useCallback(
    () => canEditRef.current === true && !isSelectedPetDeceasedRef.current,
    [],
  );
  const [pendingOwnerChange, setPendingOwnerChange] = useState<PendingOwnerChange | null>(null);
  const recordVersion = existingRecord?.version;

  const updateOwner = useCallback(
    (newOwner: PendingOwnerChange) => {
      if (!recordId || !isMutationAllowed()) return;
      startSaveTransition(async () => {
        if (!isMutationAllowed()) return;
        try {
          await updateMutation.mutateAsync({
            id: recordId,
            req: {
              owner_id: Number(newOwner.id),
              version: recordVersion,
            } as UpdateMedicalRecordRequest,
          });
          toast.success(`飼主を ${newOwner.name} に変更しました`);
        } catch (error) {
          handleApiError(error, "飼主変更");
        }
      });
    },
    [recordId, recordVersion, startSaveTransition, updateMutation, isMutationAllowed],
  );

  const requestOwnerChange = useCallback(
    (newOwner: OwnerChangeRequest) => {
      if (!isMutationAllowed()) return;
      const needsConfirm =
        !owner ||
        owner.discountRate !== newOwner.discountRate ||
        owner.membershipType !== newOwner.membershipType;
      if (needsConfirm) {
        setPendingOwnerChange({ id: newOwner.id, name: newOwner.name });
        return;
      }
      updateOwner(newOwner);
    },
    [owner, updateOwner, isMutationAllowed],
  );

  const confirmOwnerChange = useCallback(() => {
    if (!pendingOwnerChange) return;
    const newOwner = pendingOwnerChange;
    setPendingOwnerChange(null);
    updateOwner(newOwner);
  }, [pendingOwnerChange, updateOwner]);

  const cancelOwnerChange = useCallback(() => setPendingOwnerChange(null), []);

  return {
    pendingOwnerChange,
    requestOwnerChange,
    confirmOwnerChange,
    cancelOwnerChange,
  };
}
