import { formatCurrency } from "@/lib/format/number";
import type { Cage, CageSize, CageType } from "../api/cages";

export const CAGE_TYPE_LABELS: Record<CageType, string> = {
  icu: "ICU",
  dog: "犬舎",
  cat: "猫舎",
  general: "汎用",
};

export const CAGE_SIZE_LABELS: Record<CageSize, string> = {
  small: "小型",
  medium: "中型",
  large: "大型",
};

export const CAGE_TYPE_OPTIONS = [
  { value: "icu", label: "ICU" },
  { value: "dog", label: "犬舎" },
  { value: "cat", label: "猫舎" },
  { value: "general", label: "汎用" },
] satisfies { value: CageType; label: string }[];

export const CAGE_SIZE_OPTIONS = [
  { value: "small", label: "小型" },
  { value: "medium", label: "中型" },
  { value: "large", label: "大型" },
] satisfies { value: CageSize; label: string }[];

export interface CageFormData {
  name: string;
  cageType: CageType;
  cageSize: CageSize;
  price: number;
  description: string;
  isActive: boolean;
}

export function cageToFormData(item: Cage | null): CageFormData {
  return {
    name: item?.name ?? "",
    cageType: item?.cageType ?? "general",
    cageSize: item?.cageSize ?? "medium",
    price: item?.price ?? 0,
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  };
}

export function formatCagePrice(price: number | null | undefined) {
  return formatCurrency(price);
}

/**
 * EMR-306: 連結メンバーケージの代表ケージを一覧から解決する。
 * cage が未連結（mergedIntoCageId=null）か、代表ケージが一覧に存在しない
 * 場合は null。代表ケージ自身（他行が自分を指す）ではなく、あくまで
 * 「この cage が所属するグループの代表」を返す。
 */
export function findCageRepresentative(cages: readonly Cage[], cage: Cage | null): Cage | null {
  if (cage == null || cage.mergedIntoCageId == null) return null;
  return cages.find((c) => c.id === String(cage.mergedIntoCageId)) ?? null;
}
