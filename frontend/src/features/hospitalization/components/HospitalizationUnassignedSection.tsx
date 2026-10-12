// React/Framework
import { memo } from "react";

// Internal
import { C, BADGE } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import { Card, CardHeader, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { getHospitalizationTypeColor } from "@/lib/status-helpers";

// Relative
import { H_STYLES } from "../lib/styles";

// Types
import type { Hospitalization } from "@/types";

interface HospitalizationUnassignedSectionProps {
  /** cage_id が未設定（「選択なし」で登録された）入院。ケージ設定済みのものは含めない。 */
  hospitalizations: Hospitalization[];
  onNavigateToForm: (id?: string) => void;
  canEdit: boolean;
}

const UNASSIGNED_HEADING_ID = "hospitalization-unassigned-heading";

interface UnassignedCardProps {
  hospitalization: Hospitalization;
  onNavigateToForm: (id?: string) => void;
  canEdit: boolean;
}

const UnassignedCard = memo(function UnassignedCard({
  hospitalization,
  onNavigateToForm,
  canEdit,
}: UnassignedCardProps) {
  const isDeceased = hospitalization.petIsDeceased ?? false;
  // CageCard と同じ規則: 詳細への遷移は編集権限があり死亡していない場合のみ。
  const canShowDetailButton = !isDeceased && canEdit;

  return (
    <Card
      className={`relative flex flex-col h-40 border border-l-4 ${
        isDeceased ? `${C.bgPage} ${C.borderPrimary20}` : `${C.bgWhite} ${C.borderLMedicalBlue}`
      }`}
    >
      <CardHeader
        className={`${H_STYLES.padding.card} pb-0 flex flex-row items-center justify-between space-y-0`}
      >
        <span className={`${H_STYLES.text.sm} ${C.text60} font-semibold`}>ケージなし</span>
        <Badge
          variant="outline"
          className={`${getHospitalizationTypeColor(hospitalization.hospitalizationType)} ${H_STYLES.text.xs} px-1.5 py-0 h-5 border-none`}
        >
          {hospitalization.hospitalizationType}
        </Badge>
      </CardHeader>
      <CardContent
        className={`${H_STYLES.padding.card} flex-1 flex flex-col justify-center items-start text-left`}
      >
        {hospitalization.startDate ? (
          <div className={`text-xs font-mono ${C.text60} w-full`}>
            {formatDate(hospitalization.startDate)}
          </div>
        ) : null}
        <div
          className={`font-bold ${C.text} ${H_STYLES.text.sm} truncate w-full`}
          title={hospitalization.ownerName}
        >
          {hospitalization.ownerName}
        </div>
        <div className={`${H_STYLES.text.sm} ${C.text} truncate w-full`}>
          {[hospitalization.species, hospitalization.petName].filter(Boolean).join(" ")}
        </div>
        {isDeceased ? (
          <Badge
            variant="outline"
            className={`${BADGE.gray} ${H_STYLES.text.xs} px-1.5 py-0 h-5 font-medium`}
          >
            死亡
          </Badge>
        ) : null}
        {canShowDetailButton ? (
          <button
            type="button"
            aria-label={`${hospitalization.petName}の詳細`}
            className={`mt-2 flex min-h-11 items-center justify-center gap-1 text-2xs ${C.textBrand} ${C.bgBrandLight30} border ${C.borderBrandLight} rounded px-1.5 ${C.hoverBgBrandLight60} transition-colors`}
            onClick={() => onNavigateToForm(hospitalization.id)}
          >
            詳細
          </button>
        ) : null}
      </CardContent>
    </Card>
  );
});

/**
 * EMR-316: ケージ未設定の入院を一覧する。ボードはケージ単位で描画するため、ここへ出さないと
 * 「選択なし」で登録した入院がボードから見えなくなる（入院中の動物の見落とし）。
 * 表示と詳細遷移のみで、ドラッグによるケージ割当は行わない。
 */
export const HospitalizationUnassignedSection = memo(function HospitalizationUnassignedSection({
  hospitalizations,
  onNavigateToForm,
  canEdit,
}: HospitalizationUnassignedSectionProps) {
  if (hospitalizations.length === 0) return null;

  return (
    <section aria-labelledby={UNASSIGNED_HEADING_ID} className="min-w-0">
      <h3
        id={UNASSIGNED_HEADING_ID}
        className={`${H_STYLES.text.lg} font-bold ${C.text} mb-3 border-b pb-1 ${C.borderPrimary10}`}
      >
        ケージ未割当
      </h3>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
        {hospitalizations.map((hospitalization) => (
          <UnassignedCard
            key={hospitalization.id}
            hospitalization={hospitalization}
            onNavigateToForm={onNavigateToForm}
            canEdit={canEdit}
          />
        ))}
      </div>
    </section>
  );
});
