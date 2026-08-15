import type { MedicalRecordResponse } from "@/types/generated/medicalrecord-responses";
import type { RecommendationReason } from "../constants/recommendation-reason";

/**
 * Backend medical-record list/detail/create/update wire.
 * Source of truth: MedicalRecordResponse (domain-owned DTO / tygo), not models.MedicalRecord.
 */
export type BackendMedicalRecord = MedicalRecordResponse;

// ── CreateMedicalRecordRequest ──
// Backend handler (medical_record_request.go) createMedicalRecordRequest に準拠。
// Go モデル (MedicalRecord) にはない visit_date, ClinicalPlan 関連フィールドを
// フラットに受け付ける。手動定義が必要。
export interface CreateMedicalRecordRequest {
  // 基本フィールド
  record_no?: string; // optional: BE で自動生成
  date?: string; // optional: ISO 8601
  visit_date?: string; // FE 送信フィールド（"YYYY-MM-DD"形式）
  visit_type?: string; // FE 送信フィールド
  owner_id: string; // FE は string → BE で uint64 変換
  pet_id: string; // FE は string → BE で uint64 変換
  doctor_id?: string; // FE は string → BE で uint64 変換
  appointment_id?: string;
  status?: string;
  recommendation_reason?: RecommendationReason | ""; // optional; "" は NULL 扱い

  // ClinicalPlan 関連（原子的作成用）
  chief_complaint?: string;
  chief_complaint_type_id?: number;
  plan?: string;
  assessment?: string;
  notes?: string;
  diagnosis_1_type_id?: number;
  diagnosis_1_name_id?: number;
  diagnosis_2_type_id?: number;
  diagnosis_2_name_id?: number;
}

// ── UpdateMedicalRecordRequest ──
// Backend updateMedicalRecordRequest に準拠（medical_record_request.go）。
// models.MedicalRecord からの Omit 導出は wire と乖離するため使わない。
export interface UpdateMedicalRecordRequest {
  date?: string;
  owner_id?: number | string;
  pet_id?: number | string;
  doctor_id?: number | string;
  appointment_id?: number | string;
  status?: string;
  /** 楽観的ロック用 CAS トークン。GET 応答の version をそのまま返す。 */
  version?: number;
  next_visit_recommended_date?: string | null;
  visit_type?: string;
}
