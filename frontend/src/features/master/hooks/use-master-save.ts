import {
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
  type TransitionStartFunction,
} from "react";
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
  /** Set false when the caller must clear local dirty state before closing. */
  closeOnSuccess?: boolean;
  showSuccessToast?: boolean;
  /** Action-specific permissions enforced at the mutation boundary. */
  permissions: MasterSavePermissions;
}

// ─────────────────────────────────────────────────
// Hook
// ─────────────────────────────────────────────────

/**
 * Validates the form, enforces the latest live action permissions at the
 * mutation boundary, and performs the create/update mutation.
 * Returns an entity-aware result method plus the legacy Boolean save method;
 * callers may own notification and close behavior via options.
 */
export function useMasterSave<T extends MasterEntity, TForm, TCreate, TUpdate>({
  crud,
  createMutation,
  updateMutation,
  validate,
  toCreateRequest,
  toUpdateRequest,
  onSuccess,
  closeOnSuccess = true,
  showSuccessToast = true,
  permissions,
}: UseMasterSaveOptions<T, TForm, TCreate, TUpdate>) {
  // rerender-dependencies: extract primitives from crud.editTarget object
  const editTargetId =
    crud.editTarget !== null && crud.editTarget !== "new" ? crud.editTarget.id : null;
  // rerender-dependencies: destructure methods to avoid object reference instability
  // NOTE: use setEditTarget(null) directly on save success to bypass discard confirmation
  // crudHandleClose calls withDirtyGuard which would open ConfirmDialog when isDirtyRef is still stale
  const crudSetEditTarget = crud.setEditTarget;
  const crudStartSave = crud.startSaveTransition;
  const canCreate = permissions.canCreate;
  const canEdit = permissions.canEdit;
  const permissionsRef = useRef<MasterSavePermissions>({
    canCreate: canCreate === true,
    canEdit: canEdit === true,
  });
  useLayoutEffect(() => {
    permissionsRef.current = {
      canCreate: canCreate === true,
      canEdit: canEdit === true,
    };
  }, [canCreate, canEdit]);

  const [validationError, setValidationError] = useState<string | null>(null);
  const { mutateAsync: createAsync } = createMutation;
  const { mutateAsync: updateAsync } = updateMutation;

  /**
   * Saves and returns the persisted entity for this invocation.
   * The optional targetId overrides the CRUD edit target for this call only,
   * selecting both the mutation and the latest permission gate without
   * changing crud.editTarget. Returns the entity after onSuccess, or null on
   * validation error, permission denial, transition start failure, mutation
   * rejection, or post-save failure. An unobserved response carries no
   * idempotency guarantee.
   */
  const handleSaveWithResult = useCallback(
    async (data: TForm, targetId?: string): Promise<T | null> => {
      const error = validate(data);
      if (error) {
        setValidationError(error);
        toast.error(error);
        return null;
      }
      setValidationError(null);

      const effectiveTargetId = targetId ?? editTargetId;
      const currentPermissions = permissionsRef.current;
      const isAllowed =
        effectiveTargetId !== null
          ? currentPermissions.canEdit === true
          : currentPermissions.canCreate === true;
      if (!isAllowed) return null;

      return new Promise<T | null>((resolve) => {
        const isUpdate = effectiveTargetId !== null;
        const actionLabel = isUpdate ? "更新" : "登録";

        /**
         * Awaits the mutation and the post-save hook, then performs the
         * optional notification/close and settles this invocation's outer
         * promise — not any global last result. API errors are already
         * notified by the mutation hooks and are not toasted twice.
         */
        const save = async () => {
          let saved: T;
          try {
            saved = isUpdate
              ? await updateAsync({
                  id: effectiveTargetId,
                  req: toUpdateRequest(data),
                })
              : await createAsync(toCreateRequest(data));
          } catch {
            // create/updateMutation の onError で handleApiError 済み（master/api/*.ts）。
            // ここで再通知すると二重 toast になるため状態遷移のみ行う。
            resolve(null);
            return;
          }

          try {
            await onSuccess?.(saved, data);
          } catch (error) {
            handleApiError(error, "保存");
            resolve(null);
            return;
          }

          if (showSuccessToast) {
            toast.success(isUpdate ? "更新しました" : "登録しました");
          }
          if (closeOnSuccess) {
            crudSetEditTarget(null);
          }
          resolve(saved);
        };

        try {
          crudStartSave(() => {
            void save();
          });
        } catch (error) {
          handleApiError(error, actionLabel);
          resolve(null);
        }
      });
    },
    [
      editTargetId,
      crudSetEditTarget,
      crudStartSave,
      createAsync,
      updateAsync,
      validate,
      toCreateRequest,
      toUpdateRequest,
      onSuccess,
      closeOnSuccess,
      showSuccessToast,
    ],
  );

  /**
   * Legacy Boolean compatibility wrapper. Identity-aware callers should use
   * handleSaveWithResult instead.
   */
  const handleSave = useCallback(
    async (data: TForm): Promise<boolean> => {
      return Boolean(await handleSaveWithResult(data));
    },
    [handleSaveWithResult],
  );

  return { handleSave, handleSaveWithResult, validationError };
}
