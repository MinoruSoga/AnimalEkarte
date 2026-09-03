import type { LucideIcon } from "lucide-react";

/** フィルタプロパティの種類 */
export type FilterType = "select" | "multi-select" | "date-range";

/** select/multi-select の選択肢 */
export interface FilterOption {
  value: string;
  label: string;
}

/** フィルタプロパティの定義（各ページが定義） */
export interface FilterProperty {
  key: string;
  label: string;
  type: FilterType;
  options?: FilterOption[];
  icon?: LucideIcon;
  /**
   * 利用可能な条件を上書きする。
   * 未指定の場合は FILTER_CONDITIONS[type] のデフォルトを使用。
   * DB 制約（NOT NULL / DEFAULT 値あり）により空が存在しないフィールドでは
   * is_empty / is_not_empty を除いた配列を渡すこと。
   */
  conditions?: FilterCondition[];
}

// ─── Filter Conditions ──────────────────────────────────────

/** フィルタ条件の種類 */
export type FilterCondition =
  | "is"
  | "is_not"
  | "contains"
  | "does_not_contain"
  | "is_before"
  | "is_after"
  | "is_between"
  | "is_empty"
  | "is_not_empty";

/** 条件ラベルのペア */
interface ConditionOption {
  value: FilterCondition;
  label: string;
}

/**
 * よく使う条件セットの定数。
 * フィールド定義の conditions フィールドで参照する。
 */
export const CONDITIONS_NO_EMPTY: FilterCondition[] = ["is", "is_not"];
export const CONDITIONS_WITH_EMPTY: FilterCondition[] = [
  "is",
  "is_not",
  "is_empty",
  "is_not_empty",
];

/** プロパティ型ごとに使える条件を定義 */
export const FILTER_CONDITIONS: Record<FilterType, ConditionOption[]> = {
  select: [
    { value: "is", label: "次と一致" },
    { value: "is_not", label: "次と不一致" },
    { value: "is_empty", label: "空" },
    { value: "is_not_empty", label: "空でない" },
  ],
  "multi-select": [
    { value: "contains", label: "含む" },
    { value: "does_not_contain", label: "含まない" },
    { value: "is_empty", label: "空" },
    { value: "is_not_empty", label: "空でない" },
  ],
  "date-range": [
    { value: "is", label: "次と一致" },
    { value: "is_before", label: "以前" },
    { value: "is_after", label: "以降" },
    { value: "is_between", label: "期間内" },
    { value: "is_empty", label: "空" },
    { value: "is_not_empty", label: "空でない" },
  ],
};

// ─── Relative Date ──────────────────────────────────────────

/** 相対日付の時点 */
export type RelativePoint = "this" | "last" | "next";

/** 相対日付の単位 */
export type RelativeUnit = "day" | "week" | "month" | "year";

// ─── Filter Logic ───────────────────────────────────────────

/** フィルタグループの論理演算 */
export type FilterLogic = "and" | "or";

// ─── Active Filter ──────────────────────────────────────────

/** アクティブなフィルタ値 */
export interface ActiveFilter {
  key: string;
  condition: FilterCondition;
  value: string | string[] | { from?: string; to?: string };
  displayValue: string;
}

// ─── Sort ───────────────────────────────────────────────────

/** ソート可能なプロパティの定義 */
export interface SortProperty {
  key: string;
  label: string;
  icon?: LucideIcon;
}

/** アクティブなソート設定 */
export interface ActiveSort {
  key: string;
  direction: "asc" | "desc";
}

// ─── Component Props ────────────────────────────────────────

/** PropertyFilter コンポーネント Props */
export interface PropertyFilterProps {
  properties: FilterProperty[];
  activeFilters: ActiveFilter[];
  onFilterChange: (filters: ActiveFilter[]) => void;
  filterLogic?: FilterLogic;
  onFilterLogicChange?: (logic: FilterLogic) => void;
  searchTerm?: string;
  onSearchChange?: (value: string) => void;
  searchPlaceholder?: string;
  count?: number;
  /** ソート可能なプロパティ一覧 */
  sortProperties?: SortProperty[];
  /** アクティブなソート設定 */
  activeSorts?: ActiveSort[];
  /** ソート変更時のコールバック */
  onSortChange?: (sorts: ActiveSort[]) => void;
}
