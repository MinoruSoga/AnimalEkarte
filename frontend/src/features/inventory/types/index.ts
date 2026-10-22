/**
 * Backend API response types
 *
 * 在庫 API の移行 HTTP DTO（非永続）。保存 model（生成 models.ts）ではなく、
 * 現行 handler が返す JSON の列をそのまま写す（omitempty は省略可 ?）。
 * enum union だけを生成 models から受け取る。在庫の wire 契約を生成型へ移すときに、
 * この DTO ごと置き換える（EMR-527）。
 */
import type { InventoryCategory, InventoryStatus } from "@/types/generated/models";

/** @see backend/internal/inventory/inventory_response.go inventoryResponse */
export interface BackendInventoryItem {
  id: number;
  clinic_id: number;
  name: string;
  category: InventoryCategory;
  quantity: number;
  unit: string;
  min_stock_level: number;
  location: string;
  expiry_date?: string;
  supplier: string;
  last_restocked?: string;
  /** 保存値ではなく quantity / min_stock_level から読み取りごとに導出される（SD-4）。 */
  status: InventoryStatus;
  created_at: string;
  updated_at: string;
}

export interface CreateInventoryItemRequest {
  name: string;
  category: string;
  quantity: number;
  unit: string;
  min_stock_level: number;
  location?: string;
  expiry_date?: string | null;
  supplier?: string;
  last_restocked?: string | null;
}

export interface UpdateInventoryItemRequest {
  name?: string;
  category?: string;
  quantity?: number;
  unit?: string;
  min_stock_level?: number;
  location?: string;
  expiry_date?: string | null;
  supplier?: string;
  last_restocked?: string | null;
}
