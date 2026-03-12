export interface BackendExaminationItem {
  id: string;
  exam_id: string;
  name: string;
  inspection_value: string;
  normal_value: string;
  result: string;
  unit: string;
  ref: string;
  status: "normal" | "high" | "low";
  sort_order: number;
  created_at: string;
}

export interface BackendExamination {
  id: string;
  medical_record_id: string;
  pet_id?: string | null;
  exam_type_id: string;
  doctor_id?: string | null;
  date: string;
  result_summary: string;
  machine: string;
  status: "依頼中" | "検査中" | "完了";
  created_at: string;
  updated_at: string;
  // Relations
  exam_type?: {
    id: string;
    name: string;
    code?: string;
  } | null;
  pet?: {
    id: string;
    name: string;
    species?: string;
    owner_id?: string;
  } | null;
  doctor?: {
    id: string;
    name: string;
  } | null;
  items?: BackendExaminationItem[];
}

export interface CreateExaminationRequest {
  medical_record_id: string;
  pet_id?: string | null;
  exam_type_id: string;
  doctor_id?: string | null;
  date: string;
  machine?: string;
  result_summary?: string;
}

export interface UpdateExaminationRequest {
  status?: "依頼中" | "検査中" | "完了";
  result_summary?: string;
  machine?: string;
  date?: string;
}
