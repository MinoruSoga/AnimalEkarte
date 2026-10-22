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

// EMR-530: 保存値・送信値は文字列のまま。preset 判定（10%/8%）は float 経由ではなく
// 十進文字列の正準化で行う（多数桁の字句が float 化で 0.1 と誤一致しない）。
function canonicalDecimal(text: string): string | null {
  if (!/^(0|[1-9][0-9]*)(\.[0-9]+)?$/.test(text)) return null;
  const [intPart, fracPart = ""] = text.split(".");
  const frac = fracPart.replace(/0+$/, "");
  if (frac === "") return intPart;
  return `${intPart}.${frac}`;
}

export function isMerchandiseTaxRate(taxRate: string, expected: string) {
  const canonical = canonicalDecimal(taxRate);
  return canonical !== null && canonical === canonicalDecimal(expected);
}

// ×100 を文字列の桁移動で exact に行う（表示専用。float の桁落ちを出さない）。
function percentLiteral(text: string): string | null {
  const m = /^(0|[1-9][0-9]*)(?:\.([0-9]+))?$/.exec(text);
  if (!m) return null;
  const [, intPart, fracPart = ""] = m;
  const digits = intPart + fracPart;
  const point = intPart.length + 2; // 小数点を2桁右へ
  const padded = digits.padEnd(point, "0");
  const whole = padded.slice(0, point).replace(/^0+(?=\d)/, "");
  const frac = padded.slice(point).replace(/0+$/, "");
  return frac === "" ? whole : `${whole}.${frac}`;
}

export function formatMerchandiseTaxRate(taxRate: string) {
  if (isMerchandiseTaxRate(taxRate, "0.1")) return "10%";
  if (isMerchandiseTaxRate(taxRate, "0.08")) return "8%";
  // 契約外の字句は捏造した率に変換せず、その文字列を表示する。
  const percent = percentLiteral(taxRate);
  return percent === null ? `${taxRate}` : `${percent}%`;
}
