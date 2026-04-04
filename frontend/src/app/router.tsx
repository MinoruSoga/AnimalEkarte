import { lazy, Suspense } from "react";
import { createBrowserRouter, Outlet } from "react-router";

import { Layout } from "@/components/shared/Layout/Layout";
import { RootErrorBoundary, RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { RequirePermission } from "@/components/shared/RequirePermission";
import { AuthProvider } from "@/features/auth";
import { ResourceDashboard, ResourceOwners, ResourceReservations, ResourceMedicalRecords, ResourceHospitalization, ResourceTrimming, ResourceExaminations, ResourceAccounting, ResourceVaccinations, ResourceCheckups, ResourceInventory, ResourceEstimates, ResourceShifts, ResourceMaster, ResourceHospitalSettings } from "@/types/generated/models";

/* bundle-dynamic-imports: ログインページは未認証ユーザー専用。認証済みユーザーのバンドルに含めない */
const Login = lazy(() =>
  import("@/features/auth/routes/Login").then((m) => ({ default: m.Login })),
);

export const router = createBrowserRouter([
  {
    // AuthProvider をアプリ全体に配置。
    // これにより /login でも useAuth() が使用可能になり、
    // LoginForm で login() を直接呼び出してから navigate() できる。
    element: (
      <Suspense fallback={null}>
        <AuthProvider>
          <Outlet />
        </AuthProvider>
      </Suspense>
    ),
    errorElement: <RootErrorBoundary />,
    children: [
      {
        path: "/login",
        element: (
          <Suspense fallback={null}>
            <Login />
          </Suspense>
        ),
      },
      {
        path: "/forgot-password",
        lazy: async () => {
          const { ForgotPasswordPage } = await import("@/features/auth/routes/ForgotPasswordPage");
          return { Component: ForgotPasswordPage };
        },
      },
      {
        path: "/reset-password",
        lazy: async () => {
          const { ResetPasswordPage } = await import("@/features/auth/routes/ResetPasswordPage");
          return { Component: ResetPasswordPage };
        },
      },
      {
        element: <Layout />,
        children: [
      // ── Dashboard ────────────────────────────────────────────────
      {
        path: "/",
        element: (
          <RequirePermission resource={ResourceDashboard}>
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { Dashboard } = await import("@/features/dashboard");
              return { Component: Dashboard };
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
                import("@/features/owners/loaders"),
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
                import("@/features/owners/loaders"),
              ]);
              return { Component: OwnerFormPage, loader: ownerLoader };
            },
          },
        ],
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
            lazy: async () => {
              const { MedicalRecordPetSelection } = await import("@/features/medical-records");
              return { Component: MedicalRecordPetSelection };
            },
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
            lazy: async () => {
              const { HospitalizationPetSelection } = await import("@/features/hospitalization");
              return { Component: HospitalizationPetSelection };
            },
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
            lazy: async () => {
              const { TrimmingPetSelection } = await import("@/features/trimming");
              return { Component: TrimmingPetSelection };
            },
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
            lazy: async () => {
              const { ExaminationPetSelection } = await import("@/features/examinations");
              return { Component: ExaminationPetSelection };
            },
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

      // ── Accounting ───────────────────────────────────────────────
      {
        path: "/accounting",
        element: (
          <RequirePermission resource={ResourceAccounting}>
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { AccountingList } = await import("@/features/accounting/routes/AccountingList");
              return { Component: AccountingList };
            },
          },
          {
            path: "select-pet",
            lazy: async () => {
              const { AccountingPetSelection } = await import(
                "@/features/accounting/routes/AccountingPetSelection"
              );
              return { Component: AccountingPetSelection };
            },
          },
          {
            // BUG-020: create 権限ガード
            path: "new",
            element: (
              <RequirePermission resource={ResourceAccounting} action="create">
                <Outlet />
              </RequirePermission>
            ),
            children: [
              {
                index: true,
                lazy: async () => {
                  const { AccountingDetailPage } = await import(
                    "@/app/pages/AccountingDetailPage"
                  );
                  return { Component: AccountingDetailPage };
                },
              },
            ],
          },
          {
            path: ":id",
            lazy: async () => {
              const { AccountingDetailPage } = await import(
                "@/app/pages/AccountingDetailPage"
              );
              return { Component: AccountingDetailPage };
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
            lazy: async () => {
              const { VaccinationPetSelection } = await import("@/features/vaccinations");
              return { Component: VaccinationPetSelection };
            },
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
        children: [
          {
            index: true,
            lazy: async () => {
              const { CheckupsList } = await import("@/features/checkups");
              return { Component: CheckupsList };
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

      // ── Settings（master） ────────────────────────────────────────
      {
        path: "/settings",
        element: (
          <RequirePermission resource={ResourceMaster}>
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { MasterSettingsIndex } = await import("@/features/master");
              return { Component: MasterSettingsIndex };
            },
          },
          {
            path: "staff",
            lazy: async () => {
              const { StaffSettings } = await import("@/features/master");
              return { Component: StaffSettings };
            },
          },
          {
            path: "treatment-items",
            lazy: async () => {
              const { TreatmentPlanMaster } = await import("@/features/master");
              return { Component: TreatmentPlanMaster };
            },
          },
          {
            path: "diagnosis",
            lazy: async () => {
              const { DiagnosisSettings } = await import("@/features/master");
              return { Component: DiagnosisSettings };
            },
          },
          {
            path: "animal-species",
            lazy: async () => {
              const { AnimalSpeciesSettings } = await import("@/features/master");
              return { Component: AnimalSpeciesSettings };
            },
          },
          {
            path: "trimming",
            lazy: async () => {
              const { TrimmingSettings } = await import("@/features/master");
              return { Component: TrimmingSettings };
            },
          },
          {
            path: "medicine",
            lazy: async () => {
              const { MedicineSettings } = await import("@/features/master");
              return { Component: MedicineSettings };
            },
          },
          {
            path: "service-type",
            lazy: async () => {
              const { ServiceTypeSettings } = await import("@/features/master");
              return { Component: ServiceTypeSettings };
            },
          },
          {
            path: "hospitalization",
            lazy: async () => {
              const { HospitalizationSettings } = await import("@/features/master");
              return { Component: HospitalizationSettings };
            },
          },
          {
            path: "cage",
            lazy: async () => {
              const { CageSettings } = await import("@/features/master");
              return { Component: CageSettings };
            },
          },
          {
            path: "merchandise-items",
            lazy: async () => {
              const { MerchandiseItemSettings } = await import("@/features/master");
              return { Component: MerchandiseItemSettings };
            },
          },
          {
            path: "insurance",
            lazy: async () => {
              const { InsuranceSettings } = await import("@/features/master");
              return { Component: InsuranceSettings };
            },
          },
          {
            path: "job-title",
            lazy: async () => {
              const { JobTitleSettings } = await import("@/features/master");
              return { Component: JobTitleSettings };
            },
          },
          {
            path: "inquiry-templates",
            lazy: async () => {
              const { InterviewTemplateSettings } = await import("@/features/master");
              return { Component: InterviewTemplateSettings };
            },
          },
          {
            path: "interview/chief-complaint",
            lazy: async () => {
              const { ChiefComplaintSettings } = await import("@/features/master");
              return { Component: ChiefComplaintSettings };
            },
          },
          {
            path: "interview/templates",
            lazy: async () => {
              const { InterviewTemplateSettings } = await import("@/features/master");
              return { Component: InterviewTemplateSettings };
            },
          },
        ],
      },

      // ── Hospital Settings（hospital-settings: 独立リソース） ───────
      {
        path: "/settings/clinic",
        element: (
          <RequirePermission resource={ResourceHospitalSettings}>
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { ClinicMasterSettings } = await import("@/features/hospital-settings");
              return { Component: ClinicMasterSettings };
            },
          },
        ],
      },

      // ── Not Found ────────────────────────────────────────────────
      {
        path: "*",
        element: (
          <div className="flex-1 p-5 flex items-center justify-center">
            <p className="text-muted-foreground">ページが見つかりません</p>
          </div>
        ),
      },
        ],
      },
    ],
  },
]);
