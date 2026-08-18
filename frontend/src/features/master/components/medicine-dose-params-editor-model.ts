import type {
  MedicineDoseBasis,
  MedicineDoseParam,
  MedicineDoseSpecies,
  MedicineRoundingMode,
} from "@/types/generated/models";
import { MedicineDoseBasisPerAdministration } from "@/types/generated/models";
import { parseOptionalNumber } from "@/lib/parse-optional-number";
import type { UpsertMedicineDoseParamRequest } from "../api/medicine-dose-params";

export interface DoseParamFormData {
  doseBasis: MedicineDoseBasis;
  dosePerKg: string;
  minMgPerKg: string;
  maxMgPerKg: string;
  absoluteMaxDose: string;
  roundingStep: string;
  roundingMode: MedicineRoundingMode | "";
  notes: string;
}

export const INITIAL_DOSE_PARAM_FORM: DoseParamFormData = {
  doseBasis: MedicineDoseBasisPerAdministration,
  dosePerKg: "",
  minMgPerKg: "",
  maxMgPerKg: "",
  absoluteMaxDose: "",
  roundingStep: "",
  roundingMode: "",
  notes: "",
};

export function isDoseParamFormEmpty(form: DoseParamFormData): boolean {
  return (
    form.dosePerKg.trim() === "" &&
    form.minMgPerKg.trim() === "" &&
    form.maxMgPerKg.trim() === "" &&
    form.absoluteMaxDose.trim() === "" &&
    form.roundingStep.trim() === "" &&
    form.roundingMode === "" &&
    form.notes.trim() === ""
  );
}

export function doseParamToFormData(param: MedicineDoseParam | undefined): DoseParamFormData {
  if (!param) return INITIAL_DOSE_PARAM_FORM;
  return {
    doseBasis: param.dose_basis,
    dosePerKg: String(param.dose_per_kg),
    minMgPerKg: param.min_mg_per_kg !== undefined ? String(param.min_mg_per_kg) : "",
    maxMgPerKg: param.max_mg_per_kg !== undefined ? String(param.max_mg_per_kg) : "",
    absoluteMaxDose: param.absolute_max_dose !== undefined ? String(param.absolute_max_dose) : "",
    roundingStep: param.rounding_step !== undefined ? String(param.rounding_step) : "",
    roundingMode: param.rounding_mode ?? "",
    notes: param.notes ?? "",
  };
}

/**
 * クライアント側の一次バリデーション（UX 用・非ブロッキング寄り）。
 * BE が最終権威（apperror メッセージは handleApiError 経由で表示される）。
 */
export interface DoseParamValidationResult {
  isValid: boolean;
  errors: string[];
}

export function validateDoseParamForm(form: DoseParamFormData): DoseParamValidationResult {
  const errors: string[] = [];
  const dosePerKg = parseOptionalNumber(form.dosePerKg);
  const minMgPerKg = parseOptionalNumber(form.minMgPerKg);
  const maxMgPerKg = parseOptionalNumber(form.maxMgPerKg);
  const absoluteMaxDose = parseOptionalNumber(form.absoluteMaxDose);

  if (dosePerKg === undefined || dosePerKg <= 0) {
    errors.push("投与量(mg/kg)は0より大きい値を入力してください");
  }
  if (minMgPerKg !== undefined && maxMgPerKg !== undefined && minMgPerKg > maxMgPerKg) {
    errors.push("下限は上限以下にしてください");
  }
  if (
    dosePerKg !== undefined &&
    minMgPerKg !== undefined &&
    dosePerKg < minMgPerKg
  ) {
    errors.push("投与量は下限以上にしてください");
  }
  if (
    dosePerKg !== undefined &&
    maxMgPerKg !== undefined &&
    dosePerKg > maxMgPerKg
  ) {
    errors.push("投与量は上限以下にしてください");
  }
  if (maxMgPerKg === undefined && absoluteMaxDose === undefined) {
    errors.push("過量防止のため上限(mg/kgまたはmg)のいずれかは必須です");
  }
  if ((form.roundingStep.trim() !== "") !== (form.roundingMode !== "")) {
    errors.push("丸め幅と丸め方向はどちらも設定するか、どちらも未設定にしてください");
  }

  return { isValid: errors.length === 0, errors };
}

export function buildUpsertDoseParamRequest(form: DoseParamFormData): UpsertMedicineDoseParamRequest {
  const dosePerKg = parseOptionalNumber(form.dosePerKg) ?? 0;
  return {
    dose_basis: form.doseBasis,
    dose_per_kg: dosePerKg,
    ...(parseOptionalNumber(form.minMgPerKg) !== undefined
      ? { min_mg_per_kg: parseOptionalNumber(form.minMgPerKg) }
      : {}),
    ...(parseOptionalNumber(form.maxMgPerKg) !== undefined
      ? { max_mg_per_kg: parseOptionalNumber(form.maxMgPerKg) }
      : {}),
    ...(parseOptionalNumber(form.absoluteMaxDose) !== undefined
      ? { absolute_max_dose: parseOptionalNumber(form.absoluteMaxDose) }
      : {}),
    ...(parseOptionalNumber(form.roundingStep) !== undefined
      ? { rounding_step: parseOptionalNumber(form.roundingStep) }
      : {}),
    ...(form.roundingMode ? { rounding_mode: form.roundingMode } : {}),
    ...(form.notes.trim() ? { notes: form.notes } : {}),
  };
}

export function findDoseParamBySpecies(
  params: MedicineDoseParam[] | undefined,
  species: MedicineDoseSpecies
): MedicineDoseParam | undefined {
  return params?.find((p) => p.species === species);
}
