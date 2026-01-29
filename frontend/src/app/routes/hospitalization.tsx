import type { RouteObject } from "react-router";
import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";

export const hospitalizationRoutes: RouteObject[] = [
  {
    path: "/hospitalization",
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { HospitalizationList } = await import("@/features/hospitalization/routes/HospitalizationList");
          return { Component: HospitalizationList };
        },
      },
      {
        path: "select-pet",
        lazy: async () => {
          const { HospitalizationPetSelection } = await import("@/features/hospitalization/routes/HospitalizationPetSelection");
          return { Component: HospitalizationPetSelection };
        },
      },
      {
        path: "new",
        lazy: async () => {
          const { HospitalizationForm } = await import("@/features/hospitalization/routes/HospitalizationForm");
          return { Component: HospitalizationForm };
        },
      },
      {
        path: ":id",
        lazy: async () => {
          const { HospitalizationDetail } = await import("@/features/hospitalization/routes/HospitalizationDetail");
          return { Component: HospitalizationDetail };
        },
      },
      {
        path: ":id/edit",
        lazy: async () => {
          const { HospitalizationForm } = await import("@/features/hospitalization/routes/HospitalizationForm");
          return { Component: HospitalizationForm };
        },
      },
    ],
  },
];
