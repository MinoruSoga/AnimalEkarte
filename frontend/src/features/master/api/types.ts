export interface CreateMasterItemRequest {
  name: string;
  category: string;
  price?: number | null;
  description?: string;
  inventory_id?: string | null;
  default_quantity?: number | null;
}

export interface UpdateMasterItemRequest {
  name?: string;
  category?: string;
  price?: number | null;
  status?: "active" | "inactive";
  description?: string;
  inventory_id?: string | null;
  default_quantity?: number | null;
}
