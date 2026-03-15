import type { MedicalRecord as ApiMedicalRecord } from "@/types/generated/models";

// APIレスポンス (medicalRecordResponse) は Go モデルと異なり accounting_id を直接返す
export type BackendMedicalRecord = ApiMedicalRecord & {
  accounting_id?: number;
};

export interface CreateMedicalRecordRequest {
  pet_id: string;
  owner_id: string;
  doctor_id?: string;
  visit_date: string;
  visit_type: string;
  chief_complaint?: string;
  subjective?: string;
  objective?: string;
  assessment?: string;
  plan?: string;
  surgery_notes?: string;
  diagnosis?: string;
  treatment?: string;
  prescription?: string;
  notes?: string;
  status?: string;
}

export interface UpdateMedicalRecordRequest {
  pet_id?: string;
  owner_id?: string;
  doctor_id?: string;
  visit_date?: string;
  visit_type?: string;
  chief_complaint?: string;
  subjective?: string;
  objective?: string;
  assessment?: string;
  plan?: string;
  surgery_notes?: string;
  diagnosis?: string;
  treatment?: string;
  prescription?: string;
  notes?: string;
  status?: string;
}
