export interface BackendExamination {
  id: string;
  pet_id: string;
  owner_id: string;
  doctor_id?: string | null;
  medical_record_id?: string | null;
  examination_date: string;
  test_type: string;
  machine?: string;
  status: "依頼中" | "検査中" | "完了";
  result_summary?: string;
  items?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
  // Preloaded relations
  pet?: {
    id: string;
    name: string;
    species?: string;
  } | null;
  owner?: {
    id: string;
    name: string;
  } | null;
}

export interface CreateExaminationRequest {
  pet_id: string;
  owner_id: string;
  doctor_id?: string | null;
  medical_record_id?: string | null;
  examination_date: string;
  test_type: string;
  machine?: string;
  result_summary?: string;
  items?: string;
  notes?: string;
}

export interface UpdateExaminationRequest {
  status?: "依頼中" | "検査中" | "完了";
  result_summary?: string;
  items?: string;
  notes?: string;
  examination_date?: string;
  test_type?: string;
  machine?: string;
}
