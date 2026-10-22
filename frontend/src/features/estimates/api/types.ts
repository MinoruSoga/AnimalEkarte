/**
 * Estimates API types
 * Backend types: {@link LegacyEstimateResponse}, {@link LegacyEstimateItemResponse}
 * （移行 HTTP DTO。保存 model ではない — EMR-527）
 */
import type {
  LegacyEstimateItemRequest,
  LegacyEstimateItemResponse,
  LegacyEstimateResponse,
} from "./legacy-wire";

// ── Backend response aliases ──
export type BackendEstimate = LegacyEstimateResponse;
export type BackendEstimateItem = LegacyEstimateItemResponse;
export interface EstimateListResponse {
  data: BackendEstimate[];
  total: number;
  page: number;
  limit: number;
}

export type EstimateItemRequest = LegacyEstimateItemRequest;

/**
 * 見積書作成リクエスト
 * @see backend/internal/billing/estimate_request.go createEstimateRequest
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
 * @see backend/internal/billing/estimate_request.go updateEstimateRequest
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
