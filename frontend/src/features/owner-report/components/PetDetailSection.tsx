import { C } from "@/lib/design-tokens";
import type { Pet } from "@/types";
import { ReportPanel } from "./ReportPanel";

interface PetDetailSectionProps {
  pet: Pet;
}

interface DetailRow {
  label: string;
  value: string;
}

/**
 * #158 ① ペット詳細セクション。
 * 既存 useGetPets(ownerId) のデータ（transformBackendPetToFrontend）を使い、新規 useGetPet は作らない。
 * 密集ワークスペースの 1 パネルとして ReportPanel に載せ、行数が増えても内部スクロールに収める。
 */
export function PetDetailSection({ pet }: PetDetailSectionProps) {
  const rows: DetailRow[] = [
    { label: "ペットNo", value: pet.petNumber || "-" },
    { label: "種別", value: pet.species || "-" },
    { label: "品種", value: pet.breed || "-" },
    { label: "性別", value: pet.gender || "-" },
    { label: "生年月日", value: pet.birthDate || "-" },
    { label: "体重", value: pet.weight ? `${pet.weight} kg` : "-" },
    { label: "毛色", value: pet.color || "-" },
    { label: "危険度", value: pet.dangerLevel || "-" },
    { label: "避妊去勢日", value: pet.neuteredDate || "-" },
    { label: "入手経路", value: pet.acquisitionType || "-" },
    { label: "フード", value: pet.food || "-" },
    { label: "飼育環境", value: pet.environment || "-" },
    { label: "保険", value: pet.insuranceName || "-" },
    { label: "保険内容", value: pet.insuranceDetails || "-" },
    { label: "生死", value: pet.status || "-" },
  ];

  return (
    <ReportPanel title="ペット詳細">
      <dl className="grid grid-cols-2 gap-x-4 gap-y-2">
        {rows.map((row) => (
          <div key={row.label} className="flex min-w-0 flex-col">
            <dt className={`text-xs ${C.text50}`}>{row.label}</dt>
            <dd className={`truncate text-sm ${C.text}`} title={row.value}>
              {row.value}
            </dd>
          </div>
        ))}
      </dl>
      {pet.remarks ? (
        <dl className="mt-3">
          <dt className={`text-xs ${C.text50}`}>備考</dt>
          <dd className={`text-sm whitespace-pre-wrap ${C.text}`}>{pet.remarks}</dd>
        </dl>
      ) : null}
    </ReportPanel>
  );
}
