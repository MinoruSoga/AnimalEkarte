import type { RouteObject } from "react-router";

type SettingsCategory =
  | "serviceType" | "examination" | "vaccine" | "medicine"
  | "staff" | "insurance" | "consultation" | "procedure"
  | "hospitalization" | "cage" | "trimmingCourse" | "trimmingOption"
  | "diagnosisCategory" | "diagnosisName";

const SETTINGS_CATEGORY_MAP: Record<string, SettingsCategory> = {
  "service-type": "serviceType",
  examination: "examination",
  vaccine: "vaccine",
  medicine: "medicine",
  staff: "staff",
  insurance: "insurance",
  consultation: "consultation",
  procedure: "procedure",
  hospitalization: "hospitalization",
  cage: "cage",
  "trimming-course": "trimmingCourse",
  "trimming-option": "trimmingOption",
  "diagnosis-category": "diagnosisCategory",
  "diagnosis-name": "diagnosisName",
};

export const settingsRoutes: RouteObject[] = [
  {
    path: "/settings/clinic",
    lazy: async () => {
      const { ClinicSettings } = await import("@/features/hospital-settings/routes/ClinicSettings");
      return { Component: ClinicSettings };
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
