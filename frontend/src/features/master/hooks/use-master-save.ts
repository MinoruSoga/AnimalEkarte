import { useCallback } from "react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import type { UseMutationResult } from "@tanstack/react-query";
import type { UseMasterCRUDReturn } from "@/features/master/hooks/use-master-crud";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

interface MasterEntity {
  id: string;
}

interface UseMasterSaveOptions<T extends MasterEntity, TForm, TCreate, TUpdate> {
  crud: UseMasterCRUDReturn<T>;
  createMutation: UseMutationResult<unknown, Error, TCreate>;
  updateMutation: UseMutationResult<unknown, Error, { id: string; req: TUpdate }>;
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
  const crudHandleClose = crud.handleClose;
  const crudStartSave = crud.startSaveTransition;

  const handleSave = useCallback(
    (data: TForm) => {
      const error = validate(data);
      if (error) {
        toast.error(error);
        return;
      }

      crudStartSave(() => {
        if (editTargetId !== null) {
          updateMutation.mutate(
            { id: editTargetId, req: toUpdateRequest(data) },
            {
              onSuccess: async (savedData) => {
                try {
                  const saved = savedData as T;
                  if (onSuccess) {
                    await onSuccess(saved, data);
                  }
                  toast.success("更新しました");
                  crudHandleClose();
                } catch (error) {
                  handleApiError(error, "保存");
                }
              },
              onError: (error) => handleApiError(error, "更新"),
            },
          );
        } else {
          createMutation.mutate(toCreateRequest(data), {
            onSuccess: async (savedData) => {
              try {
                const saved = savedData as T;
                if (onSuccess) {
                  await onSuccess(saved, data);
                }
                toast.success("登録しました");
                crudHandleClose();
              } catch (error) {
                handleApiError(error, "保存");
              }
            },
            onError: (error) => handleApiError(error, "登録"),
          });
        }
      });
    },
    [editTargetId, crudHandleClose, crudStartSave, createMutation, updateMutation, validate, toCreateRequest, toUpdateRequest, onSuccess],
  );

  return { handleSave };
}
