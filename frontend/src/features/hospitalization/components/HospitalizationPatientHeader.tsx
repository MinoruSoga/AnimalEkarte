// Internal
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { formatDate } from "@/lib/format/date";

// Types
import type { Hospitalization } from "@/types";

interface HospitalizationPatientHeaderProps {
  hospitalization: Hospitalization;
  currentWeight?: string;
  /** スタッフ向け飼主危険マーク (EMR-173)。呼び出し側が owner 危険情報を持つ時だけ指定する。 */
  ownerIsDangerous?: boolean;
  /** ペット危険度 (表示値 "高"/"中"/"低" または wire 値)。高/中のみ Popover バッジを出す。 */
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
  return (
    <PatientInfoCard
      ownerName={hospitalization.ownerName}
      petName={hospitalization.petName}
      petNumber={hospitalization.hospitalizationNo}
      weight={currentWeight || "-"}
      status={hospitalization.petIsDeceased ? "deceased" : "alive"}
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
