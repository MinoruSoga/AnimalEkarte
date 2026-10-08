// Internal
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { formatDate } from "@/lib/format/date";
import { paths } from "@/config/paths";
import { useGetPet } from "@/hooks/use-pet";
import { isPetDeceasedForClinicalWrite } from "@/lib/transforms/pet";

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
  // EMR-301: wire HospitalizationResponse には birth_date があるが transform 未伝播のため
  // Pet detail を別途取得して誕生月判定に使う（ReservationDetailModal と同型）。
  // 死亡判定は wire petIsDeceased を正本とし、取得済み Pet でも正一致なら抑制する。
  const { data: pet } = useGetPet(hospitalization.petId ?? "");
  const isDeceased =
    hospitalization.petIsDeceased || (pet ? isPetDeceasedForClinicalWrite(pet) : false);

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
      birthDate={pet?.birthDate}
      staffName={hospitalization.doctorName ?? "担当医未設定"}
      reservationType={hospitalization.hospitalizationType}
      petDetails={hospitalization.species}
      insuranceName="-"
      insuranceDetails="-"
      ownerIsDangerous={ownerIsDangerous}
      petDangerLevel={petDangerLevel}
      petDangerReason={petDangerReason}
      nextVisitDate={formatDate(hospitalization.endDate)}
      nextVisitContent="退院予定"
    />
  );
}
