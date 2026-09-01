import { useCallback, useRef, useTransition } from "react";
import type { Dispatch, RefObject, SetStateAction } from "react";
import type { NavigateFunction } from "react-router";
import { toast } from "sonner";

import { handleApiError, extractApiErrorMessage } from "@/lib/handle-api-error";
import { jstWallDateToISOString } from "@/lib/jst-date";
import { getReservationStatusLabel } from "@/lib/status-helpers";
import type { ReservationCreateMutations } from "@/types/reservation-create-mutations";

import { useCreateReservation, useCreateReservationBatch } from "../api/create-reservation";
import { useDeleteReservation } from "../api/delete-reservation";
import { transformToCreateRequest } from "../api/transforms";
import { useUpdateReservation } from "../api/update-reservation";
import type { UpdateReservationRequest } from "@/hooks/use-update-reservation";
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

/** Notes-only edits must omit schedule/doctor so BE skips on-duty conflict checks (BUG-012). */
export function buildReservationUpdateRequest(
  current: ReservationFormData,
  data: ReservationFormData,
  targetDoctor: string,
): { id: string; req: UpdateReservationRequest } | null {
  if (!current.id) return null;

  const req: UpdateReservationRequest = {};
  if (data.start) {
    const nextStart = jstWallDateToISOString(data.start);
    const prevStart = current.start ? jstWallDateToISOString(current.start) : "";
    if (nextStart !== prevStart) req.start_time = nextStart;
  }
  if (data.end) {
    const nextEnd = jstWallDateToISOString(data.end);
    const prevEnd = current.end ? jstWallDateToISOString(current.end) : "";
    if (nextEnd !== prevEnd) req.end_time = nextEnd;
  }
  const nextVisit = data.visitType || "first";
  if (nextVisit !== (current.visitType || "first")) req.visit_type = nextVisit;
  const nextType = data.type ? Number(data.type) : undefined;
  const prevType = current.type ? Number(current.type) : undefined;
  if (nextType !== undefined && nextType !== prevType) req.reservation_type_id = nextType;
  const nextDoctor = targetDoctor ? Number(targetDoctor) : undefined;
  const prevDoctor = current.doctor ? Number(current.doctor) : undefined;
  if (nextDoctor !== prevDoctor && nextDoctor !== undefined) req.doctor_id = nextDoctor;
  if ((data.isDesignated ?? false) !== (current.isDesignated ?? false)) {
    req.is_designated = data.isDesignated ?? false;
  }
  const nextStatus = data.status || "confirmed";
  if (nextStatus !== (current.status || "confirmed")) req.status = nextStatus;
  if ((data.notes ?? "") !== (current.notes ?? "")) req.notes = data.notes;

  if (Object.keys(req).length === 0) {
    req.notes = data.notes ?? "";
  }
  return { id: current.id, req };
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
  const createBatchMutation = useCreateReservationBatch();
  const updateMutation = useUpdateReservation();
  const { mutate: updateReservationFn } = updateMutation;
  const deleteMutation = useDeleteReservation();
  const { mutate: deleteReservationFn } = deleteMutation;
  const [, startUpdateTransition] = useTransition();
  const [, startDeleteTransition] = useTransition();
  const createdPetIdsRef = useRef(new Set<string>());
  const createdReservationIdsRef = useRef(new Set<string>());
  const newOwnerProgressRef = useRef<{ key: string; ownerID?: string; petID?: string } | null>(null);
  const resetCreateProgress = useCallback(() => {
    createdPetIdsRef.current.clear();
    createdReservationIdsRef.current.clear();
    newOwnerProgressRef.current = null;
  }, []);
  const handleCloseCreateForm = useCallback(() => {
    resetCreateProgress();
    handleCloseForm();
  }, [handleCloseForm, resetCreateProgress]);

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

  const handleSave = useCallback(
    async (data: ReservationFormData, selectedPets: Pick<Pet, "id" | "ownerId" | "name">[], newOwnerData?: NewOwnerFormData): Promise<string | null> => {
      if (!data.start || !data.end) return null;
      if (!newOwnerData && selectedPets.length === 0) return null;

      const currentEditing = editingAppointmentRef.current;
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
        const updatePayload = buildReservationUpdateRequest(currentEditing, data, targetDoctor);
        if (!updatePayload) return null;
        return await new Promise<string | null>((resolve) => {
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
      }

      if (newOwnerData) {
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
          await createMutation.mutateAsync(createPayload);
          toast.success("予約を作成しました", {
            description: `${newOwnerData.ownerName}様 / ${newOwnerData.petName} / 担当医: ${targetDoctor}`,
          });
          handleCloseCreateForm();
          navigateBackIfNeeded();
          return null;
        } catch (error) {
          return extractApiErrorMessage(error, "作成");
        }
      }

      try {
        // A selected multi-pet booking is one atomic server-side operation. Do not
        // issue individual creates: the second intentional overlap must not conflict.
        const { pet_id: _petID, owner_id: _ownerID, ...base } = transformToCreateRequest(data, "0", "0");
        await createBatchMutation.mutateAsync({
          ...base,
          pets: selectedPets.map((pet) => ({ pet_id: Number(pet.id), owner_id: Number(pet.ownerId) })),
        });
        toast.success(`${selectedPets.length}件の予約を作成しました`, { description: `担当医: ${targetDoctor}` });
        handleCloseCreateForm();
        navigateBackIfNeeded();
        return null;
      } catch (error) {
        return extractApiErrorMessage(error, "作成");
      }
    },
    [
      checkOverlap,
      createBatchMutation,
      createMutation,
      createMutations,
      editingAppointmentRef,
      handleCloseCreateForm,
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
    resetCreateProgress,
    handleCloseCreateForm,
  };
}
