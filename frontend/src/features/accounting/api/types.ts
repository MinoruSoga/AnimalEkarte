import type { Billing, BillingItem } from "@/types/generated/models";

export type BackendAccounting = Billing;
export type BackendAccountingItem = BillingItem;

export interface CreateAccountingRequest {
  pet_id: string;
  owner_id: string;
  medical_record_id?: string | null;
  scheduled_date: string;
  subtotal?: number | null;
  tax_total?: number | null;
  total_amount?: number | null;
  insurance_name?: string;
  insurance_ratio?: number | null;
  insurance_amount?: number | null;
  discount_amount?: number | null;
  billing_amount?: number | null;
  payment_method?: string;
  memo?: string;
}

export interface UpdateAccountingRequest {
  status?: string;
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
  payment_method?: string;
  completed_at?: string | null;
  memo?: string;
}
