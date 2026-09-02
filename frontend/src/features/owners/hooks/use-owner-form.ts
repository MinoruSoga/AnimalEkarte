import {
  useState,
  useCallback,
  useActionState,
  useEffect,
  useLayoutEffect,
  useRef,
} from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";
import type { Owner } from "@/types/owner";
import type { PetMutations } from "@/types/pet";
import type { OwnerData } from "../types";
import { createOwner } from "../api/create-owner";
import { updateOwner } from "../api/update-owner";
import { usePetFormListState } from "./use-pet-form-list-state";
import {
  buildCreateOwnerRequest,
  buildUpdateOwnerRequest,
  DEFAULT_OWNER_DATA,
  DENIED_MUTATION_PERMISSIONS,
  mapOwnerPetsToFormData,
  mapOwnerToFormData,
  resolveCreatedOwnerClinicId,
  validateOwnerForm,
  type OwnerMutationPermissions,
} from "./use-owner-form-model";

export type { OwnerMutationPermissions };

export function useOwnerForm(
  id?: string,
  initialOwner?: Owner,
  petMutations?: PetMutations,
  permissions: Readonly<OwnerMutationPermissions> = DENIED_MUTATION_PERMISSIONS,
) {
  const isEdit = !!id;
  const queryClient = useQueryClient();
  const { canCreate, canEdit, canDelete } = permissions;
  const permissionsRef = useRef(permissions);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit, canDelete };
  }, [canCreate, canDelete, canEdit]);
  const isMutationAllowed = useCallback(
    (action: keyof OwnerMutationPermissions) =>
      permissionsRef.current[action] === true,
    [],
  );

  const [ownerData, setOwnerData] = useState<OwnerData>(
    () => initialOwner ? mapOwnerToFormData(initialOwner) : DEFAULT_OWNER_DATA
  );

  const {
    pets,
    setPets,
    petModalOpen,
    setPetModalOpen,
    editingPet,
    handleAddPet,
    handleEditPet,
    handleDeletePet,
    handleSavePet,
    handlePetLifecycleChange,
  } = usePetFormListState({
    id,
    initialPets: initialOwner ? mapOwnerPetsToFormData(initialOwner) : [],
    petMutations,
    permissions,
  });

  const [manualErrors, setManualErrors] = useState<Record<string, string>>({});

  const [formState, formAction, isPending] = useActionState(
    async (prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      const errors = validateOwnerForm(ownerData);
      if (Object.keys(errors).length > 0) {
        return { success: false, fieldErrors: errors, timestamp: Date.now() };
      }

      try {
        if (isEdit && id) {
          const updateData = buildUpdateOwnerRequest(ownerData);
          if (!isMutationAllowed("canEdit")) {
            return { success: false, timestamp: Date.now() };
          }
          await updateOwner(id, updateData);
          await queryClient.invalidateQueries({ queryKey: queryKeys.owners.all() });
          toast.success("飼主情報を更新しました");
          return { success: true, timestamp: Date.now() };
        }
        const createData = buildCreateOwnerRequest(ownerData, pets);
        if (!isMutationAllowed("canCreate")) {
          return { success: false, timestamp: Date.now() };
        }
        const newOwner = await createOwner(createData);
        await queryClient.invalidateQueries({ queryKey: queryKeys.owners.all() });
        toast.success("飼主情報を登録しました");
        return {
          success: true,
          data: { id: newOwner.id, clinicId: resolveCreatedOwnerClinicId(ownerData, newOwner, createData) },
          timestamp: Date.now(),
        };
      } catch (error) {
        handleApiError(error, "保存");
        return { ...prevState, success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE
  );

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- ActionState のエラーをフォームフィールドに同期するパターン
    setManualErrors(formState.fieldErrors || {});
  }, [formState.fieldErrors, formState.timestamp]);

  const clearFieldError = useCallback((field: string) => {
    setManualErrors((prev) => {
      if (!prev[field]) return prev;
      const next = { ...prev };
      delete next[field];
      return next;
    });
  }, []);

  return {
    isEdit,
    isLoading: isPending,
    ownerData,
    setOwnerData,
    pets,
    setPets,
    petModalOpen,
    setPetModalOpen,
    editingPet,
    handleAddPet,
    handleEditPet,
    handleDeletePet,
    handleSavePet,
    handlePetLifecycleChange,
    formAction,
    formState,
    fieldErrors: manualErrors,
    clearFieldError,
  };
}
