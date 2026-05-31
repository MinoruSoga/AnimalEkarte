import type {
  AggregationSortField,
  AmountBasis,
  PeriodPreset,
} from "./api/get-aggregations";
import type { AggregationTab } from "./AggregationDashboardPage";

export const AMOUNT_BASIS_OPTIONS = [
  { value: "gross_total_amount", label: "売上総額" },
  { value: "paid_amount", label: "入金額" },
  { value: "net_paid_amount", label: "返金控除後" },
] as const satisfies readonly { value: AmountBasis; label: string }[];

export const PERIOD_PRESET_OPTIONS = [
  { value: "last_3_months", label: "直近3ヶ月" },
  { value: "last_6_months", label: "直近6ヶ月" },
  { value: "last_12_months", label: "直近12ヶ月" },
  { value: "calendar_year", label: "今年度" },
] as const satisfies readonly { value: PeriodPreset; label: string }[];

// 仕様書 §4.3 準拠: within_3m は「3ヶ月未満」(0〜89日)。「以内」は誤訳。
export const LAST_VISIT_BUCKET_OPTIONS = [
  { value: "within_3m", label: "3ヶ月未満" },
  { value: "over_3m", label: "3ヶ月以上" },
  { value: "over_6m", label: "6ヶ月以上" },
  { value: "over_1y", label: "1年以上" },
  { value: "no_visit", label: "来院なし" },
] as const;

type SortOption = { value: AggregationSortField; label: string };

// 並び替え選択肢は仕様書 §4.1〜4.3 で軸ごとに固定する。
export const SORT_OPTIONS_BY_TAB: Record<AggregationTab, SortOption[]> = {
  revenue: [
    { value: "annual_amount", label: "年間診療費" },
    { value: "period_visit_count", label: "期間内来院回数" },
    { value: "last_visit_date", label: "最終来院日" },
    { value: "days_since_last_visit", label: "経過日数" },
    { value: "owner_name", label: "飼い主名" },
  ],
  visit: [
    { value: "period_visit_count", label: "期間内来院回数" },
    { value: "last_visit_date", label: "最終来院日" },
    { value: "owner_name", label: "飼い主名" },
  ],
  last_visit: [
    { value: "last_visit_date", label: "最終来院日" },
    { value: "days_since_last_visit", label: "経過日数" },
    { value: "owner_name", label: "飼い主名" },
  ],
};

export const DEFAULT_SORT_BY_TAB: Record<AggregationTab, AggregationSortField> = {
  revenue: "annual_amount",
  visit: "period_visit_count",
  last_visit: "last_visit_date",
};
