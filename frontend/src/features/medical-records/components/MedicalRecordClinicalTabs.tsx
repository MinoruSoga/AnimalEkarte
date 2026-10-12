import { isPetDeceasedForClinicalWrite } from "@/lib/transforms/pet";
import { MedicalRecordDiagnosisPlan } from "./MedicalRecordDiagnosisPlan";
import { MedicalRecordInterview } from "./MedicalRecordInterview";
import { InterviewWeightPanel } from "./InterviewWeightPanel";
import { MedicalRecordTreatment } from "./MedicalRecordTreatment";
import { RecommendationReasonSelect } from "./RecommendationReasonSelect";
import { MedicalRecordMountedTab } from "./MedicalRecordTabsShared";
import { collectPastWeightMeasurements } from "../lib/interview-weight-model";
import type { MedicalRecordTabsAreaProps } from "../lib/medical-record-tabs-types";

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
  recommendationReason,
  diagnosis1NameIdError,
  recordClinicId,
  recordPetId,
  historyLoading,
  historyError,
  isFinalized,
  isLocked,
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
  onRecommendationReasonChange,
  onCopyRecordTreatments,
}: MedicalRecordTabsAreaProps & { isFinalized: boolean }) {
  // EMR-356 C2: 過去体重は履歴各行のローカル投影を集約（id で dedupe）。
  // `?? []` 逐次評価だと「投影未提供の医療行」を「測定なし確定」へ潰すため、
  // 未提供判定（unavailable）も同時に返す helper を必ず通す。
  // 最新体重の代理取得はしない — 行に相関した測定のみが対象。
  const { measurements: pastWeightMeasurements, unavailable: pastWeightUnavailable } =
    collectPastWeightMeasurements(historyItems);
  // 書き込み可能な medicalRecordId は、選択中カルテの petId が selectedPet.id と
  // 一致し、clinic コンテキストが解決できた時だけ渡す（切替途中の旧カルテ誤書防止）。
  const writableRecordId =
    recordId && recordClinicId && recordPetId && recordPetId === String(selectedPet.id)
      ? recordId
      : undefined;

  return (
    <>
      <MedicalRecordMountedTab
        tab="問診"
        activeTab={activeTab}
        mountedTabs={mountedTabs}
        contentClassName="min-h-0 flex flex-col"
        isLocked={isLocked}
        internalLock
      >
        <MedicalRecordInterview
          chiefComplaint={chiefComplaint}
          setChiefComplaint={onChiefComplaintChange}
          chiefComplaintTypeId={chiefComplaintTypeId}
          setChiefComplaintTypeId={onChiefComplaintTypeIdChange}
          treatmentPolicy={treatmentPolicy}
          setTreatmentPolicy={onTreatmentPolicyChange}
          historyItems={historyItems}
          isFinalized={isFinalized}
          isLocked={isLocked}
          historyLoading={historyLoading}
          historyError={historyError}
          onCopyRecordTreatments={onCopyRecordTreatments}
          weightPanel={
            <InterviewWeightPanel
              key={`${writableRecordId ?? "new"}|${recordClinicId ?? ""}|${String(selectedPet.id)}`}
              medicalRecordId={writableRecordId}
              recordClinicId={recordClinicId}
              isPetDeceased={isPetDeceasedForClinicalWrite(selectedPet)}
              pastMeasurements={pastWeightMeasurements}
              pastUnavailable={pastWeightUnavailable}
              historyLoading={historyLoading}
              historyError={historyError}
              isLocked={isLocked}
              isFinalized={isFinalized}
            />
          }
        />
      </MedicalRecordMountedTab>
      <MedicalRecordMountedTab
        tab="診察/治療プラン"
        activeTab={activeTab}
        mountedTabs={mountedTabs}
        isLocked={isLocked}
      >
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
          isPetDeceased={isPetDeceasedForClinicalWrite(selectedPet)}
        />
        <div className="px-4 pb-4 mt-4 flex flex-col gap-6">
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
      <MedicalRecordMountedTab
        tab="治療"
        activeTab={activeTab}
        mountedTabs={mountedTabs}
        isLocked={isLocked}
      >
        <MedicalRecordTreatment
          medicalRecordId={recordId ?? ""}
          isNewRecord={isNewRecord}
          petSpecies={selectedPet.species}
          recordClinicId={recordClinicId}
          isPetDeceased={isPetDeceasedForClinicalWrite(selectedPet)}
        />
      </MedicalRecordMountedTab>
    </>
  );
}
