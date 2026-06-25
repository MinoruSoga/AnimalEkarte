import { ArrowDown, ArrowUp, Search } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { C, STYLE } from "@/lib/design-tokens";

import { CPM_STAGE_OPTIONS, type CPMStage } from "@/lib/cpm-stage";

import type {
  AggregationParams,
  AggregationSortField,
  AmountBasis,
  PeriodPreset,
} from "./api/get-aggregations";
import type { AggregationTab } from "./AggregationDashboardPage";
import {
  AMOUNT_BASIS_OPTIONS,
  DEFAULT_SORT_BY_TAB,
  LAST_VISIT_BUCKET_OPTIONS,
  PERIOD_PRESET_OPTIONS,
  SORT_OPTIONS_BY_TAB,
} from "./AggregationFilterPanelModel";

interface FilterSectionProps {
  params: AggregationParams;
  inputClass: string;
  labelClass: string;
  onParamsChange: (params: Partial<AggregationParams>) => void;
}

interface AggregationSearchFilterProps {
  params: AggregationParams;
  onParamsChange: (params: Partial<AggregationParams>) => void;
}

function toOptionalNumber(value: string) {
  return value === "" ? undefined : Number(value);
}

export function AggregationSearchFilter({ params, onParamsChange }: AggregationSearchFilterProps) {
  return (
    <div className="relative min-w-[220px] flex-1 max-w-sm">
      <Search className={STYLE.searchIcon} />
      <Input
        className={STYLE.searchInput}
        placeholder="飼主名を検索..."
        value={params.search ?? ""}
        onChange={(e) => onParamsChange({ search: e.target.value || undefined, page: 1 })}
      />
    </div>
  );
}

export function RevenueFilters({
  params,
  inputClass,
  labelClass,
  onParamsChange,
}: FilterSectionProps) {
  return (
    <>
      <div className="flex flex-col gap-1 min-w-[110px]">
        <label className={labelClass}>年度</label>
        <Input
          type="number"
          className={inputClass}
          placeholder="年度"
          value={params.year ?? ""}
          onChange={(e) => onParamsChange({ year: toOptionalNumber(e.target.value), page: 1 })}
          min={2000}
        />
      </div>

      <div className="flex flex-col gap-1 min-w-[140px]">
        <label className={labelClass}>売上基準</label>
        <Select
          value={params.amount_basis ?? "gross_total_amount"}
          onValueChange={(value) => onParamsChange({ amount_basis: value as AmountBasis, page: 1 })}
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
            onChange={(e) => onParamsChange({ min_amount: toOptionalNumber(e.target.value), page: 1 })}
            min={0}
          />
          <span className={`text-sm ${C.text50}`}>〜</span>
          <Input
            type="number"
            className={`${inputClass} w-28`}
            placeholder="上限"
            value={params.max_amount ?? ""}
            onChange={(e) => onParamsChange({ max_amount: toOptionalNumber(e.target.value), page: 1 })}
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
          onChange={(e) => onParamsChange({ include_zero: e.target.checked ? true : undefined, page: 1 })}
        />
        <label htmlFor="revenue-include-zero" className={`text-sm ${C.text} cursor-pointer select-none`}>
          0円を含む
        </label>
      </div>
    </>
  );
}

export function VisitFilters({
  params,
  inputClass,
  labelClass,
  onParamsChange,
}: FilterSectionProps) {
  return (
    <>
      <div className="flex flex-col gap-1 min-w-[140px]">
        <label className={labelClass}>期間</label>
        <Select
          value={params.period_preset ?? "last_12_months"}
          onValueChange={(value) => onParamsChange({ period_preset: value as PeriodPreset, page: 1 })}
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
            onChange={(e) => onParamsChange({ from: e.target.value || undefined, page: 1 })}
          />
          <span className={`text-sm ${C.text50}`}>〜</span>
          <Input
            type="date"
            className={`${inputClass} w-36`}
            value={params.to ?? ""}
            onChange={(e) => onParamsChange({ to: e.target.value || undefined, page: 1 })}
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
            onChange={(e) => onParamsChange({ min_visit_count: toOptionalNumber(e.target.value), page: 1 })}
            min={0}
          />
          <span className={`text-sm ${C.text50}`}>〜</span>
          <Input
            type="number"
            className={`${inputClass} w-24`}
            placeholder="上限"
            value={params.max_visit_count ?? ""}
            onChange={(e) => onParamsChange({ max_visit_count: toOptionalNumber(e.target.value), page: 1 })}
            min={0}
          />
        </div>
      </div>
    </>
  );
}

export function LastVisitFilters({
  params,
  inputClass,
  labelClass,
  onParamsChange,
}: FilterSectionProps) {
  return (
    <>
      <div className="flex flex-col gap-1 min-w-[140px]">
        <label className={labelClass}>最終来院</label>
        <Select
          value={params.last_visit_bucket ?? "over_3m"}
          onValueChange={(value) => onParamsChange({ last_visit_bucket: value === "all" ? undefined : value, page: 1 })}
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
          onChange={(e) => onParamsChange({ include_no_visit: e.target.checked ? true : undefined, page: 1 })}
        />
        <label htmlFor="last-visit-include-no-visit" className={`text-sm ${C.text} cursor-pointer select-none`}>
          来院なしを含む
        </label>
      </div>
    </>
  );
}

// ISSUE-180: CPM セグメント絞り込み。全タブ共通の属性フィルタ。
// 値域・ラベルは共有定義 @/lib/cpm-stage（健診対象者抽出と単一真実源）。
export function CPMStageFilter({
  params,
  inputClass,
  labelClass,
  onParamsChange,
}: FilterSectionProps) {
  return (
    <div className="flex flex-col gap-1 min-w-[160px]">
      <label className={labelClass}>CPMセグメント</label>
      <Select
        value={params.cpm_stage ?? "all"}
        onValueChange={(value) =>
          onParamsChange({
            cpm_stage: value === "all" ? undefined : (value as CPMStage),
            page: 1,
          })
        }
      >
        <SelectTrigger className={inputClass} aria-label="CPMセグメント">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">すべて</SelectItem>
          {CPM_STAGE_OPTIONS.map((opt) => (
            <SelectItem key={opt.value} value={opt.value}>
              {opt.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}

interface SortControlsProps {
  params: AggregationParams;
  activeTab: AggregationTab;
  inputClass: string;
  labelClass: string;
  onParamsChange: (params: Partial<AggregationParams>) => void;
}

export function SortControls({
  params,
  activeTab,
  inputClass,
  labelClass,
  onParamsChange,
}: SortControlsProps) {
  const sortOptions = SORT_OPTIONS_BY_TAB[activeTab];
  const sortValue = sortOptions.some((opt) => opt.value === params.sort)
    ? params.sort
    : DEFAULT_SORT_BY_TAB[activeTab];
  const orderValue = params.order ?? "desc";

  return (
    <div className="flex flex-col gap-1 min-w-[160px]">
      <label className={labelClass}>並び替え</label>
      <div className="flex items-center gap-1">
        <Select
          value={sortValue}
          onValueChange={(value) => onParamsChange({ sort: value as AggregationSortField, page: 1 })}
        >
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
          onClick={() => onParamsChange({ order: orderValue === "asc" ? "desc" : "asc", page: 1 })}
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
  );
}
