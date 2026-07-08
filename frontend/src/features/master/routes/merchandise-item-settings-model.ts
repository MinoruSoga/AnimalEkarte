import type {
  CreateMerchandiseItemRequest,
  UpdateMerchandiseItemRequest,
} from "../api/merchandise-items";
import type { MerchandiseFormData } from "../components/merchandise-side-panel-model";

export function buildMerchandiseCreateRequest(
  data: MerchandiseFormData,
): CreateMerchandiseItemRequest {
  return {
    name: data.name,
    category: data.category,
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
