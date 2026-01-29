import type { RouteObject } from "react-router";
import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";

export const trimmingRoutes: RouteObject[] = [
  {
    path: "/trimming",
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { TrimmingList } = await import("@/features/trimming/routes/TrimmingList");
          return { Component: TrimmingList };
        },
      },
      {
        path: "select-pet",
        lazy: async () => {
          const { TrimmingPetSelection } = await import("@/features/trimming/routes/TrimmingPetSelection");
          return { Component: TrimmingPetSelection };
        },
      },
      {
        path: "new",
        lazy: async () => {
          const { TrimmingForm } = await import("@/features/trimming/routes/TrimmingForm");
          return { Component: TrimmingForm };
        },
      },
      {
        path: ":id",
        lazy: async () => {
          const { TrimmingForm } = await import("@/features/trimming/routes/TrimmingForm");
          return { Component: TrimmingForm };
        },
      },
    ],
  },
];
