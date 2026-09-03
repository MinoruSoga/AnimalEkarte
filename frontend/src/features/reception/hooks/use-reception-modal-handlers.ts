import { useCallback, useLayoutEffect, useRef, useState } from "react";

import { toast } from "sonner";

import {
  useUpdateReservation,
  type UpdateReservationRequest,
} from "@/hooks/use-update-reservation";
import {
  buildJSTWallDateTime,
  formatJSTWallDate,
  formatJSTWallTime,
  jstWallDateToISOString,
} from "@/lib/jst-date";
import type { Pet, Reservation } from "@/types";
import {
  DangerLevelHigh,
  DangerLevelLow,
  DangerLevelMedium,
  PetStatusAlive,
  PetStatusDeceased,
  type DangerLevel,
  type PetStatus,
} from "@/types/generated/models";

import type { ReceptionAppointment } from "../api/types";

interface UseReceptionModalHandlersParams {
  advanceStatus: (appointment: ReceptionAppointment) => unknown;
  cancelAppointment: (appointmentId: string) => Promise<boolean>;
  updateAppointment: (appointment: ReceptionAppointment) => unknown;
  canEditReservation?: boolean;
  canDeleteReservation?: boolean;
}

function optionalNumericID(value: string | undefined): number | undefined {
  if (!value) return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function preserveEditableStatus(appointment: ReceptionAppointment): Reservation["status"] {
  return appointment.status === "pending" ? "confirmed" : appointment.status;
}

function toPetStatus(status: Pet["status"]): PetStatus | undefined {
  switch (status) {
    case "生存":
      return PetStatusAlive;
    case "死亡":
      return PetStatusDeceased;
    default:
      return undefined;
  }
}

function toPetDangerLevel(dangerLevel: Pet["dangerLevel"]): DangerLevel | undefined {
  switch (dangerLevel) {
    case "低":
      return DangerLevelLow;
    case "中":
      return DangerLevelMedium;
    case "高":
      return DangerLevelHigh;
    default:
      return undefined;
  }
}

/** Notes-only edits must omit schedule/doctor so BE skips on-duty conflict checks. */
function buildReceptionReservationUpdateRequest(
  current: Partial<Reservation>,
  data: Partial<Reservation>,
): UpdateReservationRequest {
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
  const nextType = optionalNumericID(data.type);
  const prevType = optionalNumericID(current.type);
  if (nextType !== undefined && nextType !== prevType) req.reservation_type_id = nextType;
  const nextDoctor = optionalNumericID(data.doctor);
  const prevDoctor = optionalNumericID(current.doctor);
  if (nextDoctor !== undefined && nextDoctor !== prevDoctor) req.doctor_id = nextDoctor;
  if ((data.isDesignated ?? false) !== (current.isDesignated ?? false)) {
    req.is_designated = data.isDesignated ?? false;
  }
  if ((data.notes ?? "") !== (current.notes ?? "")) req.notes = data.notes;

  if (Object.keys(req).length === 0) {
    req.notes = data.notes ?? "";
  }
  return req;
}

export function useReceptionModalHandlers({
  advanceStatus,
  cancelAppointment,
  updateAppointment,
  canEditReservation,
  canDeleteReservation,
}: UseReceptionModalHandlersParams) {
  const updateReservationMutation = useUpdateReservation();
  const { mutate } = updateReservationMutation;
  const [modalOpen, setModalOpen] = useState(false);
  const [selectedAppointment, setSelectedAppointment] = useState<ReceptionAppointment | null>(null);
  const [selectedAppointmentId, setSelectedAppointmentId] = useState<string | null>(null);

  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [editingAppointment, setEditingAppointment] = useState<Partial<Reservation> | null>(null);
  const [editingAppointmentId, setEditingAppointmentId] = useState<string | null>(null);

  const [cancelConfirmOpen, setCancelConfirmOpen] = useState(false);
  const [cancelTarget, setCancelTarget] = useState<ReceptionAppointment | null>(null);
  const [cancelTargetId, setCancelTargetId] = useState<string | null>(null);

  const selectedAppointmentRef = useRef(selectedAppointment);
  const editingAppointmentRef = useRef(editingAppointment);
  const cancelTargetRef = useRef(cancelTarget);
  const permissionsRef = useRef({ canEditReservation, canDeleteReservation });
  useLayoutEffect(() => {
    selectedAppointmentRef.current = selectedAppointment;
  }, [selectedAppointment]);
  useLayoutEffect(() => {
    editingAppointmentRef.current = editingAppointment;
  }, [editingAppointment]);
  useLayoutEffect(() => {
    cancelTargetRef.current = cancelTarget;
  }, [cancelTarget]);
  useLayoutEffect(() => {
    permissionsRef.current = { canEditReservation, canDeleteReservation };
  }, [canEditReservation, canDeleteReservation]);

  const handleCardClick = useCallback((appointment: ReceptionAppointment) => {
    if (appointment.petStatus === PetStatusDeceased) return;
    setSelectedAppointment(appointment);
    setSelectedAppointmentId(appointment.id);
    setModalOpen(true);
  }, []);

  const handleAdvanceStatus = useCallback(() => {
    if (permissionsRef.current.canEditReservation !== true) return;
    if (!selectedAppointmentId) return;
    const appointment = selectedAppointmentRef.current;
    if (!appointment || appointment.petStatus === PetStatusDeceased) return;
    advanceStatus(appointment);
    setModalOpen(false);
    setSelectedAppointment(null);
    setSelectedAppointmentId(null);
  }, [selectedAppointmentId, advanceStatus]);

  const handleEditAppointment = useCallback((appointment: ReceptionAppointment) => {
    if (permissionsRef.current.canEditReservation !== true) return;
    if (appointment.petStatus === PetStatusDeceased) return;
    selectedAppointmentRef.current = appointment;
    const start = buildJSTWallDateTime(appointment.visitDate, appointment.time);

    const reservationFormData: Partial<Reservation> = {
      id: appointment.id,
      start,
      end: appointment.end,
      visitType: appointment.visitType === "初診" ? "first" : "revisit",
      type: appointment.reservationTypeId,
      doctor: appointment.doctorId || "",
      isDesignated: appointment.isDesignated || false,
      status: preserveEditableStatus(appointment),
      petId: appointment.petId,
      ownerId: appointment.ownerId,
      ownerName: appointment.ownerName,
      petName: appointment.petName,
      notes: appointment.notes,
    };

    editingAppointmentRef.current = reservationFormData;
    setEditingAppointment(reservationFormData);
    setEditingAppointmentId(appointment.id);
    setIsEditModalOpen(true);
    setModalOpen(false);
  }, []);

  const handleEditSave = useCallback(
    (data: Partial<Reservation>, selectedPets: Pet[]) => {
      const selectedAppointmentSnapshot = selectedAppointmentRef.current;
      const selectedPetSnapshot = selectedPets[0];
      if (permissionsRef.current.canEditReservation !== true) return;
      if (!editingAppointmentId || !data.start) return;
      if (selectedAppointmentSnapshot?.petStatus === PetStatusDeceased) return;
      if (selectedPetSnapshot?.status === "死亡") return;
      const resolvedPetId =
        selectedPetSnapshot?.id || data.petId || selectedAppointmentSnapshot?.petId;
      const resolvedOwnerId =
        selectedPetSnapshot?.ownerId || data.ownerId || selectedAppointmentSnapshot?.ownerId;
      const selectedPetDangerLevel = toPetDangerLevel(selectedPetSnapshot?.dangerLevel);
      const petSentinelFields: {
        petStatus?: PetStatus;
        petDangerLevel?: DangerLevel;
        petDangerReason?: string;
      } = selectedPetSnapshot
        ? {
            petStatus: toPetStatus(selectedPetSnapshot.status),
            petDangerLevel: selectedPetDangerLevel,
            petDangerReason: selectedPetSnapshot.dangerReason,
          }
        : {
            petStatus: selectedAppointmentSnapshot?.petStatus,
            petDangerLevel: selectedAppointmentSnapshot?.petDangerLevel,
            petDangerReason: selectedAppointmentSnapshot?.petDangerReason,
          };
      const shouldUpdateAppointmentLocally =
        !selectedPetSnapshot || selectedPetDangerLevel !== undefined;

      const current = editingAppointmentRef.current;
      if (!current) return;
      const nextEnd = data.end ?? selectedAppointmentSnapshot?.end ?? current.end;
      if (nextEnd === undefined) return;

      const req = buildReceptionReservationUpdateRequest(current, data);
      req.status =
        data.status ||
        (selectedAppointmentSnapshot
          ? preserveEditableStatus(selectedAppointmentSnapshot)
          : "confirmed");
      const nextPetId = optionalNumericID(resolvedPetId);
      const nextOwnerId = optionalNumericID(resolvedOwnerId);
      if (nextPetId !== undefined) req.pet_id = nextPetId;
      if (nextOwnerId !== undefined) req.owner_id = nextOwnerId;

      const updatedAppointment: ReceptionAppointment = {
        id: editingAppointmentId,
        time: formatJSTWallTime(data.start),
        visitDate: formatJSTWallDate(data.start),
        end: nextEnd,
        ownerName: selectedPetSnapshot?.ownerName || data.ownerName || "",
        petName: selectedPetSnapshot?.name || data.petName || "",
        petType: selectedPetSnapshot?.species || "犬",
        ...petSentinelFields,
        visitType: data.visitType === "first" ? "初診" : "再診",
        reservationType: selectedAppointmentSnapshot?.reservationType || "診療",
        reservationTypeId: data.type || "",
        reservationCategory: selectedAppointmentSnapshot?.reservationCategory || "general",
        doctor: selectedAppointmentSnapshot?.doctor,
        doctorId: data.doctor || "",
        isDesignated: data.isDesignated ?? false,
        petId: selectedPetSnapshot?.id || data.petId || "",
        ownerId: selectedPetSnapshot?.ownerId || selectedAppointmentSnapshot?.ownerId || "",
        status:
          data.status ||
          (selectedAppointmentSnapshot
            ? preserveEditableStatus(selectedAppointmentSnapshot)
            : "confirmed"),
        notes: data.notes,
        source: selectedAppointmentSnapshot?.source || "manual",
        // 予約内容の編集(時間/医師変更等)では checked_in_at は不変。ローカルの楽観更新でも既存値を保持する。
        checkedInAt: selectedAppointmentSnapshot?.checkedInAt,
      };

      mutate(
        {
          id: editingAppointmentId,
          req,
        },
        {
          onSuccess: () => {
            if (shouldUpdateAppointmentLocally) {
              updateAppointment(updatedAppointment);
            }
            setIsEditModalOpen(false);
            editingAppointmentRef.current = null;
            setEditingAppointment(null);
            setEditingAppointmentId(null);
          },
        },
      );
    },
    [editingAppointmentId, updateAppointment, mutate],
  );

  const handleCancelAppointment = useCallback((appointment: ReceptionAppointment) => {
    if (permissionsRef.current.canDeleteReservation !== true) return;
    if (appointment.petStatus === PetStatusDeceased) return;
    cancelTargetRef.current = appointment;
    setCancelTarget(appointment);
    setCancelTargetId(appointment.id);
    setCancelConfirmOpen(true);
  }, []);

  const executeCancel = useCallback(async (): Promise<void> => {
    if (permissionsRef.current.canDeleteReservation !== true) return;
    if (!cancelTargetId) return;
    if (cancelTargetRef.current?.petStatus === PetStatusDeceased) return;
    const succeeded = await cancelAppointment(cancelTargetId);
    if (!succeeded) return;
    toast.success("予約を取り消しました");
    setModalOpen(false);
    setCancelConfirmOpen(false);
    setCancelTarget(null);
    setCancelTargetId(null);
  }, [cancelTargetId, cancelAppointment]);

  const closeEditModal = useCallback(() => {
    setIsEditModalOpen(false);
    editingAppointmentRef.current = null;
    setEditingAppointment(null);
  }, []);

  const closeCancelConfirm = useCallback(() => {
    setCancelConfirmOpen(false);
    setCancelTarget(null);
  }, []);

  return {
    cancelConfirmOpen,
    cancelTarget,
    closeCancelConfirm,
    closeEditModal,
    editingAppointment,
    executeCancel,
    handleAdvanceStatus,
    handleCancelAppointment,
    handleCardClick,
    handleEditAppointment,
    handleEditSave,
    isEditModalOpen,
    modalOpen,
    selectedAppointment,
    setModalOpen,
  };
}
