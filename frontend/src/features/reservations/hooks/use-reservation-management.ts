import { useState, useCallback, useMemo } from "react";
import { useNavigate, useLocation } from "react-router";
import { addHours, startOfWeek, endOfWeek, startOfMonth, endOfMonth, addDays, format } from "date-fns";
import { ja } from "date-fns/locale";
import type {
  CalendarView,
  Reservation,
  ReservationFormData,
  NavigationState,
} from "../types";
import type { ReservationCreateMutations } from "@/types/reservation-create-mutations";

import { useGetReservations } from "../api/get-reservations";
import {
  useReservationActions,
  type StatusConfirmTarget,
} from "./use-reservation-actions";
import { useReservationModalState } from "./use-reservation-modal-state";
import { useReservationRecordNavigation } from "./use-reservation-record-navigation";

const EMPTY_RESERVATIONS: Reservation[] = [];

interface UseReservationManagementParams {
  currentDate: Date;
  view: CalendarView;
  days: 5 | 7;
  /** #86: 拠点横断表示。複数医院IDのとき clinic_ids をAPIに送信する */
  clinicIds?: string[];
  createMutations: ReservationCreateMutations;
}

export function useReservationManagement({ currentDate, view, days, clinicIds, createMutations }: UseReservationManagementParams) {
  const navigate = useNavigate();
  const location = useLocation();
  const locationFrom = (location.state as NavigationState | null)?.from ?? null;

  // 表示中の週/月の期間を算出し、その範囲の予約だけを取得する。
  // WeekView (startOfWeek から days 日) / MonthView (startOfWeek(monthStart)〜endOfWeek(monthEnd)) の
  // 日付生成ロジックと一致させ、表示範囲の予約が limit から漏れないようにする (BUG #82)。
  const { startDate, endDate } = useMemo(() => {
    if (view === "month") {
      const start = startOfWeek(startOfMonth(currentDate), { locale: ja });
      const end = endOfWeek(endOfMonth(currentDate), { locale: ja });
      return { startDate: format(start, "yyyy-MM-dd"), endDate: format(end, "yyyy-MM-dd") };
    }
    const start = startOfWeek(currentDate, { locale: ja });
    const end = addDays(start, days - 1);
    return { startDate: format(start, "yyyy-MM-dd"), endDate: format(end, "yyyy-MM-dd") };
  }, [currentDate, view, days]);

  const { data: appointments = EMPTY_RESERVATIONS, isLoading } = useGetReservations({ startDate, endDate, clinicIds });

  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Reservation | null>(null);
  const [statusConfirmOpen, setStatusConfirmOpen] = useState(false);
  const [statusConfirmTarget, setStatusConfirmTarget] = useState<StatusConfirmTarget | null>(null);

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
    executeStatusChange,
    executeDelete,
    handleCloseCreateForm,
  } = useReservationActions({
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

  const handleStatusConfirmClose = useCallback(() => {
    setStatusConfirmOpen(false);
    setStatusConfirmTarget(null);
  }, []);

  return {
    // Data
    appointments,
    isLoading,

    // Form modal
    isFormOpen,
    editingAppointment,
    handleOpenForm,
    handleCloseForm: handleCloseCreateForm,
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

    // Status confirm (BUG-020 destructive statuses)
    statusConfirmOpen,
    statusConfirmTarget,
    executeStatusChange,
    handleStatusConfirmClose,

    // Pet select confirm
    petSelectConfirmOpen,
    setPetSelectConfirmOpen,
    handlePetSelectConfirm,
  };
}
