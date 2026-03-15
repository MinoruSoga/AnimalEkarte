import type { RouteObject } from "react-router";

type SettingsCategory =
  | "examination" | "vaccine"
  | "insurance" | "consultation" | "procedure"
  | "hospitalization" | "cage" | "trimmingCourse" | "trimmingOption"
  | "diagnosisCategory" | "diagnosisName" | "checkup" | "job_title";

const SETTINGS_CATEGORY_MAP: Record<string, SettingsCategory> = {
  examination: "examination",
  vaccine: "vaccine",
  // medicine は MedicineSettings 専用ルートで処理
  insurance: "insurance",
  consultation: "consultation",
  procedure: "procedure",
  hospitalization: "hospitalization",
  cage: "cage",
  "trimming-course": "trimmingCourse",
  "trimming-option": "trimmingOption",
  "diagnosis-category": "diagnosisCategory",
  "diagnosis-name": "diagnosisName",
  checkup: "checkup",
  "job-title": "job_title",
};

export const settingsRoutes: RouteObject[] = [
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
      const { StaffSettings } = await import("@/features/master/routes/StaffSettings");
      return { Component: StaffSettings };
    },
  },
  {
    path: "/settings/treatment-items",
    lazy: async () => {
      const { TreatmentItemsSettings } = await import("@/features/master/routes/TreatmentItemsSettings");
      return { Component: TreatmentItemsSettings };
    },
  },
  {
    path: "/settings/diagnosis",
    lazy: async () => {
      const { DiagnosisSettings } = await import("@/features/master/routes/DiagnosisSettings");
      return { Component: DiagnosisSettings };
    },
  },
  {
    path: "/settings/trimming",
    lazy: async () => {
      const { TrimmingSettings } = await import("@/features/master/routes/TrimmingSettings");
      return { Component: TrimmingSettings };
    },
  },
  {
    path: "/settings/medicine",
    lazy: async () => {
      const { MedicineSettings } = await import("@/features/master/routes/MedicineSettings");
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
  ...Object.entries(SETTINGS_CATEGORY_MAP).map(([slug, category]) => ({
    path: `/settings/${slug}`,
    lazy: async () => {
      const { Settings } = await import("@/features/master/routes/Settings");
      return {
        element: <Settings category={category} />,
      };
    },
  })),
];
