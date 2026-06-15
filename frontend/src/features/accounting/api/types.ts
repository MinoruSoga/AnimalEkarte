import type {
  Billing,
  BillingItem,
  BillingStatus,
  PaymentMethod,
  TaxType,
} from "@/types/generated/models";

// Backend 型エイリアス
export type BackendAccounting = Billing;

// BillingItem のレスポンス型（BE handler が計算して返す追加フィールドを含む）
export interface BackendAccountingItem extends BillingItem {
  tax_amount?: number;
  subtotal?: number;
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
}

// API リクエスト型（models.ts から導出）
export interface CreateAccountingRequest {
  pet_id: number;
  owner_id: number;
  medical_record_id?: number | null;
  scheduled_date: string;
  subtotal?: number | null;
  tax_total?: number | null;
  total_amount?: number | null;
  insurance_name?: string;
  insurance_ratio?: number | null;
  insurance_amount?: number | null;
  discount_amount?: number | null;
  billing_amount?: number | null;
  payment_method?: PaymentMethod;
  memo?: string;
}

export interface PaymentSplitRequest {
  method: PaymentMethod;
  payment_method_id?: number;
  amount: number;
  received_amount?: number;
  change_amount: number; // #119: required — cash: received-amount, non-cash: 0
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
  completed_at?: string | null;
  memo?: string;
  post_close_reason?: string; // #115: レジ締め済み期間の遡り編集理由
}
