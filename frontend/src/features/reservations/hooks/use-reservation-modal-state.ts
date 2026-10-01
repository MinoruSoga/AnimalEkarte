import { useCallback, useEffect, useRef, useState } from "react";
import { addHours } from "date-fns";

import {
  buildJSTWallDateTime,
  formatJSTWallDate,
  formatJSTWallTime,
  toJSTWallDate,
} from "@/lib/jst-date";
import { parseDateInput } from "@/components/shared/DatePicker/DatePickerModel";

import type { Reservation, ReservationFormData } from "../types";

interface UseReservationModalStateArgs {
  locationSearch: string;
}

function roundUpToNextQuarterHour(date: Date): Date {
  const rounded = new Date(date);
  rounded.setSeconds(0, 0);
  const minutes = rounded.getMinutes();
  const remainder = minutes % 15;
  if (remainder !== 0) {
    rounded.setMinutes(minutes + (15 - remainder));
  }
  return rounded;
}

export function useReservationModalState({ locationSearch }: UseReservationModalStateArgs) {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingAppointment, setEditingAppointment] = useState<ReservationFormData | null>(null);
  const [isDetailOpen, setIsDetailOpen] = useState(false);
  const [detailAppointment, setDetailAppointment] = useState<Reservation | null>(null);

  const handleOpenForm = useCallback((appointment?: ReservationFormData) => {
    setEditingAppointment(appointment ?? null);
    setIsFormOpen(true);
    setIsDetailOpen(false);
  }, []);

  const handleCloseForm = useCallback(() => {
    setIsFormOpen(false);
    setEditingAppointment(null);
  }, []);

  const handleOpenDetail = useCallback((reservation: Reservation) => {
    setDetailAppointment(reservation);
    setIsDetailOpen(true);
  }, []);

  const handleCloseDetail = useCallback(() => {
    setIsDetailOpen(false);
    setDetailAppointment(null);
  }, []);

  const isFormOpenRef = useRef(false);
  useEffect(() => {
    isFormOpenRef.current = isFormOpen;
  }, [isFormOpen]);

  const editingAppointmentRef = useRef<ReservationFormData | null>(null);
  useEffect(() => {
    editingAppointmentRef.current = editingAppointment;
  }, [editingAppointment]);

  useEffect(() => {
    const searchParams = new URLSearchParams(locationSearch);
    const petId = searchParams.get("petId");
    const isReceptionEntry = searchParams.get("reception") === "1";
    const isNewReservationEntry = searchParams.get("newReservation") === "1";

    if (isReceptionEntry && !isFormOpenRef.current) {
      const start = roundUpToNextQuarterHour(toJSTWallDate(new Date()));

      const stub: ReservationFormData = {
        start,
        end: addHours(start, 1),
        status: "checked_in",
        visitType: "first",
        doctor: "",
        isDesignated: false,
        reservationRoute: "reception",
        source: "manual",
      };
      handleOpenForm(stub);
      return;
    }

    // 受付予約ボードの「新規追加」起点: 通常の新規予約（confirmed → 受付予約カラム）。
    // 受付 walk-in と異なり reservationRoute は強制せずモーダルで選択させる。
    // EMR-243: `&date=YYYY-MM-DD` があれば stub の日付を選択日にする
    // （時刻は現在時刻の15分丸めを維持）。未指定/不正は従来通り当日。
    // reception=1 は当日受付（checked_in）専線のため date を解釈しない。
    if (isNewReservationEntry && !isFormOpenRef.current) {
      const roundedNow = roundUpToNextQuarterHour(toJSTWallDate(new Date()));
      const dateParam = searchParams.get("date");
      const parsedDate = dateParam === null ? null : parseDateInput(dateParam);
      const start =
        parsedDate === null
          ? roundedNow
          : buildJSTWallDateTime(formatJSTWallDate(parsedDate), formatJSTWallTime(roundedNow));

      const stub: ReservationFormData = {
        start,
        end: addHours(start, 1),
        status: "confirmed",
        visitType: "first",
        doctor: "",
        isDesignated: false,
      };
      handleOpenForm(stub);
      return;
    }

    if (petId && !isFormOpenRef.current) {
      const now = toJSTWallDate(new Date());
      now.setMinutes(0, 0, 0);
      const start = addHours(now, 1);

      const stub: ReservationFormData = {
        start,
        end: addHours(start, 1),
        status: "confirmed",
        visitType: "first",
        doctor: "",
        isDesignated: false,
        petId,
      };
      handleOpenForm(stub);
    }
  }, [handleOpenForm, locationSearch]);

  return {
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
  };
}
