export interface BackendInventoryItem {
  id: string;
  name: string;
  category: string;
  quantity: number;
  unit: string;
  min_stock_level: number;
  location?: string;
  expiry_date?: string | null;
  supplier?: string;
  last_restocked?: string | null;
  status: "sufficient" | "low" | "out_of_stock";
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
