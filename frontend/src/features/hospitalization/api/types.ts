import type { HospitalizationResponse } from "@/types/generated/hospitalization-responses";

/**
 * Backend hospitalization list/detail/create/update wire.
 * Source of truth: HospitalizationResponse (domain-owned DTO / tygo), not models.Hospitalization.
 * treatment_plans are NOT on this payload — use GET /hospitalizations/:id/treatment-plans.
 */
export type BackendHospitalization = HospitalizationResponse;

/** Nested plan on POST /hospitalizations (TASK-001 atomic create). Same shape as createTreatmentPlanRequest. */
interface NestedCreateTreatmentPlanRequest {
  treatment_content: string;
  memo?: string;
  is_insurance?: boolean;
  unit_price?: number;
  quantity: number;
  discount_rate?: number;
  discount_amount?: number;
  sort_order?: number;
}

export interface CreateHospitalizationRequest {
  pet_id: string;
  owner_id: string;
  cage_id?: string;
  hospitalization_type: string;
  start_date: string;
  end_date: string;
  next_visit?: string;
  owner_request?: string;
  staff_notes?: string;
  memo?: string;
  doctor_id?: string;
  is_insurance?: boolean;
  insurance_company_name?: string | null;
  insurance_number?: string | null;
  /** Optional nested plans; same TX as parent on BE (TASK-001). */
  treatment_plans?: NestedCreateTreatmentPlanRequest[];
}

export interface UpdateHospitalizationRequest {
  hospitalization_type?: string;
  cage_id?: string;
  /** true で cage_id を NULL（ケージ未割当）に戻す。cage_id との同時指定は BE が 400 で拒否する（EMR-316）。 */
  clear_cage_id?: boolean;
  expected_cage_id?: string;
  expected_cage_is_null?: boolean;
  start_date?: string;
  end_date?: string;
  next_visit?: string;
  /** true で next_visit を NULL に戻す。next_visit との同時指定は BE が 400 で拒否する（EMR-353）。 */
  clear_next_visit?: boolean;
  status?: string;
  owner_request?: string;
  staff_notes?: string;
  memo?: string;
  doctor_id?: string;
  is_insurance?: boolean;
  insurance_company_name?: string | null;
  insurance_number?: string | null;
}
