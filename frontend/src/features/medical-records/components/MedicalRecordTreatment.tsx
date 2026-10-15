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
  /** 死亡ペットの治療明細ミューテーションをブロックする（VitalsTab と同型の二重防壁） */
  isPetDeceased?: boolean;
  /** EMR-412 K5: ロック中は理由つき空状態文言。既定 false。 */
  isLocked?: boolean;
  /** EMR-412 K5: 確定済み判定（K5 文言の優先理由）。既定 false。 */
  isFinalized?: boolean;
}

export function MedicalRecordTreatment({
  medicalRecordId,
  isNewRecord = false,
  ownerDiscountRate = 0,
  petSpecies,
  recordClinicId,
  isPetDeceased = false,
  isLocked = false,
  isFinalized = false,
}: MedicalRecordTreatmentProps) {
  if (isNewRecord) {
    return (
      <div className={`flex items-center justify-center h-48 text-sm ${C.text60}`}>
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
      isPetDeceased={isPetDeceased}
      isLocked={isLocked}
      isFinalized={isFinalized}
    />
  );
}
