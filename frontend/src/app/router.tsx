import { lazy, Suspense } from "react";
import { createBrowserRouter } from "react-router";

import { Layout } from "@/components/shared/Layout/Layout";
import { RootErrorBoundary, RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
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
        lazy: async () => {
          const { Dashboard } = await import("@/features/dashboard/routes/Dashboard");
          return { Component: Dashboard };
        },
      },

      // ── Owners ───────────────────────────────────────────────────
      {
        path: "/owners",
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
        lazy: async () => {
          const { ReservationManagement } = await import(
            "@/features/reservations/routes/ReservationManagement"
          );
          return { Component: ReservationManagement };
        },
      },

      // ── Medical Records ──────────────────────────────────────────
      {
        path: "/medical-records",
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
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { Examinations } = await import(
                "@/features/examinations/routes/Examinations"
              );
              return { Component: Examinations };
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
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const [{ Accounting }, { accountingsLoader }] = await Promise.all([
                import("@/features/accounting/routes/Accounting"),
                import("@/features/accounting/loaders"),
              ]);
              return { Component: Accounting, loader: accountingsLoader };
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
              const { AccountingDetail } = await import(
                "@/features/accounting/routes/AccountingDetail"
              );
              return { Component: AccountingDetail };
            },
          },
          {
            path: ":id",
            lazy: async () => {
              const { AccountingDetail } = await import(
                "@/features/accounting/routes/AccountingDetail"
              );
              return { Component: AccountingDetail };
            },
          },
        ],
      },

      // ── Vaccinations ─────────────────────────────────────────────
      {
        path: "/vaccinations",
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

      // ── Inventory ────────────────────────────────────────────────
      {
        path: "/inventory",
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
        lazy: async () => {
          const { EstimateList } = await import(
            "@/features/estimates/routes/EstimateList"
          );
          return { Component: EstimateList };
        },
      },
      {
        path: "/estimates/new",
        lazy: async () => {
          const { EstimateForm } = await import(
            "@/features/estimates/routes/EstimateForm"
          );
          return { Component: EstimateForm };
        },
      },
      {
        path: "/estimates/:id",
        lazy: async () => {
          const { EstimateDetail } = await import(
            "@/features/estimates/routes/EstimateDetail"
          );
          return { Component: EstimateDetail };
        },
      },
      {
        path: "/estimates/:id/edit",
        lazy: async () => {
          const { EstimateForm } = await import(
            "@/features/estimates/routes/EstimateForm"
          );
          return { Component: EstimateForm };
        },
      },

      // ── Shifts ───────────────────────────────────────────────────
      {
        path: "/shifts",
        lazy: async () => {
          const { ShiftCalendarPage } = await import(
            "@/features/shifts/routes/ShiftCalendarPage"
          );
          return { Component: ShiftCalendarPage };
        },
      },

      // ── Settings ─────────────────────────────────────────────────
      {
        path: "/settings",
        lazy: async () => {
          const { MasterSettingsIndex } = await import(
            "@/features/master/routes/MasterSettingsIndex"
          );
          return { Component: MasterSettingsIndex };
        },
      },
      {
        path: "/settings/clinic",
        lazy: async () => {
          const { ClinicMasterSettings } = await import(
            "@/features/hospital-settings/routes/ClinicMasterSettings"
          );
          return { Component: ClinicMasterSettings };
        },
      },
      {
        path: "/settings/staff",
        lazy: async () => {
          const { StaffSettings } = await import(
            "@/features/master/routes/StaffSettings"
          );
          return { Component: StaffSettings };
        },
      },
      {
        path: "/settings/treatment-items",
        lazy: async () => {
          const { TreatmentPlanMaster } = await import(
            "@/features/master/routes/TreatmentPlanMaster"
          );
          return { Component: TreatmentPlanMaster };
        },
      },
      {
        path: "/settings/diagnosis",
        lazy: async () => {
          const { DiagnosisSettings } = await import(
            "@/features/master/routes/DiagnosisSettings"
          );
          return { Component: DiagnosisSettings };
        },
      },
      {
        path: "/settings/animal-species",
        lazy: async () => {
          const { AnimalSpeciesSettings } = await import(
            "@/features/master/routes/AnimalSpeciesSettings"
          );
          return { Component: AnimalSpeciesSettings };
        },
      },
      {
        path: "/settings/trimming",
        lazy: async () => {
          const { TrimmingSettings } = await import(
            "@/features/master/routes/TrimmingSettings"
          );
          return { Component: TrimmingSettings };
        },
      },
      {
        path: "/settings/medicine",
        lazy: async () => {
          const { MedicineSettings } = await import(
            "@/features/master/routes/MedicineSettings"
          );
          return { Component: MedicineSettings };
        },
      },
      {
        path: "/settings/service-type",
        lazy: async () => {
          const { ServiceTypeSettings } = await import(
            "@/features/master/routes/ServiceTypeSettings"
          );
          return { Component: ServiceTypeSettings };
        },
      },
      {
        path: "/settings/hospitalization",
        lazy: async () => {
          const { HospitalizationSettings } = await import(
            "@/features/master/routes/HospitalizationSettings"
          );
          return { Component: HospitalizationSettings };
        },
      },
      {
        path: "/settings/cage",
        lazy: async () => {
          const { CageSettings } = await import(
            "@/features/master/routes/CageSettings"
          );
          return { Component: CageSettings };
        },
      },
      {
        path: "/settings/insurance",
        lazy: async () => {
          const { InsuranceSettings } = await import(
            "@/features/master/routes/InsuranceSettings"
          );
          return { Component: InsuranceSettings };
        },
      },
      {
        path: "/settings/job-title",
        lazy: async () => {
          const { JobTitleSettings } = await import(
            "@/features/master/routes/JobTitleSettings"
          );
          return { Component: JobTitleSettings };
        },
      },
      {
        path: "/settings/inquiry-templates",
        lazy: async () => {
          const { InterviewTemplateSettings } = await import(
            "@/features/master/routes/InterviewTemplateSettings"
          );
          return { Component: InterviewTemplateSettings };
        },
      },
      {
        path: "/settings/interview/chief-complaint",
        lazy: async () => {
          const { ChiefComplaintSettings } = await import(
            "@/features/master/routes/ChiefComplaintSettings"
          );
          return { Component: ChiefComplaintSettings };
        },
      },
      {
        path: "/settings/interview/templates",
        lazy: async () => {
          const { InterviewTemplateSettings } = await import(
            "@/features/master/routes/InterviewTemplateSettings"
          );
          return { Component: InterviewTemplateSettings };
        },
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
