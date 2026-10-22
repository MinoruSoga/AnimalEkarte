import type {
  AnesthesiaType,
  ExaminationType,
  TaxType,
  Vaccine,
  CheckupType,
} from "@/types/generated/models";
import type { ReorderRequest } from "@/types/form";

// ─────────────────────────────────────────────────
// 共通: サーバーが自動生成するフィールド
// ─────────────────────────────────────────────────

type ServerFields = "id" | "clinic_id" | "created_at" | "updated_at" | "version";

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
 * generated model（`Vaccine` / `CheckupType`）由来の要求型へ is_non_insurance の受け口を足す。
 * 診察・処置は下の明示 DTO / 要求型が直接持つ（EMR-527）。
 */
type ItemInsuranceWire = {
  is_non_insurance?: boolean;
};

// ─────────────────────────────────────────────────
// 診察・処置の移行 HTTP DTO（非永続・EMR-527）
// 保存 model（生成 Consultation / Procedure）から Omit / Pick せず、現行 handler の
// response struct と binding struct の列をそのまま写す。omitempty は省略可（?）、
// omitempty なしのポインタは null 可で表す。新しい契約へ切り替えるときに置き換える。
// ─────────────────────────────────────────────────

/**
 * Go の model.ExactNumeric（pgtype.Numeric.MarshalJSON）の JSON 表現。
 * 無効値（SQL NULL）は null、NaN は文字列 "NaN"、それ以外は桁を保った JSON number。
 */
export type LegacyExactNumericJson = number | "NaN" | null;

/** @see backend/internal/medicalrecord/consultation_response.go consultationResponse */
export interface LegacyConsultationResponse {
  version: number;
  id: number;
  clinic_id: number;
  name: string;
  price?: number;
  is_active: boolean;
  description: string;
  time_condition: string;
  duration?: number;
  parent_id?: number;
  tax_type: TaxType;
  /** model.ExactNumeric のまま直列化される（float64 ではない）。 */
  tax_rate: LegacyExactNumericJson;
  sort_order: number | null;
  created_at: string;
  updated_at: string;
  /** true=保険対象外 / false=保険対象 / null=未保存・旧欠落（EMR-454）。 */
  is_non_insurance: boolean | null;
}

/**
 * tax_type は Go では string(model.TaxType) で直列化される。値域どおり TaxType で受ける。
 * @see backend/internal/medicalrecord/procedure_response.go procedureResponse
 */
export interface LegacyProcedureResponse {
  version: number;
  id: number;
  clinic_id: number;
  name: string;
  price?: number;
  is_active: boolean;
  description: string;
  duration?: number;
  anesthesia: AnesthesiaType | null;
  parent_id?: number;
  tax_type: TaxType;
  /** model.ExactNumeric のまま直列化される（float64 ではない）。 */
  tax_rate: LegacyExactNumericJson;
  sort_order: number | null;
  created_at: string;
  updated_at: string;
  /** true=保険対象外 / false=保険対象 / null=未保存・旧欠落（EMR-454）。 */
  is_non_insurance: boolean | null;
}

// ─────────────────────────────────────────────────
// Consultation
// ─────────────────────────────────────────────────

/**
 * create は新規初期値（保険対象外=true）を必ず明示する。
 * 送信側（buildConsultationCreateRequest + applyItemInsuranceForCreate）が送るキーだけを列挙する。
 * @see backend/internal/medicalrecord/consultation_request.go createConsultationRequest
 */
export interface CreateConsultationRequest {
  name: string;
  price?: number;
  is_active?: boolean;
  description?: string;
  parent_id?: number;
  tax_type?: TaxType;
  tax_rate?: number;
  is_non_insurance: boolean;
}

/** create 要求のうち保険設定を除いた形（保存経路が保険設定を明示的に足す）。 */
export type CreateConsultationRequestBase = Omit<CreateConsultationRequest, "is_non_insurance">;

/**
 * update は未指定（key なし）/ false / true を区別する。null は旧欠落なので送らない。
 * @see backend/internal/medicalrecord/consultation_request.go updateConsultationRequest
 */
export interface UpdateConsultationRequest {
  version: number;
  name?: string;
  price?: number;
  is_active?: boolean;
  description?: string;
  parent_id?: number;
  clear_parent_id?: boolean;
  tax_type?: TaxType;
  tax_rate?: number;
  is_non_insurance?: boolean;
}

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

/**
 * anesthesia と tax_type は BE create binding:"required" と一致させる。
 * 送信側（buildProcedureCreateRequest + applyItemInsuranceForCreate）が送るキーだけを列挙する。
 * @see backend/internal/medicalrecord/procedure_request.go createProcedureRequest
 */
export interface CreateProcedureRequest {
  name: string;
  price?: number;
  is_active?: boolean;
  description?: string;
  anesthesia: AnesthesiaType;
  parent_id?: number;
  tax_type: TaxType;
  tax_rate?: number;
  is_non_insurance: boolean;
}

/** create 要求のうち保険設定を除いた形（保存経路が保険設定を明示的に足す）。 */
export type CreateProcedureRequestBase = Omit<CreateProcedureRequest, "is_non_insurance">;

/**
 * update は未指定（key なし）/ false / true を区別する。null は旧欠落なので送らない。
 * @see backend/internal/medicalrecord/procedure_request.go updateProcedureRequest
 */
export interface UpdateProcedureRequest {
  version: number;
  name?: string;
  price?: number;
  is_active?: boolean;
  description?: string;
  anesthesia?: AnesthesiaType;
  parent_id?: number;
  clear_parent_id?: boolean;
  tax_type?: TaxType;
  tax_rate?: number;
  is_non_insurance?: boolean;
}

// ─────────────────────────────────────────────────
// Vaccine
// ─────────────────────────────────────────────────

type VaccineWritable = Omit<Vaccine, ServerFields | "inventory_id">;

export type CreateVaccineRequest = Pick<VaccineWritable, "name"> &
  Partial<Omit<VaccineWritable, "name">> &
  ItemInsuranceWire;

export type UpdateVaccineRequest = Partial<VaccineWritable> &
  ItemInsuranceWire & {
    version?: number;
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
    version?: number;
    clear_parent_id?: boolean;
  };

// ─────────────────────────────────────────────────
// 共通 Reorder リクエスト
// ─────────────────────────────────────────────────

export type ReorderTreatmentRequest = ReorderRequest;
