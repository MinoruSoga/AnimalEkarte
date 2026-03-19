import { useCallback } from "react";
import { toast } from "sonner";
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
}: UseMasterSaveOptions<T, TForm, TCreate, TUpdate>) {
  // rerender-dependencies: extract primitives from crud.editTarget object
  const editTargetId = crud.editTarget !== null && crud.editTarget !== "new" ? crud.editTarget.id : null;

  const handleSave = useCallback(
    (data: TForm) => {
      const error = validate(data);
      if (error) {
        toast.error(error);
        return;
      }

      crud.startSaveTransition(() => {
        if (editTargetId !== null) {
          updateMutation.mutate(
            { id: editTargetId, req: toUpdateRequest(data) },
            {
              onSuccess: () => {
                toast.success("更新しました");
                crud.handleClose();
              },
              onError: () => toast.error("更新に失敗しました"),
            },
          );
        } else {
          createMutation.mutate(toCreateRequest(data), {
            onSuccess: () => {
              toast.success("登録しました");
              crud.handleClose();
            },
            onError: () => toast.error("登録に失敗しました"),
          });
        }
      });
    },
    [editTargetId, crud.handleClose, crud.startSaveTransition, createMutation, updateMutation, validate, toCreateRequest, toUpdateRequest],
  );

  return { handleSave };
}
