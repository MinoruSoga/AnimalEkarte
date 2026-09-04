import { lazy } from "react";

export const ReceptionDetailModal = lazy(() =>
  import("../components/ReceptionDetailModal").then((m) => ({ default: m.ReceptionDetailModal })),
);

export const ReservationFormModal = lazy(() =>
  import("@/components/shared/ReservationFormModal/ReservationFormModal").then((m) => ({
    default: m.ReservationFormModal,
  })),
);
