// Medical records feature types

/** Interview (問診) history list item */
export interface InterviewHistoryItem {
  id: string;
  date: string;
  author: string;
  type: string;
  title: string;
  content: string;
}

// ── Treatment (治療明細) ──────────────────────────────────────────────

export type TreatmentItemType = 'consultation' | 'procedure' | 'medicine' | 'other';

export interface Treatment {
  id: string;
  medical_record_id: string;
  item_type: TreatmentItemType;
  consultation_id?: string | null;
  procedure_id?: string | null;
  medicine_id?: string | null;
  inventory_id?: string | null;
  unit_price: number;
  quantity: number;
  selected: boolean;
  status: string;
  content: string;
  memo: string;
  insurance: boolean;
  discount_rate: number;
  discount_amount: number;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface CreateTreatmentInput {
  item_type: TreatmentItemType;
  consultation_id?: string | null;
  procedure_id?: string | null;
  medicine_id?: string | null;
  inventory_id?: string | null;
  unit_price?: number;
  quantity?: number;
  selected?: boolean;
  status?: string;
  content?: string;
  memo?: string;
  insurance?: boolean;
  discount_rate?: number;
  discount_amount?: number;
  sort_order?: number;
}

export interface UpdateTreatmentInput {
  item_type?: TreatmentItemType;
  consultation_id?: string | null;
  procedure_id?: string | null;
  medicine_id?: string | null;
  inventory_id?: string | null;
  unit_price?: number;
  quantity?: number;
  selected?: boolean;
  status?: string;
  content?: string;
  memo?: string;
  insurance?: boolean;
  discount_rate?: number;
  discount_amount?: number;
  sort_order?: number;
}

export interface BulkReorderTreatmentsInput {
  treatments: Array<{ id: string; sort_order: number }>;
}

// ── BillingReview (会計医師確認) ──────────────────────────────────────

/** 会計医師確認ステータス */
export type BillingReviewStatus = 'pending' | 'confirmed' | 'returned';

/** 会計医師確認レコード */
export interface BillingReview {
  id: string;
  medical_record_id: string;
  status: BillingReviewStatus;
  confirmed_by?: string | null;
  confirmed_at?: string | null;
  return_reason?: string | null;
  returned_by?: string | null;
  returned_at?: string | null;
  memo?: string | null;
  created_at: string;
  updated_at: string;
}

/** 会計差し戻しリクエスト入力 */
export interface ReturnBillingReviewInput {
  return_reason: string;
}

// ── Vital (バイタル) ──────────────────────────────────────────────────

export interface Vital {
  id: string;
  medical_record_id: string;
  recorded_at: string;
  temperature?: number | null;
  heart_rate?: number | null;
  respiratory_rate?: number | null;
  body_weight?: number | null;
  note?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateVitalInput {
  recorded_at: string;
  temperature?: number | null;
  heart_rate?: number | null;
  respiratory_rate?: number | null;
  body_weight?: number | null;
  note?: string | null;
}

export interface UpdateVitalInput {
  recorded_at?: string;
  temperature?: number | null;
  heart_rate?: number | null;
  respiratory_rate?: number | null;
  body_weight?: number | null;
  note?: string | null;
}
