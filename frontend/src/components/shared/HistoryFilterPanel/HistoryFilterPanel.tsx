import { memo, useId } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DatePicker } from "@/components/shared/DatePicker";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { SortOrder } from "@/types";
import { SORT_ORDER_VALUES } from "@/types";
import { getSortOrderLabel } from "@/lib/status-helpers";
import { isOneOf } from "@/lib/type-utils";
import { C } from "@/lib/design-tokens";

interface HistoryFilterPanelProps {
  /** 日付範囲フィルターを表示するか */
  showDateRange?: boolean;
  filterStartDate?: string;
  onFilterStartDateChange?: (value: string) => void;
  filterEndDate?: string;
  onFilterEndDateChange?: (value: string) => void;
  /** 検索 */
  searchTerm: string;
  onSearchTermChange: (value: string) => void;
  searchPlaceholder?: string;
  /** ソート */
  sortOrder: SortOrder;
  onSortOrderChange: (value: SortOrder) => void;
  /** クリアボタン */
  onClear: () => void;
}

export const HistoryFilterPanel = memo(function HistoryFilterPanel({
  showDateRange = true,
  filterStartDate = "",
  onFilterStartDateChange,
  filterEndDate = "",
  onFilterEndDateChange,
  searchTerm,
  onSearchTermChange,
  searchPlaceholder = "検索...",
  sortOrder,
  onSortOrderChange,
  onClear,
}: HistoryFilterPanelProps) {
  const dateRangeId = useId();
  const startDateId = `${dateRangeId}-start-date`;
  const endDateId = `${dateRangeId}-end-date`;

  return (
    <div className={`space-y-3 bg-white p-3 rounded-lg border ${C.borderMedium}`}>
      {showDateRange ? (
        <div className="flex flex-col gap-1.5">
          <span className={`mb-2 flex items-center gap-2 text-sm leading-none select-none ${C.text60}`}>
            実施日
          </span>
          <div className="flex items-center gap-2">
            <div className="flex-1 min-w-0">
              <Label htmlFor={startDateId} className="sr-only">開始日</Label>
              <DatePicker
                id={startDateId}
                name="historyStartDate"
                value={filterStartDate}
                onChange={(val) => onFilterStartDateChange?.(val)}
                placeholder="開始日"
              />
            </div>
            <span className={`${C.text} text-sm shrink-0`}>〜</span>
            <div className="flex-1 min-w-0">
              <Label htmlFor={endDateId} className="sr-only">終了日</Label>
              <DatePicker
                id={endDateId}
                name="historyEndDate"
                value={filterEndDate}
                onChange={(val) => onFilterEndDateChange?.(val)}
                placeholder="終了日"
              />
            </div>
          </div>
        </div>
      ) : null}
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="history-filter-search" className={`text-sm ${C.text60}`}>検索単語</Label>
        <div className="flex gap-2">
          <Input
            id="history-filter-search"
            value={searchTerm}
            onChange={(e) => onSearchTermChange(e.target.value)}
            className={`flex-1 bg-white ${C.borderMedium} h-11 text-sm ${C.text}`}
            placeholder={searchPlaceholder}
          />
          <Button
            variant="outline"
            className={`h-11 text-sm ${C.text60} ${C.hoverText} ${C.hoverBgPage} ${C.borderMedium}`}
            onClick={onClear}
          >
            クリア
          </Button>
          <Select value={sortOrder} onValueChange={(val) => {
            if (isOneOf(val, SORT_ORDER_VALUES)) {
              onSortOrderChange(val);
            }
          }}>
            <SelectTrigger aria-label="並び順" className={`w-[80px] h-11 text-sm bg-white ${C.borderMedium} ${C.text}`}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {SORT_ORDER_VALUES.map(order => (
                <SelectItem key={order} value={order}>{getSortOrderLabel(order)}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
    </div>
  );
});
