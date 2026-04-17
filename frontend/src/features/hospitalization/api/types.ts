import type { Hospitalization } from "@/types/generated/models";

export type BackendHospitalization = Hospitalization;

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
}

export interface UpdateHospitalizationRequest {
  hospitalization_type?: string;
  cage_id?: string;
  end_date?: string;
  status?: string;
  owner_request?: string;
  staff_notes?: string;
  memo?: string;
}
