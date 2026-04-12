/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { Examination, ExamResult } from "@/types/generated/models";

export type BackendExamResult = ExamResult;
export type BackendExamination = Examination;

export interface CreateExaminationRequest {
  medical_record_id?: number | null;
  pet_id?: number | null;
  exam_type_id: number;
  doctor_id?: number | null;
  date: string;
  machine?: string;
  result_summary?: string;
}

export interface UpdateExaminationRequest {
  medical_record_id?: number | null;
  status?: "pending" | "in_progress" | "result_entered" | "completed" | "confirmed";
  result_summary?: string;
  machine?: string;
  date?: string;
}
