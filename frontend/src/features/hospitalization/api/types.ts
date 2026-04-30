import type { Hospitalization } from "@/types/generated/models";

// Extend generated type with fields added in migration 006 (pending codegen)
export type BackendHospitalization = Hospitalization & {
  insurance_company_name?: string | null;
  insurance_number?: string | null;
};

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
