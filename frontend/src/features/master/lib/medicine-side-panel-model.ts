import type { Medicine } from "@/types";
import type {
  MedicineCalculationType,
  MedicineDoseSpecies,
  TaxType,
} from "@/types/generated/models";
import {
  MedicineCalculationTypeNone,
  MedicineCalculationTypePerWeight,
} from "@/types/generated/models";
import { isExactDecimalString } from "@/lib/exact-decimal";
import type { UpsertMedicineDoseParamRequest } from "../api/medicine-dose-params";

type DoseParamDraft = {
  species: MedicineDoseSpecies;
  input: UpsertMedicineDoseParamRequest;
};

export interface MedicineFormData {
  name: string;
  parentId: string;
  dosageForm: string;
  medicineUnit: string;
  price: number;
  description: string;
  isActive: boolean;
  taxType: TaxType;
  taxRate: number;
  isNonInsurance: boolean;
  /** #201 投与量自動計算（製品軸）。none=手動（既定） */
  calculationType: MedicineCalculationType;
  /** 製品含量 mg/単位。テキスト入力のため文字列で保持し、送信時に parse する */
  strength: string;
  frequencyPerDay: string;
  defaultDurationDays: string;
  doseParamDrafts?: DoseParamDraft[];
}

const INITIAL_FORM: MedicineFormData = {
  name: "",
  parentId: "",
  dosageForm: "tablet",
  medicineUnit: "per_tablet",
  price: 0,
  description: "",
  isActive: true,
  taxType: "included",
  taxRate: 0.1,
  isNonInsurance: false,
  calculationType: MedicineCalculationTypeNone,
  strength: "",
  frequencyPerDay: "",
  defaultDurationDays: "",
};

export const MEDICINE_STRENGTH_REQUIRED_MESSAGE = "体重換算(mg/kg)では製品含量を入力してください";

/** BE の strength 保存尺度（numeric(10,4)。medicine_decimal.go medicineStrengthSpec と同じ） */
const MEDICINE_STRENGTH_MAX_INT_DIGITS = 6;
const MEDICINE_STRENGTH_MAX_FRAC_DIGITS = 4;

/**
 * EMR-447 Q4（SPEC-QA-20261022-447-STRENGTH-CLEAR）: per_weight の必須含量が空なら保存前に止める。
 * 送信される含量（per_weight のときだけ）は BE と同じ字句・桁・正値の規則で先に検査し、
 * 違反なら用量パラメータを含むどの書込より前に止める（値は補正しない。BE の 400 は残す）。
 */
export function getMedicineStrengthError(
  data: Pick<MedicineFormData, "calculationType" | "strength">,
): string | null {
  if (data.calculationType !== MedicineCalculationTypePerWeight) {
    return null;
  }
  const strength = data.strength.trim();
  if (!strength) {
    return MEDICINE_STRENGTH_REQUIRED_MESSAGE;
  }
  if (!isExactDecimalString(strength)) {
    return "製品含量は符号・指数のない数値（例: 12.5）で入力してください";
  }
  const [intPart, frac = ""] = strength.split(".");
  if (intPart.length > MEDICINE_STRENGTH_MAX_INT_DIGITS) {
    return `製品含量の整数部は${MEDICINE_STRENGTH_MAX_INT_DIGITS}桁までで入力してください`;
  }
  const significantFrac = frac.replace(/0+$/, "");
  if (significantFrac.length > MEDICINE_STRENGTH_MAX_FRAC_DIGITS) {
    return `製品含量は小数第${MEDICINE_STRENGTH_MAX_FRAC_DIGITS}位までで入力してください`;
  }
  if (intPart === "0" && significantFrac === "") {
    return "製品含量は0より大きい値を入力してください";
  }
  return null;
}

export function medicineToFormData(
  medicine: Medicine | null,
  defaultParentId?: string,
): MedicineFormData {
  if (!medicine) {
    return {
      ...INITIAL_FORM,
      parentId: defaultParentId && defaultParentId !== "uncategorized" ? defaultParentId : "",
    };
  }

  return {
    name: medicine.name,
    parentId: medicine.parentId ?? "",
    dosageForm: medicine.dosageForm ?? "",
    medicineUnit: medicine.medicineUnit ?? "",
    price: medicine.price,
    description: medicine.description,
    isActive: medicine.isActive,
    taxType: medicine.taxType ?? "excluded",
    taxRate: medicine.taxRate ?? 0.1,
    isNonInsurance: medicine.isNonInsurance ?? false,
    calculationType: medicine.calculationType ?? MedicineCalculationTypeNone,
    strength: medicine.strength !== undefined ? String(medicine.strength) : "",
    frequencyPerDay: medicine.frequencyPerDay !== undefined ? String(medicine.frequencyPerDay) : "",
    defaultDurationDays:
      medicine.defaultDurationDays !== undefined ? String(medicine.defaultDurationDays) : "",
  };
}
