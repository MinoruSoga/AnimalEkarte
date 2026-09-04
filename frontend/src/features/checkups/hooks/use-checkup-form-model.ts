import {
  calculateNextDate,
  resolveScheduleTypeAfterManualDate,
} from "@/components/shared/NextScheduleField";

export interface CheckupFormState {
  checkupTypeId: string;
  date: string;
  nextScheduleType: string;
  nextDate: string;
  doctorId: string;
  result: string;
}

export interface CheckupMutationPermissions {
  canCreate: boolean;
  canEdit: boolean;
}

export const DENIED_MUTATION_PERMISSIONS: Readonly<CheckupMutationPermissions> = {
  canCreate: false,
  canEdit: false,
};

export const DEFAULT_CHECKUP_FORM: CheckupFormState = {
  checkupTypeId: "",
  date: "",
  nextScheduleType: "1year",
  nextDate: "",
  doctorId: "",
  result: "",
};

export function validateCheckupForm(formData: CheckupFormState): Record<string, string> {
  const errors: Record<string, string> = {};
  if (!formData.checkupTypeId) errors.checkupTypeId = "健診種別を選択してください";
  if (!formData.date) errors.date = "実施日を入力してください";
  return errors;
}

export function buildCheckupOnMedicalRecordRequest(formData: CheckupFormState) {
  return {
    checkup_type_id: Number(formData.checkupTypeId),
    date: formData.date,
    ...(formData.nextDate ? { next_date: formData.nextDate } : {}),
    ...(formData.doctorId ? { doctor_id: Number(formData.doctorId) } : {}),
    ...(formData.result ? { result: formData.result } : {}),
  };
}

export function checkupOverridesOnDate(
  prev: Partial<CheckupFormState>,
  date: string,
): Partial<CheckupFormState> {
  const scheduleType = prev.nextScheduleType ?? DEFAULT_CHECKUP_FORM.nextScheduleType;
  const calculated = calculateNextDate(date, scheduleType);
  return { ...prev, date, ...(calculated ? { nextDate: calculated } : {}) };
}

export function checkupOverridesOnScheduleType(
  prev: Partial<CheckupFormState>,
  scheduleType: string,
): Partial<CheckupFormState> {
  const currentDate = prev.date ?? DEFAULT_CHECKUP_FORM.date;
  const calculated = calculateNextDate(currentDate, scheduleType);
  return {
    ...prev,
    nextScheduleType: scheduleType,
    ...(calculated ? { nextDate: calculated } : {}),
  };
}

export function checkupOverridesOnNextDate(
  prev: Partial<CheckupFormState>,
  nextDate: string,
): Partial<CheckupFormState> {
  const currentDate = prev.date ?? DEFAULT_CHECKUP_FORM.date;
  const currentType = prev.nextScheduleType ?? DEFAULT_CHECKUP_FORM.nextScheduleType;
  return {
    ...prev,
    nextDate,
    nextScheduleType: resolveScheduleTypeAfterManualDate(currentDate, currentType, nextDate),
  };
}
