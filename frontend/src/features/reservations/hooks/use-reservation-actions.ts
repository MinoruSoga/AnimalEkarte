import { useCallback, useTransition } from "react";
import type { Dispatch, RefObject, SetStateAction } from "react";
import type { NavigateFunction } from "react-router";
import { toast } from "sonner";

import { handleApiError } from "@/lib/handle-api-error";
import { jstWallDateToISOString } from "@/lib/jst-date";
import { getReservationStatusLabel } from "@/lib/status-helpers";
import type { ReservationCreateMutations } from "@/types/reservation-create-mutations";

import { useCreateReservation } from "../api/create-reservation";
import { useDeleteReservation } from "../api/delete-reservation";
import { transformToCreateRequest } from "../api/transforms";
import { useUpdateReservation } from "../api/update-reservation";
import type {
  NavigationState,
  NewOwnerFormData,
  Pet,
  Reservation,
  ReservationFormData,
  ReservationStatus,
} from "../types";

/** 詳細モーダルからの破壊的 status 変更（確認ダイアログ対象）。BUG-020 */
export const DESTRUCTIVE_RESERVATION_STATUSES: readonly ReservationStatus[] = [
  "cancelled",
  "no_show",
] as const;

export function isDestructiveReservationStatus(status: ReservationStatus): boolean {
  return (DESTRUCTIVE_RESERVATION_STATUSES as readonly string[]).includes(status);
}

export interface StatusConfirmTarget {
  reservation: Reservation;
  status: ReservationStatus;
}

interface UseReservationActionsArgs {
  appointments: Reservation[];
  editingAppointmentRef: RefObject<ReservationFormData | null>;
  deleteTarget: Reservation | null;
  setDeleteConfirmOpen: Dispatch<SetStateAction<boolean>>;
  setDeleteTarget: Dispatch<SetStateAction<Reservation | null>>;
  statusConfirmTarget: StatusConfirmTarget | null;
  setStatusConfirmOpen: Dispatch<SetStateAction<boolean>>;
  setStatusConfirmTarget: Dispatch<SetStateAction<StatusConfirmTarget | null>>;
  setDetailAppointment: Dispatch<SetStateAction<Reservation | null>>;
  handleCloseForm: () => void;
  handleCloseDetail: () => void;
  locationFrom: NavigationState["from"] | null;
  navigate: NavigateFunction;
  createMutations: ReservationCreateMutations;
}

function buildUpdatePayload(
  reservation: Reservation,
  start: Date,
  end: Date,
  status: ReservationStatus,
) {
  return {
    id: reservation.id,
    req: {
      start_time: jstWallDateToISOString(start),
      end_time: jstWallDateToISOString(end),
      visit_type: reservation.visitType,
      doctor_id: reservation.doctor ? Number(reservation.doctor) : undefined,
      is_designated: reservation.isDesignated,
      status,
      notes: reservation.notes,
    },
  };
}

export function useReservationActions({
  appointments,
  editingAppointmentRef,
  deleteTarget,
  setDeleteConfirmOpen,
  setDeleteTarget,
  statusConfirmTarget,
  setStatusConfirmOpen,
  setStatusConfirmTarget,
  setDetailAppointment,
  handleCloseForm,
  handleCloseDetail,
  locationFrom,
  navigate,
  createMutations,
}: UseReservationActionsArgs) {
  const createMutation = useCreateReservation();
  const updateMutation = useUpdateReservation();
  const { mutate: updateReservationFn } = updateMutation;
  const deleteMutation = useDeleteReservation();
  const { mutate: deleteReservationFn } = deleteMutation;
  const [, startUpdateTransition] = useTransition();
  const [, startDeleteTransition] = useTransition();

  const checkOverlap = useCallback(
    (newStart: Date, newEnd: Date, doctor: string, excludeId?: string): boolean =>
      appointments.some((app) => {
        if (excludeId && app.id === excludeId) return false;
        if (app.status === "cancelled") return false;
        if (app.doctor !== doctor) return false;
        return newStart < app.end && newEnd > app.start;
      }),
    [appointments],
  );

  const navigateBackIfNeeded = useCallback(() => {
    if (locationFrom) {
      navigate(locationFrom);
    }
  }, [locationFrom, navigate]);

  const handleSave = useCallback(
    async (data: ReservationFormData, selectedPets: Pet[], newOwnerData?: NewOwnerFormData) => {
      if (!data.start || !data.end) return;
      if (!newOwnerData && selectedPets.length === 0) return;

      const currentEditing = editingAppointmentRef.current;
      const targetDoctor = data.doctor || currentEditing?.doctor || "";
      const hasOverlap = checkOverlap(data.start, data.end, targetDoctor, currentEditing?.id);

      if (hasOverlap) {
        toast.error("指定された時間帯には既に予約が入っています", {
          description: "時間帯または担当医を変更してください。",
        });
        return;
      }

      if (currentEditing?.id) {
        const updatePayload = {
          id: currentEditing.id,
          req: {
            start_time: jstWallDateToISOString(data.start),
            end_time: jstWallDateToISOString(data.end),
            visit_type: data.visitType || "first",
            reservation_type_id: data.type ? Number(data.type) : undefined,
            doctor_id: targetDoctor ? Number(targetDoctor) : undefined,
            is_designated: data.isDesignated ?? false,
            status: data.status || ("confirmed" as const),
            notes: data.notes,
          },
        };
        startUpdateTransition(() => {
          updateReservationFn(updatePayload, {
            onSuccess: () => {
              toast.success("予約を更新しました", { description: `担当医: ${targetDoctor}` });
              handleCloseForm();
              navigateBackIfNeeded();
            },
          });
        });
      } else if (newOwnerData) {
        try {
          const owner = await createMutations.createOwnerFn({
            owner_name: newOwnerData.ownerName,
            phone: newOwnerData.phone,
          });
          const pet = await createMutations.createPetFn({
            owner_id: Number(owner.id),
            animal_species_id: newOwnerData.animalSpeciesId,
            name: newOwnerData.petName,
          });
          const createPayload = transformToCreateRequest(
            { ...data, notes: data.notes ?? newOwnerData.chiefComplaint },
            String(pet.id),
            String(owner.id),
          );
          await createMutation.mutateAsync(createPayload);
          toast.success("予約を作成しました", {
            description: `${newOwnerData.ownerName}様 / ${newOwnerData.petName} / 担当医: ${targetDoctor}`,
          });
          handleCloseForm();
          navigateBackIfNeeded();
        } catch (error) {
          handleApiError(error, "作成");
        }
      } else {
        try {
          const results = await Promise.allSettled(
            selectedPets.map((pet) => {
              const createPayload = transformToCreateRequest(data, pet.id, pet.ownerId);
              return createMutation.mutateAsync(createPayload);
            }),
          );
          const succeeded = results.filter((r) => r.status === "fulfilled").length;
          const failed = results.filter((r) => r.status === "rejected").length;

          if (failed > 0) {
            toast.error(`${failed}件の予約作成に失敗しました`, {
              description: `${succeeded}件は成功しました。`,
            });
          }
          if (succeeded > 0) {
            toast.success(`${succeeded}件の予約を作成しました`, {
              description: `担当医: ${targetDoctor}`,
            });
            handleCloseForm();
            navigateBackIfNeeded();
          }
        } catch (error) {
          handleApiError(error, "作成");
        }
      }
    },
    [
      checkOverlap,
      createMutation,
      createMutations,
      editingAppointmentRef,
      handleCloseForm,
      navigateBackIfNeeded,
      updateReservationFn,
    ],
  );

  const handleReservationUpdate = useCallback(
    (reservation: Reservation, newStart: Date, newEnd: Date) => {
      const hasOverlap = checkOverlap(newStart, newEnd, reservation.doctor, reservation.id);

      if (hasOverlap) {
        toast.error("移動先に予約が重複しています", {
          description: "別の時間帯にドラッグしてください。",
        });
        return;
      }

      startUpdateTransition(() => {
        updateReservationFn(buildUpdatePayload(reservation, newStart, newEnd, reservation.status), {
          onSuccess: () => {
            toast.success("予約時間を変更しました", {
              description: `${reservation.petName} / ${reservation.doctor}`,
            });
          },
        });
      });
    },
    [checkOverlap, updateReservationFn],
  );

  /** BUG-020: status のみの最小 payload（日時・担当医の再検証を避ける） */
  const applyStatusChange = useCallback(
    (reservation: Reservation, status: ReservationStatus) => {
      startUpdateTransition(() => {
        updateReservationFn(
          { id: reservation.id, req: { status } },
          {
            onSuccess: () => {
              setDetailAppointment((prev) => (prev ? { ...prev, status } : null));
              const statusLabel = getReservationStatusLabel(status);
              toast.success("ステータスを更新しました", {
                description: `${reservation.petName} → ${statusLabel}`,
              });
            },
          },
        );
      });
    },
    [setDetailAppointment, updateReservationFn],
  );

  const handleStatusChange = useCallback(
    (reservation: Reservation, status: ReservationStatus) => {
      if (status === reservation.status) return;
      if (isDestructiveReservationStatus(status)) {
        setStatusConfirmTarget({ reservation, status });
        setStatusConfirmOpen(true);
        return;
      }
      applyStatusChange(reservation, status);
    },
    [applyStatusChange, setStatusConfirmOpen, setStatusConfirmTarget],
  );

  const executeStatusChange = useCallback(() => {
    if (!statusConfirmTarget) return;
    const { reservation, status } = statusConfirmTarget;
    setStatusConfirmOpen(false);
    setStatusConfirmTarget(null);
    applyStatusChange(reservation, status);
  }, [applyStatusChange, setStatusConfirmOpen, setStatusConfirmTarget, statusConfirmTarget]);

  const executeDelete = useCallback(() => {
    if (!deleteTarget) return;
    startDeleteTransition(() => {
      deleteReservationFn(deleteTarget.id, {
        onSuccess: () => {
          setDeleteConfirmOpen(false);
          setDeleteTarget(null);
          handleCloseDetail();
          toast.success("予約を削除しました", {
            description: `${deleteTarget.petName} (${deleteTarget.ownerName}様)`,
          });
        },
        onError: (error: unknown) => {
          handleApiError(error, "削除");
        },
      });
    });
  }, [deleteReservationFn, deleteTarget, handleCloseDetail, setDeleteConfirmOpen, setDeleteTarget]);

  return {
    handleSave,
    handleReservationUpdate,
    handleStatusChange,
    executeStatusChange,
    executeDelete,
  };
}
