/**
 * 検査 API の移行 HTTP DTO（非永続）。
 *
 * 保存 model（生成 models.ts）ではなく、現行 handler が返す JSON の列をそのまま写す。
 * Go の json tag に合わせ、omitempty は省略可（?）で表す。
 * pet / doctor の入れ子は同じ Go struct の tygo 生成型（medicalrecord-responses）を使う。
 * examinations・medical-records・owner-report と shared hooks/lib が同じ wire を読むため
 * 共有層に置く（feature 間 import は禁止）。検査の wire 契約を生成型へ移すときに、
 * この adapter ごと置き換える（EMR-527）。
 */
import type {
  PetSummaryResponse,
  StaffSummaryResponse,
} from "@/types/generated/medicalrecord-responses";
// enum union は allowlist 済みの ./index から type-only で受け取る。
import type { ExaminationResultStatus, ExaminationStatus } from "./index";

/** @see backend/internal/medicalrecord/examination_response.go examTypeSummaryResponse */
interface LegacyExamTypeSummary {
  id: number;
  name: string;
  price?: number;
}

/** @see backend/internal/medicalrecord/examination_response.go examResultResponse */
export interface LegacyExamResultResponse {
  id: number;
  exam_id: number;
  exam_type_field_id?: number;
  name: string;
  inspection_value: string;
  normal_value: string;
  result: string;
  unit: string;
  reference_value: string;
  ref_min?: number;
  ref_max?: number;
  qualitative_min?: string;
  qualitative_max?: string;
  is_assessed: boolean;
  is_abnormal: boolean;
  status: ExaminationResultStatus;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

/** @see backend/internal/medicalrecord/examination_response.go examinationResponse */
export interface LegacyExaminationResponse {
  id: number;
  clinic_id: number;
  medical_record_id?: number;
  pet_id?: number;
  exam_type_id: number;
  doctor_id?: number;
  date: string;
  result_summary: string;
  machine: string;
  status: ExaminationStatus;
  current_revision_version?: number;
  created_at: string;
  updated_at: string;
  /** Preload 時のみ埋まる。 */
  pet?: PetSummaryResponse;
  doctor?: StaffSummaryResponse;
  exam_type?: LegacyExamTypeSummary;
  /** include_items=true の一覧でのみ出力される。 */
  items?: LegacyExamResultResponse[];
}

/** @see backend/internal/medicalrecord/examination_response.go examItemsResponse */
export interface LegacyExamItemsResponse {
  items: LegacyExamResultResponse[];
}
