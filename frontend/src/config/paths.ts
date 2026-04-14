/**
 * config/paths.ts — アプリケーション全URLの一元管理
 *
 * 使用例:
 *   import { paths } from "@/config/paths";
 *   navigate(paths.owners.detail.getHref(ownerId));
 */

export const paths = {
  home: {
    path: "/",
    getHref: () => "/",
  },

  auth: {
    login: {
      path: "/login",
      getHref: () => "/login",
    },
    forgotPassword: {
      path: "/forgot-password",
      getHref: () => "/forgot-password",
    },
    resetPassword: {
      path: "/reset-password",
      getHref: () => "/reset-password",
    },
  },

  owners: {
    path: "/owners",
    getHref: () => "/owners",
    new: {
      path: "/owners/new",
      getHref: () => "/owners/new",
    },
    detail: {
      path: "/owners/:id",
      getHref: (id: string | number) => `/owners/${id}`,
    },
  },

  reservations: {
    path: "/reservations",
    getHref: () => "/reservations",
  },

  medicalRecords: {
    path: "/medical-records",
    getHref: () => "/medical-records",
    selectPet: {
      path: "/medical-records/select-pet",
      getHref: () => "/medical-records/select-pet",
    },
    new: {
      path: "/medical-records/new",
      getHref: () => "/medical-records/new",
    },
    detail: {
      path: "/medical-records/:id",
      getHref: (id: string | number) => `/medical-records/${id}`,
    },
  },

  hospitalization: {
    path: "/hospitalization",
    getHref: () => "/hospitalization",
    selectPet: {
      path: "/hospitalization/select-pet",
      getHref: () => "/hospitalization/select-pet",
    },
    new: {
      path: "/hospitalization/new",
      getHref: () => "/hospitalization/new",
    },
    detail: {
      path: "/hospitalization/:id",
      getHref: (id: string | number) => `/hospitalization/${id}`,
    },
    edit: {
      path: "/hospitalization/:id/edit",
      getHref: (id: string | number) => `/hospitalization/${id}/edit`,
    },
  },

  trimming: {
    path: "/trimming",
    getHref: () => "/trimming",
    selectPet: {
      path: "/trimming/select-pet",
      getHref: () => "/trimming/select-pet",
    },
    new: {
      path: "/trimming/new",
      getHref: () => "/trimming/new",
    },
    detail: {
      path: "/trimming/:id",
      getHref: (id: string | number) => `/trimming/${id}`,
    },
  },

  examinations: {
    path: "/examinations",
    getHref: () => "/examinations",
    selectPet: {
      path: "/examinations/select-pet",
      getHref: () => "/examinations/select-pet",
    },
    new: {
      path: "/examinations/new",
      getHref: () => "/examinations/new",
    },
    detail: {
      path: "/examinations/:id",
      getHref: (id: string | number) => `/examinations/${id}`,
    },
  },

  accounting: {
    path: "/accounting",
    getHref: () => "/accounting",
    selectPet: {
      path: "/accounting/select-pet",
      getHref: () => "/accounting/select-pet",
    },
    new: {
      path: "/accounting/new",
      getHref: () => "/accounting/new",
    },
    // BUG-370: 月末未納者一覧
    unpaid: {
      path: "/accounting/unpaid",
      getHref: () => "/accounting/unpaid",
    },
    detail: {
      path: "/accounting/:id",
      getHref: (id: string | number) => `/accounting/${id}`,
    },
  },

  vaccinations: {
    path: "/vaccinations",
    getHref: () => "/vaccinations",
    selectPet: {
      path: "/vaccinations/select-pet",
      getHref: () => "/vaccinations/select-pet",
    },
    new: {
      path: "/vaccinations/new",
      getHref: () => "/vaccinations/new",
    },
    detail: {
      path: "/vaccinations/:id",
      getHref: (id: string | number) => `/vaccinations/${id}`,
    },
  },

  inventory: {
    path: "/inventory",
    getHref: () => "/inventory",
    new: {
      path: "/inventory/new",
      getHref: () => "/inventory/new",
    },
    detail: {
      path: "/inventory/:id",
      getHref: (id: string | number) => `/inventory/${id}`,
    },
  },

  estimates: {
    path: "/estimates",
    getHref: () => "/estimates",
    new: {
      path: "/estimates/new",
      getHref: () => "/estimates/new",
    },
    detail: {
      path: "/estimates/:id",
      getHref: (id: string | number) => `/estimates/${id}`,
    },
    edit: {
      path: "/estimates/:id/edit",
      getHref: (id: string | number) => `/estimates/${id}/edit`,
    },
  },

  shifts: {
    path: "/shifts",
    getHref: () => "/shifts",
  },

  lineReservation: {
    path: "/line-reservation",
    getHref: () => "/line-reservation",
    settings: { path: "/line-reservation/settings", getHref: () => "/line-reservation/settings" },
    pageEditor: { path: "/line-reservation/page-editor", getHref: () => "/line-reservation/page-editor" },
  },

  settings: {
    path: "/settings",
    getHref: () => "/settings",
    clinic: {
      path: "/settings/clinic",
      getHref: () => "/settings/clinic",
    },
    animalSpecies: {
      path: "/settings/animal-species",
      getHref: () => "/settings/animal-species",
    },
    staff: {
      path: "/settings/staff",
      getHref: () => "/settings/staff",
    },
    treatmentItems: {
      path: "/settings/treatment-items",
      getHref: () => "/settings/treatment-items",
    },
    diagnosis: {
      path: "/settings/diagnosis",
      getHref: () => "/settings/diagnosis",
    },
    trimming: {
      path: "/settings/trimming",
      getHref: () => "/settings/trimming",
    },
    medicine: {
      path: "/settings/medicine",
      getHref: () => "/settings/medicine",
    },
    reservationType: {
      path: "/settings/reservation-type",
      getHref: () => "/settings/reservation-type",
    },
    hospitalization: {
      path: "/settings/hospitalization",
      getHref: () => "/settings/hospitalization",
    },
    cage: {
      path: "/settings/cage",
      getHref: () => "/settings/cage",
    },
    insurance: {
      path: "/settings/insurance",
      getHref: () => "/settings/insurance",
    },
    occupations: {
      path: "/settings/occupations",
      getHref: () => "/settings/occupations",
    },
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
    vaccine: {
      path: "/settings/vaccine",
      getHref: () => "/settings/vaccine",
    },
    examination: {
      path: "/settings/examination",
      getHref: () => "/settings/examination",
    },
    // BUG-384: 旧パス（trimming-course / trimming-option）は dead route。
    // /settings/trimming?tab=course|option に統合済み。
    trimmingCourse: {
      path: "/settings/trimming",
      getHref: () => "/settings/trimming?tab=course",
    },
    trimmingOption: {
      path: "/settings/trimming",
      getHref: () => "/settings/trimming?tab=option",
    },
    consultation: {
      path: "/settings/consultation",
      getHref: () => "/settings/consultation",
    },
    procedure: {
      path: "/settings/procedure",
      getHref: () => "/settings/procedure",
    },
    diagnosisType: {
      path: "/settings/diagnosis-type",
      getHref: () => "/settings/diagnosis-type",
    },
    diagnosisName: {
      path: "/settings/diagnosis-name",
      getHref: () => "/settings/diagnosis-name",
    },
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
  },
} as const;
