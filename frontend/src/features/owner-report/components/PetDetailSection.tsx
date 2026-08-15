import { C } from "@/lib/design-tokens";
import type { Pet } from "@/types";
import { formatDate } from "@/lib/format/date";
import { formatPetAge } from "../lib/pet-age";
import { ReportPanel } from "./ReportPanel";

interface PetDetailSectionProps {
  pet: Pet;
  /**
   * #158 初診日（最古カルテ date 由来の派生値）。useGetPetFirstVisit が解決した値を渡す。
   * 受診歴なし／読み込み中は null/undefined で、行は "-" 表示になる（捏造しない）。
   */
  firstVisitDate?: string | null;
}

interface DetailRow {
  label: string;
  value: string;
}

/**
 * #158 ① ペット詳細セクション。
 * ペット属性は useGetPets(ownerId) のデータ（transformBackendPetToFrontend）を使う。
 * 初診日のみ medical_records 由来の派生値のため、親が useGetPetFirstVisit で取得して prop で渡す。
 * 密集ワークスペースの 1 パネルとして ReportPanel に載せ、行数が増えても内部スクロールに収める。
 */
export function PetDetailSection({ pet, firstVisitDate }: PetDetailSectionProps) {
  // 年齢はレガシー EMR（Figma 37:142）が誕生日に併記する導出値。birthDate から算出する（捏造ではない）。
  const age = pet.birthDate ? formatPetAge(pet.birthDate) : null;
  // 前回来院・初診日は date 型。一覧テーブルと同じ共有 formatDate で JST 整合の YYYY/MM/DD に揃える（未設定時は "-"）。
  const lastVisit = formatDate(pet.lastVisit);
  const firstVisit = formatDate(firstVisitDate);

  const rows: DetailRow[] = [
    { label: "ペットNo", value: pet.petNumber || "-" },
    { label: "ふりがな", value: pet.petNameKana || "-" },
    { label: "種別", value: pet.species || "-" },
    { label: "品種", value: pet.breed || "-" },
    { label: "性別", value: pet.gender || "-" },
    { label: "生年月日", value: pet.birthDate || "-" },
    { label: "年齢", value: age || "-" },
    { label: "血液型", value: pet.bloodType || "-" },
    { label: "マイクロチップ", value: pet.microchipNumber || "-" },
    { label: "体重", value: pet.weight ? `${pet.weight} kg` : "-" },
    { label: "毛色", value: pet.color || "-" },
    { label: "避妊去勢日", value: pet.neuteredDate || "-" },
    { label: "入手経路", value: pet.acquisitionType || "-" },
    { label: "フード", value: pet.food || "-" },
    { label: "飼育環境", value: pet.environment || "-" },
    { label: "初診日", value: firstVisit },
    { label: "前回来院", value: lastVisit },
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
