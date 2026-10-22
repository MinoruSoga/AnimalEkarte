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
    // EMR-528: 明示入力の 0 円を未入力へ潰さない。
    price: data.price,
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
    // EMR-528: 備考を空へ変更できるよう、空文字も省略せず送る（BE は "" を空で保存）。
    description: data.description,
    is_active: data.isActive,
    body_size: data.bodySize !== "" ? data.bodySize : null,
    billing_unit: data.billingUnit !== "" ? data.billingUnit : null,
    tax_type: data.taxType,
    tax_rate: data.taxRate,
  };
}
