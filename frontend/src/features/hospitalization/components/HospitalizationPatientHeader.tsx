// Internal
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { formatDate } from "@/lib/format/date";
import { paths } from "@/config/paths";

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
      ownerDetailHref={
        hospitalization.ownerId ? paths.owners.detail.getHref(hospitalization.ownerId) : undefined
      }
      petDetailHref={
        hospitalization.ownerId && hospitalization.petId
          ? paths.owners.detail.pet.getHref(hospitalization.ownerId, hospitalization.petId)
          : undefined
      }
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
