import { Calendar, CircleDot } from "lucide-react";
import { normalizeKana } from "@/lib/normalize-kana";
import type {
  ActiveFilter,
  ActiveSort,
  FilterProperty,
  SortProperty,
} from "@/components/shared/PropertyFilter/types";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/PropertyFilter/types";
import type { Estimate } from "../types";

export const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "draft", label: "下書き" },
      { value: "sent", label: "送付済み" },
      { value: "approved", label: "承認済み" },
      { value: "rejected", label: "却下" },
    ],
  },
  {
    key: "validUntil",
    label: "有効期限",
    type: "date-range",
    icon: Calendar,
  },
];

export const SORT_PROPERTIES: SortProperty[] = [
  { key: "estimateNo", label: "見積番号" },
  { key: "title", label: "タイトル" },
  { key: "ownerName", label: "飼主名" },
  { key: "validUntil", label: "有効期限" },
  { key: "totalAmount", label: "合計金額" },
];

export const COLUMNS = [
  { header: "見積番号", className: "w-[140px]" },
  { header: "タイトル" },
  { header: "飼主名", className: "w-[130px]" },
  { header: "有効期限", className: "w-[110px]" },
  { header: "合計金額", align: "right" as const },
  { header: "ステータス", className: "w-[100px]" },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

function matchesStatusFilter(estimate: Estimate, filter: ActiveFilter): boolean {
  if (typeof filter.value !== "string") return true;
  switch (filter.condition) {
    case "is":
      return estimate.status === filter.value;
    case "is_not":
      return estimate.status !== filter.value;
    case "is_empty":
      return !estimate.status;
    case "is_not_empty":
      return !!estimate.status;
    default:
      return estimate.status === filter.value;
  }
}

function matchesValidUntilFilter(estimate: Estimate, filter: ActiveFilter): boolean {
  if (typeof filter.value !== "object" || Array.isArray(filter.value)) return true;
  const dateVal = filter.value;
  if (!estimate.validUntil) return filter.condition === "is_empty";
  const d = estimate.validUntil.slice(0, 10);
  switch (filter.condition) {
    case "is":
      return dateVal.from ? d === dateVal.from : true;
    case "is_before":
      return dateVal.from ? d < dateVal.from : true;
    case "is_after":
      return dateVal.from ? d > dateVal.from : true;
    case "is_between":
      return (dateVal.from ? d >= dateVal.from : true) && (dateVal.to ? d <= dateVal.to : true);
    case "is_empty":
      return false;
    case "is_not_empty":
      return true;
    default:
      return true;
  }
}

function compareEstimates(a: Estimate, b: Estimate, sorts: ActiveSort[]): number {
  for (const sort of sorts) {
    if (sort.key === "totalAmount") {
      const cmp = a.totalAmount - b.totalAmount;
      if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
    } else {
      const aVal = String(a[sort.key as keyof Estimate] ?? "");
      const bVal = String(b[sort.key as keyof Estimate] ?? "");
      const cmp = aVal.localeCompare(bVal, "ja");
      if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
    }
  }
  return 0;
}

export function filterAndSortEstimates(
  source: Estimate[],
  activeFilters: ActiveFilter[],
  deferredSearch: string,
  activeSorts: ActiveSort[],
): Estimate[] {
  let items = [...source];

  for (const filter of activeFilters) {
    if (filter.key === "status") {
      items = items.filter((estimate) => matchesStatusFilter(estimate, filter));
    }
    if (filter.key === "validUntil") {
      items = items.filter((estimate) => matchesValidUntilFilter(estimate, filter));
    }
  }

  if (deferredSearch) {
    const normalizedTerm = normalizeKana(deferredSearch).toLowerCase();
    items = items.filter(
      (estimate) =>
        normalizeKana(estimate.title).toLowerCase().includes(normalizedTerm) ||
        normalizeKana(estimate.ownerName ?? "").toLowerCase().includes(normalizedTerm) ||
        estimate.estimateNo.toLowerCase().includes(normalizedTerm),
    );
  }

  if (activeSorts.length > 0) {
    items.sort((a, b) => compareEstimates(a, b, activeSorts));
  }

  return items;
}
