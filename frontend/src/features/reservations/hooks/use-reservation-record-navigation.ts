import { useCallback, useState } from "react";
import type { NavigateFunction } from "react-router";

import { paths } from "@/config/paths";
import { formatJSTWallDate } from "@/lib/jst-date";

import type { Reservation } from "../types";

function isHospitalizationReservation(type: string): boolean {
  return type.includes("入院") || type.includes("ホテル");
}

function resolveRecordPath(reservation: Reservation): string {
  if (reservation.category === "trimming") return "/trimming/new";
  if (isHospitalizationReservation(reservation.type)) return "/hospitalization/new";
  return "/medical-records/new";
}

function resolveSelectPath(reservation: Reservation): string {
  if (reservation.category === "trimming") return "/trimming/select-pet";
  if (isHospitalizationReservation(reservation.type)) return "/hospitalization/select-pet";
  return "/medical-records/select-pet";
}

interface UseReservationRecordNavigationArgs {
  navigate: NavigateFunction;
}

interface PetSelectNavigationState {
  from: string;
  appointmentId: string;
  visitDate: string;
}

export function useReservationRecordNavigation({ navigate }: UseReservationRecordNavigationArgs) {
  const [petSelectConfirmOpen, setPetSelectConfirmOpen] = useState(false);
  const [petSelectPath, setPetSelectPath] = useState("");
  const [petSelectState, setPetSelectState] = useState<PetSelectNavigationState | null>(null);

  const handleCreateRecord = useCallback(
    (reservation: Reservation) => {
      const targetPath = resolveRecordPath(reservation);
      const visitDate = formatJSTWallDate(reservation.start);

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
