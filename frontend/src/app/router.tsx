import { lazy, Suspense } from "react";
import { createBrowserRouter, Outlet } from "react-router";

import { Layout } from "@/components/shared/Layout/Layout";
import { RootErrorBoundary, RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { RequirePermission } from "@/components/shared/RequirePermission";
import { AuthProvider } from "@/features/auth";

/* bundle-dynamic-imports: ログインページは未認証ユーザー専用。認証済みユーザーのバンドルに含めない */
const Login = lazy(() =>
  import("@/features/auth/routes/Login").then((m) => ({ default: m.Login })),
);

export const router = createBrowserRouter([
  {
    path: "/login",
    element: (
      <Suspense fallback={null}>
        <Login />
      </Suspense>
    ),
  },
  {
    // AuthProvider を保護ルート側にのみ配置。
    // /login は上のルートで AuthProvider 外に定義されているため GET /v1/me は実行されない。
    element: (
      <AuthProvider>
        <Layout />
      </AuthProvider>
    ),
    errorElement: <RootErrorBoundary />,
    children: [
      // ── Dashboard ────────────────────────────────────────────────
      {
        path: "/",
        element: (
          <RequirePermission resource="dashboard">
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { Dashboard } = await import("@/features/dashboard/routes/Dashboard");
              return { Component: Dashboard };
            },
          },
        ],
      },

      // ── Owners ───────────────────────────────────────────────────
      {
        path: "/owners",
        element: (
          <RequirePermission resource="owners">
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
            path: "new",
            lazy: async () => {
              const { OwnerFormPage } = await import("@/app/pages/OwnerFormPage");
              return { Component: OwnerFormPage };
            },
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
          <RequirePermission resource="reservations">
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { ReservationManagement } = await import(
                "@/features/reservations/routes/ReservationManagement"
              );
              return { Component: ReservationManagement };
            },
          },
        ],
      },

      // ── Medical Records ──────────────────────────────────────────
      {
        path: "/medical-records",
        element: (
          <RequirePermission resource="medical-records">
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { MedicalRecords } = await import(
                "@/features/medical-records/routes/MedicalRecords"
              );
              return { Component: MedicalRecords };
            },
          },
          {
            path: "select-pet",
            lazy: async () => {
              const { MedicalRecordPetSelection } = await import(
                "@/features/medical-records/routes/MedicalRecordPetSelection"
              );
              return { Component: MedicalRecordPetSelection };
            },
          },
          {
            path: "new",
            lazy: async () => {
              const { MedicalRecordForm } = await import(
                "@/features/medical-records/routes/MedicalRecordForm"
              );
              return { Component: MedicalRecordForm };
            },
          },
          {
            path: ":id",
            lazy: async () => {
              const { MedicalRecordForm } = await import(
                "@/features/medical-records/routes/MedicalRecordForm"
              );
              return { Component: MedicalRecordForm };
            },
          },
        ],
      },

      // ── Hospitalization ──────────────────────────────────────────
      {
        path: "/hospitalization",
        element: (
          <RequirePermission resource="hospitalization">
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { HospitalizationList } = await import(
                "@/features/hospitalization/routes/HospitalizationList"
              );
              return { Component: HospitalizationList };
            },
          },
          {
            path: "select-pet",
            lazy: async () => {
              const { HospitalizationPetSelection } = await import(
                "@/features/hospitalization/routes/HospitalizationPetSelection"
              );
              return { Component: HospitalizationPetSelection };
            },
          },
          {
            path: "new",
            lazy: async () => {
              const { HospitalizationForm } = await import(
                "@/features/hospitalization/routes/HospitalizationForm"
              );
              return { Component: HospitalizationForm };
            },
          },
          {
            path: ":id",
            lazy: async () => {
              const { HospitalizationDetail } = await import(
                "@/features/hospitalization/routes/HospitalizationDetail"
              );
              return { Component: HospitalizationDetail };
            },
          },
          {
            path: ":id/edit",
            lazy: async () => {
              const { HospitalizationForm } = await import(
                "@/features/hospitalization/routes/HospitalizationForm"
              );
              return { Component: HospitalizationForm };
            },
          },
        ],
      },

      // ── Trimming ─────────────────────────────────────────────────
      {
        path: "/trimming",
        element: (
          <RequirePermission resource="trimming">
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { TrimmingList } = await import(
                "@/features/trimming/routes/TrimmingList"
              );
              return { Component: TrimmingList };
            },
          },
          {
            path: "select-pet",
            lazy: async () => {
              const { TrimmingPetSelection } = await import(
                "@/features/trimming/routes/TrimmingPetSelection"
              );
              return { Component: TrimmingPetSelection };
            },
          },
          {
            path: "new",
            lazy: async () => {
              const { TrimmingForm } = await import(
                "@/features/trimming/routes/TrimmingForm"
              );
              return { Component: TrimmingForm };
            },
          },
          {
            path: ":id",
            lazy: async () => {
              const { TrimmingForm } = await import(
                "@/features/trimming/routes/TrimmingForm"
              );
              return { Component: TrimmingForm };
            },
          },
        ],
      },

      // ── Examinations ─────────────────────────────────────────────
      {
        path: "/examinations",
        element: (
          <RequirePermission resource="examinations">
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { ExaminationsList } = await import(
                "@/features/examinations/routes/ExaminationsList"
              );
              return { Component: ExaminationsList };
            },
          },
          {
            path: "select-pet",
            lazy: async () => {
              const { ExaminationPetSelection } = await import(
                "@/features/examinations/routes/ExaminationPetSelection"
              );
              return { Component: ExaminationPetSelection };
            },
          },
          {
            path: "new",
            lazy: async () => {
              const { ExaminationForm } = await import(
                "@/features/examinations/routes/ExaminationForm"
              );
              return { Component: ExaminationForm };
            },
          },
          {
            path: ":id",
            lazy: async () => {
              const { ExaminationForm } = await import(
                "@/features/examinations/routes/ExaminationForm"
              );
              return { Component: ExaminationForm };
            },
          },
        ],
      },

      // ── Accounting ───────────────────────────────────────────────
      {
        path: "/accounting",
        element: (
          <RequirePermission resource="accounting">
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
            path: "new",
            lazy: async () => {
              const { AccountingDetailPage } = await import(
                "@/app/pages/AccountingDetailPage"
              );
              return { Component: AccountingDetailPage };
            },
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
          <RequirePermission resource="vaccinations">
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { VaccinationList } = await import(
                "@/features/vaccinations/routes/VaccinationList"
              );
              return { Component: VaccinationList };
            },
          },
          {
            path: "select-pet",
            lazy: async () => {
              const { VaccinationPetSelection } = await import(
                "@/features/vaccinations/routes/VaccinationPetSelection"
              );
              return { Component: VaccinationPetSelection };
            },
          },
          {
            path: "new",
            lazy: async () => {
              const { VaccinationForm } = await import(
                "@/features/vaccinations/routes/VaccinationForm"
              );
              return { Component: VaccinationForm };
            },
          },
          {
            path: ":id",
            lazy: async () => {
              const { VaccinationForm } = await import(
                "@/features/vaccinations/routes/VaccinationForm"
              );
              return { Component: VaccinationForm };
            },
          },
        ],
      },

      // ── Checkups ─────────────────────────────────────────────────
      {
        path: "/checkups",
        element: (
          <RequirePermission resource="checkups">
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { CheckupsList } = await import(
                "@/features/checkups/routes/CheckupsList"
              );
              return { Component: CheckupsList };
            },
          },
        ],
      },

      // ── Inventory ────────────────────────────────────────────────
      {
        path: "/inventory",
        element: (
          <RequirePermission resource="inventory">
            <Outlet />
          </RequirePermission>
        ),
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { InventoryList } = await import(
                "@/features/inventory/routes/InventoryList"
              );
              return { Component: InventoryList };
            },
          },
          {
            path: "new",
            lazy: async () => {
              const { InventoryForm } = await import(
                "@/features/inventory/routes/InventoryForm"
              );
              return { Component: InventoryForm };
            },
          },
          {
            path: ":id",
            lazy: async () => {
              const { InventoryForm } = await import(
                "@/features/inventory/routes/InventoryForm"
              );
              return { Component: InventoryForm };
            },
          },
        ],
      },

      // ── Estimates ────────────────────────────────────────────────
      {
        path: "/estimates",
        element: (
          <RequirePermission resource="estimates">
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { EstimateList } = await import(
                "@/features/estimates/routes/EstimateList"
              );
              return { Component: EstimateList };
            },
          },
          {
            path: "new",
            lazy: async () => {
              const { EstimateForm } = await import(
                "@/features/estimates/routes/EstimateForm"
              );
              return { Component: EstimateForm };
            },
          },
          {
            path: ":id",
            lazy: async () => {
              const { EstimateDetail } = await import(
                "@/features/estimates/routes/EstimateDetail"
              );
              return { Component: EstimateDetail };
            },
          },
          {
            path: ":id/edit",
            lazy: async () => {
              const { EstimateForm } = await import(
                "@/features/estimates/routes/EstimateForm"
              );
              return { Component: EstimateForm };
            },
          },
        ],
      },

      // ── Shifts ───────────────────────────────────────────────────
      {
        path: "/shifts",
        element: (
          <RequirePermission resource="shifts">
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { ShiftCalendarPage } = await import(
                "@/features/shifts/routes/ShiftCalendarPage"
              );
              return { Component: ShiftCalendarPage };
            },
          },
        ],
      },

      // ── Settings（master） ────────────────────────────────────────
      {
        path: "/settings",
        element: (
          <RequirePermission resource="master">
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { MasterSettingsIndex } = await import(
                "@/features/master/routes/MasterSettingsIndex"
              );
              return { Component: MasterSettingsIndex };
            },
          },
          {
            path: "staff",
            lazy: async () => {
              const { StaffSettings } = await import(
                "@/features/master/routes/StaffSettings"
              );
              return { Component: StaffSettings };
            },
          },
          {
            path: "treatment-items",
            lazy: async () => {
              const { TreatmentPlanMaster } = await import(
                "@/features/master/routes/TreatmentPlanMaster"
              );
              return { Component: TreatmentPlanMaster };
            },
          },
          {
            path: "diagnosis",
            lazy: async () => {
              const { DiagnosisSettings } = await import(
                "@/features/master/routes/DiagnosisSettings"
              );
              return { Component: DiagnosisSettings };
            },
          },
          {
            path: "animal-species",
            lazy: async () => {
              const { AnimalSpeciesSettings } = await import(
                "@/features/master/routes/AnimalSpeciesSettings"
              );
              return { Component: AnimalSpeciesSettings };
            },
          },
          {
            path: "trimming",
            lazy: async () => {
              const { TrimmingSettings } = await import(
                "@/features/master/routes/TrimmingSettings"
              );
              return { Component: TrimmingSettings };
            },
          },
          {
            path: "medicine",
            lazy: async () => {
              const { MedicineSettings } = await import(
                "@/features/master/routes/MedicineSettings"
              );
              return { Component: MedicineSettings };
            },
          },
          {
            path: "service-type",
            lazy: async () => {
              const { ServiceTypeSettings } = await import(
                "@/features/master/routes/ServiceTypeSettings"
              );
              return { Component: ServiceTypeSettings };
            },
          },
          {
            path: "hospitalization",
            lazy: async () => {
              const { HospitalizationSettings } = await import(
                "@/features/master/routes/HospitalizationSettings"
              );
              return { Component: HospitalizationSettings };
            },
          },
          {
            path: "cage",
            lazy: async () => {
              const { CageSettings } = await import(
                "@/features/master/routes/CageSettings"
              );
              return { Component: CageSettings };
            },
          },
          {
            path: "merchandise-items",
            lazy: async () => {
              const { MerchandiseItemSettings } = await import(
                "@/features/master/routes/MerchandiseItemSettings"
              );
              return { Component: MerchandiseItemSettings };
            },
          },
          {
            path: "insurance",
            lazy: async () => {
              const { InsuranceSettings } = await import(
                "@/features/master/routes/InsuranceSettings"
              );
              return { Component: InsuranceSettings };
            },
          },
          {
            path: "job-title",
            lazy: async () => {
              const { JobTitleSettings } = await import(
                "@/features/master/routes/JobTitleSettings"
              );
              return { Component: JobTitleSettings };
            },
          },
          {
            path: "inquiry-templates",
            lazy: async () => {
              const { InterviewTemplateSettings } = await import(
                "@/features/master/routes/InterviewTemplateSettings"
              );
              return { Component: InterviewTemplateSettings };
            },
          },
          {
            path: "interview/chief-complaint",
            lazy: async () => {
              const { ChiefComplaintSettings } = await import(
                "@/features/master/routes/ChiefComplaintSettings"
              );
              return { Component: ChiefComplaintSettings };
            },
          },
          {
            path: "interview/templates",
            lazy: async () => {
              const { InterviewTemplateSettings } = await import(
                "@/features/master/routes/InterviewTemplateSettings"
              );
              return { Component: InterviewTemplateSettings };
            },
          },
          // ── User Accounts（権限グループ割当） ──
          {
            path: "user-accounts",
            lazy: async () => {
              const { UserAccountSettings } = await import(
                "@/features/master/routes/UserAccountSettings"
              );
              return { Component: UserAccountSettings };
            },
          },
          // ── Permission Groups ──
          {
            path: "permission-groups",
            lazy: async () => {
              const { PermissionGroupSettings } = await import(
                "@/features/master/routes/PermissionGroupSettings"
              );
              return { Component: PermissionGroupSettings };
            },
          },
        ],
      },

      // ── Hospital Settings（hospital-settings: 独立リソース） ───────
      {
        path: "/settings/clinic",
        element: (
          <RequirePermission resource="hospital-settings">
            <Outlet />
          </RequirePermission>
        ),
        children: [
          {
            index: true,
            lazy: async () => {
              const { ClinicMasterSettings } = await import(
                "@/features/hospital-settings/routes/ClinicMasterSettings"
              );
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
]);
