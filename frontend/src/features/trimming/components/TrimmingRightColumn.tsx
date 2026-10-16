import { memo, useMemo } from "react";

import { PastRecordHistoryPanel } from "@/components/shared/PastRecordHistoryPanel/PastRecordHistoryPanel";
import { LoadingFallback } from "@/components/shared/DataStates";

import type { TrimmingRightColumnProps } from "../lib/trimming-form-column-types";

/**
 * EMR-412 T2: トリミング編集の履歴は共有 PastRecordHistoryPanel へ統一。
 * 日付はパネル内 formatDate、見出し「過去の施術履歴」もパネル側へ移管。
 * クリック時は従来どおりコース・オプション等をフォームへ写す。
 * PR679 F7: 外部フィルタ props はパネル内フィルタに完全統一のため削除済み。
 */
export const TrimmingRightColumn = memo(function TrimmingRightColumn({
  sortedHistory,
  isHistoryLoading,
  onHistoryClick,
}: TrimmingRightColumnProps) {
  const itemsById = useMemo(
    () => new Map(sortedHistory.map((history) => [history.id, history])),
    [sortedHistory],
  );
  const historyItems = useMemo(
    () =>
      sortedHistory.map((history) => ({
        id: history.id,
        date: history.date,
        title: history.styleRequest,
        subtitle: history.staff,
      })),
    [sortedHistory],
  );

  return (
    <div className="lg:col-span-2">
      {isHistoryLoading ? (
        <LoadingFallback />
      ) : (
        <PastRecordHistoryPanel
          title="過去の施術履歴"
          searchPlaceholder="スタイル希望で検索..."
          items={historyItems}
          onItemClick={(item) => {
            const history = itemsById.get(item.id);
            if (!history) return;
            onHistoryClick({
              courseId: history.courseId,
              optionIds: history.optionIds,
              styleRequest: history.styleRequest,
              usedShampoo: history.usedShampoo,
              usedRibbon: history.usedRibbon,
              remarks: history.remarks,
              staffName: history.staff,
            });
          }}
        />
      )}
    </div>
  );
});
