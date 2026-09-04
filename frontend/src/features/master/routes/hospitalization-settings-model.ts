import type {
  CreateHospitalizationPlanRequest,
  UpdateHospitalizationPlanRequest,
} from "../api/hospitalization-plans";
import type { HospitalizationFormData } from "../lib/hospitalization-side-panel-model";

export function buildHospitalizationCreateRequest(
  data: HospitalizationFormData,
): CreateHospitalizationPlanRequest {
  return {
    name: data.name,
    price: data.price || undefined,
    description: data.description || undefined,
    is_active: data.isActive,
    body_size: data.bodySize !== "" ? data.bodySize : undefined,
    billing_unit: data.billingUnit !== "" ? data.billingUnit : undefined,
    tax_type: data.taxType,
    tax_rate: data.taxRate,
  };
}

export function buildHospitalizationUpdateRequest(
  data: HospitalizationFormData,
): UpdateHospitalizationPlanRequest {
  return {
    name: data.name,
    price: data.price,
    description: data.description || undefined,
    is_active: data.isActive,
    body_size: data.bodySize !== "" ? data.bodySize : null,
    billing_unit: data.billingUnit !== "" ? data.billingUnit : null,
    tax_type: data.taxType,
    tax_rate: data.taxRate,
  };
}
