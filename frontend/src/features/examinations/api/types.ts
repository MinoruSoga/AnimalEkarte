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

/**
 * 検査項目 1 行分の入力（PUT /examinations/:id/items のリクエスト要素）。
 * status / is_abnormal はサーバ側で ref_min/ref_max から導出するため送信しない。
 */
export interface UpsertExamItemRequest {
  exam_type_field_id?: number | null;
  name: string;
  inspection_value?: string;
  normal_value?: string;
  unit?: string;
  reference_value?: string;
  ref_min?: number | null;
  ref_max?: number | null;
  sort_order?: number;
}

/** PUT /api/v1/examinations/:id/items のリクエスト */
export interface ReplaceExamItemsRequest {
  items: UpsertExamItemRequest[];
}

/** GET / PUT /api/v1/examinations/:id/items のレスポンスエンベロープ */
export interface ExamItemsResponse {
  items: BackendExamResult[];
}
