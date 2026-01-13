export const HOSPITALIZATION_STATUS = {
  ACTIVE: "入院中",
  DISCHARGED: "退院済",
  RESERVED: "予約",
  TEMP_DISCHARGE: "一時帰宅",
} as const;

export const HOSPITALIZATION_FILTER_STATUS = {
  ALL: "all",
  ACTIVE: "active",
  DISCHARGED: "discharged",
  RESERVED: "reserved",
} as const;

export type HospitalizationFilterStatus = typeof HOSPITALIZATION_FILTER_STATUS[keyof typeof HOSPITALIZATION_FILTER_STATUS];
