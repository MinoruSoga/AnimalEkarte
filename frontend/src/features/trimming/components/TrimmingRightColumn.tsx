import { memo, useMemo } from "react";

import { Label } from "@/components/ui/label";
import { HistoryFilterPanel } from "@/components/shared/HistoryFilterPanel";
import { LoadingFallback } from "@/components/shared/DataStates";
import { C } from "@/lib/design-tokens";

import type { TrimmingRightColumnProps } from "./trimming-form-column-types";

export const TrimmingRightColumn = memo(function TrimmingRightColumn({
  sortedHistory,
  isHistoryLoading,
  historySearchTerm,
  historySortOrder,
  historyDateRange,
  onSearchTermChange,
  onSortOrderChange,
  onClear,
  onFilterStartDateChange,
  onFilterEndDateChange,
  onHistoryClick,
}: TrimmingRightColumnProps) {
  const historyCards = useMemo(
    () =>
      sortedHistory.map((history) => (
        <button
          key={history.id}
          type="button"
          className={`block w-full text-left p-3 border ${C.borderMedium} rounded-lg ${C.bgWhite} ${C.hoverBgPage} transition-colors cursor-pointer`}
          onClick={() =>
            onHistoryClick({
              courseId: history.courseId,
              optionIds: history.optionIds,
              styleRequest: history.styleRequest,
              usedShampoo: history.usedShampoo,
              usedRibbon: history.usedRibbon,
              remarks: history.remarks,
              staffName: history.staff,
            })
          }
        >
          <div className="flex items-start justify-between">
            <div className="flex-1 min-w-0">
              <div className={`text-xs ${C.text60} mb-1`}>{history.date}</div>
              <div className={`text-sm ${C.text} font-medium truncate`}>{history.styleRequest}</div>
              <div className={`text-xs ${C.text60} mt-1`}>{history.staff}</div>
            </div>
          </div>
        </button>
      )),
    [sortedHistory, onHistoryClick]
  );

  return (
    <div className={`${C.bgWhite} rounded-lg border ${C.borderMedium} p-3 space-y-4`}>
      <div>
        <Label className={`text-sm ${C.text60} mb-2 block`}>施術履歴</Label>
        <HistoryFilterPanel
          searchTerm={historySearchTerm}
          onSearchTermChange={onSearchTermChange}
          sortOrder={historySortOrder}
          onSortOrderChange={onSortOrderChange}
          onClear={onClear}
          showDateRange={true}
          filterStartDate={historyDateRange.from}
          onFilterStartDateChange={onFilterStartDateChange}
          filterEndDate={historyDateRange.to}
          onFilterEndDateChange={onFilterEndDateChange}
          searchPlaceholder="スタイル希望で検索..."
        />
      </div>

      <div className="space-y-2 max-h-[600px] overflow-y-auto">
        {isHistoryLoading ? (
          <LoadingFallback />
        ) : sortedHistory.length === 0 ? (
          <div className={`text-center py-8 text-sm ${C.text40}`}>
            施術履歴がありません
          </div>
        ) : (
          historyCards
        )}
      </div>
    </div>
  );
});
