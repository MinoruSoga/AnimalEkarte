import type { RouteObject } from "react-router";
import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";

export const vaccinationsRoutes: RouteObject[] = [
  {
    path: "/vaccinations",
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { VaccinationList } = await import("@/features/vaccinations/routes/VaccinationList");
          return { Component: VaccinationList };
        },
      },
      {
        path: "select-pet",
        lazy: async () => {
          const { VaccinationPetSelection } = await import("@/features/vaccinations/routes/VaccinationPetSelection");
          return { Component: VaccinationPetSelection };
        },
      },
      {
        path: "new",
        lazy: async () => {
          const { VaccinationForm } = await import("@/features/vaccinations/routes/VaccinationForm");
          return { Component: VaccinationForm };
        },
      },
      {
        path: ":id",
        lazy: async () => {
          const { VaccinationForm } = await import("@/features/vaccinations/routes/VaccinationForm");
          return { Component: VaccinationForm };
        },
      },
    ],
  },
];
