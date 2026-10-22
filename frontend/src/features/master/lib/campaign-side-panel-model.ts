import type { Campaign, CampaignDiscountType } from "../api/campaign";

export interface CampaignFormData {
  name: string;
  startDate: string; // YYYY-MM-DD
  endDate: string; // YYYY-MM-DD
  discountType: CampaignDiscountType;
  // EMR-530: 入力字句をそのまま保持する exact decimal 文字列（number 化しない）。
  discountValue: string;
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
    discountValue: item?.discountValue ?? "0",
    isActive: item?.isActive ?? true,
    targetCategories: item?.targetCategories ?? [],
    targetItemIds: item?.targetItemIds ?? [],
  };
}
