import { Outlet, type RouteObject } from "react-router";

import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { RequirePermission } from "@/components/shared/RequirePermission";
import {
  ResourceCheckups,
  ResourceExaminations,
  ResourceHospitalization,
  ResourceMedicalRecords,
  ResourceTrimming,
  ResourceVaccinations,
} from "@/types/generated/models";

export const clinicalCareRoutes: RouteObject[] = [
// ── Medical Records ──────────────────────────────────────────
      {
        path: "/medical-records",
        element: (
          <RequirePermission resource={ResourceMedicalRecords}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { MedicalRecords } = await import("@/features/medical-records");
              return { Component: MedicalRecords };
            },
          },
          {
            path: "select-pet",
            element: <RequirePermission resource={ResourceMedicalRecords} action="create"><Outlet /></RequirePermission>,
            children: [{
              index: true,
              lazy: async () => {
                const { MedicalRecordPetSelection } = await import("@/features/medical-records");
                return { Component: MedicalRecordPetSelection };
              },
            }],
          },
          {
            // BUG-020: create 権限ガード
            path: "new",
            element: (
              <RequirePermission resource={ResourceMedicalRecords} action="create">
                <Outlet />
              </RequirePermission>
            ),
            children: [
              {
                index: true,
                lazy: async () => {
                  const { MedicalRecordForm } = await import("@/features/medical-records");
                  return { Component: MedicalRecordForm };
                },
              },
            ],
          },
          {
            path: ":id",
            lazy: async () => {
              const { MedicalRecordForm } = await import("@/features/medical-records");
              return { Component: MedicalRecordForm };
            },
          },
        ],
      },

      // ── Hospitalization ──────────────────────────────────────────
      {
        path: "/hospitalization",
        element: (
          <RequirePermission resource={ResourceHospitalization}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { HospitalizationList } = await import("@/features/hospitalization");
              return { Component: HospitalizationList };
            },
          },
          {
            path: "select-pet",
            element: <RequirePermission resource={ResourceHospitalization} action="create"><Outlet /></RequirePermission>,
            children: [{
              index: true,
              lazy: async () => {
                const { HospitalizationPetSelection } = await import("@/features/hospitalization");
                return { Component: HospitalizationPetSelection };
              },
            }],
          },
          {
            // BUG-020: create 権限ガード
            path: "new",
            element: (
              <RequirePermission resource={ResourceHospitalization} action="create">
                <Outlet />
              </RequirePermission>
            ),
            children: [
              {
                index: true,
                lazy: async () => {
                  const { HospitalizationForm } = await import("@/features/hospitalization");
                  return { Component: HospitalizationForm };
                },
              },
            ],
          },
          {
            path: ":id",
            lazy: async () => {
              const { HospitalizationDetail } = await import("@/features/hospitalization");
              return { Component: HospitalizationDetail };
            },
          },
          {
            // BUG-020: edit 権限ガード
            path: ":id/edit",
            element: (
              <RequirePermission resource={ResourceHospitalization} action="edit">
                <Outlet />
              </RequirePermission>
            ),
            children: [
              {
                index: true,
                lazy: async () => {
                  const { HospitalizationForm } = await import("@/features/hospitalization");
                  return { Component: HospitalizationForm };
                },
              },
            ],
          },
        ],
      },

      // ── Trimming ─────────────────────────────────────────────────
      {
        path: "/trimming",
        element: (
          <RequirePermission resource={ResourceTrimming}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { TrimmingList } = await import("@/features/trimming");
              return { Component: TrimmingList };
            },
          },
          {
            path: "select-pet",
            element: <RequirePermission resource={ResourceTrimming} action="create"><Outlet /></RequirePermission>,
            children: [{
              index: true,
              lazy: async () => {
                const { TrimmingPetSelection } = await import("@/features/trimming");
                return { Component: TrimmingPetSelection };
              },
            }],
          },
          {
            // BUG-020: create 権限ガード
            path: "new",
            element: (
              <RequirePermission resource={ResourceTrimming} action="create">
                <Outlet />
              </RequirePermission>
            ),
            children: [
              {
                index: true,
                lazy: async () => {
                  const { TrimmingForm } = await import("@/features/trimming");
                  return { Component: TrimmingForm };
                },
              },
            ],
          },
          {
            path: ":id",
            lazy: async () => {
              const { TrimmingForm } = await import("@/features/trimming");
              return { Component: TrimmingForm };
            },
          },
        ],
      },

      // ── Examinations ─────────────────────────────────────────────
      {
        path: "/examinations",
        element: (
          <RequirePermission resource={ResourceExaminations}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { ExaminationsList } = await import("@/features/examinations");
              return { Component: ExaminationsList };
            },
          },
          {
            path: "select-pet",
            element: <RequirePermission resource={ResourceExaminations} action="create"><Outlet /></RequirePermission>,
            children: [{
              index: true,
              lazy: async () => {
                const { ExaminationPetSelection } = await import("@/features/examinations");
                return { Component: ExaminationPetSelection };
              },
            }],
          },
          {
            // BUG-020: create 権限ガード
            path: "new",
            element: (
              <RequirePermission resource={ResourceExaminations} action="create">
                <Outlet />
              </RequirePermission>
            ),
            children: [
              {
                index: true,
                lazy: async () => {
                  const { ExaminationForm } = await import("@/features/examinations");
                  return { Component: ExaminationForm };
                },
              },
            ],
          },
          {
            path: ":id",
            lazy: async () => {
              const { ExaminationForm } = await import("@/features/examinations");
              return { Component: ExaminationForm };
            },
          },
        ],
      },

      // ── Vaccinations ─────────────────────────────────────────────
      {
        path: "/vaccinations",
        element: (
          <RequirePermission resource={ResourceVaccinations}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { VaccinationList } = await import("@/features/vaccinations");
              return { Component: VaccinationList };
            },
          },
          {
            path: "select-pet",
            element: <RequirePermission resource={ResourceVaccinations} action="create"><Outlet /></RequirePermission>,
            children: [{
              index: true,
              lazy: async () => {
                const { VaccinationPetSelection } = await import("@/features/vaccinations");
                return { Component: VaccinationPetSelection };
              },
            }],
          },
          {
            // BUG-020: create 権限ガード
            path: "new",
            element: (
              <RequirePermission resource={ResourceVaccinations} action="create">
                <Outlet />
              </RequirePermission>
            ),
            children: [
              {
                index: true,
                lazy: async () => {
                  const { VaccinationForm } = await import("@/features/vaccinations");
                  return { Component: VaccinationForm };
                },
              },
            ],
          },
          {
            path: ":id",
            lazy: async () => {
              const { VaccinationForm } = await import("@/features/vaccinations");
              return { Component: VaccinationForm };
            },
          },
        ],
      },

      // ── Checkups ─────────────────────────────────────────────────
      {
        path: "/checkups",
        element: (
          <RequirePermission resource={ResourceCheckups}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { CheckupsList } = await import("@/features/checkups");
              return { Component: CheckupsList };
            },
          },
          {
            path: "select-pet",
            element: (
              <RequirePermission resource={ResourceMedicalRecords} action="create">
                <RequirePermission resource={ResourceMedicalRecords} action="edit">
                  <Outlet />
                </RequirePermission>
              </RequirePermission>
            ),
            children: [{
              index: true,
              lazy: async () => {
                const { CheckupPetSelection } = await import("@/features/checkups");
                return { Component: CheckupPetSelection };
              },
            }],
          },
          {
            path: "new",
            element: (
              <RequirePermission resource={ResourceMedicalRecords} action="create">
                <RequirePermission resource={ResourceMedicalRecords} action="edit">
                  <Outlet />
                </RequirePermission>
              </RequirePermission>
            ),
            children: [{
              index: true,
              lazy: async () => {
                const { CheckupForm } = await import("@/features/checkups");
                return { Component: CheckupForm };
              },
            }],
          },
        ],
      },
];
