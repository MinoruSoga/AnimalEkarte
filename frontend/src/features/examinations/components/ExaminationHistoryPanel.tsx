import { HistoryFilterPanel } from "@/components/shared/HistoryFilterPanel";
import { C } from "@/lib/design-tokens";
import type { SortOrder } from "@/types";
import type { ExaminationRecord } from "../api/transforms";
import { ExaminationCard } from "./ExaminationCard";

interface ExaminationHistoryPanelProps {
  filteredHistory: ExaminationRecord[];
  currentPetId: string | undefined;
  historyStartDate: string;
  historyEndDate: string;
  historySearchTerm: string;
  historySortOrder: SortOrder;
  onHistoryStartDateChange: (value: string) => void;
  onHistoryEndDateChange: (value: string) => void;
  onHistorySearchTermChange: (value: string) => void;
  onHistorySortOrderChange: (value: SortOrder) => void;
  onHistoryClear: () => void;
}

export function ExaminationHistoryPanel({
  filteredHistory,
  currentPetId,
  historyStartDate,
  historyEndDate,
  historySearchTerm,
  historySortOrder,
  onHistoryStartDateChange,
  onHistoryEndDateChange,
  onHistorySearchTermChange,
  onHistorySortOrderChange,
  onHistoryClear,
}: ExaminationHistoryPanelProps) {
  return (
    <div className="lg:col-span-2 space-y-3">
      <h3 className={`text-sm font-medium ${C.text60} px-1`}>過去の検査履歴</h3>
      <HistoryFilterPanel
        showDateRange={true}
        filterStartDate={historyStartDate}
        onFilterStartDateChange={onHistoryStartDateChange}
        filterEndDate={historyEndDate}
        onFilterEndDateChange={onHistoryEndDateChange}
        searchTerm={historySearchTerm}
        onSearchTermChange={onHistorySearchTermChange}
        searchPlaceholder="検査種別・所見で検索..."
        sortOrder={historySortOrder}
        onSortOrderChange={onHistorySortOrderChange}
        onClear={onHistoryClear}
      />
      <div className="space-y-2 max-h-[600px] overflow-y-auto">
        {filteredHistory.length > 0 ? (
          filteredHistory.map((examination) => (
            <ExaminationCard key={examination.id} examination={examination} />
          ))
        ) : (
          <p className={`text-sm ${C.text45} text-center py-6`}>
            {currentPetId ? "検査記録がありません" : "ペットを選択してください"}
          </p>
        )}
      </div>
    </div>
  );
}
