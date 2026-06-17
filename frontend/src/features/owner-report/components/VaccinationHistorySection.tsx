import { C } from "@/lib/design-tokens";
import { usePermission } from "@/hooks/use-permission";
import { ResourceVaccinations } from "@/types/generated/models";
import { useGetPetVaccinations } from "@/features/medical-records";
import { ReportSection } from "./ReportSection";
import { HistoryTable } from "./HistoryTable";

interface VaccinationHistorySectionProps {
  petId: string;
}

/**
 * #158 ② 予防接種履歴。既存 useGetPetVaccinations（pet_id API）を流用。
 * バックエンドが vaccinations.date DESC で返すため日付降順。
 */
export function VaccinationHistorySection({ petId }: VaccinationHistorySectionProps) {
  const { canView } = usePermission(ResourceVaccinations);
  const { data, isLoading, isError } = useGetPetVaccinations(canView ? petId : undefined);
  const items = data ?? [];

  return (
    <ReportSection
      title="予防接種履歴"
      canView={canView}
      isLoading={isLoading}
      isError={isError}
      isEmpty={items.length === 0}
      emptyMessage="予防接種の履歴はありません"
    >
      <HistoryTable headers={["接種日", "ワクチン", "次回予定"]}>
        {items.map((v) => (
          <tr key={v.id}>
            <td className={`py-1.5 pr-3 font-mono whitespace-nowrap ${C.text}`}>{v.date}</td>
            <td className={`py-1.5 pr-3 ${C.text}`}>{v.name}</td>
            <td className={`py-1.5 pr-3 font-mono whitespace-nowrap ${C.text50}`}>{v.next}</td>
          </tr>
        ))}
      </HistoryTable>
    </ReportSection>
  );
}
