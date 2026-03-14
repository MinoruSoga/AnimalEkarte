import type { RouteObject } from "react-router";

export const shiftsRoutes: RouteObject[] = [
  {
    path: "/shifts",
    lazy: async () => {
      const { ShiftCalendarPage } = await import("@/features/shifts/routes/ShiftCalendarPage");
      return { Component: ShiftCalendarPage };
    },
  },
];
