import { useCallback, useLayoutEffect, useRef, useState } from "react";

import { addHours } from "date-fns";
import { toast } from "sonner";

import { useUpdateReservation } from "@/hooks/use-update-reservation";
import {
  buildJSTWallDateTime,
  formatJSTWallDate,
  formatJSTWallTime,
  jstWallDateToISOString,
} from "@/lib/jst-date";
import type { Pet, Reservation } from "@/types";
import { PetStatusDeceased } from "@/types/generated/models";

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

export function useReceptionModalHandlers({
  advanceStatus,
  cancelAppointment,
  updateAppointment,
  canEditReservation,
  canDeleteReservation,
}: UseReceptionModalHandlersParams) {
  const updateReservationMutation = useUpdateReservation();
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
  const cancelTargetRef = useRef(cancelTarget);
  const permissionsRef = useRef({ canEditReservation, canDeleteReservation });
  useLayoutEffect(() => {
    selectedAppointmentRef.current = selectedAppointment;
  }, [selectedAppointment]);
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
      end: addHours(start, 1),
      visitType: appointment.visitType === "初診" ? "first" : "revisit",
      type: appointment.reservationTypeId,
      doctor: appointment.doctorId || "",
      isDesignated: appointment.isDesignated || false,
      status: preserveEditableStatus(appointment),
      petId: appointment.petId,
      ownerName: appointment.ownerName,
      petName: appointment.petName,
    };

    setEditingAppointment(reservationFormData);
    setEditingAppointmentId(appointment.id);
    setIsEditModalOpen(true);
    setModalOpen(false);
  }, []);

  const handleEditSave = useCallback(
    (data: Partial<Reservation>, selectedPets: Pet[]) => {
      if (permissionsRef.current.canEditReservation !== true) return;
      if (!editingAppointmentId || !data.start) return;
      if (selectedAppointmentRef.current?.petStatus === PetStatusDeceased) return;
      if (selectedPets[0]?.status === "死亡") return;
      const resolvedPetId = selectedPets[0]?.id || data.petId || selectedAppointmentRef.current?.petId;
      const resolvedOwnerId = selectedPets[0]?.ownerId || data.ownerId || selectedAppointmentRef.current?.ownerId;

      const updatedAppointment: ReceptionAppointment = {
        id: editingAppointmentId,
        time: formatJSTWallTime(data.start),
        visitDate: formatJSTWallDate(data.start),
        ownerName: selectedPets[0]?.ownerName || data.ownerName || "",
        petName: selectedPets[0]?.name || data.petName || "",
        petType: selectedPets[0]?.species || "犬",
        visitType: data.visitType === "first" ? "初診" : "再診",
        reservationType: selectedAppointmentRef.current?.reservationType || "診療",
        reservationTypeId: data.type || "",
        reservationCategory: selectedAppointmentRef.current?.reservationCategory || "general",
        doctor: selectedAppointmentRef.current?.doctor,
        doctorId: data.doctor || "",
        isDesignated: data.isDesignated ?? false,
        petId: selectedPets[0]?.id || data.petId || "",
        ownerId: selectedPets[0]?.ownerId || selectedAppointmentRef.current?.ownerId || "",
        status: data.status || (selectedAppointmentRef.current ? preserveEditableStatus(selectedAppointmentRef.current) : "confirmed"),
        notes: data.notes,
        source: selectedAppointmentRef.current?.source || "manual",
        // 予約内容の編集(時間/医師変更等)では checked_in_at は不変。ローカルの楽観更新でも既存値を保持する。
        checkedInAt: selectedAppointmentRef.current?.checkedInAt,
      };

      updateReservationMutation.mutate(
        {
          id: editingAppointmentId,
          req: {
            start_time: jstWallDateToISOString(data.start),
            end_time: jstWallDateToISOString(data.end ?? addHours(data.start, 1)),
            pet_id: optionalNumericID(resolvedPetId),
            owner_id: optionalNumericID(resolvedOwnerId),
            visit_type: data.visitType || "first",
            reservation_type_id: optionalNumericID(data.type),
            doctor_id: optionalNumericID(data.doctor),
            is_designated: data.isDesignated ?? false,
            status: data.status || (selectedAppointmentRef.current ? preserveEditableStatus(selectedAppointmentRef.current) : "confirmed"),
            notes: data.notes,
          },
        },
        {
          onSuccess: () => {
            updateAppointment(updatedAppointment);
            setIsEditModalOpen(false);
            setEditingAppointment(null);
            setEditingAppointmentId(null);
          },
        },
      );
    },
    [editingAppointmentId, updateAppointment, updateReservationMutation],
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
