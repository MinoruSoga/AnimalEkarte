import type { BodySize, BillingUnit, TaxType } from "@/types/generated/models";

import type { HospitalizationPlan } from "../api/hospitalization-plans";

export interface HospitalizationFormData {
  name: string;
  price: number;
  description: string;
  isActive: boolean;
  bodySize: BodySize | "";
  billingUnit: BillingUnit | "";
  taxType: TaxType;
  taxRate: number;
}

export function hospitalizationToFormData(
  item: HospitalizationPlan | null,
): HospitalizationFormData {
  return {
    name: item?.name ?? "",
    price: item?.price ?? 0,
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
    bodySize: item?.bodySize ?? "",
    billingUnit: item?.billingUnit ?? "",
    taxType: item?.taxType ?? "excluded",
    taxRate: item?.taxRate ?? 0.1,
  };
}
