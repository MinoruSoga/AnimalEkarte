import React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { SortOrder } from "@/types";
import { SORT_ORDER_VALUES } from "@/types";
import { getSortOrderLabel } from "@/utils/status-helpers";
import { isOneOf } from "@/lib/type-utils";

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

export const HistoryFilterPanel = React.memo(function HistoryFilterPanel({
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
    <div className="space-y-3 bg-white p-3 rounded-lg border border-[rgba(55,53,47,0.16)] shadow-sm">
      {showDateRange && (
        <div className="flex flex-col gap-1.5">
          <Label className="text-sm text-[#37352F]/60">実施日</Label>
          <div className="flex items-center gap-2">
            <NotionDatePicker
              value={filterStartDate}
              onChange={(val) => onFilterStartDateChange?.(val)}
              className="flex-1"
              placeholder="開始日"
            />
            <span className="text-[#37352F] text-sm shrink-0">〜</span>
            <NotionDatePicker
              value={filterEndDate}
              onChange={(val) => onFilterEndDateChange?.(val)}
              className="flex-1"
              placeholder="終了日"
            />
          </div>
        </div>
      )}
      <div className="flex flex-col gap-1.5">
        <Label className="text-sm text-[#37352F]/60">検索単語</Label>
        <div className="flex gap-2">
          <Input
            value={searchTerm}
            onChange={(e) => onSearchTermChange(e.target.value)}
            className="flex-1 bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm text-[#37352F]"
            placeholder={searchPlaceholder}
          />
          <Button
            variant="outline"
            className="h-10 text-sm text-[#37352F]/60 hover:text-[#37352F] hover:bg-[#F7F6F3] border-[rgba(55,53,47,0.16)]"
            onClick={onClear}
          >
            クリア
          </Button>
          <Select value={sortOrder} onValueChange={(val) => {
            if (isOneOf(val, SORT_ORDER_VALUES)) {
              onSortOrderChange(val);
            }
          }}>
            <SelectTrigger className="w-[80px] h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] text-[#37352F]">
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
