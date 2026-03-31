import type { MedicalRecord as ApiMedicalRecord } from "@/types/generated/models";

// APIレスポンス (medicalRecordResponse) は Go モデルと異なり accounting_id を直接返す
export type BackendMedicalRecord = ApiMedicalRecord & {
  accounting_id?: number;
};

// ── CreateMedicalRecordRequest ──
// Backend handler (medical_record_request.go) createMedicalRecordRequest に準拠。
// Go モデル (MedicalRecord) にはない visit_date, ClinicalPlan 関連フィールドを
// フラットに受け付ける。tygo 生成の ApiMedicalRecord からの Omit 導出では
// これらのフィールドが含まれないため、手動定義が必要。
export interface CreateMedicalRecordRequest {
  // 基本フィールド
  record_no?: string;       // optional: BE で自動生成
  date?: string;            // optional: ISO 8601
  visit_date?: string;      // FE 送信フィールド（"YYYY-MM-DD"形式）
  visit_type?: string;      // FE 送信フィールド
  owner_id: string;         // FE は string → BE で uint64 変換
  pet_id: string;           // FE は string → BE で uint64 変換
  doctor_id?: string;       // FE は string → BE で uint64 変換
  reservation_appointment_id?: string;
  status?: string;

  // ClinicalPlan 関連（原子的作成用）
  chief_complaint?: string;
  chief_complaint_category_id?: number;
  plan?: string;
  assessment?: string;
  notes?: string;
  diagnosis_1_category_id?: number;
  diagnosis_1_name_id?: number;
  diagnosis_2_category_id?: number;
  diagnosis_2_name_id?: number;
}

// ★ Partial で全フィールド optional に（編集時は部分更新）
export type UpdateMedicalRecordRequest = Partial<
  Omit<ApiMedicalRecord, "id" | "clinic_id" | "created_at" | "updated_at" | "owner" | "pet" | "doctor" | "clinical_plan" | "inquiry" | "treatments" | "vitals" | "exams" | "vaccinations" | "checkups" | "estimates" | "billing_review" | "billing">
>;
