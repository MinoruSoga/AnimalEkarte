import { C } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { ResourceMedicalRecords } from "@/types/generated/models";
import { ReportSection } from "./ReportSection";
import { HistoryTable } from "./HistoryTable";
import {
  useGetPetTreatmentHistory,
  type TreatmentHistoryFilter,
} from "../api/get-pet-treatment-history";

interface TreatmentHistorySectionProps {
  petId: string;
  title: string;
  filter: TreatmentHistoryFilter;
  /** 麻酔列を出す（手術・処置セクション用）。 */
  showAnesthesia?: boolean;
  emptyMessage?: string;
}

/**
 * #158 ④投薬 / ⑤手術・処置 / ⑥治療 履歴。
 * いずれも treatments 由来（prescriptions は参照しない）。日付は medical_records.date 由来で降順。
 */
export function TreatmentHistorySection({
  petId,
  title,
  filter,
  showAnesthesia = false,
  emptyMessage,
}: TreatmentHistorySectionProps) {
  const { canView } = usePermission(ResourceMedicalRecords);
  const { data, isLoading, isError } = useGetPetTreatmentHistory(
    canView ? petId : undefined,
    filter,
  );
  const items = data ?? [];

  const headers = showAnesthesia
    ? ["診療日", "内容", "麻酔"]
    : filter === "medicine"
      ? ["診療日", "薬剤", "投与経路", "数量"]
      : ["診療日", "内容", "数量"];

  return (
    <ReportSection
      title={title}
      canView={canView}
      isLoading={isLoading}
      isError={isError}
      isEmpty={items.length === 0}
      count={items.length}
      emptyMessage={emptyMessage ?? "履歴はありません"}
    >
      <HistoryTable headers={headers}>
        {items.map((item) => (
          <tr key={item.id}>
            <td className={`py-1.5 pr-3 font-mono whitespace-nowrap ${C.text}`}>{item.date}</td>
            <td className={`py-1.5 pr-3 ${C.text}`}>{item.name}</td>
            {showAnesthesia ? (
              <td className={`py-1.5 pr-3 whitespace-nowrap ${C.text50}`}>
                {item.anesthesia ?? "-"}
              </td>
            ) : filter === "medicine" ? (
              <>
                <td className={`py-1.5 pr-3 whitespace-nowrap ${C.text50}`}>
                  {item.adminRoute || "-"}
                </td>
                <td className={`py-1.5 pr-3 font-mono whitespace-nowrap ${C.text50}`}>
                  {item.quantity}
                </td>
              </>
            ) : (
              <td className={`py-1.5 pr-3 font-mono whitespace-nowrap ${C.text50}`}>
                {item.quantity}
              </td>
            )}
          </tr>
        ))}
      </HistoryTable>
    </ReportSection>
  );
}
