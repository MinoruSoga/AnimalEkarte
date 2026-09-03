import { useCallback, useDeferredValue, useMemo, useState } from "react";
import { normalizeKana } from "@/lib/normalize-kana";
import type { SortOrder } from "@/types";
import { useGetExaminations } from "../api/get-examinations";

interface UseExaminationHistoryFiltersInput {
  currentPetId: string | undefined;
  isEdit: boolean;
  excludeId: string | undefined;
}

// FE-RC-045/046: ExaminationForm.tsx から履歴検索/絞り込み関連の state・派生値を分離。
// 表示（ExaminationHistoryPanel）とは独立して振る舞いを検証できるようにする。
export function useExaminationHistoryFilters({
  currentPetId,
  isEdit,
  excludeId,
}: UseExaminationHistoryFiltersInput) {
  const [historySearchTerm, setHistorySearchTerm] = useState("");
  const [historySortOrder, setHistorySortOrder] = useState<SortOrder>("desc");
  const [historyStartDate, setHistoryStartDate] = useState("");
  const [historyEndDate, setHistoryEndDate] = useState("");

  const handleHistoryClear = useCallback(() => {
    setHistorySearchTerm("");
    setHistorySortOrder("desc");
    setHistoryStartDate("");
    setHistoryEndDate("");
  }, []);

  const { data: allExaminations = [] } = useGetExaminations({
    petId: currentPetId,
    startDate: historyStartDate || undefined,
    endDate: historyEndDate || undefined,
  });

  const deferredHistorySearch = useDeferredValue(historySearchTerm);

  const petHistory = useMemo(() => {
    if (!currentPetId) return [];
    return allExaminations.filter((e) => e.petId === currentPetId);
  }, [allExaminations, currentPetId]);

  const searchedPetHistory = useMemo(() => {
    if (!deferredHistorySearch) return petHistory;

    const searchValue = normalizeKana(deferredHistorySearch).toLowerCase();
    return petHistory.filter(
      (examination) =>
        normalizeKana(examination.testType)
          .toLowerCase()
          .includes(searchValue) ||
        normalizeKana(examination.resultSummary ?? "")
          .toLowerCase()
          .includes(searchValue),
    );
  }, [deferredHistorySearch, petHistory]);

  // js-cache-function-results: カード履歴フィルタ結果をメモ化
  const filteredHistory = useMemo(() => {
    let result = searchedPetHistory;
    // 編集中の記録自体は除外
    if (isEdit && excludeId) {
      result = result.filter((e) => e.id !== excludeId);
    }
    return [...result].sort((a, b) => {
      const cmp = a.date.localeCompare(b.date);
      return historySortOrder === "asc" ? cmp : -cmp;
    });
  }, [searchedPetHistory, isEdit, excludeId, historySortOrder]);

  return {
    historySearchTerm,
    setHistorySearchTerm,
    historySortOrder,
    setHistorySortOrder,
    historyStartDate,
    setHistoryStartDate,
    historyEndDate,
    setHistoryEndDate,
    handleHistoryClear,
    searchedPetHistory,
    filteredHistory,
  };
}
