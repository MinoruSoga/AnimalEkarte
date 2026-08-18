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
}
