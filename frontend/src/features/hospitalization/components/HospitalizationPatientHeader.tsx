// Internal
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { formatDate } from "@/lib/format/date";

// Types
import type { Hospitalization } from "@/types";

interface HospitalizationPatientHeaderProps {
  hospitalization: Hospitalization;
  currentWeight?: string;
}

export function HospitalizationPatientHeader({
  hospitalization,
  currentWeight,
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
      nextVisitDate={formatDate(hospitalization.endDate)}
      nextVisitContent="退院予定"
    />
  );
}
