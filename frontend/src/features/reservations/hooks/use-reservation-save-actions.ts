import { useCallback, useLayoutEffect, useRef, useTransition } from "react";
import type { RefObject } from "react";
import { toast } from "sonner";

import { extractApiErrorMessage } from "@/lib/handle-api-error";
import type { ReservationCreateMutations } from "@/types/reservation-create-mutations";

import { useCreateReservation, useCreateReservationBatch } from "../api/create-reservation";
import { transformToCreateRequest } from "../api/transforms";
import { useUpdateReservation } from "../api/update-reservation";
import { buildReservationUpdateRequest } from "../lib/reservation-actions-model";
import type { NewOwnerFormData, Pet, ReservationFormData } from "../types";

/** FE-RC-204: action 別の最新権限値。mutation 直前に isMutationAllowed() で再検査する。 */
export interface ReservationMutationPermissions {
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
}

export const DENIED_RESERVATION_MUTATION_PERMISSIONS: Readonly<ReservationMutationPermissions> = {
  canCreate: false,
  canEdit: false,
  canDelete: false,
};

export const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

interface UseReservationSaveActionsArgs {
  editingAppointmentRef: RefObject<ReservationFormData | null>;
  checkOverlap: (
    newStart: Date,
    newEnd: Date,
    doctor: string,
    excludeId?: string,
    excludeIds?: ReadonlySet<string>,
  ) => boolean;
  handleCloseForm: () => void;
  navigateBackIfNeeded: () => void;
  createMutations: ReservationCreateMutations;
  /** FE-RC-204: 未指定は fail-closed（全拒否）。 */
  permissions?: Readonly<ReservationMutationPermissions>;
}

/** FE-RC-045: use-reservation-actions.ts から保存系（更新/新規飼主/既存ペット）を分離。 */
export function useReservationSaveActions({
  editingAppointmentRef,
  checkOverlap,
  handleCloseForm,
  navigateBackIfNeeded,
  createMutations,
  permissions: permissionsArg,
}: UseReservationSaveActionsArgs) {
  const permissions = permissionsArg ?? DENIED_RESERVATION_MUTATION_PERMISSIONS;
  const { canCreate, canEdit, canDelete } = permissions;
  const permissionsRef = useRef(permissions);
  useLayoutEffect(() => {
    permissionsRef.current = { canCreate, canEdit, canDelete };
  }, [canCreate, canDelete, canEdit]);
  const isMutationAllowed = useCallback(
    (action: keyof ReservationMutationPermissions) => permissionsRef.current[action] === true,
    [],
  );

  const createMutation = useCreateReservation();
  const createBatchMutation = useCreateReservationBatch();
  const updateMutation = useUpdateReservation();
  const { mutateAsync: createReservationAsync } = createMutation;
  const { mutateAsync: createBatchReservationAsync } = createBatchMutation;
  const { mutate: updateReservationFn } = updateMutation;
  const [, startUpdateTransition] = useTransition();
  const createdPetIdsRef = useRef(new Set<string>());
  const createdReservationIdsRef = useRef(new Set<string>());
  const newOwnerProgressRef = useRef<{ key: string; ownerID?: string; petID?: string } | null>(
    null,
  );
  const resetCreateProgress = useCallback(() => {
    createdPetIdsRef.current.clear();
    createdReservationIdsRef.current.clear();
    newOwnerProgressRef.current = null;
  }, []);
  const handleCloseCreateForm = useCallback(() => {
    resetCreateProgress();
    handleCloseForm();
  }, [handleCloseForm, resetCreateProgress]);

  /** 既存予約の日時/内容変更を保存する（FE-RC-046: handleSave から分離した1経路）。 */
  const saveUpdateReservation = useCallback(
    (
      currentEditing: ReservationFormData,
      data: ReservationFormData,
      targetDoctor: string,
    ): Promise<string | null> => {
      const updatePayload = buildReservationUpdateRequest(currentEditing, data, targetDoctor);
      if (!updatePayload) return Promise.resolve(null);
      return new Promise<string | null>((resolve) => {
        startUpdateTransition(() => {
          updateReservationFn(updatePayload, {
            onSuccess: () => {
              toast.success("予約を更新しました", { description: `担当医: ${targetDoctor}` });
              handleCloseForm();
              navigateBackIfNeeded();
              resolve(null);
            },
            onError: (error: unknown) => {
              resolve(extractApiErrorMessage(error, "更新"));
            },
          });
        });
      });
    },
    [handleCloseForm, navigateBackIfNeeded, updateReservationFn],
  );

  /** 新規飼主/ペットを作成してから予約を作成する（FE-RC-046: handleSave から分離した1経路）。 */
  const saveNewOwnerReservation = useCallback(
    async (
      data: ReservationFormData,
      newOwnerData: NewOwnerFormData,
      targetDoctor: string,
    ): Promise<string | null> => {
      try {
        const progressKey = JSON.stringify(newOwnerData);
        if (newOwnerProgressRef.current?.key !== progressKey) {
          newOwnerProgressRef.current = { key: progressKey };
        }
        const progress = newOwnerProgressRef.current;
        if (!progress.ownerID) {
          const owner = await createMutations.createOwnerFn({
            owner_name: newOwnerData.ownerName,
            phone: newOwnerData.phone,
          });
          progress.ownerID = String(owner.id);
        }
        if (!progress.petID) {
          const pet = await createMutations.createPetFn({
            owner_id: Number(progress.ownerID),
            animal_species_id: newOwnerData.animalSpeciesId,
            name: newOwnerData.petName,
          });
          progress.petID = String(pet.id);
        }
        const createPayload = transformToCreateRequest(
          { ...data, notes: data.notes ?? newOwnerData.chiefComplaint },
          progress.petID,
          progress.ownerID,
        );
        await createReservationAsync(createPayload);
        toast.success("予約を作成しました", {
          description: `${newOwnerData.ownerName}様 / ${newOwnerData.petName} / 担当医: ${targetDoctor}`,
        });
        handleCloseCreateForm();
        navigateBackIfNeeded();
        return null;
      } catch (error) {
        return extractApiErrorMessage(error, "作成");
      }
    },
    [createReservationAsync, createMutations, handleCloseCreateForm, navigateBackIfNeeded],
  );

  /** 既存の飼主/ペット（単体・複数）に予約を作成する（FE-RC-046: handleSave から分離した1経路）。 */
  const saveExistingPetsReservation = useCallback(
    async (
      data: ReservationFormData,
      selectedPets: Pick<Pet, "id" | "ownerId" | "name">[],
      targetDoctor: string,
    ): Promise<string | null> => {
      try {
        if (selectedPets.length === 1) {
          const pet = selectedPets[0];
          await createReservationAsync(transformToCreateRequest(data, pet.id, pet.ownerId));
        } else {
          // A selected multi-pet booking is one atomic server-side operation. Do not
          // issue individual creates: the second intentional overlap must not conflict.
          const {
            pet_id: _petID,
            owner_id: _ownerID,
            ...base
          } = transformToCreateRequest(data, "0", "0");
          await createBatchReservationAsync({
            ...base,
            pets: selectedPets.map((pet) => ({
              pet_id: Number(pet.id),
              owner_id: Number(pet.ownerId),
            })),
          });
        }
        toast.success(`${selectedPets.length}件の予約を作成しました`, {
          description: `担当医: ${targetDoctor}`,
        });
        handleCloseCreateForm();
        navigateBackIfNeeded();
        return null;
      } catch (error) {
        return extractApiErrorMessage(error, "作成");
      }
    },
    [
      createBatchReservationAsync,
      createReservationAsync,
      handleCloseCreateForm,
      navigateBackIfNeeded,
    ],
  );

  const handleSave = useCallback(
    async (
      data: ReservationFormData,
      selectedPets: Pick<Pet, "id" | "ownerId" | "name">[],
      newOwnerData?: NewOwnerFormData,
    ): Promise<string | null> => {
      if (!data.start || !data.end) return null;
      if (!newOwnerData && selectedPets.length === 0) return null;

      const currentEditing = editingAppointmentRef.current;
      if (currentEditing?.id) {
        if (!isMutationAllowed("canEdit")) {
          toast.error(PERMISSION_DENIED_MESSAGE);
          return PERMISSION_DENIED_MESSAGE;
        }
      } else if (!isMutationAllowed("canCreate")) {
        toast.error(PERMISSION_DENIED_MESSAGE);
        return PERMISSION_DENIED_MESSAGE;
      }

      const targetDoctor = data.doctor || currentEditing?.doctor || "";
      const hasOverlap = checkOverlap(
        data.start,
        data.end,
        targetDoctor,
        currentEditing?.id,
        currentEditing?.id ? undefined : createdReservationIdsRef.current,
      );

      if (hasOverlap) {
        // FE precheck — keep modal open with inline message (same surface as API 409).
        return "指定された時間帯には既に予約が入っています";
      }

      if (currentEditing?.id) {
        return saveUpdateReservation(currentEditing, data, targetDoctor);
      }

      if (newOwnerData) {
        return saveNewOwnerReservation(data, newOwnerData, targetDoctor);
      }

      return saveExistingPetsReservation(data, selectedPets, targetDoctor);
    },
    [
      checkOverlap,
      editingAppointmentRef,
      isMutationAllowed,
      saveExistingPetsReservation,
      saveNewOwnerReservation,
      saveUpdateReservation,
    ],
  );

  return {
    handleSave,
    resetCreateProgress,
    handleCloseCreateForm,
  };
}
