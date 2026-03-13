/**
 * Backend API response types
 * Generated from backend/docs/api.yaml via openapi-typescript
 * DO NOT EDIT manually — run `make codegen` to regenerate
 */
import type { components } from "@/types/generated/api";

export type BackendInventoryItem = components["schemas"]["InventoryItem"];

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
