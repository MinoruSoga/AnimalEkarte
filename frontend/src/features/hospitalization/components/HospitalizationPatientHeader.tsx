// Internal
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { formatDate } from "@/utils/format/date";

// Types
import type { Hospitalization } from "@/types";

interface HospitalizationPatientHeaderProps {
    hospitalization: Hospitalization;
    currentWeight?: string;
}

export function HospitalizationPatientHeader({ hospitalization, currentWeight }: HospitalizationPatientHeaderProps) {
    return (
        <PatientInfoCard
            ownerName={hospitalization.ownerName}
            petName={hospitalization.petName}
            petNumber={hospitalization.hospitalizationNo}
            weight={currentWeight || "-"}
            staffName={hospitalization.doctorName ?? "担当医未設定"}
            serviceType={hospitalization.hospitalizationType}
            petDetails={hospitalization.species}
            insuranceName="-"
            insuranceDetails="-"
            nextVisitDate={formatDate(hospitalization.endDate)}
            nextVisitContent="退院予定"
        />
    );
}
