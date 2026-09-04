import { useDeferredValue, useCallback, useMemo, useState } from "react";
import { normalizeKana } from "@/lib/normalize-kana";

import type { SortOrder } from "@/types";
import type { TrimmingFormData } from "@/types/trimming";

import { useGetTrimmingsByPetId } from "../api/get-trimming";

export function useTrimmingHistory(petId: string) {
  const [historySearchTerm, setHistorySearchTerm] = useState("");
  const [historySortOrder, setHistorySortOrder] = useState<SortOrder>("desc");
  const [historyDateRangeFrom, setHistoryDateRangeFrom] = useState("");
  const [historyDateRangeTo, setHistoryDateRangeTo] = useState("");
  const deferredHistorySearch = useDeferredValue(historySearchTerm);

  const { data: petTrimmings = [], isLoading: isHistoryLoading } = useGetTrimmingsByPetId(petId);

  const sortedHistory = useMemo(() => {
    const filtered = petTrimmings.filter((trimming) => {
      if (
        deferredHistorySearch &&
        !normalizeKana(trimming.styleRequest)
          .toLowerCase()
          .includes(normalizeKana(deferredHistorySearch).toLowerCase())
      ) {
        return false;
      }
      const recordDate = trimming.date.slice(0, 10);
      if (historyDateRangeFrom && recordDate < historyDateRangeFrom) return false;
      if (historyDateRangeTo && recordDate > historyDateRangeTo) return false;
      return true;
    });
    return filtered.toSorted((a, b) => {
      const dateA = new Date(a.date).getTime();
      const dateB = new Date(b.date).getTime();
      return historySortOrder === "desc" ? dateB - dateA : dateA - dateB;
    });
  }, [
    petTrimmings,
    deferredHistorySearch,
    historyDateRangeFrom,
    historyDateRangeTo,
    historySortOrder,
  ]);

  const handleHistoryClear = useCallback(() => {
    setHistorySearchTerm("");
    setHistorySortOrder("desc");
    setHistoryDateRangeFrom("");
    setHistoryDateRangeTo("");
  }, []);

  const handleHistoryClick = useCallback((updates: Partial<TrimmingFormData>) => updates, []);
  const historyDateRange = useMemo(
    () => ({ from: historyDateRangeFrom, to: historyDateRangeTo }),
    [historyDateRangeFrom, historyDateRangeTo],
  );

  return {
    handleHistoryClear,
    handleHistoryClick,
    historyDateRange,
    historySearchTerm,
    historySortOrder,
    isHistoryLoading,
    setHistoryDateRangeFrom,
    setHistoryDateRangeTo,
    setHistorySearchTerm,
    setHistorySortOrder,
    sortedHistory,
  };
}
