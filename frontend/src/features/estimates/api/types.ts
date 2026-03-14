// バックエンドレスポンス型（snake_case）

export interface BackendEstimateItem {
  id: number;
  estimate_id: number;
  name: string;
  category: string;
  unit_price: number;
  quantity: number;
  tax_rate: number;
  discount_rate: number;
  discount_amount: number;
  is_insurance_applicable: boolean;
  sort_order: number;
  created_at: string;
}

export interface BackendOwnerSummary {
  id: number;
  owner_name: string;
  phone: string;
}

export interface BackendEstimate {
  id: number;
  clinic_id: number;
  estimate_no: string;
  medical_record_id?: number | null;
  title: string;
  owner_id?: number | null;
  owner?: BackendOwnerSummary | null;
  status: string;
  subtotal: number;
  tax_total: number;
  total_amount: number;
  insurance_amount: number;
  discount_amount: number;
  valid_until?: string | null;
  comment: string;
  notes: string;
  created_by?: number | null;
  items?: BackendEstimateItem[];
  created_at: string;
  updated_at: string;
}

export interface EstimateListResponse {
  data: BackendEstimate[];
  total: number;
  page: number;
  limit: number;
}

export interface CreateEstimateRequest {
  title: string;
  owner_id?: number | null;
  medical_record_id?: number | null;
  status?: string;
  subtotal?: number;
  tax_total?: number;
  total_amount?: number;
  insurance_amount?: number;
  discount_amount?: number;
  valid_until?: string | null;
  comment?: string;
  notes?: string;
}

export interface UpdateEstimateRequest {
  title?: string;
  status?: string;
  subtotal?: number;
  tax_total?: number;
  total_amount?: number;
  insurance_amount?: number;
  discount_amount?: number;
  valid_until?: string | null;
  clear_valid_until?: boolean;
  comment?: string;
  notes?: string;
}
