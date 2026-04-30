import { useCallback } from "react";
import { ArrowDown, ArrowUp, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { C, STYLE } from "@/lib/design-tokens";
import type {
  AggregationParams,
  AggregationSortField,
  AmountBasis,
  PeriodPreset,
} from "./api/get-aggregations";
import type { AggregationTab } from "./AggregationDashboardPage";

interface AggregationFilterPanelProps {
  params: AggregationParams;
  onParamsChange: (params: Partial<AggregationParams>) => void;
  activeTab: AggregationTab;
}

const AMOUNT_BASIS_OPTIONS = [
  { value: "gross_total_amount", label: "売上総額" },
  { value: "paid_amount", label: "入金額" },
  { value: "net_paid_amount", label: "返金控除後" },
] as const;

const PERIOD_PRESET_OPTIONS = [
  { value: "last_3_months", label: "直近3ヶ月" },
  { value: "last_6_months", label: "直近6ヶ月" },
  { value: "last_12_months", label: "直近12ヶ月" },
  { value: "calendar_year", label: "今年度" },
] as const;

// 仕様書 §4.3 準拠: within_3m は「3ヶ月未満」(0〜89日)。「以内」は誤訳。
const LAST_VISIT_BUCKET_OPTIONS = [
  { value: "within_3m", label: "3ヶ月未満" },
  { value: "over_3m", label: "3ヶ月以上" },
  { value: "over_6m", label: "6ヶ月以上" },
  { value: "over_1y", label: "1年以上" },
  { value: "no_visit", label: "来院なし" },
] as const;

// 並び替え選択肢は仕様書 §4.1〜4.3 で軸ごとに固定する。
// 軸を跨いで存在しない sort 値を送るとBEで400になるため、画面側で軸別にホワイトリスト化する。
type SortOption = { value: AggregationSortField; label: string };

const SORT_OPTIONS_BY_TAB: Record<AggregationTab, SortOption[]> = {
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

const DEFAULT_SORT_BY_TAB: Record<AggregationTab, AggregationSortField> = {
  revenue: "annual_amount",
  visit: "period_visit_count",
  last_visit: "last_visit_date",
};

export function AggregationFilterPanel({ params, onParamsChange, activeTab }: AggregationFilterPanelProps) {
  const handleSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onParamsChange({ search: e.target.value || undefined, page: 1 });
    },
    [onParamsChange]
  );

  // Revenue tab handlers
  const handleYearChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const val = e.target.value === "" ? undefined : Number(e.target.value);
      onParamsChange({ year: val, page: 1 });
    },
    [onParamsChange]
  );

  const handleAmountBasisChange = useCallback(
    (value: string) => {
      onParamsChange({ amount_basis: value as AmountBasis, page: 1 });
    },
    [onParamsChange]
  );

  const handleMinAmountChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const val = e.target.value === "" ? undefined : Number(e.target.value);
      onParamsChange({ min_amount: val, page: 1 });
    },
    [onParamsChange]
  );

  const handleMaxAmountChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const val = e.target.value === "" ? undefined : Number(e.target.value);
      onParamsChange({ max_amount: val, page: 1 });
    },
    [onParamsChange]
  );

  const handleIncludeZeroChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onParamsChange({ include_zero: e.target.checked ? true : undefined, page: 1 });
    },
    [onParamsChange]
  );

  // Visit tab handlers
  const handlePeriodPresetChange = useCallback(
    (value: string) => {
      onParamsChange({ period_preset: value as PeriodPreset, page: 1 });
    },
    [onParamsChange]
  );

  const handleFromChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onParamsChange({ from: e.target.value || undefined, page: 1 });
    },
    [onParamsChange]
  );

  const handleToChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onParamsChange({ to: e.target.value || undefined, page: 1 });
    },
    [onParamsChange]
  );

  const handleMinVisitCountChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const val = e.target.value === "" ? undefined : Number(e.target.value);
      onParamsChange({ min_visit_count: val, page: 1 });
    },
    [onParamsChange]
  );

  const handleMaxVisitCountChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const val = e.target.value === "" ? undefined : Number(e.target.value);
      onParamsChange({ max_visit_count: val, page: 1 });
    },
    [onParamsChange]
  );

  // Last visit tab handlers
  const handleLastVisitBucketChange = useCallback(
    (value: string) => {
      onParamsChange({ last_visit_bucket: value === "all" ? undefined : value, page: 1 });
    },
    [onParamsChange]
  );

  const handleIncludeNoVisitChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onParamsChange({ include_no_visit: e.target.checked ? true : undefined, page: 1 });
    },
    [onParamsChange]
  );

  // 並び替えハンドラ。タブごとの sort オプションは SORT_OPTIONS_BY_TAB に固定。
  const handleSortChange = useCallback(
    (value: string) => {
      onParamsChange({ sort: value as AggregationSortField, page: 1 });
    },
    [onParamsChange]
  );

  const handleOrderToggle = useCallback(() => {
    onParamsChange({ order: params.order === "asc" ? "desc" : "asc", page: 1 });
  }, [onParamsChange, params.order]);

  const sortOptions = SORT_OPTIONS_BY_TAB[activeTab];
  // 現在の sort が当該タブで有効でない場合は軸の既定値にフォールバック表示（state は触らない）。
  const sortValue = sortOptions.some((opt) => opt.value === params.sort)
    ? params.sort
    : DEFAULT_SORT_BY_TAB[activeTab];
  const orderValue = params.order ?? "desc";

  // 入力スタイルは NotionFilter のツールバー高 (h-9) に揃え、
  // 余計な装飾（外枠ボックス・大きめラベル）を排除して他ページと密度を統一する。
  const inputClass = `h-9 ${C.borderMedium} ${C.text} bg-white text-base`;
  const labelClass = `text-xs ${C.text50}`;

  return (
    <div className="flex flex-wrap items-end gap-x-3 gap-y-2">
      {/* 飼い主名検索 (他ページの NotionFilter 同様、ラベルは置かず placeholder で示す) */}
      <div className="relative min-w-[220px] flex-1 max-w-sm">
        <Search className={STYLE.searchIcon} />
        <Input
          className={STYLE.searchInput}
          placeholder="飼い主名を検索..."
          value={params.search ?? ""}
          onChange={handleSearchChange}
        />
      </div>

      {/* Revenue tab filters */}
      {activeTab === "revenue" ? (
        <>
          <div className="flex flex-col gap-1 min-w-[110px]">
            <label className={labelClass}>年度</label>
            <Input
              type="number"
              className={inputClass}
              placeholder="年度"
              value={params.year ?? ""}
              onChange={handleYearChange}
              min={2000}
            />
          </div>

          <div className="flex flex-col gap-1 min-w-[140px]">
            <label className={labelClass}>売上基準</label>
            <Select
              value={params.amount_basis ?? "gross_total_amount"}
              onValueChange={handleAmountBasisChange}
            >
              <SelectTrigger className={inputClass}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {AMOUNT_BASIS_OPTIONS.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-1">
            <label className={labelClass}>売上額</label>
            <div className="flex items-center gap-2">
              <Input
                type="number"
                className={`${inputClass} w-28`}
                placeholder="下限"
                value={params.min_amount ?? ""}
                onChange={handleMinAmountChange}
                min={0}
              />
              <span className={`text-sm ${C.text50}`}>〜</span>
              <Input
                type="number"
                className={`${inputClass} w-28`}
                placeholder="上限"
                value={params.max_amount ?? ""}
                onChange={handleMaxAmountChange}
                min={0}
              />
            </div>
          </div>

          <div className="flex items-center gap-2 h-9">
            <input
              type="checkbox"
              id="revenue-include-zero"
              className="size-4 cursor-pointer"
              checked={params.include_zero === true}
              onChange={handleIncludeZeroChange}
            />
            <label htmlFor="revenue-include-zero" className={`text-sm ${C.text} cursor-pointer select-none`}>
              0円を含む
            </label>
          </div>
        </>
      ) : null}

      {/* Visit tab filters */}
      {activeTab === "visit" ? (
        <>
          <div className="flex flex-col gap-1 min-w-[140px]">
            <label className={labelClass}>期間</label>
            <Select
              value={params.period_preset ?? "last_12_months"}
              onValueChange={handlePeriodPresetChange}
            >
              <SelectTrigger className={inputClass}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {PERIOD_PRESET_OPTIONS.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-1">
            <label className={labelClass}>日付範囲</label>
            <div className="flex items-center gap-2">
              <Input
                type="date"
                className={`${inputClass} w-36`}
                value={params.from ?? ""}
                onChange={handleFromChange}
              />
              <span className={`text-sm ${C.text50}`}>〜</span>
              <Input
                type="date"
                className={`${inputClass} w-36`}
                value={params.to ?? ""}
                onChange={handleToChange}
              />
            </div>
          </div>

          <div className="flex flex-col gap-1">
            <label className={labelClass}>来院回数</label>
            <div className="flex items-center gap-2">
              <Input
                type="number"
                className={`${inputClass} w-24`}
                placeholder="下限"
                value={params.min_visit_count ?? ""}
                onChange={handleMinVisitCountChange}
                min={0}
              />
              <span className={`text-sm ${C.text50}`}>〜</span>
              <Input
                type="number"
                className={`${inputClass} w-24`}
                placeholder="上限"
                value={params.max_visit_count ?? ""}
                onChange={handleMaxVisitCountChange}
                min={0}
              />
            </div>
          </div>
        </>
      ) : null}

      {/* Last visit tab filters */}
      {activeTab === "last_visit" ? (
        <>
          <div className="flex flex-col gap-1 min-w-[140px]">
            <label className={labelClass}>最終来院</label>
            <Select
              value={params.last_visit_bucket ?? "over_3m"}
              onValueChange={handleLastVisitBucketChange}
            >
              <SelectTrigger className={inputClass}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {LAST_VISIT_BUCKET_OPTIONS.map((opt) => (
                  <SelectItem key={opt.value} value={opt.value}>
                    {opt.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2 h-9">
            <input
              type="checkbox"
              id="last-visit-include-no-visit"
              className="size-4 cursor-pointer"
              checked={params.include_no_visit === true}
              onChange={handleIncludeNoVisitChange}
            />
            <label htmlFor="last-visit-include-no-visit" className={`text-sm ${C.text} cursor-pointer select-none`}>
              来院なしを含む
            </label>
          </div>
        </>
      ) : null}

      {/* 並び替え (sort / order) — 軸ごとに選択肢が変わる。order はトグル風ボタンで asc/desc を切替。 */}
      <div className="flex flex-col gap-1 min-w-[160px]">
        <label className={labelClass}>並び替え</label>
        <div className="flex items-center gap-1">
          <Select value={sortValue} onValueChange={handleSortChange}>
            <SelectTrigger className={inputClass} aria-label="並び替え">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {sortOptions.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            type="button"
            variant="outline"
            size="icon"
            className={`h-9 w-9 shrink-0 ${C.borderMedium}`}
            onClick={handleOrderToggle}
            aria-label={orderValue === "asc" ? "昇順 (クリックで降順)" : "降順 (クリックで昇順)"}
            title={orderValue === "asc" ? "昇順" : "降順"}
          >
            {orderValue === "asc" ? (
              <ArrowUp className="size-4" />
            ) : (
              <ArrowDown className="size-4" />
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}
