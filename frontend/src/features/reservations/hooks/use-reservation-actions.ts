import { useCallback, useTransition } from "react";
import type { Dispatch, RefObject, SetStateAction } from "react";
import type { NavigateFunction } from "react-router";
import { toast } from "sonner";

import { getReservationStatusLabel } from "@/lib/status-helpers";
import type { ReservationCreateMutations } from "@/types/reservation-create-mutations";

import { useDeleteReservation } from "../api/delete-reservation";
import { useUpdateReservation } from "../api/update-reservation";
import {
  buildUpdatePayload,
  isDestructiveReservationStatus,
  type StatusConfirmTarget,
} from "../lib/reservation-actions-model";
import type { NavigationState, ReservationFormData, Reservation, ReservationStatus } from "../types";
import { useReservationSaveActions } from "./use-reservation-save-actions";

export { isDestructiveReservationStatus } from "../lib/reservation-actions-model";
export type { StatusConfirmTarget } from "../lib/reservation-actions-model";
export { buildReservationUpdateRequest } from "../lib/reservation-actions-model";

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
  const updateMutation = useUpdateReservation();
  const { mutate: updateReservationFn } = updateMutation;
  const deleteMutation = useDeleteReservation();
  const { mutate: deleteReservationFn } = deleteMutation;
  const [, startUpdateTransition] = useTransition();
  const [, startDeleteTransition] = useTransition();

  const checkOverlap = useCallback(
    (
      newStart: Date,
      newEnd: Date,
      doctor: string,
      excludeId?: string,
      excludeIds?: ReadonlySet<string>,
    ): boolean =>
      appointments.some((app) => {
        if ((excludeId && app.id === excludeId) || excludeIds?.has(app.id)) return false;
        if (app.status === "cancelled") return false;
        if (app.doctorId !== doctor) return false;
        return newStart < app.end && newEnd > app.start;
      }),
    [appointments],
  );

  const navigateBackIfNeeded = useCallback(() => {
    if (locationFrom) {
      navigate(locationFrom);
    }
  }, [locationFrom, navigate]);

  const { handleSave, resetCreateProgress, handleCloseCreateForm } = useReservationSaveActions({
    editingAppointmentRef,
    checkOverlap,
    handleCloseForm,
    navigateBackIfNeeded,
    createMutations,
  });

  const handleReservationUpdate = useCallback(
    (reservation: Reservation, newStart: Date, newEnd: Date) => {
      const hasOverlap = checkOverlap(newStart, newEnd, reservation.doctorId ?? "", reservation.id);

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
      if (status === "in_consultation") {
        toast.error("カルテ作成が必要です", {
          description: "このステータスに変更するには、詳細画面からカルテを作成してください。",
        });
        return;
      }
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
        // useDeleteReservation の onError が handleApiError 済み。ここでは再通知しない。
      });
    });
  }, [deleteReservationFn, deleteTarget, handleCloseDetail, setDeleteConfirmOpen, setDeleteTarget]);

  return {
    handleSave,
    handleReservationUpdate,
    handleStatusChange,
    executeStatusChange,
    executeDelete,
    resetCreateProgress,
    handleCloseCreateForm,
  };
}
