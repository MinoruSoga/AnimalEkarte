/**
 * config/paths.ts — アプリケーション全URLの一元管理
 *
 * 使用例:
 *   import { paths } from "@/config/paths";
 *   navigate(paths.owners.detail.getHref(ownerId));
 */

interface PathEntry {
  path: string;
  getHref: () => string;
}

/**
 * BUG-383/384/385: dead route の旧パスを、実在ルートの `?tab=` エイリアスとして
 * 表現するヘルパー。base と同じ path を指しつつ getHref にだけ tab クエリを付与する。
 */
function withTab(base: PathEntry, tab: string): PathEntry {
  return {
    path: base.path,
    getHref: () => `${base.getHref()}?tab=${tab}`,
  };
}

const settingsTreatmentItemsBase: PathEntry = {
  path: "/settings/treatment-items",
  getHref: () => "/settings/treatment-items",
};

const settingsTrimmingBase: PathEntry = {
  path: "/settings/trimming",
  getHref: () => "/settings/trimming",
};

const settingsDiagnosisBase: PathEntry = {
  path: "/settings/diagnosis",
  getHref: () => "/settings/diagnosis",
};

export const paths = {
  home: { path: "/", getHref: () => "/" },

  auth: {
    login: { path: "/login", getHref: () => "/login" },
    forgotPassword: { path: "/forgot-password", getHref: () => "/forgot-password" },
    resetPassword: { path: "/reset-password", getHref: () => "/reset-password" },
  },

  owners: {
    path: "/owners",
    getHref: () => "/owners",
    new: { path: "/owners/new", getHref: () => "/owners/new" },
    detail: {
      path: "/owners/:id",
      getHref: (id: string | number) => `/owners/${encodeURIComponent(id)}`,
      // #158: 飼主単位カルテレポート（別ウィンドウ / Layout 外スタンドアロン）
      report: {
        path: "/owners/:id/report",
        getHref: (id: string | number) => `/owners/${encodeURIComponent(id)}/report`,
      },
    },
  },

  aggregation: { path: "/aggregation", getHref: () => "/aggregation" },

  reservations: { path: "/reservations", getHref: () => "/reservations" },

  medicalRecords: {
    path: "/medical-records",
    getHref: () => "/medical-records",
    selectPet: {
      path: "/medical-records/select-pet",
      getHref: () => "/medical-records/select-pet",
    },
    new: { path: "/medical-records/new", getHref: () => "/medical-records/new" },
    detail: {
      path: "/medical-records/:id",
      getHref: (id: string | number) => `/medical-records/${encodeURIComponent(id)}`,
    },
  },

  hospitalization: {
    path: "/hospitalization",
    getHref: () => "/hospitalization",
    selectPet: {
      path: "/hospitalization/select-pet",
      getHref: () => "/hospitalization/select-pet",
    },
    new: { path: "/hospitalization/new", getHref: () => "/hospitalization/new" },
    detail: {
      path: "/hospitalization/:id",
      getHref: (id: string | number) => `/hospitalization/${encodeURIComponent(id)}`,
    },
    edit: {
      path: "/hospitalization/:id/edit",
      getHref: (id: string | number) => `/hospitalization/${encodeURIComponent(id)}/edit`,
    },
  },

  trimming: {
    path: "/trimming",
    getHref: () => "/trimming",
    selectPet: { path: "/trimming/select-pet", getHref: () => "/trimming/select-pet" },
    new: { path: "/trimming/new", getHref: () => "/trimming/new" },
    detail: {
      path: "/trimming/:id",
      getHref: (id: string | number) => `/trimming/${encodeURIComponent(id)}`,
    },
  },

  labDevice: { path: "/lab-device", getHref: () => "/lab-device" },

  identityLinks: { path: "/identity-links", getHref: () => "/identity-links" },

  examinations: {
    path: "/examinations",
    getHref: () => "/examinations",
    selectPet: { path: "/examinations/select-pet", getHref: () => "/examinations/select-pet" },
    new: { path: "/examinations/new", getHref: () => "/examinations/new" },
    detail: {
      path: "/examinations/:id",
      getHref: (id: string | number) => `/examinations/${encodeURIComponent(id)}`,
    },
  },

  checkups: {
    path: "/checkups",
    getHref: () => "/checkups",
    selectPet: { path: "/checkups/select-pet", getHref: () => "/checkups/select-pet" },
    new: { path: "/checkups/new", getHref: () => "/checkups/new" },
  },

  accounting: {
    path: "/accounting",
    getHref: () => "/accounting",
    selectPet: { path: "/accounting/select-pet", getHref: () => "/accounting/select-pet" },
    new: { path: "/accounting/new", getHref: () => "/accounting/new" },
    detail: {
      path: "/accounting/:id",
      getHref: (id: string | number) => `/accounting/${encodeURIComponent(id)}`,
    },
    // FEAT-368: 締め・集計
    close: { path: "/accounting/close", getHref: () => "/accounting/close" },
    closeHistory: { path: "/accounting/close/history", getHref: () => "/accounting/close/history" },
    reports: { path: "/accounting/reports", getHref: () => "/accounting/reports" },
  },

  vaccinations: {
    path: "/vaccinations",
    getHref: () => "/vaccinations",
    selectPet: { path: "/vaccinations/select-pet", getHref: () => "/vaccinations/select-pet" },
    new: { path: "/vaccinations/new", getHref: () => "/vaccinations/new" },
    detail: {
      path: "/vaccinations/:id",
      getHref: (id: string | number) => `/vaccinations/${encodeURIComponent(id)}`,
    },
  },

  inventory: {
    path: "/inventory",
    getHref: () => "/inventory",
    new: { path: "/inventory/new", getHref: () => "/inventory/new" },
    detail: {
      path: "/inventory/:id",
      getHref: (id: string | number) => `/inventory/${encodeURIComponent(id)}`,
    },
  },

  estimates: {
    path: "/estimates",
    getHref: () => "/estimates",
    new: { path: "/estimates/new", getHref: () => "/estimates/new" },
    detail: {
      path: "/estimates/:id",
      getHref: (id: string | number) => `/estimates/${encodeURIComponent(id)}`,
    },
    edit: {
      path: "/estimates/:id/edit",
      getHref: (id: string | number) => `/estimates/${encodeURIComponent(id)}/edit`,
    },
  },

  shifts: { path: "/shifts", getHref: () => "/shifts" },

  manual: {
    path: "/manual",
    getHref: () => "/manual",
    article: {
      path: "/manual/:category/:slug",
      getHref: (category: string, slug: string) =>
        `/manual/${encodeURIComponent(category)}/${encodeURIComponent(slug)}`,
    },
  },

  lineReservation: {
    path: "/line-reservation",
    getHref: () => "/line-reservation",
    settings: { path: "/line-reservation/settings", getHref: () => "/line-reservation/settings" },
    pageEditor: {
      path: "/line-reservation/page-editor",
      getHref: () => "/line-reservation/page-editor",
    },
    slots: { path: "/line-reservation/slots", getHref: () => "/line-reservation/slots" },
  },

  lstep: {
    settings: {
      path: "/settings/integrations/lstep",
      getHref: () => "/settings/integrations/lstep",
    },
    tags: { path: "/settings/lstep/tags", getHref: () => "/settings/lstep/tags" },
    checkupSync: { path: "/lstep/checkup-sync", getHref: () => "/lstep/checkup-sync" },
    analytics: { path: "/lstep/analytics", getHref: () => "/lstep/analytics" },
    deliveryMonitor: { path: "/lstep/delivery-monitor", getHref: () => "/lstep/delivery-monitor" },
  },

  settings: {
    path: "/settings",
    getHref: () => "/settings",
    clinic: { path: "/settings/clinic", getHref: () => "/settings/clinic" },
    animalSpecies: { path: "/settings/animal-species", getHref: () => "/settings/animal-species" },
    staff: { path: "/settings/staff", getHref: () => "/settings/staff" },
    treatmentItems: settingsTreatmentItemsBase,
    diagnosis: settingsDiagnosisBase,
    trimming: settingsTrimmingBase,
    trimmingCourseType: {
      path: "/settings/trimming-course-type",
      getHref: () => "/settings/trimming-course-type",
    },
    medicine: { path: "/settings/medicine", getHref: () => "/settings/medicine" },
    reservationType: {
      path: "/settings/reservation-type",
      getHref: () => "/settings/reservation-type",
    },
    hospitalization: {
      path: "/settings/hospitalization",
      getHref: () => "/settings/hospitalization",
    },
    cage: { path: "/settings/cage", getHref: () => "/settings/cage" },
    insurance: { path: "/settings/insurance", getHref: () => "/settings/insurance" },
    occupations: { path: "/settings/occupations", getHref: () => "/settings/occupations" },
    permissionGroups: {
      path: "/settings/permission-groups",
      getHref: () => "/settings/permission-groups",
    },
    inquiryTemplates: {
      path: "/settings/inquiry-templates",
      getHref: () => "/settings/inquiry-templates",
    },
    merchandiseItems: {
      path: "/settings/merchandise-items",
      getHref: () => "/settings/merchandise-items",
    },
    // FE4-13: BUG-383 旧URL redirect先。settings-routes.tsx の shift-template → shift-templates
    shiftTemplates: {
      path: "/settings/shift-templates",
      getHref: () => "/settings/shift-templates",
    },
    // BUG-385: /settings/vaccine|examination|consultation|procedure は dead route。
    // router.tsx の実在ルートに合わせ、/settings/treatment-items?tab=xxx へ直接リンク。
    vaccine: withTab(settingsTreatmentItemsBase, "vaccine"),
    examination: withTab(settingsTreatmentItemsBase, "examination"),
    // BUG-384: 旧パス（trimming-course / trimming-option）は dead route。
    // /settings/trimming?tab=course|option に統合済み。
    trimmingCourse: withTab(settingsTrimmingBase, "course"),
    trimmingOption: withTab(settingsTrimmingBase, "option"),
    consultation: withTab(settingsTreatmentItemsBase, "consultation"),
    procedure: withTab(settingsTreatmentItemsBase, "procedure"),
    diagnosisType: withTab(settingsDiagnosisBase, "diagnosis_type"),
    diagnosisName: withTab(settingsDiagnosisBase, "diagnosis_name"),
    interview: {
      chiefComplaint: {
        path: "/settings/interview/chief-complaint",
        getHref: () => "/settings/interview/chief-complaint",
      },
      interviewTemplate: {
        path: "/settings/interview/templates",
        getHref: () => "/settings/interview/templates",
      },
    },
    // FEAT-368: 締め設定・支払方法マスタ
    closingTime: { path: "/settings/closing-time", getHref: () => "/settings/closing-time" },
    paymentMethods: {
      path: "/settings/payment-methods",
      getHref: () => "/settings/payment-methods",
    },
    campaigns: { path: "/settings/campaigns", getHref: () => "/settings/campaigns" },
    labDeviceItemMasters: {
      path: "/settings/lab-device-item-masters",
      getHref: () => "/settings/lab-device-item-masters",
    },
  },
} as const;
