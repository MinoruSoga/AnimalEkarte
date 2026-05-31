import type {
  CreateInsuranceRequest,
  UpdateInsuranceRequest,
} from "../api/insurances";
import type { InsuranceFormData } from "../components/InsuranceSidePanelModel";

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
