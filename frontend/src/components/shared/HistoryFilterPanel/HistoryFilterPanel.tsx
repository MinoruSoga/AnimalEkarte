import { memo } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DatePicker } from "@/components/shared/DatePicker";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { SortOrder } from "@/types";
import { SORT_ORDER_VALUES } from "@/types";
import { getSortOrderLabel } from "@/utils/status-helpers";
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
  return (
    <div className={`space-y-3 bg-white p-3 rounded-lg border ${C.borderMedium} shadow-sm`}>
      {showDateRange ? (
        <div className="flex flex-col gap-1.5">
          <Label className={`text-sm ${C.text60}`}>実施日</Label>
          <div className="flex items-center gap-2">
            <DatePicker
              value={filterStartDate}
              onChange={(val) => onFilterStartDateChange?.(val)}
              className="flex-1"
              placeholder="開始日"
            />
            <span className={`${C.text} text-sm shrink-0`}>〜</span>
            <DatePicker
              value={filterEndDate}
              onChange={(val) => onFilterEndDateChange?.(val)}
              className="flex-1"
              placeholder="終了日"
            />
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
            className={`flex-1 bg-white ${C.borderMedium} h-10 text-sm ${C.text}`}
            placeholder={searchPlaceholder}
          />
          <Button
            variant="outline"
            className={`h-10 text-sm ${C.text60} ${C.hoverText} ${C.hoverBgPage} ${C.borderMedium}`}
            onClick={onClear}
          >
            クリア
          </Button>
          <Select value={sortOrder} onValueChange={(val) => {
            if (isOneOf(val, SORT_ORDER_VALUES)) {
              onSortOrderChange(val);
            }
          }}>
            <SelectTrigger aria-label="並び順" className={`w-[80px] h-10 text-sm bg-white ${C.borderMedium} ${C.text}`}>
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
