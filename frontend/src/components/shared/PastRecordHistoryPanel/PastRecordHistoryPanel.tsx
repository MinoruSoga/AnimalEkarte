import { memo, useDeferredValue, useMemo, useState } from "react";

import { HistoryFilterPanel } from "@/components/shared/HistoryFilterPanel";
import { C } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import { normalizeKana } from "@/lib/normalize-kana";
import type { SortOrder } from "@/types";

export interface PastRecordHistoryItem {
  id: string;
  date: string;
  title: string;
  subtitle?: string;
}

interface PastRecordHistoryPanelProps {
  title: string;
  searchPlaceholder: string;
  items: PastRecordHistoryItem[];
  isLoading?: boolean;
}

export const PastRecordHistoryPanel = memo(function PastRecordHistoryPanel({
  title,
  searchPlaceholder,
  items,
  isLoading = false,
}: PastRecordHistoryPanelProps) {
  const [filterStartDate, setFilterStartDate] = useState("");
  const [filterEndDate, setFilterEndDate] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState<SortOrder>("desc");
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    const needle = normalizeKana(deferredSearch).toLowerCase();
    const next = items.filter((item) => {
      const recordDate = item.date.slice(0, 10);
      if (filterStartDate && recordDate < filterStartDate) return false;
      if (filterEndDate && recordDate > filterEndDate) return false;
      if (!needle) return true;
      const haystack = normalizeKana(`${item.title} ${item.subtitle ?? ""}`).toLowerCase();
      return haystack.includes(needle);
    });
    return next.toSorted((a, b) => {
      const cmp = a.date.localeCompare(b.date);
      return sortOrder === "desc" ? -cmp : cmp;
    });
  }, [deferredSearch, filterEndDate, filterStartDate, items, sortOrder]);

  return (
    <div className="flex flex-col gap-3 lg:col-span-2">
      <h3 className={`text-base font-semibold ${C.text}`}>{title}</h3>
      <HistoryFilterPanel
        showDateRange
        filterStartDate={filterStartDate}
        onFilterStartDateChange={setFilterStartDate}
        filterEndDate={filterEndDate}
        onFilterEndDateChange={setFilterEndDate}
        searchTerm={searchTerm}
        onSearchTermChange={setSearchTerm}
        searchPlaceholder={searchPlaceholder}
        sortOrder={sortOrder}
        onSortOrderChange={setSortOrder}
        onClear={() => {
          setFilterStartDate("");
          setFilterEndDate("");
          setSearchTerm("");
          setSortOrder("desc");
        }}
      />
      <div className="flex max-h-[600px] flex-col gap-2 overflow-y-auto">
        {isLoading ? (
          <p className={`py-4 text-center text-sm ${C.text45}`}>読み込み中...</p>
        ) : filteredItems.length === 0 ? (
          <p className={`py-4 text-center text-sm ${C.text45}`}>履歴がありません</p>
        ) : (
          filteredItems.map((item) => (
            <div
              key={item.id}
              className={`rounded-lg border ${C.borderLight} ${C.bgWhite} px-3 py-2`}
            >
              <div className={`text-sm font-medium ${C.text}`}>{item.title}</div>
              <div className={`text-xs ${C.text60}`}>{formatDate(item.date)}</div>
              {item.subtitle ? (
                <div className={`mt-0.5 truncate text-xs ${C.text50}`}>{item.subtitle}</div>
              ) : null}
            </div>
          ))
        )}
      </div>
    </div>
  );
});
