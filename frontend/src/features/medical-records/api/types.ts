// Backend Medical Record types (snake_case to match Go model JSON tags)
export interface BackendMedicalRecord {
  id: string;
  record_no: string;
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
  status: string;
  created_at: string;
  updated_at: string;
}

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
