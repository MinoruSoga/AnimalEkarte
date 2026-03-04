export interface BackendPetSummary {
  id: string;
  name: string;
  species?: string;
}

export interface BackendOwnerSummary {
  id: string;
  name: string;
}

export interface BackendAccountingItem {
  id: string;
  accounting_id: string;
  master_id?: string | null;
  code?: string;
  category: string;
  name: string;
  unit_price?: number | null;
  quantity: number;
  tax_rate?: number | null;
  is_insurance_applicable: boolean;
  source: string;
  created_at: string;
}

export interface BackendAccounting {
  id: string;
  medical_record_id?: string | null;
  pet_id: string;
  owner_id: string;
  scheduled_date: string;
  completed_at?: string | null;
  status: "未収" | "保留" | "回収済" | "キャンセル";
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
  memo?: string;
  created_at: string;
  updated_at: string;
  // Preloaded relations
  pet?: BackendPetSummary | null;
  owner?: BackendOwnerSummary | null;
  accounting_items?: BackendAccountingItem[];
}

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
  status?: "未収" | "保留" | "回収済" | "キャンセル";
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
