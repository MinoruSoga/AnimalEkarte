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

interface UseReservationRecordNavigationArgs {
  navigate: NavigateFunction;
}

export function useReservationRecordNavigation({
  navigate,
}: UseReservationRecordNavigationArgs) {
  const [petSelectConfirmOpen, setPetSelectConfirmOpen] = useState(false);
  const [petSelectPath, setPetSelectPath] = useState("");

  const handleCreateRecord = useCallback(
    (reservation: Reservation) => {
      const targetPath = RECORD_PATH[reservation.type] || "/medical-records/new";

      const queryParams = new URLSearchParams();
      if (reservation.petId) {
        queryParams.append("petId", reservation.petId);
      }
      if (reservation.doctorId) {
        queryParams.append("doctorId", reservation.doctorId);
      }

      if (reservation.petId) {
        navigate(`${targetPath}?${queryParams.toString()}`, { state: { from: paths.reservations.getHref() } });
      } else {
        const selectPath = SELECT_PATH[reservation.type] || "/medical-records/select-pet";
        setPetSelectPath(selectPath);
        setPetSelectConfirmOpen(true);
      }
    },
    [navigate],
  );

  const handlePetSelectConfirm = useCallback(() => {
    setPetSelectConfirmOpen(false);
    navigate(petSelectPath, { state: { from: paths.reservations.getHref() } });
  }, [navigate, petSelectPath]);

  return {
    petSelectConfirmOpen,
    setPetSelectConfirmOpen,
    handleCreateRecord,
    handlePetSelectConfirm,
  };
}
