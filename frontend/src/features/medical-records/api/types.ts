import type { MedicalRecord as ApiMedicalRecord } from "@/types/generated/models";

// APIレスポンス (medicalRecordResponse) は Go モデルと異なり accounting_id を直接返す
export type BackendMedicalRecord = ApiMedicalRecord & {
  accounting_id?: number;
};

// ★ Omit で自動生成モデルから導出（CLAUDE.md準拠）
// Backend API (backend/internal/handler/medical_record_request.go) は
// basic medical record フィールドのみを受け付ける
//
// NOTE: 詳細情報（chief_complaint, plan, assessment, notes, diagnosis_*）は
// 関連エンティティに属する：
//   - Inquiry: chief_complaint, chief_complaint_category_id, notes
//   - ClinicalPlan: diagnosis_category_id, diagnosis_name_id, assessment
// これらは別のエンドポイント (inquiries, treatment-plans) で更新する
//
// ⚠️  Backend Issue BE-015: createMedicalRecordRequest と frontend の
// 期待値に乖離あり。backend 修正待ち（record_no, visit_date等の名前、型）
// https://github.com/animal-ekarte/AnimalEkarte/blob/main/backend/issues/open/BE-015-medical-record-create-api-contract.md
export type CreateMedicalRecordRequest = Omit<
  ApiMedicalRecord,
  | "id"
  | "clinic_id"
  | "created_at"
  | "updated_at"
  | "owner"
  | "pet"
  | "doctor"
  | "clinical_plan"
  | "inquiry"
  | "treatments"
  | "vitals"
  | "exams"
  | "vaccinations"
  | "checkups"
  | "estimates"
  | "billing_review"
  | "billing"
>;

// ★ Partial で全フィールド optional に（編集時は部分更新）
export type UpdateMedicalRecordRequest = Partial<CreateMedicalRecordRequest>;
