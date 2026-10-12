import type {
  MedicalRecordResponse,
  MedicalRecordAddendumResponse,
} from "@/types/generated/medicalrecord-responses";
import type { RecommendationReason } from "../constants/recommendation-reason";

/**
 * Backend medical-record list/detail/create/update wire.
 * Source of truth: MedicalRecordResponse (domain-owned DTO / tygo), not models.MedicalRecord.
 */
export type BackendMedicalRecord = MedicalRecordResponse;

// ── MedicalRecordAddendum + author_name (EMR-326) ──
// generated の MedicalRecordAddendumResponse.author_name は `?: string`。
// wire 上は omitempty により「キー省略 or 文字列」のみが来るが、非生成経路
// （旧キャッシュ・他クライアント応答）で null が来ても UI フォールバックが潰れないよう
// `| null` に広げて受ける。Omit 必須: 通常の交差型だと string | null が string に
// 潰れて null を表せない。
// author_name: 未解決（削除・異動・他医院）時はキー省略または null →
// UI は author_user_id へフォールバック表示する。
export type MedicalRecordAddendumWithAuthorName = Omit<
  MedicalRecordAddendumResponse,
  "author_name"
> & {
  author_name?: string | null;
};

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
  visit_type?: string;
}
