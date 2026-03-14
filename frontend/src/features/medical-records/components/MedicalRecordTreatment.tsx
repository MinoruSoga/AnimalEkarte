// Internal
import { C } from "@/lib/design-tokens";

// Relative
import { TreatmentsTab } from "./TreatmentsTab/TreatmentsTab";

interface MedicalRecordTreatmentProps {
  medicalRecordId: string;
  isNewRecord?: boolean;
}

export function MedicalRecordTreatment({
  medicalRecordId,
  isNewRecord = false,
}: MedicalRecordTreatmentProps) {
  if (isNewRecord) {
    return (
      <div className={`flex items-center justify-center h-48 text-sm ${C.text40}`}>
        カルテを保存してから治療明細を追加できます
      </div>
    );
  }

  return <TreatmentsTab medicalRecordId={medicalRecordId} />;
}
