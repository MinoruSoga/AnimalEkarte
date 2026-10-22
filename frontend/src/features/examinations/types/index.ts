/**
 * Backend API response types
 * Source: @/types/examination-legacy-wire（現行 handler の JSON を写す移行 HTTP DTO。EMR-527）
 */
import type {
  LegacyExaminationResponse,
  LegacyExamItemsResponse,
  LegacyExamResultResponse,
} from "@/types/examination-legacy-wire";
import type { UpdateExaminationRequest as SharedUpdateExaminationRequest } from "@/hooks/use-update-examination";

export type BackendExamResult = LegacyExamResultResponse;
export type BackendExamination = LegacyExaminationResponse;

export interface CreateExaminationRequest {
  medical_record_id?: number | null;
  pet_id?: number | null;
  exam_type_id: number;
  doctor_id?: number | null;
  date: string;
  machine?: string;
  result_summary?: string;
  items?: UpsertExamItemRequest[];
}

/** Feature-local PATCH contract extends the shared parent-only mutation with atomic items. */
export type UpdateExaminationRequest = SharedUpdateExaminationRequest & {
  items?: UpsertExamItemRequest[];
};

export interface UnconfirmExaminationRequest {
  reason: string;
}

/**
 * 検査項目 1 行分の入力（POST/PATCH の nested items と PUT items で共通）。
 * status / is_abnormal と判定基準値はサーバ側でマスタから導出するため送信しない。
 */
export interface UpsertExamItemRequest {
  exam_type_field_id?: number | null;
  name: string;
  inspection_value?: string;
  normal_value?: string;
  unit?: string;
  reference_value?: string;
  sort_order?: number;
}

/** PUT /api/v1/examinations/:id/items のリクエスト */
export interface ReplaceExamItemsRequest {
  items: UpsertExamItemRequest[];
}

/** GET / PUT /api/v1/examinations/:id/items のレスポンスエンベロープ */
export type ExamItemsResponse = LegacyExamItemsResponse;
