import { SYNTHETIC_CREATED_AT as CREATED_AT, SYNTHETIC_IDS } from "./ui-design-clinical-constants";

// Detail-page read-only response stubs for the medicalRecordCreate synthetic
// scenario. Split from ui-design-clinical.ts to keep each fixture module under
// the 800-line source limit.

export const SYNTHETIC_CLINICAL_PLAN = {
  id: "syn-plan-990013",
  medical_record_id: String(SYNTHETIC_IDS.medicalRecord),
  physical_exam: "",
  diagnoses: [],
  diagnosis_details: "",
  treatment_policy: "",
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
  version: 1,
} satisfies {
  id: string;
  medical_record_id: string;
  physical_exam: string;
  diagnoses: Array<{
    id: string;
    version: number;
    diagnosis_slot: 1 | 2;
    diagnosis_type_id: string | null;
    diagnosis_name_id: string | null;
    diagnosis_name_snapshot: string | null;
    category_name_snapshot: string | null;
    sort_order: number;
  }>;
  diagnosis_details: string;
  treatment_policy: string;
  created_at: string;
  updated_at: string;
  version: number;
};

// Backend BillingConfirmation marshals uint64 IDs as JSON numbers
// (backend/internal/model/billing_confirmation.go) — keep the wire shape numeric.
export const SYNTHETIC_BILLING_CONFIRMATION = {
  id: SYNTHETIC_IDS.billingConfirmation,
  medical_record_id: SYNTHETIC_IDS.medicalRecord,
  status: "pending",
  return_reason: "",
  memo: "",
  created_at: CREATED_AT,
  updated_at: CREATED_AT,
} satisfies {
  id: number;
  medical_record_id: number;
  status: "pending" | "confirmed" | "returned";
  confirmed_by?: number;
  confirmed_at?: string;
  returned_by?: number;
  returned_at?: string;
  return_reason: string;
  memo: string;
  created_at: string;
  updated_at: string;
};
