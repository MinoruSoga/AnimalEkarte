import { Outlet, type RouteObject } from "react-router";

import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { RequirePermission } from "@/components/shared/RequirePermission";
import {
  ResourceCheckups,
  ResourceEstimates,
  ResourceExaminations,
  ResourceHospitalization,
  ResourceInventory,
  ResourceMedicalRecords,
  ResourceOwners,
  ResourceReception,
  ResourceReservations,
  ResourceShifts,
  ResourceTrimming,
  ResourceVaccinations,
} from "@/types/generated/models";

export const clinicalRoutes: RouteObject[] = [
// ── Reception（当日の受付）────────────────────────────────────
      {
        path: "/",
        element: (
          <RequirePermission resource={ResourceReception}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { Reception } = await import("@/features/reception");
              return { Component: Reception };
            },
          },

        ],
      },

      // ── Owners ───────────────────────────────────────────────────
      {
        path: "/owners",
        element: (
          <RequirePermission resource={ResourceOwners}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const [{ OwnersListPage }, { ownersLoader }] = await Promise.all([
                import("@/app/pages/OwnersListPage"),
                import("@/features/owners"),
              ]);
              return { Component: OwnersListPage, loader: ownersLoader };
            },
          },
          {
            // BUG-020: create 権限がないユーザーは新規作成フォームへアクセス不可
            path: "new",
            element: (
              <RequirePermission resource={ResourceOwners} action="create">
                <Outlet />
              </RequirePermission>
            ),
            children: [
              {
                index: true,
                lazy: async () => {
                  const { OwnerFormPage } = await import("@/app/pages/OwnerFormPage");
                  return { Component: OwnerFormPage };
                },
              },
            ],
          },
          {
            path: ":id",
            lazy: async () => {
              const [{ OwnerFormPage }, { ownerLoader }] = await Promise.all([
                import("@/app/pages/OwnerFormPage"),
                import("@/features/owners"),
              ]);
              return { Component: OwnerFormPage, loader: ownerLoader };
            },
          },
        ],
      },

      // ── Aggregation ──────────────────────────────────────────────
      {
        path: "/aggregation",
        element: (
          <RequirePermission resource={ResourceOwners}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        lazy: async () => {
          const { AggregationDashboardPage } = await import("@/features/aggregation");
          return { Component: AggregationDashboardPage };
        },
      },

// ── Reservations ─────────────────────────────────────────────
{
  path: "/reservations",
  element: (
    <RequirePermission resource={ResourceReservations}>
      <Outlet />
    </RequirePermission>
  ),
  errorElement: <RouteErrorBoundary />,
  children: [
    {
      index: true,
      lazy: async () => {
        const { ReservationManagement } = await import("@/features/reservations");
        return { Component: ReservationManagement };
      },
    },
  ],
},


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
            lazy: async () => {
              const { CheckupPetSelection } = await import("@/features/checkups");
              return { Component: CheckupPetSelection };
            },
          },
          {
            path: "new",
            lazy: async () => {
              const { CheckupForm } = await import("@/features/checkups");
              return { Component: CheckupForm };
            },
          },
        ],
      },

      // ── Inventory ────────────────────────────────────────────────
      {
        path: "/inventory",
        element: (
          <RequirePermission resource={ResourceInventory}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { InventoryList } = await import("@/features/inventory");
              return { Component: InventoryList };
            },
          },
          {
            // BUG-020: create 権限ガード
            path: "new",
            element: (
              <RequirePermission resource={ResourceInventory} action="create">
                <Outlet />
              </RequirePermission>
            ),
            children: [
              {
                index: true,
                lazy: async () => {
                  const { InventoryForm } = await import("@/features/inventory");
                  return { Component: InventoryForm };
                },
              },
            ],
          },
          {
            path: ":id",
            lazy: async () => {
              const { InventoryForm } = await import("@/features/inventory");
              return { Component: InventoryForm };
            },
          },
        ],
      },

      // ── Estimates ────────────────────────────────────────────────
      {
        path: "/estimates",
        element: (
          <RequirePermission resource={ResourceEstimates}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { EstimateList } = await import("@/features/estimates");
              return { Component: EstimateList };
            },
          },
          {
            // BUG-020: create 権限ガード
            path: "new",
            element: (
              <RequirePermission resource={ResourceEstimates} action="create">
                <Outlet />
              </RequirePermission>
            ),
            children: [
              {
                index: true,
                lazy: async () => {
                  const { EstimateForm } = await import("@/features/estimates");
                  return { Component: EstimateForm };
                },
              },
            ],
          },
          {
            path: ":id",
            lazy: async () => {
              const { EstimateDetail } = await import("@/features/estimates");
              return { Component: EstimateDetail };
            },
          },
          {
            // BUG-020: edit 権限ガード
            path: ":id/edit",
            element: (
              <RequirePermission resource={ResourceEstimates} action="edit">
                <Outlet />
              </RequirePermission>
            ),
            children: [
              {
                index: true,
                lazy: async () => {
                  const { EstimateForm } = await import("@/features/estimates");
                  return { Component: EstimateForm };
                },
              },
            ],
          },
        ],
      },

      // ── Shifts ───────────────────────────────────────────────────
      {
        path: "/shifts",
        element: (
          <RequirePermission resource={ResourceShifts}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { ShiftCalendarPage } = await import("@/features/shifts");
              return { Component: ShiftCalendarPage };
            },
          },
        ],
      },

      ];
