import { useCallback, useLayoutEffect, useRef, useState, type TransitionStartFunction } from "react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import type { UseMutationResult } from "@tanstack/react-query";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

interface MasterEntity {
  id: string;
}

interface MasterSaveCrud<T extends MasterEntity> {
  editTarget: T | "new" | null;
  setEditTarget: (target: T | "new" | null) => void;
  startSaveTransition: TransitionStartFunction;
}

interface MasterSavePermissions {
  canCreate: boolean;
  canEdit: boolean;
}

interface UseMasterSaveOptions<T extends MasterEntity, TForm, TCreate, TUpdate> {
  crud: MasterSaveCrud<T>;
  createMutation: UseMutationResult<T, Error, TCreate>;
  updateMutation: UseMutationResult<T, Error, { id: string; req: TUpdate }>;
  /** Return error message string to show via toast.error. Return null if valid. */
  validate: (data: TForm) => string | null;
  /** FormData → CreateRequest conversion */
  toCreateRequest: (data: TForm) => TCreate;
  /** FormData → UpdateRequest conversion */
  toUpdateRequest: (data: TForm) => TUpdate;
  /** Optional: post-save hook for additional operations (e.g., setting related data) */
  onSuccess?: (saved: T, formData: TForm) => Promise<void> | void;
  /** When provided, engage action-specific permission guards at the mutation boundary. */
  permissions?: MasterSavePermissions;
}

// ─────────────────────────────────────────────────
// Hook
// ─────────────────────────────────────────────────

export function useMasterSave<T extends MasterEntity, TForm, TCreate, TUpdate>({
  crud,
  createMutation,
  updateMutation,
  validate,
  toCreateRequest,
  toUpdateRequest,
  onSuccess,
  permissions,
}: UseMasterSaveOptions<T, TForm, TCreate, TUpdate>) {
  // rerender-dependencies: extract primitives from crud.editTarget object
  const editTargetId = crud.editTarget !== null && crud.editTarget !== "new" ? crud.editTarget.id : null;
  // rerender-dependencies: destructure methods to avoid object reference instability
  // NOTE: use setEditTarget(null) directly on save success to bypass confirmDiscard()
  // crudHandleClose calls confirmDiscard() which shows window.confirm when isDirtyRef is still stale
  const crudSetEditTarget = crud.setEditTarget;
  const crudStartSave = crud.startSaveTransition;
  const permissionsEngaged = permissions !== undefined;
  const canCreate = permissions?.canCreate;
  const canEdit = permissions?.canEdit;
  const permissionsRef = useRef<MasterSavePermissions | undefined>(permissions);
  useLayoutEffect(() => {
    permissionsRef.current = permissionsEngaged
      ? { canCreate: canCreate === true, canEdit: canEdit === true }
      : undefined;
  }, [permissionsEngaged, canCreate, canEdit]);

  const [validationError, setValidationError] = useState<string | null>(null);

  const handleSave = useCallback(
    (data: TForm) => {
      const error = validate(data);
      if (error) {
        setValidationError(error);
        toast.error(error);
        return;
      }
      setValidationError(null);

      const currentPermissions = permissionsRef.current;
      if (currentPermissions !== undefined) {
        const isAllowed = editTargetId !== null
          ? currentPermissions.canEdit === true
          : currentPermissions.canCreate === true;
        if (!isAllowed) return;
      }

      crudStartSave(() => {
        if (editTargetId !== null) {
          updateMutation.mutate(
            { id: editTargetId, req: toUpdateRequest(data) },
            {
              onSuccess: async (saved) => {
                try {
                  if (onSuccess) {
                    await onSuccess(saved, data);
                  }
                  toast.success("更新しました");
                  crudSetEditTarget(null);
                } catch (error) {
                  handleApiError(error, "保存");
                }
              },
              onError: (error) => handleApiError(error, "更新"),
            },
          );
        } else {
          createMutation.mutate(toCreateRequest(data), {
            onSuccess: async (saved) => {
              try {
                if (onSuccess) {
                  await onSuccess(saved, data);
                }
                toast.success("登録しました");
                crudSetEditTarget(null);
              } catch (error) {
                handleApiError(error, "保存");
              }
            },
            onError: (error) => handleApiError(error, "登録"),
          });
        }
      });
    },
    [editTargetId, crudSetEditTarget, crudStartSave, createMutation, updateMutation, validate, toCreateRequest, toUpdateRequest, onSuccess],
  );

  return { handleSave, validationError };
}
