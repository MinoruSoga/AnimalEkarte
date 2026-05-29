import { useCallback, useState } from "react";
import type { NavigateFunction } from "react-router";

import { paths } from "@/config/paths";

import type { Reservation } from "../types";

const RECORD_PATH: Record<string, string> = {
  "トリミング": "/trimming/new",
  "入院": "/hospitalization/new",
  "ホテル": "/hospitalization/new",
};

const SELECT_PATH: Record<string, string> = {
  "トリミング": "/trimming/select-pet",
  "入院": "/hospitalization/select-pet",
  "ホテル": "/hospitalization/select-pet",
};

const JST_OFFSET_MS = 9 * 60 * 60 * 1000;

function formatJSTDate(date: Date): string {
  const jstDate = new Date(date.getTime() + JST_OFFSET_MS);
  const year = jstDate.getUTCFullYear();
  const month = String(jstDate.getUTCMonth() + 1).padStart(2, "0");
  const day = String(jstDate.getUTCDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function resolveRecordPath(reservation: Reservation): string {
  if (reservation.category === "trimming") return "/trimming/new";
  return RECORD_PATH[reservation.type] || "/medical-records/new";
}

function resolveSelectPath(reservation: Reservation): string {
  if (reservation.category === "trimming") return "/trimming/select-pet";
  return SELECT_PATH[reservation.type] || "/medical-records/select-pet";
}

interface UseReservationRecordNavigationArgs {
  navigate: NavigateFunction;
}

interface PetSelectNavigationState {
  from: string;
  appointmentId: string;
  visitDate: string;
}

export function useReservationRecordNavigation({
  navigate,
}: UseReservationRecordNavigationArgs) {
  const [petSelectConfirmOpen, setPetSelectConfirmOpen] = useState(false);
  const [petSelectPath, setPetSelectPath] = useState("");
  const [petSelectState, setPetSelectState] = useState<PetSelectNavigationState | null>(null);

  const handleCreateRecord = useCallback(
    (reservation: Reservation) => {
      const targetPath = resolveRecordPath(reservation);
      const visitDate = formatJSTDate(reservation.start);

      const queryParams = new URLSearchParams();
      if (reservation.petId) {
        queryParams.append("petId", reservation.petId);
      }
      if (reservation.doctorId) {
        queryParams.append("doctorId", reservation.doctorId);
      }
      queryParams.append("appointmentId", reservation.id);
      queryParams.append("visitDate", visitDate);

      const navigationState = {
        from: paths.reservations.getHref(),
        appointmentId: reservation.id,
        visitDate,
      };

      if (reservation.petId) {
        navigate(`${targetPath}?${queryParams.toString()}`, { state: navigationState });
      } else {
        const selectPath = resolveSelectPath(reservation);
        setPetSelectPath(`${selectPath}?${queryParams.toString()}`);
        setPetSelectState(navigationState);
        setPetSelectConfirmOpen(true);
      }
    },
    [navigate],
  );

  const handlePetSelectConfirm = useCallback(() => {
    setPetSelectConfirmOpen(false);
    navigate(petSelectPath, { state: petSelectState ?? { from: paths.reservations.getHref() } });
  }, [navigate, petSelectPath, petSelectState]);

  return {
    petSelectConfirmOpen,
    setPetSelectConfirmOpen,
    handleCreateRecord,
    handlePetSelectConfirm,
  };
}
