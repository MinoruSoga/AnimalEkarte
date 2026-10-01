import { useActionState, useLayoutEffect, useMemo, useRef, useState } from "react";

import { extractApiErrorMessage, handleApiError } from "@/lib/handle-api-error";
import type { UsePermissionResult } from "@/hooks/use-permission";
import type { PetChronicCondition } from "@/types/generated/models";

import {
  useCreatePetChronicCondition,
  useDeletePetChronicCondition,
  useGetPetChronicConditions,
  useUpdatePetChronicCondition,
} from "../api/pet-chronic-conditions";

export interface ChronicConditionDraft {
  conditionCode: string;
  conditionName: string;
  /** DatePicker 由来の strict YYYY-MM-DD（空文字 = 未選択） */
  diagnosedAt: string;
  notes: string;
  isActive: boolean;
}

export interface ChronicConditionActionState {
  kind: "idle" | "success" | "error";
  message: string;
}

const EMPTY_DRAFT: ChronicConditionDraft = {
  conditionCode: "",
  conditionName: "",
  diagnosedAt: "",
  notes: "",
  isActive: true,
};

const INITIAL_ACTION_STATE: ChronicConditionActionState = { kind: "idle", message: "" };

const DIAGNOSED_AT_PATTERN = /^\d{4}-\d{2}-\d{2}$/;
const CONDITION_NAME_MAX_LENGTH = 255;

function validateDraft(draft: ChronicConditionDraft): string | null {
  const nameLength = Array.from(draft.conditionName.trim()).length;
  if (nameLength < 1) {
    return "疾患名を入力してください。";
  }
  if (nameLength > CONDITION_NAME_MAX_LENGTH) {
    return "疾患名は255文字以内で入力してください。";
  }
  if (draft.conditionCode.trim() === "") {
    return "疾患コードを入力してください。";
  }
  if (!DIAGNOSED_AT_PATTERN.test(draft.diagnosedAt)) {
    return "診断日を選択してください。";
  }
  return null;
}

/** 空白のみ → null（BE 契約: notes は nullable） */
function toNotesPayload(notes: string): string | null {
  const trimmed = notes.trim();
  return trimmed === "" ? null : trimmed;
}

interface PermissionSnapshot {
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
  canEditFieldset: boolean;
}

export function usePetChronicConditionsSection(
  petId: string,
  canEdit: boolean,
  permissions: UsePermissionResult,
) {
  const conditionsQuery = useGetPetChronicConditions(petId);
  const createMutation = useCreatePetChronicCondition();
  const updateMutation = useUpdatePetChronicCondition();
  const deleteMutation = useDeletePetChronicCondition();

  // FE12-02: commit 直後にも発火し得る取得済み action は mutation 直前に最新権限を再検査する
  const { canCreate, canEdit: canEditPermission, canDelete } = permissions;
  const permissionsRef = useRef<PermissionSnapshot>({
    canCreate,
    canEdit: canEditPermission,
    canDelete,
    canEditFieldset: canEdit,
  });
  useLayoutEffect(() => {
    permissionsRef.current = {
      canCreate,
      canEdit: canEditPermission,
      canDelete,
      canEditFieldset: canEdit,
    };
  }, [canCreate, canEditPermission, canDelete, canEdit]);

  const isActionAllowed = (action: "canCreate" | "canEdit" | "canDelete") =>
    permissionsRef.current[action] === true && permissionsRef.current.canEditFieldset === true;

  const conditions = useMemo(() => conditionsQuery.data ?? [], [conditionsQuery.data]);

  const [addDraft, setAddDraft] = useState<ChronicConditionDraft>(EMPTY_DRAFT);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editDraft, setEditDraft] = useState<ChronicConditionDraft>(EMPTY_DRAFT);
  const [deletingId, setDeletingId] = useState<number | null>(null);

  const [addState, addAction, isAddPending] = useActionState<ChronicConditionActionState, FormData>(
    async () => {
      if (!isActionAllowed("canCreate")) {
        return { kind: "error", message: "慢性疾患を登録する権限がありません。" };
      }
      const validationMessage = validateDraft(addDraft);
      if (validationMessage !== null) {
        return { kind: "error", message: validationMessage };
      }
      try {
        await createMutation.mutateAsync({
          petId,
          request: {
            condition_code: addDraft.conditionCode.trim(),
            condition_name: addDraft.conditionName.trim(),
            diagnosed_at: addDraft.diagnosedAt,
            notes: toNotesPayload(addDraft.notes),
            is_active: true,
          },
        });
        setAddDraft(EMPTY_DRAFT);
        return { kind: "success", message: "慢性疾患を登録しました" };
      } catch (error: unknown) {
        handleApiError(error, "慢性疾患の登録");
        return { kind: "error", message: extractApiErrorMessage(error, "慢性疾患の登録") };
      }
    },
    INITIAL_ACTION_STATE,
  );

  const [editState, editAction, isEditPending] = useActionState<
    ChronicConditionActionState,
    FormData
  >(async () => {
    const conditionId = editingId;
    if (conditionId === null) {
      return { kind: "error", message: "更新対象の慢性疾患が選択されていません。" };
    }
    if (!isActionAllowed("canEdit")) {
      return { kind: "error", message: "慢性疾患を更新する権限がありません。" };
    }
    const validationMessage = validateDraft(editDraft);
    if (validationMessage !== null) {
      return { kind: "error", message: validationMessage };
    }
    try {
      await updateMutation.mutateAsync({
        petId,
        conditionId,
        request: {
          condition_code: editDraft.conditionCode.trim(),
          condition_name: editDraft.conditionName.trim(),
          diagnosed_at: editDraft.diagnosedAt,
          notes: toNotesPayload(editDraft.notes),
          // PATCH は nil フィールドを省略するため、無効化は明示的に false を送る
          is_active: editDraft.isActive,
        },
      });
      setEditingId(null);
      return { kind: "success", message: "慢性疾患を更新しました" };
    } catch (error: unknown) {
      handleApiError(error, "慢性疾患の更新");
      return { kind: "error", message: extractApiErrorMessage(error, "慢性疾患の更新") };
    }
  }, INITIAL_ACTION_STATE);

  const startEdit = (condition: PetChronicCondition) => {
    setEditingId(condition.id);
    setEditDraft({
      conditionCode: condition.condition_code,
      conditionName: condition.condition_name,
      diagnosedAt: condition.diagnosed_at,
      notes: condition.notes ?? "",
      isActive: condition.is_active,
    });
  };

  const cancelEdit = () => {
    setEditingId(null);
  };

  const handleDeleteCondition = (conditionId: number) => {
    if (deletingId !== null || !isActionAllowed("canDelete")) {
      return;
    }
    setDeletingId(conditionId);
    deleteMutation
      .mutateAsync({ petId, conditionId })
      .catch((error: unknown) => {
        handleApiError(error, "慢性疾患の削除");
      })
      .finally(() => {
        setDeletingId(null);
      });
  };

  const isBusy = isAddPending || isEditPending || deletingId !== null;
  const isAddDirty =
    addDraft.conditionCode !== "" ||
    addDraft.conditionName !== "" ||
    addDraft.diagnosedAt !== "" ||
    addDraft.notes !== "";
  const loadError = conditionsQuery.error;

  return {
    conditionsQuery,
    conditions,
    addDraft,
    setAddDraft,
    isAddDirty,
    addState,
    addAction,
    isAddPending,
    editingId,
    editDraft,
    setEditDraft,
    startEdit,
    cancelEdit,
    editState,
    editAction,
    isEditPending,
    deletingId,
    handleDeleteCondition,
    isBusy,
    loadError,
    /** fieldset の canEdit と action 別権限の AND。UI disabled 用 */
    canCreateCondition: canEdit && canCreate,
    canEditCondition: canEdit && canEditPermission,
    canDeleteCondition: canEdit && canDelete,
  };
}
