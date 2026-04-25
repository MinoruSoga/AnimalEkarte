import { useCallback } from "react";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { C, STYLE } from "@/lib/design-tokens";
import type { LtvOwnersParams } from "../api/get-ltv-owners";

interface LtvFilterPanelProps {
  params: LtvOwnersParams;
  onParamsChange: (params: Partial<LtvOwnersParams>) => void;
}

const CPM_STAGE_OPTIONS = [
  { value: "all", label: "すべて" },
  { value: "core", label: "コア" },
  { value: "growing", label: "グロウィング" },
  { value: "new", label: "新規" },
  { value: "dormant", label: "休眠" },
  { value: "lost", label: "離脱" },
] as const;

export function LtvFilterPanel({ params, onParamsChange }: LtvFilterPanelProps) {
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

  const handleSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onParamsChange({ search: e.target.value || undefined, page: 1 });
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

      {/* 累計診療費 */}
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
