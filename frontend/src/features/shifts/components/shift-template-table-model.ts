import { CircleDot } from "lucide-react";

import type { ActiveFilter, FilterProperty } from "@/components/shared/PropertyFilter/types";
import { normalizedIncludes } from "@/lib/normalize-kana";

import { SHIFT_TYPE_LABELS, type ShiftTemplate } from "../types";

export const SHIFT_STATUS_FILTER: FilterProperty = {
  key: "status",
  label: "ステータス",
  type: "select",
  icon: CircleDot,
  options: [
    { value: "active", label: "有効" },
    { value: "inactive", label: "無効" },
  ],
};

export const SHIFT_TEMPLATE_COLUMNS = [
  { header: "", className: "w-11 px-0" },
  { header: "テンプレート名" },
  { header: "種別", className: "w-[80px]" },
  { header: "時間", className: "w-[140px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

export function filterShiftTemplates(
  items: ShiftTemplate[],
  searchTerm: string,
  filters: ActiveFilter[],
): ShiftTemplate[] {
  const term = searchTerm.trim();
  return items.filter((item) => {
    const typeLabel = SHIFT_TYPE_LABELS[item.shift_type] ?? item.shift_type;
    if (term && !normalizedIncludes(item.name, term) && !normalizedIncludes(typeLabel, term)) {
      return false;
    }
    for (const filter of filters) {
      if (filter.key !== "status" || typeof filter.value !== "string") {
        continue;
      }
      const wantActive = filter.value === "active";
      if (filter.condition === "is" && item.is_active !== wantActive) {
        return false;
      }
      if (filter.condition === "is_not" && item.is_active === wantActive) {
        return false;
      }
    }
    return true;
  });
}
