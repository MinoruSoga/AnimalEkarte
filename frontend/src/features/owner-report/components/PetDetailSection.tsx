import { C } from "@/lib/design-tokens";
import type { Pet } from "@/types";

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
    <section className={`rounded-lg border ${C.borderLight} ${C.bgWhite} p-4`}>
      <h2 className={`text-sm font-semibold ${C.text} mb-3`}>ペット詳細</h2>
      <dl className="grid grid-cols-2 gap-x-6 gap-y-2 sm:grid-cols-3">
        {rows.map((row) => (
          <div key={row.label} className="flex flex-col">
            <dt className={`text-xs ${C.text50}`}>{row.label}</dt>
            <dd className={`text-sm ${C.text}`}>{row.value}</dd>
          </div>
        ))}
      </dl>
      {pet.remarks ? (
        <dl className="mt-3">
          <dt className={`text-xs ${C.text50}`}>備考</dt>
          <dd className={`text-sm ${C.text} whitespace-pre-wrap`}>{pet.remarks}</dd>
        </dl>
      ) : null}
    </section>
  );
}
