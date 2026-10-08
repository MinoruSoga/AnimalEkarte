import { useCallback, useLayoutEffect, useRef, useTransition } from "react";
import type { RefObject } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { requireStoredClinicId } from "@/lib/current-clinic";
import { formatJSTWallDate, todayJSTISO } from "@/lib/jst-date";
import { getReceptionCandidates } from "../api/get-reception-candidates";
import { extractApiErrorMessage, handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
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

  const queryClient = useQueryClient();
  const createMutation = useCreateReservation();
  const createBatchMutation = useCreateReservationBatch();
  const updateMutation = useUpdateReservation();
  const { mutateAsync: createReservationAsync } = createMutation;
  const { mutateAsync: createBatchReservationAsync } = createBatchMutation;
  const { mutate: updateReservationFn } = updateMutation;
  const [, startUpdateTransition] = useTransition();
  const savingReceptionRef = useRef(false);
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
          // EMR-264: 注入される生関数は useMutation の invalidate を経由しない。
          // 部分成功（飼主作成済み・予約未作成）でも検索キャッシュを正す。
          await queryClient.invalidateQueries({ queryKey: queryKeys.owners.all() });
        }
        if (!progress.petID) {
          const pet = await createMutations.createPetFn({
            owner_id: Number(progress.ownerID),
            animal_species_id: newOwnerData.animalSpeciesId,
            name: newOwnerData.petName,
          });
          progress.petID = String(pet.id);
          await queryClient.invalidateQueries({ queryKey: queryKeys.pets.list() });
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
    [
      createReservationAsync,
      createMutations,
      handleCloseCreateForm,
      navigateBackIfNeeded,
      queryClient,
    ],
  );

  /** 既存の飼主/ペット（単体・複数）に予約を作成する（FE-RC-046: handleSave から分離した1経路）。 */
  const saveExistingPetsReservation = useCallback(
    async (
      data: ReservationFormData,
      selectedPets: Pick<Pet, "id" | "ownerId" | "name">[],
      targetDoctor: string,
      preventDuplicateReception = false,
    ): Promise<string | null> => {
      try {
        const duplicateGuard = preventDuplicateReception
          ? { prevent_duplicate_reception: true }
          : {};
        if (selectedPets.length === 1) {
          const pet = selectedPets[0];
          await createReservationAsync({
            ...transformToCreateRequest(data, pet.id, pet.ownerId),
            ...duplicateGuard,
          });
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
            ...duplicateGuard,
            pets: selectedPets.map((pet) => ({
              pet_id: Number(pet.id),
              owner_id: Number(pet.ownerId),
            })),
          });
        }
        toast.success(`${selectedPets.length.toLocaleString()}件の予約を作成しました`, {
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

  const saveReception = useCallback(
    async (
      data: ReservationFormData,
      selectedPets: Pick<Pet, "id" | "ownerId" | "name">[],
      targetDoctor: string,
    ): Promise<string | null> => {
      if (savingReceptionRef.current) return "受付を処理中です";
      const receptionDate = todayJSTISO();
      if (!data.start || formatJSTWallDate(data.start) !== receptionDate)
        return "受付日は当日を指定してください";
      savingReceptionRef.current = true;
      try {
        const receptionClinicId = requireStoredClinicId();
        const candidates = await Promise.all(
          selectedPets.map((pet) => getReceptionCandidates(receptionDate, pet.id, pet.ownerId)),
        );
        if (requireStoredClinicId() !== receptionClinicId)
          return "医院が変わりました。再読み込みしてください";
        if (todayJSTISO() !== receptionDate) return "受付日が変わりました。再読み込みしてください";
        if (!isMutationAllowed("canCreate")) return PERMISSION_DENIED_MESSAGE;
        if (candidates.some((rows) => rows.length > 1))
          return "当日の予約が複数あります。受付画面で対象の予約を選択してください";
        const existing = candidates.flat();
        if (selectedPets.length > 1 && existing.length > 0)
          return "当日の既存予約があります。受付画面でペットごとに受付してください";
        const reservation = existing[0];
        if (!reservation)
          return await saveExistingPetsReservation(data, selectedPets, targetDoctor, true);
        if (
          !["confirmed", "checked_in", "in_consultation", "accounting"].includes(reservation.status)
        )
          return "予約の状態を受付画面で確認してください";
        if (reservation.status === "confirmed") {
          if (!isMutationAllowed("canEdit")) return PERMISSION_DENIED_MESSAGE;
          const error = await new Promise<string | null>((resolve) => {
            startUpdateTransition(() =>
              updateReservationFn(
                {
                  id: String(reservation.id),
                  req: { status: "checked_in", expected_updated_at: reservation.updated_at },
                },
                {
                  onSuccess: () => resolve(null),
                  onError: (cause: unknown) => resolve(extractApiErrorMessage(cause, "受付")),
                },
              ),
            );
          });
          if (error) return error;
        }
        // Already checked-in or later: do not stamp checked_in_at again or alter the schedule.
        toast.success("当日の予約を受付画面で確認できます");
        handleCloseCreateForm();
        navigateBackIfNeeded();
        return null;
      } catch (error) {
        handleApiError(error, "受付");
        return extractApiErrorMessage(error, "受付");
      } finally {
        savingReceptionRef.current = false;
      }
    },
    [
      handleCloseCreateForm,
      isMutationAllowed,
      navigateBackIfNeeded,
      saveExistingPetsReservation,
      updateReservationFn,
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
      if (
        !currentEditing?.id &&
        !newOwnerData &&
        data.reservationRoute === "reception" &&
        data.status === "checked_in"
      ) {
        return saveReception(data, selectedPets, targetDoctor);
      }
      const hasOverlap = checkOverlap(
        data.start,
        data.end,
        targetDoctor,
        currentEditing?.id,
        currentEditing?.id ? undefined : createdReservationIdsRef.current,
      );

      if (hasOverlap) {
        // FE precheck — keep modal open with the same message the API 409 returns
        // (EMR-76: server-side conflict body is 既に予約が存在します).
        return "既に予約が存在します";
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
      saveReception,
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
