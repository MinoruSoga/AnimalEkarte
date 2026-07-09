import { C } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { ResourceCheckups } from "@/types/generated/models";
import { useGetPetCheckupResults, type PetCheckupResult } from "@/hooks/use-pet-checkup-results";
import { ReportSection } from "./ReportSection";

interface CheckupHistorySectionProps {
  petId: string;
}

interface CheckupGroup {
  checkupId: number;
  date: string;
  checkupTypeName: string;
  results: PetCheckupResult[];
}

// checkup 単位に結果値をまとめる（BE は date DESC で返すため初出順を維持する）。
function groupByCheckup(results: PetCheckupResult[]): CheckupGroup[] {
  const order: number[] = [];
  const map = new Map<number, CheckupGroup>();
  for (const r of results) {
    let group = map.get(r.checkupId);
    if (!group) {
      group = { checkupId: r.checkupId, date: r.date, checkupTypeName: r.checkupTypeName, results: [] };
      map.set(r.checkupId, group);
      order.push(r.checkupId);
    }
    group.results.push(r);
  }
  return order.map((id) => map.get(id)!);
}

// field_type に応じた表示値を組み立てる。
function displayValue(r: PetCheckupResult): string {
  switch (r.fieldType) {
    case "boolean":
      return r.valueBool ? "はい" : "いいえ";
    case "number":
      if (r.valueNumber === undefined) return "-";
      return r.unit ? `${r.valueNumber} ${r.unit}` : String(r.valueNumber);
    case "multi_select":
    case "checklist":
      return r.valueList.length > 0 ? r.valueList.join("、") : "-";
    default:
      return r.valueText || "-";
  }
}

/**
 * #211 健診（パッケージ）履歴。pet_id 単位取得・checkup 単位グルーピング。
 * 歯科検診など型付きフィールドを持つ健診結果を読み取り表示する。
 */
export function CheckupHistorySection({ petId }: CheckupHistorySectionProps) {
  const { canView } = usePermission(ResourceCheckups);
  const { data, isLoading, isError } = useGetPetCheckupResults(canView ? petId : undefined);
  const groups = groupByCheckup(data ?? []);

  return (
    <ReportSection
      title="健診（パッケージ）履歴"
      canView={canView}
      isLoading={isLoading}
      isError={isError}
      isEmpty={groups.length === 0}
      count={groups.length}
      emptyMessage="健診結果の履歴はありません"
    >
      <ul className="flex flex-col gap-3">
        {groups.map((group) => (
          <li key={group.checkupId} className={`rounded border ${C.borderLight} p-2.5`}>
            <div className="mb-1.5 flex flex-wrap items-center gap-x-3 gap-y-1">
              <span className={`text-sm font-medium ${C.text}`}>
                {group.checkupTypeName || "健診"}
              </span>
              <span className={`font-mono text-xs ${C.text50}`}>{group.date || "-"}</span>
            </div>
            <table className="w-full text-sm">
              <thead>
                <tr className={`text-left text-xs ${C.text50}`}>
                  <th scope="col" className="py-0.5 pr-3 font-medium">
                    項目
                  </th>
                  <th scope="col" className="py-0.5 font-medium whitespace-nowrap">
                    結果
                  </th>
                </tr>
              </thead>
              <tbody>
                {group.results.map((r) => (
                  <tr key={r.id}>
                    <td className={`py-0.5 pr-3 ${C.text}`}>{r.fieldName}</td>
                    <td className={`py-0.5 ${r.isAbnormal ? C.danger : C.text}`}>
                      {r.isAbnormal ? (
                        <span aria-label="要注意" className="mr-1">
                          ⚠
                        </span>
                      ) : null}
                      <span>{displayValue(r)}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </li>
        ))}
      </ul>
    </ReportSection>
  );
}
