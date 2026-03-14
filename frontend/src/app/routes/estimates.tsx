import type { RouteObject } from "react-router";

export const estimatesRoutes: RouteObject[] = [
  {
    path: "/estimates",
    lazy: async () => {
      const { EstimateList } = await import("@/features/estimates/routes/EstimateList");
      return { Component: EstimateList };
    },
  },
  {
    path: "/estimates/new",
    lazy: async () => {
      const { EstimateForm } = await import("@/features/estimates/routes/EstimateForm");
      return { Component: EstimateForm };
    },
  },
  {
    path: "/estimates/:id",
    lazy: async () => {
      const { EstimateDetail } = await import("@/features/estimates/routes/EstimateDetail");
      return { Component: EstimateDetail };
    },
  },
  {
    path: "/estimates/:id/edit",
    lazy: async () => {
      const { EstimateForm } = await import("@/features/estimates/routes/EstimateForm");
      return { Component: EstimateForm };
    },
  },
];
