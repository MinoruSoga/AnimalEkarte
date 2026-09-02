import { calculateNextDate as calculateSharedNextDate, resolveScheduleTypeAfterManualDate } from "@/components/shared/NextScheduleField";
import { jstDateStartISOString, todayJSTISO } from "@/lib/jst-date";
import type { CreateVaccinationRequest, UpdateVaccinationRequest } from "../api/types";
import type { VaccinationRecord } from "@/types";

export interface VaccinationFormState {
  vaccineId: string;
  date: string;
  supplemental: string;
  lot1: string;
  lot2: string;
  lot3: string;
  lot4: string;
  nextScheduleType: string;
  nextDate: string;
  remarks: string;
}

export interface VaccinationMutationPermissions {
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
}

export const DENIED_MUTATION_PERMISSIONS: Readonly<VaccinationMutationPermissions> = {
  canCreate: false,
  canEdit: false,
  canDelete: false,
};

export const DEFAULT_NEXT_SCHEDULE_TYPE = "1year" as const;

const DEFAULT_VACCINATION_FORM: VaccinationFormState = {
  vaccineId: "",
  date: "",
  supplemental: "",
  lot1: "",
  lot2: "",
  lot3: "",
  lot4: "",
  nextScheduleType: DEFAULT_NEXT_SCHEDULE_TYPE,
  nextDate: "",
  remarks: "",
};

export function calculateNextDate(vaccinationDate: string, scheduleType: string): string {
  return calculateSharedNextDate(vaccinationDate, scheduleType);
}

// BUG-401/BUG-026: vaccine interval (vaccines master 実データ) → schedule type。
function scheduleTypeForInterval(interval: string | undefined): string {
  switch (interval) {
    case "1年":
      return "1year";
    case "1ヶ月":
      return "4weeks";
    default:
      return DEFAULT_NEXT_SCHEDULE_TYPE;
  }
}

export function mergeVaccinationFormData(
  isEdit: boolean,
  existingVaccination: VaccinationRecord | undefined,
  localOverrides: Partial<VaccinationFormState>,
): VaccinationFormState {
  if (isEdit && existingVaccination) {
    return {
      vaccineId: existingVaccination.vaccineId,
      date: existingVaccination.date ? existingVaccination.date.slice(0, 10) : "",
      supplemental: existingVaccination.supplemental ?? "",
      lot1: existingVaccination.lot1 ?? "",
      lot2: existingVaccination.lot2 ?? "",
      lot3: existingVaccination.lot3 ?? "",
      lot4: existingVaccination.lot4 ?? "",
      nextScheduleType: existingVaccination.nextScheduleType ?? DEFAULT_NEXT_SCHEDULE_TYPE,
      nextDate: existingVaccination.nextDate ? existingVaccination.nextDate.slice(0, 10) : "",
      remarks: existingVaccination.remarks ?? "",
      ...localOverrides,
    };
  }
  return { ...DEFAULT_VACCINATION_FORM, date: todayJSTISO(), ...localOverrides };
}

export function validateVaccinationForm(
  isEdit: boolean,
  formData: VaccinationFormState,
): Record<string, string> {
  const errors: Record<string, string> = {};
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  if (!isEdit) {
    if (!formData.vaccineId || formData.vaccineId === "0") {
      errors.vaccineId = "ワクチン種別を選択してください";
    }
    if (!formData.date) {
      errors.date = "接種日を入力してください";
    } else if (new Date(formData.date + "T00:00:00") > today) {
      errors.date = "接種日は今日以前の日付を入力してください";
    }
  } else if (!formData.date) {
    errors.date = "接種日を入力してください";
  } else if (new Date(formData.date + "T00:00:00") > today) {
    errors.date = "接種日は今日以前の日付を入力してください";
  }
  if (!isEdit && formData.nextDate) {
    if (new Date(formData.nextDate + "T00:00:00") < today) {
      errors.nextDate = "次回予定日は本日以降の日付を入力してください";
    }
  }
  if (formData.date && formData.nextDate) {
    const dateVal = new Date(formData.date + "T00:00:00");
    const nextDateVal = new Date(formData.nextDate + "T00:00:00");
    if (!isNaN(dateVal.getTime()) && !isNaN(nextDateVal.getTime()) && nextDateVal <= dateVal) {
      errors.nextDate = "次回予定日は接種日より後の日付を入力してください";
    }
  }
  return errors;
}

export function buildUpdateVaccinationRequest(
  formData: VaccinationFormState,
): UpdateVaccinationRequest {
  const toRFC3339 = (d: string) => d ? jstDateStartISOString(d) : undefined;
  return {
    date: toRFC3339(formData.date),
    next_date: formData.nextDate ? jstDateStartISOString(formData.nextDate) : null,
    lot1: formData.lot1 || undefined,
    lot2: formData.lot2 || undefined,
    lot3: formData.lot3 || undefined,
    lot4: formData.lot4 || undefined,
    remarks: formData.remarks || undefined,
    supplemental: formData.supplemental || undefined,
    next_schedule_type: formData.nextScheduleType || undefined,
  };
}

export function buildCreateVaccinationRequest(
  formData: VaccinationFormState,
  petId: string,
): CreateVaccinationRequest {
  return {
    medical_record_id: null,
    pet_id: Number(petId),
    vaccine_id: Number(formData.vaccineId),
    date: jstDateStartISOString(formData.date || todayJSTISO()),
    next_date: formData.nextDate ? jstDateStartISOString(formData.nextDate) : null,
    lot1: formData.lot1 || undefined,
    lot2: formData.lot2 || undefined,
    lot3: formData.lot3 || undefined,
    lot4: formData.lot4 || undefined,
    remarks: formData.remarks || undefined,
    supplemental: formData.supplemental || undefined,
    next_schedule_type: formData.nextScheduleType || undefined,
  };
}

export function vaccinationOverridesOnVaccineId(
  prev: Partial<VaccinationFormState>,
  vaccineId: string,
  currentDate: string,
  interval: string | undefined,
): Partial<VaccinationFormState> {
  const scheduleType = scheduleTypeForInterval(interval);
  const calculated = calculateNextDate(currentDate, scheduleType);
  return {
    ...prev,
    vaccineId,
    nextScheduleType: scheduleType,
    ...(calculated ? { nextDate: calculated } : {}),
  };
}

export function vaccinationOverridesOnDate(
  prev: Partial<VaccinationFormState>,
  date: string,
  scheduleType: string,
): Partial<VaccinationFormState> {
  const calculated = calculateNextDate(date, scheduleType);
  return { ...prev, date, ...(calculated ? { nextDate: calculated } : {}) };
}

export function vaccinationOverridesOnScheduleType(
  prev: Partial<VaccinationFormState>,
  scheduleType: string,
  currentDate: string,
): Partial<VaccinationFormState> {
  const calculated = calculateNextDate(currentDate, scheduleType);
  return { ...prev, nextScheduleType: scheduleType, ...(calculated ? { nextDate: calculated } : {}) };
}

export function vaccinationOverridesOnNextDate(
  prev: Partial<VaccinationFormState>,
  nextDate: string,
  vaccinationDate: string,
  currentType: string,
): Partial<VaccinationFormState> {
  return {
    ...prev,
    nextDate,
    nextScheduleType: resolveScheduleTypeAfterManualDate(vaccinationDate, currentType, nextDate),
  };
}
