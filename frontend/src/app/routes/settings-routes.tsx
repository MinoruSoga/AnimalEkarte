import { Navigate, Outlet, type RouteObject } from "react-router";

import { RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { RequirePermission } from "@/components/shared/RequirePermission";
import {
  ResourceClosingSettings,
  ResourceHospitalSettings,
  ResourceMasterAnimalSpecies,
  ResourceMasterHospitalization,
  ResourceMasterInsurance,
  ResourceMasterMedical,
  ResourceMasterMerchandise,
  ResourceMasterPermission,
  ResourceMasterReservationType,
  ResourceMasterStaff,
  ResourceMasterTrimming,
  ResourcePaymentMethod,
  ResourceShifts,
} from "@/types/generated/models";

export const settingsRoute: RouteObject = {
  path: "/settings",
  element: <Outlet />,
  errorElement: <RouteErrorBoundary />,
  children: [
    {
      // MasterSettingsIndex — ガード不要（BUG-123 でカードフィルタリング対応）
      index: true,
      lazy: async () => {
        const { MasterSettingsIndex } = await import("@/features/master");
        return { Component: MasterSettingsIndex };
      },
    },
    // BUG-382 / BUG-384: 旧ルートからの互換 redirect。
    // カルテ編集画面・トリミング編集画面等、既存 UI からの遷移先を救済する。
    { path: "job-title", element: <Navigate to="/settings/occupations" replace /> },
    { path: "service-type", element: <Navigate to="/settings/reservation-type" replace /> },
    { path: "diagnosis-type", element: <Navigate to="/settings/diagnosis?tab=diagnosis_type" replace /> },
    { path: "diagnosis-name", element: <Navigate to="/settings/diagnosis?tab=diagnosis_name" replace /> },
    { path: "trimming-course", element: <Navigate to="/settings/trimming?tab=course" replace /> },
    { path: "trimming-option", element: <Navigate to="/settings/trimming?tab=option" replace /> },
    { path: "examination", element: <Navigate to="/settings/treatment-items?tab=examination" replace /> },
    { path: "vaccine", element: <Navigate to="/settings/treatment-items?tab=vaccine" replace /> },
    { path: "consultation", element: <Navigate to="/settings/treatment-items?tab=consultation" replace /> },
    { path: "procedure", element: <Navigate to="/settings/treatment-items?tab=procedure" replace /> },
    { path: "inquiry-template", element: <Navigate to="/settings/inquiry-templates" replace /> },
    {
      path: "staff",
      element: <RequirePermission resource={ResourceMasterStaff}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { StaffSettings } = await import("@/features/master");
          return { Component: StaffSettings };
        },
      }],
    },
    {
      path: "treatment-items",
      element: <RequirePermission resource={ResourceMasterMedical}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { TreatmentPlanMaster } = await import("@/features/master");
          return { Component: TreatmentPlanMaster };
        },
      }],
    },
    {
      path: "diagnosis",
      element: <RequirePermission resource={ResourceMasterMedical}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { DiagnosisSettings } = await import("@/features/master");
          return { Component: DiagnosisSettings };
        },
      }],
    },
    {
      path: "animal-species",
      element: <RequirePermission resource={ResourceMasterAnimalSpecies}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { AnimalSpeciesSettings } = await import("@/features/master");
          return { Component: AnimalSpeciesSettings };
        },
      }],
    },
    {
      path: "trimming",
      element: <RequirePermission resource={ResourceMasterTrimming}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { TrimmingSettings } = await import("@/features/master");
          return { Component: TrimmingSettings };
        },
      }],
    },
    {
      path: "medicine",
      element: <RequirePermission resource={ResourceMasterMedical}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { MedicineSettings } = await import("@/features/master");
          return { Component: MedicineSettings };
        },
      }],
    },
    {
      path: "reservation-type",
      element: <RequirePermission resource={ResourceMasterReservationType}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { ReservationTypeSettings } = await import("@/features/master");
          return { Component: ReservationTypeSettings };
        },
      }],
    },
    {
      path: "hospitalization",
      element: <RequirePermission resource={ResourceMasterHospitalization}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { HospitalizationSettings } = await import("@/features/master");
          return { Component: HospitalizationSettings };
        },
      }],
    },
    {
      path: "cage",
      element: <RequirePermission resource={ResourceMasterHospitalization}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { CageSettings } = await import("@/features/master");
          return { Component: CageSettings };
        },
      }],
    },
    {
      path: "merchandise-items",
      element: <RequirePermission resource={ResourceMasterMerchandise}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { MerchandiseItemSettings } = await import("@/features/master");
          return { Component: MerchandiseItemSettings };
        },
      }],
    },
    {
      path: "insurance",
      element: <RequirePermission resource={ResourceMasterInsurance}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { InsuranceSettings } = await import("@/features/master");
          return { Component: InsuranceSettings };
        },
      }],
    },
    {
      path: "occupations",
      element: <RequirePermission resource={ResourceMasterStaff}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { OccupationSettings } = await import("@/features/master");
          return { Component: OccupationSettings };
        },
      }],
    },
    {
      path: "permission-groups",
      element: <RequirePermission resource={ResourceMasterPermission}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { PermissionGroupSettings } = await import("@/features/master");
          return { Component: PermissionGroupSettings };
        },
      }],
    },
    {
      path: "inquiry-templates",
      element: <RequirePermission resource={ResourceMasterMedical}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { InterviewTemplateSettings } = await import("@/features/master");
          return { Component: InterviewTemplateSettings };
        },
      }],
    },
    {
      path: "interview/chief-complaint",
      element: <RequirePermission resource={ResourceMasterMedical}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { ChiefComplaintSettings } = await import("@/features/master");
          return { Component: ChiefComplaintSettings };
        },
      }],
    },
    {
      path: "interview/templates",
      element: <RequirePermission resource={ResourceMasterMedical}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { InterviewTemplateSettings } = await import("@/features/master");
          return { Component: InterviewTemplateSettings };
        },
      }],
    },
    {
      // BUG-383: 旧URL redirect
      path: "shift-template",
      element: <Navigate to="/settings/shift-templates" replace />,
    },
    {
      path: "shift-templates",
      element: <RequirePermission resource={ResourceShifts}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { ShiftTemplateSettings } = await import("@/features/shifts");
          return { Component: ShiftTemplateSettings };
        },
      }],
    },
    // FEAT-368: 締め時間設定
    // STG-BLOCKER-002: RequirePermission 追加。Sidebar は ResourceClosingSettings で
    // フィルタしているが、URL 直叩きで権限なしユーザーが到達できる脆弱性を塞ぐ。
    {
      path: "closing-time",
      element: <RequirePermission resource={ResourceClosingSettings}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { ClosingSettingsPage } = await import("@/features/closing-settings");
          return { Component: ClosingSettingsPage };
        },
      }],
    },
    // FEAT-368: 支払方法マスタ
    // STG-BLOCKER-002: RequirePermission 追加（理由は closing-time と同じ）
    {
      path: "payment-methods",
      element: <RequirePermission resource={ResourcePaymentMethod}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { PaymentMethodSettings } = await import("@/features/master");
          return { Component: PaymentMethodSettings };
        },
      }],
    },
    // FE-001: Lステップ連携設定
    {
      path: "integrations/lstep",
      element: <RequirePermission resource={ResourceHospitalSettings}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { LstepSettingsPage } = await import("@/features/settings");
          return { Component: LstepSettingsPage };
        },
      }],
    },
    // FE-007: Lステップタグ管理
    {
      path: "lstep/tags",
      element: <RequirePermission resource={ResourceHospitalSettings}><Outlet /></RequirePermission>,
      children: [{
        index: true,
        lazy: async () => {
          const { LstepTagManagementPage } = await import("@/features/lstep");
          return { Component: LstepTagManagementPage };
        },
      }],
    },
  ],
};
