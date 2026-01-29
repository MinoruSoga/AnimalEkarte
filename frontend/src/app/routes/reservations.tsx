import type { RouteObject } from "react-router";

export const reservationsRoutes: RouteObject[] = [
  {
    path: "/reservations",
    lazy: async () => {
      const { ReservationManagement } = await import("@/features/reservations/routes/ReservationManagement");
      return { Component: ReservationManagement };
    },
  },
];
