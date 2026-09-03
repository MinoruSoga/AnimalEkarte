import type { Insurance } from "../api/insurances";

export interface InsuranceFormData {
  name: string;
  description: string;
  coverageRate: string;
  contactPhone: string;
  isActive: boolean;
}

export function insuranceToFormData(item: Insurance | null): InsuranceFormData {
  return {
    name: item?.name ?? "",
    description: item?.description ?? "",
    coverageRate: item?.coverageRate != null ? String(item.coverageRate) : "0",
    contactPhone: item?.contactPhone ?? "",
    isActive: item?.isActive ?? true,
  };
}
