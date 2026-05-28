import { useCallback, useEffect, useRef, useState } from "react";

import { addHours, format } from "date-fns";
import { toast } from "sonner";

import type { Pet, Reservation } from "@/types";

import type { ReceptionAppointment } from "../api/types";

interface UseReceptionModalHandlersParams {
  advanceStatus: (appointment: ReceptionAppointment) => unknown;
  cancelAppointment: (appointmentId: string) => unknown;
  updateAppointment: (appointment: ReceptionAppointment) => unknown;
}

export function useReceptionModalHandlers({
  advanceStatus,
  cancelAppointment,
  updateAppointment,
}: UseReceptionModalHandlersParams) {
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
  useEffect(() => {
    selectedAppointmentRef.current = selectedAppointment;
  }, [selectedAppointment]);

  const handleCardClick = useCallback((appointment: ReceptionAppointment) => {
    setSelectedAppointment(appointment);
    setSelectedAppointmentId(appointment.id);
    setModalOpen(true);
  }, []);

  const handleAdvanceStatus = useCallback(() => {
    if (!selectedAppointmentId) return;
    const appointment = selectedAppointmentRef.current;
    if (!appointment) return;
    advanceStatus(appointment);
    setModalOpen(false);
    setSelectedAppointment(null);
    setSelectedAppointmentId(null);
  }, [selectedAppointmentId, advanceStatus]);

  const handleEditAppointment = useCallback((appointment: ReceptionAppointment) => {
    const now = new Date();
    const [hours, minutes] = appointment.time.split(":").map(Number);
    const start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), hours, minutes);

    const reservationFormData: Partial<Reservation> = {
      id: appointment.id,
      start,
      end: addHours(start, 1),
      status: "confirmed",
      visitType: appointment.visitType === "初診" ? "first" : "revisit",
      type: appointment.reservationType,
      doctor: appointment.doctor || "医師A",
      isDesignated: appointment.isDesignated || false,
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
      if (!editingAppointmentId || !data.start) return;

      const updatedAppointment: ReceptionAppointment = {
        id: editingAppointmentId,
        time: format(data.start, "HH:mm"),
        ownerName: selectedPets[0]?.ownerName || data.ownerName || "",
        petName: selectedPets[0]?.name || data.petName || "",
        petType: selectedPets[0]?.species || "犬",
        visitType: data.visitType === "first" ? "初診" : "再診",
        reservationType: data.type || "診療",
        doctor: data.doctor,
        isDesignated: data.isDesignated ?? false,
        petId: selectedPets[0]?.id || data.petId || "",
        ownerId: "",
        status: "confirmed",
        notes: undefined,
        source: "manual" as const,
      };

      updateAppointment(updatedAppointment);
      setIsEditModalOpen(false);
      setEditingAppointment(null);
      setEditingAppointmentId(null);
    },
    [editingAppointmentId, updateAppointment],
  );

  const handleCancelAppointment = useCallback((appointment: ReceptionAppointment) => {
    setCancelTarget(appointment);
    setCancelTargetId(appointment.id);
    setCancelConfirmOpen(true);
  }, []);

  const executeCancel = useCallback(() => {
    if (!cancelTargetId) return;
    cancelAppointment(cancelTargetId);
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
