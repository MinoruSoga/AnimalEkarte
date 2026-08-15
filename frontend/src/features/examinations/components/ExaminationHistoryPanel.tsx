import { HistoryFilterPanel } from "@/components/shared/HistoryFilterPanel";
import { Button } from "@/components/ui/button";
import { C } from "@/lib/design-tokens";
import type { SortOrder } from "@/types";
import type { ExaminationRecord } from "../api/transforms";
import { ExamPivotTable } from "./ExamPivotTable";
import { ExaminationCard } from "./ExaminationCard";

interface ExaminationHistoryPanelProps {
  filteredHistory: ExaminationRecord[];
  pivotHistory: ExaminationRecord[];
  currentPetId: string | undefined;
  historyStartDate: string;
  historyEndDate: string;
  historySearchTerm: string;
  historySortOrder: SortOrder;
  historyView: "cards" | "pivot";
  onHistoryStartDateChange: (value: string) => void;
  onHistoryEndDateChange: (value: string) => void;
  onHistorySearchTermChange: (value: string) => void;
  onHistorySortOrderChange: (value: SortOrder) => void;
  onHistoryViewChange: (value: "cards" | "pivot") => void;
  onHistoryClear: () => void;
}

export function ExaminationHistoryPanel({
  filteredHistory,
  pivotHistory,
  currentPetId,
  historyStartDate,
  historyEndDate,
  historySearchTerm,
  historySortOrder,
  historyView,
  onHistoryStartDateChange,
  onHistoryEndDateChange,
  onHistorySearchTermChange,
  onHistorySortOrderChange,
  onHistoryViewChange,
  onHistoryClear,
}: ExaminationHistoryPanelProps) {
  return (
    <div className="lg:col-span-2 space-y-3">
      <div className="flex items-center justify-between gap-2 px-1">
        <h3 className={`text-sm font-medium ${C.text60}`}>過去の検査履歴</h3>
        <div
          className="flex items-center gap-1"
          role="group"
          aria-label="履歴表示形式"
        >
          <Button
            type="button"
            variant={historyView === "cards" ? "secondary" : "ghost"}
            size="sm"
            aria-pressed={historyView === "cards"}
            onClick={() => onHistoryViewChange("cards")}
          >
            カード
          </Button>
          <Button
            type="button"
            variant={historyView === "pivot" ? "secondary" : "ghost"}
            size="sm"
            aria-pressed={historyView === "pivot"}
            onClick={() => onHistoryViewChange("pivot")}
          >
            時系列
          </Button>
        </div>
      </div>
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
      {historyView === "pivot" ? (
        currentPetId ? (
          <ExamPivotTable
            examinations={pivotHistory}
            sortOrder={historySortOrder}
          />
        ) : (
          <p className={`py-6 text-center text-sm ${C.text45}`}>
            ペットを選択してください
          </p>
        )
      ) : (
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
      )}
    </div>
  );
}
