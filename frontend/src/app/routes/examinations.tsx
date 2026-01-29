import type { RouteObject } from "react-router";
import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";

export const examinationsRoutes: RouteObject[] = [
  {
    path: "/examinations",
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { Examinations } = await import("@/features/examinations/routes/Examinations");
          return { Component: Examinations };
        },
      },
      {
        path: "select-pet",
        lazy: async () => {
          const { ExaminationPetSelection } = await import("@/features/examinations/routes/ExaminationPetSelection");
          return { Component: ExaminationPetSelection };
        },
      },
      {
        path: "new",
        lazy: async () => {
          const { ExaminationForm } = await import("@/features/examinations/routes/ExaminationForm");
          return { Component: ExaminationForm };
        },
      },
      {
        path: ":id",
        lazy: async () => {
          const { ExaminationForm } = await import("@/features/examinations/routes/ExaminationForm");
          return { Component: ExaminationForm };
        },
      },
    ],
  },
];
