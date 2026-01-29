import type { RouteObject } from "react-router";
import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";

export const medicalRecordsRoutes: RouteObject[] = [
  {
    path: "/medical-records",
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        lazy: async () => {
          const { MedicalRecords } = await import("@/features/medical-records/routes/MedicalRecords");
          return { Component: MedicalRecords };
        },
      },
      {
        path: "select-pet",
        lazy: async () => {
          const { MedicalRecordPetSelection } = await import("@/features/medical-records/routes/MedicalRecordPetSelection");
          return { Component: MedicalRecordPetSelection };
        },
      },
      {
        path: "new",
        lazy: async () => {
          const { MedicalRecordForm } = await import("@/features/medical-records/routes/MedicalRecordForm");
          return { Component: MedicalRecordForm };
        },
      },
      {
        path: ":id",
        lazy: async () => {
          const { MedicalRecordForm } = await import("@/features/medical-records/routes/MedicalRecordForm");
          return { Component: MedicalRecordForm };
        },
      },
    ],
  },
];
