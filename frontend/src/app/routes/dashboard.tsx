import type { RouteObject } from "react-router";

export const dashboardRoutes: RouteObject[] = [
  {
    path: "/",
    lazy: async () => {
      const { Dashboard } = await import("@/features/dashboard/routes");
      return { Component: Dashboard };
    },
  },
];
