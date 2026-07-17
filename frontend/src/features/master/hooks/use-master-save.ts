import { useCallback, useState, type TransitionStartFunction } from "react";
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
}: UseMasterSaveOptions<T, TForm, TCreate, TUpdate>) {
  // rerender-dependencies: extract primitives from crud.editTarget object
  const editTargetId = crud.editTarget !== null && crud.editTarget !== "new" ? crud.editTarget.id : null;
  // rerender-dependencies: destructure methods to avoid object reference instability
  // NOTE: use setEditTarget(null) directly on save success to bypass confirmDiscard()
  // crudHandleClose calls confirmDiscard() which shows window.confirm when isDirtyRef is still stale
  const crudSetEditTarget = crud.setEditTarget;
  const crudStartSave = crud.startSaveTransition;

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
