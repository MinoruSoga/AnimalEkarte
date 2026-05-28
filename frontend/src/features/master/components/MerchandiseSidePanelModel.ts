import type { FrontendMerchandiseItem } from "../api/merchandise-items";
import type { TaxType } from "@/types/generated/models";

export const MERCHANDISE_CATEGORY_LABELS: Record<string, string> = {
  food: "フード",
  goods: "物販",
  other: "その他",
};

export const MERCHANDISE_CATEGORY_OPTIONS = [
  { value: "food", label: "フード" },
  { value: "goods", label: "物販" },
  { value: "other", label: "その他" },
] as const;

export interface MerchandiseFormData {
  name: string;
  category: string;
  unitPrice: number;
  taxType: TaxType;
  taxRate: number;
  isActive: boolean;
}

export function merchandiseToFormData(item: FrontendMerchandiseItem | null): MerchandiseFormData {
  return {
    name: item?.name ?? "",
    category: item?.category ?? "goods",
    unitPrice: item?.unitPrice ?? 0,
    taxType: item?.taxType ?? "excluded",
    taxRate: item?.taxRate ?? 0.1,
    isActive: item?.isActive ?? true,
  };
}

export function formatMerchandiseTaxRate(taxRate: number) {
  if (taxRate === 0.1) return "10%";
  if (taxRate === 0.08) return "8%";
  return `${taxRate * 100}%`;
}
