/**
 * Backend API response types
 * Source: frontend/src/types/generated/models.ts (tygo generated)
 */
import type { InventoryItem } from "@/types/generated/models";

export type BackendInventoryItem = InventoryItem;

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
