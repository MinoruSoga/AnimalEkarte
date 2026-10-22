/**
 * 会計 API の移行 HTTP DTO（非永続）。
 *
 * 保存 model（生成 models.ts）ではなく、現行 handler が返す JSON の列をそのまま写す。
 * Go の json tag に合わせ、omitempty は省略可（?）、omitempty なしのポインタは null 可で表す。
 * 新しい資金・返金 event 契約の consumer を実装するときに、この adapter ごと置き換える（EMR-527）。
 */
import type {
  AccountingStatus,
  ItemCategory,
  ItemSource,
  PaymentMethod,
  TaxType,
} from "../types";

/** @see backend/internal/billing/accounting_response.go paymentResponse */
export interface LegacyPaymentResponse {
  id: number;
  billing_id: number;
  subtotal: number;
  tax_total: number;
  total_amount: number;
  insurance_name: string;
  insurance_ratio: number;
  insurance_amount: number;
  discount_amount: number;
  billing_amount: number;
  received_amount: number;
  change_amount: number;
  method: PaymentMethod;
  paid_by: number | null;
  paid_by_name: string;
  created_at: string;
  updated_at: string;
}

/** @see backend/internal/billing/accounting_response.go paymentSplitResponse */
export interface LegacyPaymentSplitResponse {
  id: number;
  clinic_id: number;
  billing_id: number;
  method: PaymentMethod;
  payment_method_id?: number;
  amount: number;
  received_amount: number;
  change_amount: number;
  paid_by?: number;
  paid_by_name: string;
  created_at: string;
}

/** @see backend/internal/billing/refund_request.go RefundResponse */
export interface LegacyRefundResponse {
  id: number;
  clinic_id: number;
  billing_id: number;
  amount: number;
  reason: string;
  refunded_by: number | null;
  refunded_by_name: string;
  payment_method?: PaymentMethod;
  refunded_at: string;
  created_at: string;
}

/** @see backend/internal/billing/billing_item_request.go BillingItemResponse */
export interface LegacyAccountingItemResponse {
  id: number;
  billing_id: number;
  category: ItemCategory;
  name: string;
  unit_price: number;
  quantity: number;
  discount_rate: number;
  discount_amount: number;
  subtotal: number;
  tax_type: TaxType;
  tax_rate: number;
  tax_amount: number;
  /** null = legacy unset。明示的な false とは区別する。 */
  is_insurance_applicable: boolean | null;
  source: ItemSource;
  other_reason?: string;
  treatment_id?: number;
  /** 未請求候補など treatment 由来の親カルテ（仮想。DB 列ではない） */
  medical_record_id?: number;
  /** EMR-246: 未請求候補の発生日（仮想。DB 列ではない）。当日/過去区分表示用 */
  service_date?: string;
  vaccination_id?: number;
  exam_id?: number;
  appointment_id?: number;
  trimming_course_id?: number;
  trimming_option_id?: number;
  merchandise_item_id?: number;
  sort_order: number;
  created_at: string;
}

/** @see backend/internal/billing/accounting_response.go accountingOwnerSummary */
export interface LegacyAccountingOwnerSummary {
  id: number;
  name: string;
}

/**
 * 動物種は含まない。会計の petSpecies は HTTP に無いため補完しない。
 * @see backend/internal/billing/accounting_response.go accountingPetSummary
 */
export interface LegacyAccountingPetSummary {
  id: number;
  name: string;
}

/** @see backend/internal/billing/accounting_response.go accountingResponse */
export interface LegacyAccountingResponse {
  id: number;
  clinic_id: number;
  medical_record_id?: number;
  hospitalization_id?: number;
  owner_id?: number;
  pet_id?: number;
  owner?: LegacyAccountingOwnerSummary;
  pet?: LegacyAccountingPetSummary;
  subtotal: number;
  tax_total: number;
  total_amount: number;
  total_refunded_amount: number;
  /** BUG-007: 未収残高（waiting 全額 or クレジット訂正差額） */
  outstanding_amount: number;
  has_insurance: boolean;
  status: AccountingStatus;
  scheduled_date: string;
  completed_at?: string;
  memo: string;
  items?: LegacyAccountingItemResponse[];
  payments?: LegacyPaymentResponse[];
  payment_splits?: LegacyPaymentSplitResponse[];
  refunds?: LegacyRefundResponse[];
  created_at: string;
  updated_at: string;
}
