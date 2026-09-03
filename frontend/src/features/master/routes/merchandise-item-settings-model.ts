import type {
  CreateMerchandiseItemRequest,
  UpdateMerchandiseItemRequest,
} from "../api/merchandise-items";
import type { MerchandiseFormData } from "../lib/merchandise-side-panel-model";
import type { ItemCategory } from "@/types/generated/models";

export function buildMerchandiseCreateRequest(
  data: MerchandiseFormData,
): CreateMerchandiseItemRequest {
  return {
    name: data.name,
    // FE6-2: category は Select（ItemCategory の値域のみ）で選択されるため実行時は常に安全。
    category: data.category as ItemCategory,
    unit_price: data.unitPrice,
    tax_type: data.taxType,
    tax_rate: data.taxRate,
    is_active: data.isActive,
  };
}

export function buildMerchandiseUpdateRequest(
  data: MerchandiseFormData,
): UpdateMerchandiseItemRequest {
  return buildMerchandiseCreateRequest(data);
}
