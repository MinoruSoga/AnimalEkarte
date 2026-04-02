import type {
  Billing,
  BillingItem,
  BillingStatus,
  Payment,
  PaymentMethod,
  TaxType,
} from "@/types/generated/models";

// Backend 型エイリアス
export type BackendAccounting = Billing;
export type BackendPayment = Payment;

// BillingItem のレスポンス型（BE handler が計算して返す追加フィールドを含む）
export interface BackendAccountingItem extends BillingItem {
  tax_amount?: number;
  subtotal?: number;
}

// BillingItem の更新リクエスト
export interface UpdateBillingItemRequest {
  unit_price?: number;
  quantity?: number;
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
  completed_at?: string | null;
  memo?: string;
}
