import {
  calculateNextDate as calculateSharedNextDate,
  resolveScheduleTypeAfterManualDate,
} from "@/components/shared/NextScheduleField";
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

// EMR-430（2026-10-20 本人回答「ワクチン種類ごとの設定を使う。設定がなければ空欄。」）:
// 次回予定日を一律の既定間隔（1年後/4週後）で補完しない。種別のマスタ interval 設定に
// 従属し、種別未選択・設定なし種別は空欄（nextScheduleType: "" / nextDate: ""）とする。
// 共有定数 DEFAULT_NEXT_SCHEDULE_TYPE（@/constants/next-schedule）は健診フォーム側のみで使用。

const DEFAULT_VACCINATION_FORM: VaccinationFormState = {
  vaccineId: "",
  date: "",
  supplemental: "",
  lot1: "",
  lot2: "",
  lot3: "",
  lot4: "",
  nextScheduleType: "",
  nextDate: "",
  remarks: "",
};

export function calculateNextDate(vaccinationDate: string, scheduleType: string): string {
  return calculateSharedNextDate(vaccinationDate, scheduleType);
}

// BUG-401/BUG-026: vaccine interval (vaccines master 実データ) → schedule type。
// EMR-430（2026-10-20 本人回答）: 既知マッピング（"1年"→1year / "1ヶ月"→4weeks）以外は
// 設定なしとして空欄を返す。旧 1year へのフォールバックは廃止（既定日の捏造防止）。
function scheduleTypeForInterval(interval: string | undefined): string {
  switch (interval) {
    case "1年":
      return "1year";
    case "1ヶ月":
      return "4weeks";
    default:
      return "";
  }
}

export function mergeVaccinationFormData(
  isEdit: boolean,
  existingVaccination: VaccinationRecord | undefined,
  localOverrides: Partial<VaccinationFormState>,
): VaccinationFormState {
  if (isEdit && existingVaccination) {
    const date = existingVaccination.date ? existingVaccination.date.slice(0, 10) : "";
    const nextDate = existingVaccination.nextDate ? existingVaccination.nextDate.slice(0, 10) : "";
    // EMR-430（2026-10-20 本人回答）: 保存済み type が無い場合も 1year にフォールバックせず
    // 空欄として扱う。保存済み nextDate がある場合は後段の再判定で解決可能な間隔が無いため
    // other に落ちるが、nextDate の保存値は上書きしない。
    const nextScheduleType = existingVaccination.nextScheduleType ?? "";
    return {
      vaccineId: existingVaccination.vaccineId,
      date,
      supplemental: existingVaccination.supplemental ?? "",
      lot1: existingVaccination.lot1 ?? "",
      lot2: existingVaccination.lot2 ?? "",
      lot3: existingVaccination.lot3 ?? "",
      lot4: existingVaccination.lot4 ?? "",
      nextScheduleType: nextDate
        ? resolveScheduleTypeAfterManualDate(date, nextScheduleType, nextDate)
        : nextScheduleType,
      nextDate,
      remarks: existingVaccination.remarks ?? "",
      ...localOverrides,
    };
  }
  // EMR-430（2026-10-20 本人回答）: 新規フォームは種別未選択のため次回予定を空欄のまま
  // 開く。旧 EMR-431 の「当日＋既定間隔」自動入力は予防接種では撤回し、種別選択時に
  // マスタ interval に従属して初めて計算する（設定なし種別では空欄のまま保存できる）。
  const date = todayJSTISO();
  return {
    ...DEFAULT_VACCINATION_FORM,
    date,
    ...localOverrides,
  };
}

export function validateVaccinationForm(
  isEdit: boolean,
  formData: VaccinationFormState,
): Record<string, string> {
  const errors: Record<string, string> = {};
  // FE-RC-003: date-only 契約は JST の文字列比較で判定する（臨床安全境界）。
  // new Date() のローカル時刻比較は TZ != JST の実行環境で当日を未来日と誤判定する。
  const today = todayJSTISO();
  if (!isEdit) {
    if (!formData.vaccineId || formData.vaccineId === "0") {
      errors.vaccineId = "ワクチン種別を選択してください";
    }
    if (!formData.date) {
      errors.date = "接種日を入力してください";
    } else if (formData.date > today) {
      errors.date = "接種日は今日以前の日付を入力してください";
    }
  } else if (!formData.date) {
    errors.date = "接種日を入力してください";
  } else if (formData.date > today) {
    errors.date = "接種日は今日以前の日付を入力してください";
  }
  if (!isEdit && formData.nextDate) {
    if (formData.nextDate < today) {
      errors.nextDate = "次回予定日は本日以降の日付を入力してください";
    }
  }
  if (formData.date && formData.nextDate && formData.nextDate <= formData.date) {
    errors.nextDate = "次回予定日は接種日より後の日付を入力してください";
  }
  return errors;
}

export function buildUpdateVaccinationRequest(
  formData: VaccinationFormState,
  version: number,
): UpdateVaccinationRequest {
  const toRFC3339 = (d: string) => (d ? jstDateStartISOString(d) : undefined);
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
    // UAT-R2-EXCLUSIVE-LOCK: 読取済み version を expectedVersion として必ず同送する。
    version,
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

// EMR-430（2026-10-20 本人回答）: 種別選択時の次回予定はマスタ interval に従属する。
// interval === undefined（マスタ取得中・失敗・遅延で種別が解決できない）では既存の
// type/date を保持する。解決できた場合は設定の有無にかかわらず type と再計算した
// nextDate で上書きする（設定なし種別への切替時は空欄で上書きし、旧日付を残さない）。
export function vaccinationOverridesOnVaccineId(
  prev: Partial<VaccinationFormState>,
  vaccineId: string,
  currentDate: string,
  interval: string | undefined,
): Partial<VaccinationFormState> {
  if (interval === undefined) {
    return { ...prev, vaccineId };
  }
  const scheduleType = scheduleTypeForInterval(interval);
  return {
    ...prev,
    vaccineId,
    nextScheduleType: scheduleType,
    nextDate: calculateNextDate(currentDate, scheduleType),
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
  // レビュー指摘（EMR-431）: 間隔の切替時は次回予定日を必ず再計算する。
  // 計算できない場合（以外（手動）・実施日空）は既定で自動計算した日付を
  // 「手動」の値として残さないよう、空文字で上書きする。
  return {
    ...prev,
    nextScheduleType: scheduleType,
    nextDate: calculateNextDate(currentDate, scheduleType),
  };
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
