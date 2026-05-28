import type { Medicine } from "@/types";
import type { TaxType } from "@/types/generated/models";

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
}

const INITIAL_FORM: MedicineFormData = {
  name: "",
  parentId: "",
  dosageForm: "tablet",
  medicineUnit: "per_tablet",
  price: 0,
  description: "",
  isActive: true,
  taxType: "excluded",
  taxRate: 0.1,
  isNonInsurance: false,
};

export function medicineToFormData(medicine: Medicine | null, defaultParentId?: string): MedicineFormData {
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
  };
}
