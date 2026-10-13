// Internal
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { formatDate } from "@/lib/format/date";
import { paths } from "@/config/paths";

// Types
import type { Hospitalization } from "@/types";

interface HospitalizationPatientHeaderProps {
  hospitalization: Hospitalization;
  currentWeight?: string;
  /** スタッフ向け飼主特記マーク (EMR-173/231)。呼び出し側が owner 特記情報を持つ時だけ指定する。 */
  ownerIsDangerous?: boolean;
  /** ペット特記レベル (表示値 "高"/"中"/"低" または wire 値)。高/中のみ Popover バッジを出す。 */
  petDangerLevel?: string;
  petDangerReason?: string;
}

export function HospitalizationPatientHeader({
  hospitalization,
  currentWeight,
  ownerIsDangerous,
  petDangerLevel,
  petDangerReason,
}: HospitalizationPatientHeaderProps) {
  // EMR-301: 誕生月・死亡判定は wire HospitalizationResponse の pet サマリを transform
  // 伝播した値を使う（/pets/:id は owners:view 権限が別途必要で、入院閲覧者には失敗する）。
  const isDeceased = hospitalization.petIsDeceased;

  return (
    <PatientInfoCard
      ownerName={hospitalization.ownerName}
      petName={hospitalization.petName}
      petNumber={hospitalization.hospitalizationNo}
      ownerDetailHref={
        hospitalization.ownerId ? paths.owners.detail.getHref(hospitalization.ownerId) : undefined
      }
      petDetailHref={
        hospitalization.ownerId && hospitalization.petId
          ? paths.owners.detail.pet.getHref(hospitalization.ownerId, hospitalization.petId)
          : undefined
      }
      weight={currentWeight || "-"}
      status={isDeceased ? "deceased" : "alive"}
      birthDate={hospitalization.petBirthDate}
      // EMR-352: 未設定 fallback は「未設定」のみ。staffLabel="担当医" は
      // PatientInfoCard 側が付与するため、ここでラベルを含めると二重表示になる。
      staffName={hospitalization.doctorName ?? "未設定"}
      reservationType={hospitalization.hospitalizationType}
      petDetails={hospitalization.species}
      insuranceName="-"
      insuranceDetails="-"
      ownerIsDangerous={ownerIsDangerous}
      petDangerLevel={petDangerLevel}
      petDangerReason={petDangerReason}
      nextVisitDate={formatDate(hospitalization.nextVisit || hospitalization.endDate)}
      nextVisitContent={hospitalization.nextVisit ? "次回予定" : "退院予定"}
    />
  );
}
