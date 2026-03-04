export interface BackendMasterItem {
  id: string;
  code: string;
  name: string;
  category: string;
  price?: number | null;
  status: "active" | "inactive";
  description?: string;
  inventory_id?: string | null;
  default_quantity?: number | null;
  created_at: string;
  updated_at: string;
}

export interface CreateMasterItemRequest {
  code: string;
  name: string;
  category: string;
  price?: number | null;
  description?: string;
  inventory_id?: string | null;
  default_quantity?: number | null;
}

export interface UpdateMasterItemRequest {
  code?: string;
  name?: string;
  category?: string;
  price?: number | null;
  status?: "active" | "inactive";
  description?: string;
  inventory_id?: string | null;
  default_quantity?: number | null;
}
