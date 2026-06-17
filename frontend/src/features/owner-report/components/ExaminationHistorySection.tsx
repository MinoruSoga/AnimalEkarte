import { C } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { ResourceExaminations } from "@/types/generated/models";
import { ReportSection } from "./ReportSection";
import { useGetPetExaminations } from "../api/get-pet-examinations";

interface ExaminationHistorySectionProps {
  petId: string;
}

/**
 * #158 ③ 健康診断（検査）履歴。Examination-first。
 * pet_id 単位取得・下書き除外・日付降順（バックエンド exams.date DESC）。
 */
export function ExaminationHistorySection({ petId }: ExaminationHistorySectionProps) {
  const { canView } = usePermission(ResourceExaminations);
  const { data, isLoading, isError } = useGetPetExaminations(canView ? petId : undefined);
  const exams = data ?? [];

  return (
    <ReportSection
      title="健康診断（検査）履歴"
      canView={canView}
      isLoading={isLoading}
      isError={isError}
      isEmpty={exams.length === 0}
      emptyMessage="検査の履歴はありません"
    >
      <ul className="flex flex-col gap-3">
        {exams.map((exam) => (
          <li key={exam.id} className={`rounded border ${C.borderLight} p-2.5`}>
            <div className="flex flex-wrap items-center gap-x-3 gap-y-1 mb-1.5">
              <span className={`text-sm font-medium ${C.text}`}>{exam.testType || "検査"}</span>
              <span className={`text-xs font-mono ${C.text50}`}>{exam.date || "-"}</span>
              <span className={`text-xs ${C.text50}`}>{exam.status}</span>
            </div>
            {exam.items && exam.items.length > 0 ? (
              <table className="w-full text-sm">
                <tbody>
                  {exam.items.map((item) => (
                    <tr key={item.id}>
                      <td className={`py-0.5 pr-3 ${C.text}`}>{item.name}</td>
                      <td
                        className={`py-0.5 pr-3 font-mono whitespace-nowrap ${
                          item.status !== "normal" ? C.danger : C.text
                        }`}
                      >
                        {item.inspectionValue || item.result || "-"}
                        {item.unit ? ` ${item.unit}` : ""}
                      </td>
                      <td className={`py-0.5 font-mono whitespace-nowrap ${C.text50}`}>
                        {item.referenceValue || "-"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <p className={`text-xs ${C.text50}`}>
                {exam.resultSummary || "検査項目なし"}
              </p>
            )}
          </li>
        ))}
      </ul>
    </ReportSection>
  );
}
