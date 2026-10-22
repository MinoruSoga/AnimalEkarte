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

/**
 * EMR-447 Q4（SPEC-QA-20261022-447-STRENGTH-CLEAR）: per_weight の必須含量が空なら保存前に止める。
 * 0・形式不正は補正せず BE の exact decimal 検証（400）に任せる。
 */
export function getMedicineStrengthError(
  data: Pick<MedicineFormData, "calculationType" | "strength">,
): string | null {
  if (data.calculationType === MedicineCalculationTypePerWeight && !data.strength.trim()) {
    return MEDICINE_STRENGTH_REQUIRED_MESSAGE;
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
