import type { MasterItem } from "@/types";
import type { BackendMasterItem } from "./types";

export function transformMasterItem(data: BackendMasterItem): MasterItem {
  return {
    id: data.id,
    code: data.code,
    name: data.name,
    category: data.category,
    price: data.price || 0,
    status: data.status,
    description: data.description,
    inventoryId: data.inventory_id || undefined,
    defaultQuantity: data.default_quantity || undefined,
  };
}
