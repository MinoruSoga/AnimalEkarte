import type {
  CreateInsuranceRequest,
  UpdateInsuranceRequest,
} from "../api/insurances";
import type { InsuranceFormData } from "../components/insurance-side-panel-model";

/** BUG-026: FE/BE 同一境界（0〜100）。範囲外は POST せずエラー表示する。 */
export function validateInsuranceForm(data: InsuranceFormData): string | null {
  if (!data.name.trim()) {
    return "名称は必須です";
  }
  const raw = data.coverageRate.trim();
  if (raw === "") {
    return null;
  }
  const rate = Number(raw);
  if (!Number.isFinite(rate) || !Number.isInteger(rate)) {
    return "補償率は0〜100の整数で入力してください";
  }
  if (rate < 0 || rate > 100) {
    return "補償率は0〜100の範囲で入力してください";
  }
  return null;
}

export function buildInsuranceCreateRequest(
  data: InsuranceFormData,
): CreateInsuranceRequest {
  return {
    name: data.name,
    description: data.description || undefined,
    coverage_rate: data.coverageRate !== "" ? Number(data.coverageRate) : 0,
    contact_phone: data.contactPhone || undefined,
    is_active: data.isActive,
  };
}

export function buildInsuranceUpdateRequest(
  data: InsuranceFormData,
): UpdateInsuranceRequest {
  return buildInsuranceCreateRequest(data);
}
