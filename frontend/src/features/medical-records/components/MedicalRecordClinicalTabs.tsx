import { MedicalRecordDiagnosisPlan } from "./MedicalRecordDiagnosisPlan";
import { MedicalRecordInterview } from "./MedicalRecordInterview";
import { MedicalRecordTreatment } from "./MedicalRecordTreatment";
import { NextVisitDateField } from "./NextVisitDateField";
import { RecommendationReasonSelect } from "./RecommendationReasonSelect";
import { MedicalRecordMountedTab } from "./medical-record-tabs-shared";
import type { MedicalRecordTabsAreaProps } from "./medical-record-tabs-types";

export function MedicalRecordClinicalTabs({
  activeTab,
  mountedTabs,
  isNewRecord,
  recordId,
  selectedPet,
  chiefComplaint,
  chiefComplaintTypeId,
  treatmentPolicy,
  historyItems,
  physicalExam,
  plan,
  assessment,
  diagnosis1CategoryId,
  diagnosis1NameId,
  diagnosis2CategoryId,
  diagnosis2NameId,
  ownerDiscountRate,
  nextVisitDate,
  hasLineIntegration,
  recommendationReason,
  diagnosis1NameIdError,
  recordClinicId,
  isFinalized,
  onChiefComplaintChange,
  onChiefComplaintTypeIdChange,
  onTreatmentPolicyChange,
  onPhysicalExamChange,
  onPlanChange,
  onAssessmentChange,
  onDiagnosis1CategoryIdChange,
  onDiagnosis1NameIdChange,
  onDiagnosis2CategoryIdChange,
  onDiagnosis2NameIdChange,
  onNextVisitDateChange,
  onNextVisitDateValidChange,
  onRecommendationReasonChange,
}: MedicalRecordTabsAreaProps & { isFinalized: boolean }) {
  return (
    <>
      <MedicalRecordMountedTab tab="問診" activeTab={activeTab} mountedTabs={mountedTabs} contentClassName="min-h-0 flex flex-col">
        <MedicalRecordInterview
          chiefComplaint={chiefComplaint}
          setChiefComplaint={onChiefComplaintChange}
          chiefComplaintTypeId={chiefComplaintTypeId}
          setChiefComplaintTypeId={onChiefComplaintTypeIdChange}
          treatmentPolicy={treatmentPolicy}
          setTreatmentPolicy={onTreatmentPolicyChange}
          historyItems={historyItems}
          isFinalized={isFinalized}
        />
      </MedicalRecordMountedTab>
      <MedicalRecordMountedTab tab="診察/治療プラン" activeTab={activeTab} mountedTabs={mountedTabs}>
        <MedicalRecordDiagnosisPlan
          isNewRecord={isNewRecord}
          chiefComplaint={chiefComplaint}
          physicalExam={physicalExam}
          setPhysicalExam={onPhysicalExamChange}
          plan={plan}
          setPlan={onPlanChange}
          assessment={assessment}
          setAssessment={onAssessmentChange}
          diagnosis1CategoryId={diagnosis1CategoryId}
          setDiagnosis1CategoryId={onDiagnosis1CategoryIdChange}
          diagnosis1NameId={diagnosis1NameId}
          setDiagnosis1NameId={onDiagnosis1NameIdChange}
          diagnosis2CategoryId={diagnosis2CategoryId}
          setDiagnosis2CategoryId={onDiagnosis2CategoryIdChange}
          diagnosis2NameId={diagnosis2NameId}
          setDiagnosis2NameId={onDiagnosis2NameIdChange}
          medicalRecordId={recordId}
          ownerDiscountRate={ownerDiscountRate}
          diagnosis1NameIdError={diagnosis1NameIdError}
          recordClinicId={recordClinicId}
        />
        <div className="px-4 pb-4 mt-4 flex flex-col gap-6">
          <NextVisitDateField
            value={nextVisitDate}
            onChange={onNextVisitDateChange}
            onValidationChange={onNextVisitDateValidChange}
            hasLineIntegration={hasLineIntegration}
            disabled={isNewRecord || isFinalized}
          />
          {recordId ? (
            <RecommendationReasonSelect
              mode="edit"
              medicalRecordId={recordId}
              value={recommendationReason}
              disabled={isFinalized}
            />
          ) : (
            <RecommendationReasonSelect
              mode="create"
              value={recommendationReason}
              onChange={onRecommendationReasonChange}
              disabled={isFinalized}
            />
          )}
        </div>
      </MedicalRecordMountedTab>
      <MedicalRecordMountedTab tab="治療" activeTab={activeTab} mountedTabs={mountedTabs}>
        <MedicalRecordTreatment
          medicalRecordId={recordId ?? ""}
          isNewRecord={isNewRecord}
          petSpecies={selectedPet.species}
          recordClinicId={recordClinicId}
        />
      </MedicalRecordMountedTab>
    </>
  );
}
