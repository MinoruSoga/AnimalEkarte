import { C } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { ResourceTrimming } from "@/types/generated/models";
import { ReportSection } from "./ReportSection";
import { HistoryTable } from "./HistoryTable";
import { useGetPetTrimmingHistory } from "../api/get-pet-trimming-history";

interface TrimmingHistorySectionProps {
  petId: string;
}

/**
 * #158 ⑦ トリミング履歴 = 施術「実施履歴」。GET /v1/trimmings?pet_id（appointments ベース）由来。
 * 取得フックが status="完了"（施術実施済み）のみを実施日降順で返す（予約・進行中・キャンセルは除外）。
 * コース・担当を併記する。値は API 由来のみで、未設定セルは捏造せず "-" を出す。
 */
export function TrimmingHistorySection({ petId }: TrimmingHistorySectionProps) {
  const { canView } = usePermission(ResourceTrimming);
  const { data, isLoading, isError } = useGetPetTrimmingHistory(canView ? petId : undefined);
  const items = data ?? [];

  return (
    <ReportSection
      title="トリミング履歴"
      canView={canView}
      isLoading={isLoading}
      isError={isError}
      isEmpty={items.length === 0}
      count={items.length}
      emptyMessage="トリミングの履歴はありません"
    >
      <HistoryTable headers={["実施日", "状態", "コース", "担当"]}>
        {items.map((t) => (
          <tr key={t.id}>
            <td className={`py-1.5 pr-3 font-mono whitespace-nowrap ${C.text}`}>{t.date || "-"}</td>
            <td className={`py-1.5 pr-3 whitespace-nowrap ${C.text}`}>{t.status}</td>
            <td className={`py-1.5 pr-3 ${C.text}`}>{t.courseName || "-"}</td>
            <td className={`py-1.5 pr-3 whitespace-nowrap ${C.text50}`}>{t.staff || "-"}</td>
          </tr>
        ))}
      </HistoryTable>
    </ReportSection>
  );
}
