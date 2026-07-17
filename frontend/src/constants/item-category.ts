import type { ItemCategory } from "@/types/generated/models";

// FE5-26: 会計側の訳語を正本化(見積側の「薬剤/食事/物品」はここへ統一)。
// FE4-1由来: vaccine/hotel/training を意図的に含まない部分集合マップ(消費側は `?? item.category` 等でフォールバック)
export const CATEGORY_LABELS: Partial<Record<ItemCategory, string>> = {
  examination: "診察",
  test: "検査",
  procedure: "処置",
  surgery: "手術",
  medicine: "処方",
  food: "フード",
  goods: "物販",
  trimming: "トリミング",
  other: "その他",
};
