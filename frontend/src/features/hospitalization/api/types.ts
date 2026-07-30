import type { HospitalizationResponse } from "@/types/generated/hospitalization-responses";

/**
 * Backend hospitalization list/detail/create/update wire.
 * Source of truth: HospitalizationResponse (domain-owned DTO / tygo), not models.Hospitalization.
 * treatment_plans are NOT on this payload — use GET /hospitalizations/:id/treatment-plans.
 */
export type BackendHospitalization = HospitalizationResponse;

export interface CreateHospitalizationRequest {
  pet_id: string;
  owner_id: string;
  cage_id?: string;
  hospitalization_type: string;
  start_date: string;
  end_date: string;
  owner_request?: string;
  staff_notes?: string;
  memo?: string;
  is_insurance?: boolean;
  insurance_company_name?: string | null;
  insurance_number?: string | null;
}

export interface UpdateHospitalizationRequest {
  hospitalization_type?: string;
  cage_id?: string;
  end_date?: string;
  status?: string;
  owner_request?: string;
  staff_notes?: string;
  memo?: string;
  is_insurance?: boolean;
  insurance_company_name?: string | null;
  insurance_number?: string | null;
}
