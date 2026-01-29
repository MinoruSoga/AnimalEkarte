// Internal
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";

// Types
import type { Hospitalization } from "@/types";

interface HospitalizationPatientHeaderProps {
    hospitalization: Hospitalization;
    currentWeight?: string;
}

export const HospitalizationPatientHeader = ({ hospitalization, currentWeight }: HospitalizationPatientHeaderProps) => {
    return (
        <PatientInfoCard
            ownerName={hospitalization.ownerName}
            petName={hospitalization.petName}
            petNumber={hospitalization.hospitalizationNo}
            weight={currentWeight || "-"}
            staffName="担当医"
            serviceType={hospitalization.hospitalizationType}
            petDetails={hospitalization.species}
            insuranceName="-"
            insuranceDetails="-"
            nextVisitDate={hospitalization.endDate}
            nextVisitContent="退院予定"
        />
    );
};
