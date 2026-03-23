/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { Examination, ExaminationItem } from "@/types/generated/models";

export type BackendExaminationItem = ExaminationItem;
export type BackendExamination = Examination;

export interface CreateExaminationRequest {
  medical_record_id: number;
  pet_id?: number | null;
  exam_type_id: number;
  doctor_id?: number | null;
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
