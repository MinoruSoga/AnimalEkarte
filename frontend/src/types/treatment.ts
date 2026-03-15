import type {
  Consultation,
  ExaminationType,
  Procedure,
  Vaccine,
  CheckupType,
} from "@/types/generated/models";

// ─────────────────────────────────────────────────
// 共通: サーバーが自動生成するフィールド
// ─────────────────────────────────────────────────

type ServerFields = "id" | "clinic_id" | "created_at" | "updated_at";

// ─────────────────────────────────────────────────
// Consultation
// ─────────────────────────────────────────────────

type ConsultationWritable = Omit<Consultation, ServerFields>;

export type CreateConsultationRequest = Pick<ConsultationWritable, "name"> &
  Partial<Omit<ConsultationWritable, "name">>;

export type UpdateConsultationRequest = Partial<ConsultationWritable> & {
  clear_parent_id?: boolean;
};

// ─────────────────────────────────────────────────
// ExaminationType
// ─────────────────────────────────────────────────

// ExaminationType は items リレーションを持つため除外
type ExaminationTypeWritable = Omit<ExaminationType, ServerFields | "items">;

export type CreateExaminationTypeRequest = Pick<ExaminationTypeWritable, "name"> &
  Partial<Omit<ExaminationTypeWritable, "name">>;

export type UpdateExaminationTypeRequest = Partial<ExaminationTypeWritable> & {
  clear_parent_id?: boolean;
};

// ─────────────────────────────────────────────────
// Procedure
// ─────────────────────────────────────────────────

type ProcedureWritable = Omit<Procedure, ServerFields>;

export type CreateProcedureRequest = Pick<ProcedureWritable, "name"> &
  Partial<Omit<ProcedureWritable, "name">>;

export type UpdateProcedureRequest = Partial<ProcedureWritable> & {
  clear_parent_id?: boolean;
};

// ─────────────────────────────────────────────────
// Vaccine
// ─────────────────────────────────────────────────

type VaccineWritable = Omit<Vaccine, ServerFields | "inventory_id">;

export type CreateVaccineRequest = Pick<VaccineWritable, "name"> &
  Partial<Omit<VaccineWritable, "name">>;

export type UpdateVaccineRequest = Partial<VaccineWritable> & {
  clear_parent_id?: boolean;
};

// ─────────────────────────────────────────────────
// CheckupType
// ─────────────────────────────────────────────────

type CheckupTypeWritable = Omit<CheckupType, ServerFields>;

export type CreateCheckupTypeRequest = Pick<CheckupTypeWritable, "name"> &
  Partial<Omit<CheckupTypeWritable, "name">>;

export type UpdateCheckupTypeRequest = Partial<CheckupTypeWritable> & {
  clear_parent_id?: boolean;
};

// ─────────────────────────────────────────────────
// 共通 Reorder リクエスト
// ─────────────────────────────────────────────────

export interface ReorderTreatmentRequest {
  ids: number[];
}
