import type {
  Billing,
  BillingItem,
  BillingStatus,
  PaymentMethod,
  TaxType,
} from "@/types/generated/models";

// Backend 型エイリアス
export type BackendAccounting = Billing & {
  /** BUG-007: 未収残高（waiting 全額 or クレジット訂正差額） */
  outstanding_amount?: number | null;
};

// BillingItem のレスポンス型（BE handler が計算して返す追加フィールドを含む）
export interface BackendAccountingItem extends BillingItem {
  vaccination_id?: number;
  exam_id?: number;
  other_reason?: string;
  tax_amount?: number;
  subtotal?: number;
  /** 未請求候補など treatment 由来の親カルテ（仮想。DB 列ではない） */
  medical_record_id?: number;
}

// BillingItem の更新リクエスト
export interface UpdateBillingItemRequest {
  unit_price?: number;
  quantity?: number;
  discount_rate?: number;
  discount_amount?: number;
  tax_type?: TaxType;
  tax_rate?: number;
  is_insurance_applicable?: boolean;
  /** #115 / BUG-021 / BUG-009: レジ締め済み・確定済み明細更新理由（BE が必須検証） */
  post_close_reason?: string;
}

/** 明細削除 body（締め後のみ post_close_reason を送る） */
export interface DeleteBillingItemRequest {
  post_close_reason?: string;
}

/** BUG-018: POST /v1/accountings/complete の明細1行 */
export interface CompleteAccountingItemRequest {
  category: string;
  name: string;
  unit_price: number;
  quantity: number;
  discount_rate?: number;
  discount_amount?: number;
  tax_type: string;
  tax_rate: number;
  is_insurance_applicable: boolean;
  source: string;
  other_reason?: string;
  merchandise_item_id?: number;
  vaccination_id?: number;
  exam_id?: number;
  treatment_id?: number;
  appointment_id?: number;
  trimming_course_id?: number;
  trimming_option_id?: number;
}

/** BUG-018: 原子的会計確定 command body（client total は送らない） */
export interface CompleteAccountingRequest {
  pet_id: number;
  owner_id: number;
  medical_record_id?: number | null;
  hospitalization_id?: number | null;
  scheduled_date: string;
  memo?: string;
  has_insurance?: boolean;
  insurance_ratio?: number | null;
  insurance_name?: string;
  insurance_amount?: number | null;
  discount_amount?: number | null;
  items: CompleteAccountingItemRequest[];
  payment_splits?: PaymentSplitRequest[];
  post_close_reason?: string;
}

export interface PaymentSplitRequest {
  method: PaymentMethod;
  payment_method_id?: number;
  amount: number;
  received_amount?: number;
  change_amount: number; // #119: required — cash: received-amount, non-cash: 0
  // #188: お釣り直接上書きモード。true の場合のみ BE が change == received-amount 整合検証を緩和する。
  change_override?: boolean;
}

export interface UpdateAccountingRequest {
  status?: BillingStatus;
  subtotal?: number | null;
  tax_total?: number | null;
  total_amount?: number | null;
  insurance_name?: string;
  insurance_ratio?: number | null;
  insurance_amount?: number | null;
  discount_amount?: number | null;
  billing_amount?: number | null;
  received_amount?: number | null;
  change_amount?: number | null;
  payment_method?: PaymentMethod;
  payment_splits?: PaymentSplitRequest[];
  memo?: string;
  post_close_reason?: string; // #115: レジ締め済み期間の遡り編集理由
}

// #189: 確定済み会計のクレジット（カード）金額の確定後訂正リクエスト。
// 専用エンドポイント POST /v1/accountings/:id/credit-correction 用。
export interface CorrectCreditPaymentRequest {
  // 生成型 PaymentMethod は string 別名のため Extract が never に潰れる。リテラルユニオンで直接定義する。
  method: "credit_card" | "electronic_money";
  amount: number;
  reason: string;
  memo?: string;
}
