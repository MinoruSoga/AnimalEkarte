import type {
  Consultation,
  ExaminationType,
  Procedure,
  Vaccine,
  CheckupType,
} from "@/types/generated/models";
import type { ReorderRequest } from "@/types/form";

// ─────────────────────────────────────────────────
// 共通: サーバーが自動生成するフィールド
// ─────────────────────────────────────────────────

type ServerFields = "id" | "clinic_id" | "created_at" | "updated_at";

// ─────────────────────────────────────────────────
// 品目別保険設定（EMR-454 / wire: is_non_insurance）
// ─────────────────────────────────────────────────

/**
 * true = 保険対象外 / false = 保険対象 / null = 旧欠落（未保存）。
 *
 * 欠落と明示値を分けるため、null を false へ補完しない
 * （AC-accounting-item-insurance-legacy-missing）。
 * カテゴリ・他 master・既存 boolean の偶然の初期値からも補完しない。
 */
export type ItemInsuranceSetting = boolean | null;

/**
 * generated model（`Consultation` / `Procedure`）は is_non_insurance 未反映のため、
 * 診察・処置だけ応答/要求の受け口をローカルに拡張する。
 */
type ItemInsuranceWire = {
  is_non_insurance?: boolean;
};

// ─────────────────────────────────────────────────
// Consultation
// ─────────────────────────────────────────────────

type ConsultationWritable = Omit<Consultation, ServerFields>;

/** create は新規初期値（保険対象外=true）を必ず明示する。 */
export type CreateConsultationRequest = Pick<ConsultationWritable, "name"> &
  Partial<Omit<ConsultationWritable, "name">> & {
    is_non_insurance: boolean;
  };

/** create 要求のうち保険設定を除いた形（保存経路が保険設定を明示的に足す）。 */
export type CreateConsultationRequestBase = Omit<CreateConsultationRequest, "is_non_insurance">;

/** update は未指定（key なし）/ false / true を区別する。null は旧欠落なので送らない。 */
export type UpdateConsultationRequest = Partial<ConsultationWritable> &
  ItemInsuranceWire & {
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

// anesthesia は BE create binding:"required" と一致させる（BUG-028）
export type CreateProcedureRequest = Pick<ProcedureWritable, "name" | "anesthesia"> &
  Partial<Omit<ProcedureWritable, "name" | "anesthesia">> & {
    is_non_insurance: boolean;
  };

/** create 要求のうち保険設定を除いた形（保存経路が保険設定を明示的に足す）。 */
export type CreateProcedureRequestBase = Omit<CreateProcedureRequest, "is_non_insurance">;

/** update は未指定（key なし）/ false / true を区別する。null は旧欠落なので送らない。 */
export type UpdateProcedureRequest = Partial<ProcedureWritable> &
  ItemInsuranceWire & {
    clear_parent_id?: boolean;
  };

// ─────────────────────────────────────────────────
// Vaccine
// ─────────────────────────────────────────────────

type VaccineWritable = Omit<Vaccine, ServerFields | "inventory_id">;

export type CreateVaccineRequest = Pick<VaccineWritable, "name"> &
  Partial<Omit<VaccineWritable, "name">> &
  ItemInsuranceWire;

export type UpdateVaccineRequest = Partial<VaccineWritable> &
  ItemInsuranceWire & {
    clear_parent_id?: boolean;
  };

// ─────────────────────────────────────────────────
// CheckupType
// ─────────────────────────────────────────────────

type CheckupTypeWritable = Omit<CheckupType, ServerFields>;

export type CreateCheckupTypeRequest = Pick<CheckupTypeWritable, "name"> &
  Partial<Omit<CheckupTypeWritable, "name">> &
  ItemInsuranceWire;

export type UpdateCheckupTypeRequest = Partial<CheckupTypeWritable> &
  ItemInsuranceWire & {
    version: number;
    clear_parent_id?: boolean;
  };

// ─────────────────────────────────────────────────
// 共通 Reorder リクエスト
// ─────────────────────────────────────────────────

export type ReorderTreatmentRequest = ReorderRequest;
