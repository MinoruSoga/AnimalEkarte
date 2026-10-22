/**
 * 見積 API の移行 HTTP DTO（非永続）。
 *
 * 保存 model（生成 models.ts）ではなく、現行 handler の JSON 列をそのまま写す。
 * request と response は別の Go struct なので、片方から Pick せず独立に列挙する。
 * estimates と medical-records（カルテ見積）が同じ wire を読むため共有層に置く
 * （feature 間 import は禁止）。新しい固定版・金銭計算の契約へ切り替えるときに、
 * この adapter ごと置き換える（EMR-527）。
 */
import type { EstimateStatus, ItemCategory, TaxType } from "./generated/models";

/** @see backend/internal/billing/estimate_response.go estimateItemResponse */
export interface LegacyEstimateItemResponse {
  id: number;
  estimate_id: number;
  name: string;
  category: ItemCategory;
  unit_price: number;
  quantity: number;
  subtotal: number;
  tax_type: TaxType;
  tax_rate: number;
  tax_amount: number;
  discount_rate: number;
  discount_amount: number;
  is_insurance_applicable: boolean;
  sort_order: number;
  created_at: string;
}

/** @see backend/internal/billing/nested_summary_response.go ownerSummaryResponse */
export interface LegacyEstimateOwnerSummary {
  id: number;
  name: string;
}

/** @see backend/internal/billing/nested_summary_response.go petSummaryResponse */
export interface LegacyEstimatePetSummary {
  id: number;
  name: string;
}

/** @see backend/internal/billing/estimate_response.go estimateResponse */
export interface LegacyEstimateResponse {
  id: number;
  clinic_id: number;
  estimate_no: string;
  medical_record_id?: number;
  title: string;
  owner_id?: number;
  owner?: LegacyEstimateOwnerSummary;
  pet_id?: number;
  pet?: LegacyEstimatePetSummary;
  status: EstimateStatus;
  subtotal: number;
  tax_total: number;
  total_amount: number;
  insurance_amount: number;
  discount_amount: number;
  valid_until?: string;
  comment: string;
  notes: string;
  created_by?: number;
  supersedes_estimate_id?: number;
  items?: LegacyEstimateItemResponse[];
  created_at: string;
  updated_at: string;
}

/**
 * 見積明細の送信列。作成・更新（明細置換）で共通。
 * @see backend/internal/billing/estimate_request.go createEstimateItemRequest
 */
export interface LegacyEstimateItemRequest {
  name: string;
  category: ItemCategory;
  unit_price: number;
  quantity: number;
  discount_rate: number;
  discount_amount: number;
  is_insurance_applicable: boolean;
  sort_order: number;
  tax_type: TaxType;
  tax_rate: number;
  /**
   * 既存行との一致ヒント（PATCH の明細置換のみ）。この見積の item id と 1:1 で
   * 解決できない ID・重複 ID はサーバ側で 400 になる。新規行では送らない。
   */
  existing_item_id?: number;
}
