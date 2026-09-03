import type { CreateCampaignRequest, UpdateCampaignRequest } from "../api/campaign";
import type { CampaignFormData } from "../lib/campaign-side-panel-model";

export function buildCampaignCreateRequest(data: CampaignFormData): CreateCampaignRequest {
  return {
    name: data.name,
    start_date: data.startDate,
    end_date: data.endDate,
    discount_type: data.discountType,
    discount_value: data.discountValue,
    is_active: data.isActive,
    target_categories: data.targetCategories,
    target_item_ids: data.targetItemIds.map(Number),
  };
}

export function buildCampaignUpdateRequest(data: CampaignFormData): UpdateCampaignRequest {
  return buildCampaignCreateRequest(data);
}
