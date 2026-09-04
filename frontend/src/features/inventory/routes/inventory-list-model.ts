import { CircleDot, FolderOpen } from "lucide-react";
import type {
  ActiveFilter,
  FilterProperty,
  SortProperty,
} from "@/components/shared/PropertyFilter/types";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/PropertyFilter/types";
import type { InventoryItem } from "@/types";

type CategoryFilter = InventoryItem["category"] | "all";
type StatusFilter = InventoryItem["status"] | "all";

export const CATEGORY_LABELS: Record<InventoryItem["category"], string> = {
  medicine: "医薬品",
  consumable: "消耗品",
  food: "フード",
  other: "その他",
};

export const INVENTORY_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "category",
    label: "カテゴリ",
    type: "select",
    icon: FolderOpen,
    // inventory_items.category NOT NULL — 空値は存在しない
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "medicine", label: "医薬品" },
      { value: "consumable", label: "消耗品" },
      { value: "food", label: "フード" },
      { value: "other", label: "その他" },
    ],
  },
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    // inventory_items.status DEFAULT 'sufficient' — 空値は存在しない
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "sufficient", label: "十分" },
      { value: "low", label: "残少" },
      { value: "out_of_stock", label: "在庫切れ" },
    ],
  },
];

// rendering-hoist-jsx: 静的ソートプロパティ定義
export const INVENTORY_SORT_PROPERTIES: SortProperty[] = [
  { key: "name", label: "品名" },
  { key: "category", label: "カテゴリ" },
  { key: "quantity", label: "在庫数" },
  { key: "status", label: "ステータス" },
];

// BUG-412: サーバサイドページネーションのページサイズ。旧 usePagination() の既定値(20件)と
// backend parsePagination の既定 limit(20) を踏襲する（BUG-411/ACCOUNTING_PAGE_SIZEと同型）。
export const INVENTORY_PAGE_SIZE = 20;

export interface InventoryListFilterResolution {
  category: CategoryFilter;
  excludeCategory: InventoryItem["category"] | null;
  statusFilter: StatusFilter;
  excludeStatus: InventoryItem["status"] | null;
}

export function resolveInventoryListFilters(
  activeFilters: ActiveFilter[],
): InventoryListFilterResolution {
  const categoryFilter = activeFilters.find((f) => f.key === "category");
  const category: CategoryFilter =
    categoryFilter?.condition === "is" ? (categoryFilter.value as CategoryFilter) : "all";
  const excludeCategory: InventoryItem["category"] | null =
    categoryFilter?.condition === "is_not"
      ? (categoryFilter.value as InventoryItem["category"])
      : null;
  const statusFilterEntry = activeFilters.find((f) => f.key === "status");
  const statusFilter: StatusFilter =
    statusFilterEntry?.condition === "is" ? (statusFilterEntry.value as StatusFilter) : "all";
  const excludeStatus: InventoryItem["status"] | null =
    statusFilterEntry?.condition === "is_not"
      ? (statusFilterEntry.value as InventoryItem["status"])
      : null;
  return { category, excludeCategory, statusFilter, excludeStatus };
}

export function excludeInventoryItems(
  items: InventoryItem[],
  excludeCategory: InventoryItem["category"] | null,
  excludeStatus: InventoryItem["status"] | null,
): InventoryItem[] {
  if (!excludeCategory && !excludeStatus) return items;
  return items.filter(
    (item) =>
      (!excludeCategory || item.category !== excludeCategory) &&
      (!excludeStatus || item.status !== excludeStatus),
  );
}

export interface ServerPagePagination<T> {
  paginatedData: T[];
  totalPages: number;
  totalCount: number;
  startIndex: number;
  endIndex: number;
  currentPage: number;
}

export function buildServerPagePagination<T>(input: {
  rows: T[];
  total: number;
  page: number;
  limit: number;
}): ServerPagePagination<T> {
  const totalPages = Math.max(1, Math.ceil(input.total / input.limit));
  return {
    paginatedData: input.rows,
    totalPages,
    totalCount: input.total,
    startIndex: input.total === 0 ? 0 : (input.page - 1) * input.limit + 1,
    endIndex: Math.min(input.page * input.limit, input.total),
    currentPage: input.page,
  };
}

export function nextListSearchParamsWithPage(prev: URLSearchParams, page: number): URLSearchParams {
  const next = new URLSearchParams(prev);
  if (page === 1) {
    next.delete("page");
  } else {
    next.set("page", String(page));
  }
  return next;
}

export function nextListSearchParamsWithoutPage(prev: URLSearchParams): URLSearchParams {
  const next = new URLSearchParams(prev);
  next.delete("page");
  return next;
}
