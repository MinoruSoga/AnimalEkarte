/**
 * Estimates API types
 * Backend types: {@link Estimate}, {@link EstimateItem} from models.ts (tygo generated)
 */
import type { Estimate, EstimateItem } from "@/types/generated/models";

// ── Backend response aliases ──
export type BackendEstimate = Estimate & {
  pet_id?: number;
  pet?: { id: number; name?: string };
};
export type BackendEstimateItem = EstimateItem;
export interface EstimateListResponse {
  data: BackendEstimate[];
  total: number;
  page: number;
  limit: number;
}

export type EstimateItemRequest = Pick<
  EstimateItem,
  | "name"
  | "category"
  | "unit_price"
  | "quantity"
  | "discount_rate"
  | "discount_amount"
  | "is_insurance_applicable"
  | "sort_order"
  | "tax_type"
  | "tax_rate"
> & {
  /**
   * 既存行との一致ヒント（PATCH の明細置換のみ）。この見積の item id と 1:1 で
   * 解決できない ID・重複 ID はサーバ側で 400 になる。新規行では送らない。
   */
  existing_item_id?: number;
};

/**
 * 見積書作成リクエスト
 * @see {@link Estimate} — models.ts 由来の Backend 型
 */
export interface CreateEstimateRequest {
  title: string;
  owner_id?: number | null;
  pet_id?: number | null;
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
  items?: EstimateItemRequest[];
}

/**
 * 見積書更新リクエスト
 * @see {@link Estimate} — models.ts 由来の Backend 型
 */
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
  items?: EstimateItemRequest[];
}
