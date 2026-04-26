import { useCallback } from "react";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { C, STYLE } from "@/lib/design-tokens";
import type { LtvOwnersParams, AmountBasis, PeriodPreset } from "../api/get-ltv-owners";
import type { LtvTab } from "./LtvOwnerTable";

interface LtvFilterPanelProps {
  params: LtvOwnersParams;
  onParamsChange: (params: Partial<LtvOwnersParams>) => void;
  activeTab: LtvTab;
}

const CPM_STAGE_OPTIONS = [
  { value: "all", label: "すべて" },
  { value: "encounter", label: "エンカウンター" },
  { value: "growing", label: "グロウィング" },
  { value: "core", label: "コア" },
  { value: "noah", label: "ノア" },
  { value: "spot", label: "スポット" },
  { value: "dormant", label: "休眠" },
] as const;

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

const LAST_VISIT_BUCKET_OPTIONS = [
  { value: "within_3m", label: "3ヶ月以内" },
  { value: "over_3m", label: "3ヶ月以上" },
  { value: "over_6m", label: "6ヶ月以上" },
  { value: "over_1y", label: "1年以上" },
  { value: "no_visit", label: "来院なし" },
] as const;

export function LtvFilterPanel({ params, onParamsChange, activeTab }: LtvFilterPanelProps) {
  const handleCpmStageChange = useCallback(
    (value: string) => {
      onParamsChange({ cpm_stage: value === "all" ? undefined : value, page: 1 });
    },
    [onParamsChange]
  );

  const handleHasLineChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onParamsChange({ has_line: e.target.checked ? true : undefined, page: 1 });
    },
    [onParamsChange]
  );

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

  // LTV tab handlers
  const handleMinFeeChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const val = e.target.value === "" ? undefined : Number(e.target.value);
      onParamsChange({ min_total_fee: val, page: 1 });
    },
    [onParamsChange]
  );

  const handleMaxFeeChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const val = e.target.value === "" ? undefined : Number(e.target.value);
      onParamsChange({ max_total_fee: val, page: 1 });
    },
    [onParamsChange]
  );

  return (
    <div className={`bg-white border ${C.borderLight} rounded-[4px] p-4 flex flex-wrap gap-4 items-end`}>
      {/* 飼い主名検索 */}
      <div className="flex flex-col gap-1 min-w-[200px]">
        <label className={`text-sm ${C.text65}`}>飼い主名</label>
        <div className="relative">
          <Search className={STYLE.searchIcon} />
          <Input
            className={`${STYLE.searchInput} pl-8`}
            placeholder="飼い主名を検索..."
            value={params.search ?? ""}
            onChange={handleSearchChange}
          />
        </div>
      </div>

      {/* CPMステージ */}
      <div className="flex flex-col gap-1 min-w-[160px]">
        <label className={`text-sm ${C.text65}`}>CPMステージ</label>
        <Select
          value={params.cpm_stage ?? "all"}
          onValueChange={handleCpmStageChange}
        >
          <SelectTrigger className={`h-11 ${C.borderMedium} ${C.text} bg-white`}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {CPM_STAGE_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Revenue tab filters */}
      {activeTab === "revenue" ? (
        <>
          <div className="flex flex-col gap-1 min-w-[120px]">
            <label className={`text-sm ${C.text65}`}>年度</label>
            <Input
              type="number"
              className={`h-11 ${C.borderMedium} ${C.text} bg-white`}
              placeholder="年度"
              value={params.year ?? ""}
              onChange={handleYearChange}
              min={2000}
            />
          </div>

          <div className="flex flex-col gap-1 min-w-[140px]">
            <label className={`text-sm ${C.text65}`}>売上基準</label>
            <Select
              value={params.amount_basis ?? "gross_total_amount"}
              onValueChange={handleAmountBasisChange}
            >
              <SelectTrigger className={`h-11 ${C.borderMedium} ${C.text} bg-white`}>
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
            <label className={`text-sm ${C.text65}`}>売上額</label>
            <div className="flex items-center gap-2">
              <Input
                type="number"
                className={`h-11 w-28 ${C.borderMedium} ${C.text} bg-white`}
                placeholder="下限"
                value={params.min_amount ?? ""}
                onChange={handleMinAmountChange}
                min={0}
              />
              <span className={`text-sm ${C.text50}`}>〜</span>
              <Input
                type="number"
                className={`h-11 w-28 ${C.borderMedium} ${C.text} bg-white`}
                placeholder="上限"
                value={params.max_amount ?? ""}
                onChange={handleMaxAmountChange}
                min={0}
              />
            </div>
          </div>

          <div className="flex items-center gap-2 pb-1">
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
            <label className={`text-sm ${C.text65}`}>期間</label>
            <Select
              value={params.period_preset ?? "last_12_months"}
              onValueChange={handlePeriodPresetChange}
            >
              <SelectTrigger className={`h-11 ${C.borderMedium} ${C.text} bg-white`}>
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
            <label className={`text-sm ${C.text65}`}>日付範囲</label>
            <div className="flex items-center gap-2">
              <Input
                type="date"
                className={`h-11 w-32 ${C.borderMedium} ${C.text} bg-white`}
                value={params.from ?? ""}
                onChange={handleFromChange}
              />
              <span className={`text-sm ${C.text50}`}>〜</span>
              <Input
                type="date"
                className={`h-11 w-32 ${C.borderMedium} ${C.text} bg-white`}
                value={params.to ?? ""}
                onChange={handleToChange}
              />
            </div>
          </div>

          <div className="flex flex-col gap-1">
            <label className={`text-sm ${C.text65}`}>来院回数</label>
            <div className="flex items-center gap-2">
              <Input
                type="number"
                className={`h-11 w-28 ${C.borderMedium} ${C.text} bg-white`}
                placeholder="下限"
                value={params.min_visit_count ?? ""}
                onChange={handleMinVisitCountChange}
                min={0}
              />
              <span className={`text-sm ${C.text50}`}>〜</span>
              <Input
                type="number"
                className={`h-11 w-28 ${C.borderMedium} ${C.text} bg-white`}
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
            <label className={`text-sm ${C.text65}`}>最終来院</label>
            <Select
              value={params.last_visit_bucket ?? "over_3m"}
              onValueChange={handleLastVisitBucketChange}
            >
              <SelectTrigger className={`h-11 ${C.borderMedium} ${C.text} bg-white`}>
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

          <div className="flex items-center gap-2 pb-1">
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

      {/* LTV tab filters */}
      {activeTab === "ltv" ? (
        <div className="flex flex-col gap-1">
          <label className={`text-sm ${C.text65}`}>累計診療費</label>
          <div className="flex items-center gap-2">
            <Input
              type="number"
              className={`h-11 w-28 ${C.borderMedium} ${C.text} bg-white`}
              placeholder="下限"
              value={params.min_total_fee ?? ""}
              onChange={handleMinFeeChange}
              min={0}
            />
            <span className={`text-sm ${C.text50}`}>〜</span>
            <Input
              type="number"
              className={`h-11 w-28 ${C.borderMedium} ${C.text} bg-white`}
              placeholder="上限"
              value={params.max_total_fee ?? ""}
              onChange={handleMaxFeeChange}
              min={0}
            />
          </div>
        </div>
      ) : null}

      {/* LINE連携済みのみ */}
      <div className="flex items-center gap-2 pb-1">
        <input
          type="checkbox"
          id="ltv-has-line"
          className="size-4 cursor-pointer"
          checked={params.has_line === true}
          onChange={handleHasLineChange}
        />
        <label htmlFor="ltv-has-line" className={`text-sm ${C.text} cursor-pointer select-none`}>
          LINE連携済みのみ
        </label>
      </div>
    </div>
  );
}
