import { useCallback } from "react";

import type { TrimmingFormData } from "@/types/trimming";

// NO32: useGetTrimmingsByPetId は cross-feature 共有のため @/hooks へ昇格済み
import { useGetTrimmingsByPetId } from "@/hooks/use-pet-trimmings";

// PR679 F7: 検索/ソート/日付フィルタは共有 PastRecordHistoryPanel 内蔵になったため
// フック側のフィルタ state・ハンドラは削除済み（データ取得とクリック受け渡しのみ残す）。
export function useTrimmingHistory(petId: string) {
  const { data: petTrimmings = [], isLoading: isHistoryLoading } = useGetTrimmingsByPetId(petId);

  const handleHistoryClick = useCallback((updates: Partial<TrimmingFormData>) => updates, []);

  // sortedHistory という公開名は呼び出し側の churn を抑えるため維持する
  // （実際のソート/フィルタは PastRecordHistoryPanel 内で行われる）。
  return {
    handleHistoryClick,
    isHistoryLoading,
    sortedHistory: petTrimmings,
  };
}
