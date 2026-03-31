/**
 * 汎用マスタアイテム作成リクエスト
 * 個別マスタ（Consultation, Procedure, Medicine 等）は専用 Request 型（src/types/treatment.ts 等）を使用。
 * この型は旧 master_items 汎用CRUD 用の legacy 型。
 */
export interface CreateMasterItemRequest {
  name: string;
  category: string;
  code?: string;
  price?: number | null;
  description?: string;
  inventory_id?: string | null;
  default_quantity?: number | null;
  // Consultation-specific fields
  timeCondition?: string;
  duration?: number | null;
}

export interface UpdateMasterItemRequest {
  name?: string;
  category?: string;
  code?: string;
  price?: number | null;
  status?: "active" | "inactive";
  description?: string;
  inventory_id?: string | null;
  default_quantity?: number | null;
  // Consultation-specific fields
  timeCondition?: string;
  duration?: number | null;
}
