import { useState, useCallback } from "react";
import { useNavigate, useLocation } from "react-router";
import { addHours } from "date-fns";
import type {
  Reservation,
  ReservationFormData,
  NavigationState,
} from "../types";

import { useGetReservations } from "../api/get-reservations";
import { useReservationActions } from "./use-reservation-actions";
import { useReservationModalState } from "./use-reservation-modal-state";
import { useReservationRecordNavigation } from "./use-reservation-record-navigation";

const EMPTY_RESERVATIONS: Reservation[] = [];

export function useReservationManagement() {
  const navigate = useNavigate();
  const location = useLocation();
  const locationFrom = (location.state as NavigationState | null)?.from ?? null;

  const { data: appointments = EMPTY_RESERVATIONS, isLoading } = useGetReservations();

  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Reservation | null>(null);

  const {
    isFormOpen,
    editingAppointment,
    editingAppointmentRef,
    handleOpenForm,
    handleCloseForm,
    isDetailOpen,
    detailAppointment,
    setDetailAppointment,
    handleOpenDetail,
    handleCloseDetail,
  } = useReservationModalState({ locationSearch: location.search });

  const handleTimeSlotClick = useCallback(
    (date: Date) => {
      const stub: ReservationFormData = {
        start: date,
        end: addHours(date, 1),
        status: "confirmed",
        visitType: "first",
        doctor: "",
        isDesignated: false,
      };
      handleOpenForm(stub);
    },
    [handleOpenForm]
  );

  const {
    handleSave,
    handleReservationUpdate,
    handleStatusChange,
    executeDelete,
  } = useReservationActions({
    appointments,
    editingAppointmentRef,
    deleteTarget,
    setDeleteConfirmOpen,
    setDeleteTarget,
    setDetailAppointment,
    handleCloseForm,
    handleCloseDetail,
    locationFrom,
    navigate,
  });

  const handleDelete = useCallback((reservation: Reservation) => {
    setDeleteTarget(reservation);
    setDeleteConfirmOpen(true);
  }, []);

  const {
    petSelectConfirmOpen,
    setPetSelectConfirmOpen,
    handleCreateRecord,
    handlePetSelectConfirm,
  } = useReservationRecordNavigation({ navigate });

  const handleDeleteConfirmClose = useCallback(() => {
    setDeleteConfirmOpen(false);
    setDeleteTarget(null);
  }, []);

  return {
    // Data
    appointments,
    isLoading,

    // Form modal
    isFormOpen,
    editingAppointment,
    handleOpenForm,
    handleCloseForm,
    handleSave,

    // Detail modal
    isDetailOpen,
    detailAppointment,
    handleOpenDetail,
    handleCloseDetail,
    handleStatusChange,
    handleDelete,
    handleCreateRecord,

    // Calendar interactions
    handleTimeSlotClick,
    handleReservationUpdate,

    // Delete confirm
    deleteConfirmOpen,
    deleteTarget,
    executeDelete,
    handleDeleteConfirmClose,

    // Pet select confirm
    petSelectConfirmOpen,
    setPetSelectConfirmOpen,
    handlePetSelectConfirm,
  };
}
