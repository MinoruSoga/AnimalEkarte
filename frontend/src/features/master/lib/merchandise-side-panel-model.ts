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
  // EMR-530: 入力字句をそのまま保持する exact decimal 文字列（10% は "0.1"、number 化しない）。
  taxRate: string;
  isActive: boolean;
}

export function merchandiseToFormData(item: FrontendMerchandiseItem | null): MerchandiseFormData {
  return {
    name: item?.name ?? "",
    category: item?.category ?? "goods",
    unitPrice: item?.unitPrice ?? 0,
    taxType: item ? (item.taxType ?? "excluded") : "included",
    taxRate: item?.taxRate ?? "0.1",
    isActive: item?.isActive ?? true,
  };
}

// EMR-530: 保存値・送信値は文字列のまま。この比較は数値として行う（表記ではなく値で判定）。
export function isMerchandiseTaxRate(taxRate: string, expected: string) {
  return taxRateValue(taxRate) !== null && taxRateValue(taxRate) === taxRateValue(expected);
}

function taxRateValue(text: string): number | null {
  if (!/^(0|[1-9][0-9]*)(\.[0-9]+)?$/.test(text)) return null;
  return Number(text);
}

export function formatMerchandiseTaxRate(taxRate: string) {
  const value = taxRateValue(taxRate);
  // 契約外の字句は捏造した率に変換せず、その文字列を表示する。
  if (value === null) return `${taxRate}`;
  if (value === 0.1) return "10%";
  if (value === 0.08) return "8%";
  return `${value * 100}%`;
}
