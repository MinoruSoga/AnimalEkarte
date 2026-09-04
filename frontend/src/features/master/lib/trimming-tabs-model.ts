import type { ReactNode } from "react";

import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { normalizeKana } from "@/lib/normalize-kana";
import { MASTER_TABLE_COL } from "../constants/styles";

interface DataTableColumn {
  header: ReactNode;
  className?: string;
  align?: "left" | "center" | "right";
}

export const TRIMMING_COURSE_COLUMNS: DataTableColumn[] = [
  { header: "コース名" },
  { header: "対象サイズ", className: "w-[120px]" },
  { header: "所要時間", className: "w-[100px]" },
  { header: "単価(税込)", className: "w-[110px]", align: "right" },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" },
];

export const TRIMMING_OPTION_COLUMNS: DataTableColumn[] = [
  { header: "オプション名" },
  { header: "所要時間", className: "w-[100px]" },
  { header: "組合せ可否", className: "w-[110px]", align: "center" },
  { header: "単価(税込)", className: "w-[110px]", align: "right" },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" },
];

interface TrimmingFilterableItem {
  name: string;
  isActive: boolean;
}

export function filterTrimmingItems<T extends TrimmingFilterableItem>({
  items,
  activeFilters,
  searchTerm,
}: {
  items: T[] | undefined;
  activeFilters: ActiveFilter[];
  searchTerm: string;
}): T[] {
  let filtered = items ?? [];

  for (const filter of activeFilters) {
    if (filter.key === "status" && typeof filter.value === "string") {
      const want = filter.value === "active";
      filtered = filtered.filter((item) =>
        filter.condition === "is" ? item.isActive === want : item.isActive !== want,
      );
    }
  }

  if (!searchTerm) return filtered;

  const lower = normalizeKana(searchTerm).toLowerCase();
  return filtered.filter((item) => normalizeKana(item.name).toLowerCase().includes(lower));
}
