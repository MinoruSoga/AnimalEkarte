import { CheckupsTab } from "./CheckupsTab/CheckupsTab";
import { MedicalRecordBillCheck } from "./MedicalRecordBillCheck";
import { MedicalRecordEstimate } from "./MedicalRecordEstimate";
import { MedicalRecordExamination } from "./MedicalRecordExamination";
import { MedicalRecordImage } from "./MedicalRecordImage";
import { MedicalRecordVaccination } from "./MedicalRecordVaccination";
import { MedicalRecordMountedTab, MedicalRecordSaveRequired } from "./MedicalRecordTabsShared";
import type { MedicalRecordTabsAreaProps } from "../lib/medical-record-tabs-types";

export function MedicalRecordServiceTabs({
  activeTab,
  mountedTabs,
  isNewRecord,
  recordId,
  selectedPet,
  ownerDiscountRate,
  lstepStatus,
  recordClinicId,
  isFinalized,
  onRegisterEstimateSave,
}: MedicalRecordTabsAreaProps & { isFinalized: boolean }) {
  const saveRequired = isNewRecord || !recordId;
  return (
    <>
      <MedicalRecordMountedTab tab="予防接種" activeTab={activeTab} mountedTabs={mountedTabs}>
        <MedicalRecordSaveRequired show={saveRequired}>
          <MedicalRecordVaccination
            petId={selectedPet.id}
            medicalRecordId={recordId ?? ""}
            lstepStatus={lstepStatus}
          />
        </MedicalRecordSaveRequired>
      </MedicalRecordMountedTab>
      <MedicalRecordMountedTab tab="定期健診" activeTab={activeTab} mountedTabs={mountedTabs}>
        <MedicalRecordSaveRequired show={saveRequired}>
          <CheckupsTab
            medicalRecordId={recordId ?? ""}
            lstepStatus={lstepStatus}
            isFinalized={isFinalized}
          />
        </MedicalRecordSaveRequired>
      </MedicalRecordMountedTab>
      <MedicalRecordMountedTab tab="検査" activeTab={activeTab} mountedTabs={mountedTabs}>
        <MedicalRecordExamination
          isNewRecord={isNewRecord}
          petId={selectedPet.id}
          medicalRecordId={recordId}
        />
      </MedicalRecordMountedTab>
      <MedicalRecordMountedTab tab="画像" activeTab={activeTab} mountedTabs={mountedTabs}>
        <MedicalRecordImage
          isNewRecord={isNewRecord}
          medicalRecordId={recordId}
          recordClinicId={recordClinicId}
          isPetDeceased={selectedPet.status === "死亡"}
        />
      </MedicalRecordMountedTab>
      <MedicalRecordMountedTab tab="見積書" activeTab={activeTab} mountedTabs={mountedTabs}>
        <MedicalRecordEstimate
          isNewRecord={isNewRecord}
          ownerDiscountRate={ownerDiscountRate}
          medicalRecordId={recordId}
          onRegisterSave={onRegisterEstimateSave}
        />
      </MedicalRecordMountedTab>
      <MedicalRecordMountedTab tab="会計(医師確認)" activeTab={activeTab} mountedTabs={mountedTabs}>
        <MedicalRecordBillCheck
          isNewRecord={isNewRecord}
          medicalRecordId={recordId}
          petId={selectedPet.id}
          ownerDiscountRate={ownerDiscountRate}
          recordClinicId={recordClinicId}
        />
      </MedicalRecordMountedTab>
    </>
  );
}
