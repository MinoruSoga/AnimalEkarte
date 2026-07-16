// Internal
import { C } from "@/lib/design-tokens";

// Relative
import { TreatmentsTab } from "./TreatmentsTab/TreatmentsTab";

interface MedicalRecordTreatmentProps {
  medicalRecordId: string;
  isNewRecord?: boolean;
  ownerDiscountRate?: number;
  /** #201: 投与量自動計算の species 解決に使う free-text ペット種 */
  petSpecies?: string | null;
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId */
  recordClinicId?: string;
}

export function MedicalRecordTreatment({
  medicalRecordId,
  isNewRecord = false,
  ownerDiscountRate = 0,
  petSpecies,
  recordClinicId,
}: MedicalRecordTreatmentProps) {
  if (isNewRecord) {
    return (
      <div className={`flex items-center justify-center h-48 text-sm ${C.text40}`}>
        カルテを保存してから治療明細を追加できます
      </div>
    );
  }

  return (
    <TreatmentsTab
      medicalRecordId={medicalRecordId}
      ownerDiscountRate={ownerDiscountRate}
      petSpecies={petSpecies}
      recordClinicId={recordClinicId}
    />
  );
}
