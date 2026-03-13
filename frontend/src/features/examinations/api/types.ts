/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { Exam, ExamItem } from "@/types/generated/models";

export type BackendExaminationItem = ExamItem;
export type BackendExamination = Exam;

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
