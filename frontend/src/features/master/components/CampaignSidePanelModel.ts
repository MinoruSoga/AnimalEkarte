import type { Campaign, CampaignDiscountType } from "../api/campaign";

export interface CampaignFormData {
  name: string;
  startDate: string; // YYYY-MM-DD
  endDate: string; // YYYY-MM-DD
  discountType: CampaignDiscountType;
  discountValue: number;
  isActive: boolean;
  targetCategories: string[];
  targetItemIds: string[];
}

export function campaignToFormData(item: Campaign | null): CampaignFormData {
  return {
    name: item?.name ?? "",
    startDate: item?.startDate ?? "",
    endDate: item?.endDate ?? "",
    discountType: item?.discountType ?? "rate",
    discountValue: item?.discountValue ?? 0,
    isActive: item?.isActive ?? true,
    targetCategories: item?.targetCategories ?? [],
    targetItemIds: item?.targetItemIds ?? [],
  };
}
