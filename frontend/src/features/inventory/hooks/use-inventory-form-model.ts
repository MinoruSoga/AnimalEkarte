import type {
  CreateInventoryItemRequest,
  UpdateInventoryItemRequest,
} from "../api/types";

function readFormString(formData: FormData, key: string): string {
  const value = formData.get(key);
  return typeof value === "string" ? value : "";
}

export interface InventoryFormFields {
  name: string;
  unit: string;
  quantityStr: string;
  minStockLevelStr: string;
  expiryDateStr: string;
  lastRestockedStr: string;
  location: string | undefined;
  supplier: string | undefined;
}

export function readInventoryFormFields(formData: FormData): InventoryFormFields {
  return {
    name: readFormString(formData, "name").trim(),
    unit: readFormString(formData, "unit").trim(),
    quantityStr: readFormString(formData, "quantity"),
    minStockLevelStr: readFormString(formData, "minStockLevel"),
    expiryDateStr: readFormString(formData, "expiryDate"),
    lastRestockedStr: readFormString(formData, "lastRestocked"),
    location: readFormString(formData, "location") || undefined,
    supplier: readFormString(formData, "supplier") || undefined,
  };
}

export function validateInventoryFormFields(
  fields: InventoryFormFields,
): Record<string, string> {
  const quantity = Number(fields.quantityStr);
  const minStockLevel = Number(fields.minStockLevelStr);
  return {
    ...(fields.name === "" ? { name: "品名を入力してください" } : {}),
    ...(fields.unit === "" ? { unit: "単位を入力してください" } : {}),
    ...(fields.quantityStr.trim() === "" || !Number.isInteger(quantity) || quantity < 0
      ? { quantity: "現在庫数は0以上の整数で入力してください" }
      : {}),
    ...(fields.minStockLevelStr.trim() === "" || !Number.isInteger(minStockLevel) || minStockLevel < 0
      ? { minStockLevel: "最低在庫数は0以上の整数で入力してください" }
      : {}),
  };
}

export function buildInventoryItemRequest(
  fields: InventoryFormFields,
  category: string,
): CreateInventoryItemRequest & UpdateInventoryItemRequest {
  return {
    name: fields.name,
    category,
    quantity: Number(fields.quantityStr),
    unit: fields.unit,
    min_stock_level: Number(fields.minStockLevelStr),
    location: fields.location,
    expiry_date: fields.expiryDateStr || undefined,
    supplier: fields.supplier,
    last_restocked: fields.lastRestockedStr || undefined,
  };
}
