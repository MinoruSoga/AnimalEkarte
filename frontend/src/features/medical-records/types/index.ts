/**
 * Medical records feature types (UI-facing: string IDs)
 * Backend types: {@link import("@/types/generated/models").Treatment},
 * {@link import("@/types/generated/models").Vital},
 * {@link import("@/types/generated/models").BillingConfirmation} from models.ts
 */
// FE6-3: tygo enum_style: "union"（FE6-1/FE6-2）により生成定数が真の literal union になったため、
// 手書き union を生成型からの re-export へ移行した。drift テストは union-drift.test.ts から削除済み。
import type { TreatmentItemType, BodyWeightUnit } from "@/types/generated/models";
export type { TreatmentItemType, BodyWeightUnit };

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
  is_selected: boolean;
  status: string;
  content: string;
  memo: string;
  is_insurance: boolean;
  discount_rate: number;
  discount_amount: number;
  sort_order: number;
  // #201 投与量自動計算の計算根拠スナップショット（全て nullable・BE treatment.go:52-56 に追随）
  dose_weight_kg?: number | null;
  dose_weight_source?: string | null;
  dose_amount_mg?: number | null;
  dose_amount_unit?: string | null;
  dose_param_snapshot?: Record<string, unknown> | null;
  created_at: string;
  updated_at: string;
}

export interface CreateTreatmentInput {
  item_type: TreatmentItemType;
  consultation_id?: string | null;
  procedure_id?: string | null;
  // ts-review-201 CRITICAL: BE は *uint64（JSON number）を期待する。他の *_id は今のところ
  // どの呼び出し元からも書き込まれていない（未使用の潜在バグ）ため型を温存するが、#201 で
  // 実際に書き込む medicine_id だけは正しい書込型（number）に直す。
  medicine_id?: number | null;
  inventory_id?: string | null;
  unit_price?: number;
  quantity?: number;
  is_selected?: boolean;
  status?: string;
  content?: string;
  memo?: string;
  is_insurance?: boolean;
  discount_rate?: number;
  discount_amount?: number;
  sort_order?: number;
  /** TASK-377: 上限内の下限割れ/著しい乖離時に必須の free-text 逸脱理由（1〜500 Unicode） */
  dose_deviation_reason?: string;
}

export interface UpdateTreatmentInput {
  item_type?: TreatmentItemType;
  consultation_id?: string | null;
  procedure_id?: string | null;
  medicine_id?: string | null;
  inventory_id?: string | null;
  unit_price?: number;
  quantity?: number;
  is_selected?: boolean;
  status?: string;
  content?: string;
  memo?: string;
  is_insurance?: boolean;
  discount_rate?: number;
  discount_amount?: number;
  sort_order?: number;
  /** TASK-377: 上限内の下限割れ/著しい乖離時に必須の free-text 逸脱理由（1〜500 Unicode） */
  dose_deviation_reason?: string;
}

export interface BulkReorderTreatmentsInput {
  treatments: Array<{ id: string; sort_order: number }>;
}

// ── BillingConfirmation (会計医師確認) ──────────────────────────────────────

/** 会計医師確認ステータス @see {@link import("@/types/generated/models").ConfirmationStatus} */
type ConfirmationStatus = "pending" | "confirmed" | "returned";

/** 会計医師確認レコード */
export interface BillingConfirmation {
  id: string;
  medical_record_id: string;
  status: ConfirmationStatus;
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
export interface ReturnBillingConfirmationInput {
  return_reason: string;
}

// ── Vital (バイタル) ──────────────────────────────────────────────────

export interface Vital {
  id: string;
  medical_record_id: string;
  recorded_at: string;
  temperature?: number | null;
  heart_rate?: number | null;
  respiration_rate?: number | null;
  weight?: number | null;
  weight_unit: BodyWeightUnit;
  note?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateVitalInput {
  recorded_at: string;
  temperature?: number | null;
  heart_rate?: number | null;
  respiration_rate?: number | null;
  weight?: number | null;
  weight_unit: BodyWeightUnit;
  note?: string | null;
}

export interface UpdateVitalInput {
  recorded_at?: string;
  temperature?: number | null;
  heart_rate?: number | null;
  respiration_rate?: number | null;
  weight?: number | null;
  weight_unit?: BodyWeightUnit;
  note?: string | null;
}
