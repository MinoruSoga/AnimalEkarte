import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { UnifiedTabsContent, UnifiedTabsList } from "@/components/shared/UnifiedTabs";
import { C, LAYOUT } from "@/lib/design-tokens";
import type { Pet } from "@/types";
import { CheckupsTab } from "./CheckupsTab/CheckupsTab";
import { MedicalRecordBillCheck } from "./MedicalRecordBillCheck";
import { MedicalRecordDiagnosisPlan } from "./MedicalRecordDiagnosisPlan";
import { MedicalRecordEstimate } from "./MedicalRecordEstimate";
import { MedicalRecordExamination } from "./MedicalRecordExamination";
import { MedicalRecordImage } from "./MedicalRecordImage";
import { MedicalRecordInterview } from "./MedicalRecordInterview";
import { MedicalRecordTreatment } from "./MedicalRecordTreatment";
import { MedicalRecordVaccination } from "./MedicalRecordVaccination";
import { NextVisitDateField } from "./NextVisitDateField";
import { RecommendationReasonSelect } from "./RecommendationReasonSelect";
import type { RecommendationReason } from "../constants/recommendation-reason";
import type { InterviewHistoryItem } from "../types";

interface MedicalRecordStickyHeaderProps {
  selectedPet: Pet;
  staffName: string;
  visitType: string;
  visitCount: number;
  canEdit: boolean;
  isNewRecord: boolean;
  tabs: { value: string; label: string }[];
  onVisitTypeClick: () => void;
  onStaffClick: () => void;
  onOwnerClick: () => void;
}

export function MedicalRecordStickyHeader({
  selectedPet,
  staffName,
  visitType,
  visitCount,
  canEdit,
  isNewRecord,
  tabs,
  onVisitTypeClick,
  onStaffClick,
  onOwnerClick,
}: MedicalRecordStickyHeaderProps) {
  return (
    <div className={`sticky top-0 z-10 ${C.bgPage}`}>
      <PatientInfoCard
        ownerName={selectedPet.ownerName}
        petName={`${selectedPet.name}${selectedPet.species ? `(${selectedPet.species})` : ""}`}
        petNumber={selectedPet.petNumber || selectedPet.id}
        weight={selectedPet.weight || "-"}
        status={selectedPet.status === "死亡" ? "deceased" : "alive"}
        staffName={staffName}
        staffLabel="担当医: "
        reservationType={visitType}
        reservationTypeLabel="来院種別"
        onReservationTypeClick={onVisitTypeClick}
        onStaffClick={canEdit ? onStaffClick : undefined}
        onOwnerClick={!isNewRecord ? onOwnerClick : undefined}
        petDetails={`${selectedPet.birthDate ? `${selectedPet.birthDate}生` : ""} / ${selectedPet.species}`}
        insuranceName={selectedPet.insuranceName || "保険情報未登録"}
        insuranceDetails={selectedPet.insuranceDetails || "-"}
        nextVisitDate="-"
        nextVisitContent="-"
        visitCount={visitCount}
        sticky={false}
      />

      <div className={`flex shrink-0 overflow-x-auto ${C.bgPage}`}>
        <UnifiedTabsList items={tabs} />
      </div>
    </div>
  );
}

interface MedicalRecordTabsAreaProps {
  activeTab: string;
  mountedTabs: Set<string>;
  isNewRecord: boolean;
  recordId: string | undefined;
  selectedPet: Pet;
  chiefComplaint: string;
  chiefComplaintTypeId: number | null;
  treatmentPolicy: string;
  historyItems: InterviewHistoryItem[];
  plan: string;
  assessment: string;
  diagnosis1CategoryId: number | null;
  diagnosis1NameId: number | null;
  diagnosis2CategoryId: number | null;
  diagnosis2NameId: number | null;
  ownerDiscountRate: number;
  nextVisitDate: string;
  hasLineIntegration: boolean;
  recommendationReason: RecommendationReason | null;
  lstepStatus: "synced" | "not-linked" | "opt-out" | undefined;
  recordStatus: string;
  diagnosis1NameIdError: string | undefined;
  onChiefComplaintChange: (value: string) => void;
  onChiefComplaintTypeIdChange: (id: number | null) => void;
  onTreatmentPolicyChange: (value: string) => void;
  onPlanChange: (value: string) => void;
  onAssessmentChange: (value: string) => void;
  onDiagnosis1CategoryIdChange: (id: number | null) => void;
  onDiagnosis1NameIdChange: (id: number | null) => void;
  onDiagnosis2CategoryIdChange: (id: number | null) => void;
  onDiagnosis2NameIdChange: (id: number | null) => void;
  onNextVisitDateChange: (value: string) => void;
  onNextVisitDateValidChange: (valid: boolean) => void;
  onRecommendationReasonChange: (value: RecommendationReason | null) => void;
  onRegisterClinicalPlanSave: (fn: () => Promise<void>) => void;
  onRegisterEstimateSave: (fn: () => Promise<void>) => void;
}

export function MedicalRecordTabsArea({
  activeTab,
  mountedTabs,
  isNewRecord,
  recordId,
  selectedPet,
  chiefComplaint,
  chiefComplaintTypeId,
  treatmentPolicy,
  historyItems,
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
  lstepStatus,
  recordStatus,
  diagnosis1NameIdError,
  onChiefComplaintChange,
  onChiefComplaintTypeIdChange,
  onTreatmentPolicyChange,
  onPlanChange,
  onAssessmentChange,
  onDiagnosis1CategoryIdChange,
  onDiagnosis1NameIdChange,
  onDiagnosis2CategoryIdChange,
  onDiagnosis2NameIdChange,
  onNextVisitDateChange,
  onNextVisitDateValidChange,
  onRecommendationReasonChange,
  onRegisterClinicalPlanSave,
  onRegisterEstimateSave,
}: MedicalRecordTabsAreaProps) {
  return (
    <div className={`mt-4 ${LAYOUT.fullHeight}`}>
      {mountedTabs.has("問診") ? (
        <UnifiedTabsContent value="問診">
          <div className={`${LAYOUT.fullHeight} ${activeTab === "問診" ? "" : "hidden"}`}>
            <MedicalRecordInterview
              chiefComplaint={chiefComplaint}
              setChiefComplaint={onChiefComplaintChange}
              chiefComplaintTypeId={chiefComplaintTypeId}
              setChiefComplaintTypeId={onChiefComplaintTypeIdChange}
              treatmentPolicy={treatmentPolicy}
              setTreatmentPolicy={onTreatmentPolicyChange}
              historyItems={historyItems}
            />
          </div>
        </UnifiedTabsContent>
      ) : null}
      {mountedTabs.has("診察/治療プラン") ? (
        <UnifiedTabsContent value="診察/治療プラン">
          <div className={`${LAYOUT.fullHeight} ${activeTab === "診察/治療プラン" ? "" : "hidden"}`}>
            <MedicalRecordDiagnosisPlan
              isNewRecord={isNewRecord}
              chiefComplaint={chiefComplaint}
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
              onRegisterClinicalPlanSave={onRegisterClinicalPlanSave}
              diagnosis1NameIdError={diagnosis1NameIdError}
            />
            <div className="px-4 pb-4 mt-4 flex flex-col gap-6">
              <NextVisitDateField
                value={nextVisitDate}
                onChange={onNextVisitDateChange}
                onValidationChange={onNextVisitDateValidChange}
                hasLineIntegration={hasLineIntegration}
                disabled={isNewRecord}
              />
              {recordId ? (
                <RecommendationReasonSelect
                  mode="edit"
                  medicalRecordId={recordId}
                  value={recommendationReason}
                  disabled={false}
                />
              ) : (
                <RecommendationReasonSelect
                  mode="create"
                  value={recommendationReason}
                  onChange={onRecommendationReasonChange}
                  disabled={false}
                />
              )}
            </div>
          </div>
        </UnifiedTabsContent>
      ) : null}
      {mountedTabs.has("治療") ? (
        <UnifiedTabsContent value="治療">
          <div className={`${LAYOUT.fullHeight} ${activeTab === "治療" ? "" : "hidden"}`}>
            <MedicalRecordTreatment medicalRecordId={recordId ?? ""} isNewRecord={isNewRecord} />
          </div>
        </UnifiedTabsContent>
      ) : null}
      {mountedTabs.has("予防接種") ? (
        <UnifiedTabsContent value="予防接種">
          <div className={`${LAYOUT.fullHeight} ${activeTab === "予防接種" ? "" : "hidden"}`}>
            <MedicalRecordVaccination petId={selectedPet.id} lstepStatus={lstepStatus} />
          </div>
        </UnifiedTabsContent>
      ) : null}
      {mountedTabs.has("定期健診") ? (
        <UnifiedTabsContent value="定期健診">
          <div className={`${LAYOUT.fullHeight} ${activeTab === "定期健診" ? "" : "hidden"}`}>
            {isNewRecord || !recordId ? (
              <div className={`flex items-center justify-center h-48 text-sm ${C.text40}`}>
                カルテを保存してから使用できます
              </div>
            ) : (
              <CheckupsTab medicalRecordId={recordId} lstepStatus={lstepStatus} isFinalized={recordStatus === "確定済"} />
            )}
          </div>
        </UnifiedTabsContent>
      ) : null}
      {mountedTabs.has("検査") ? (
        <UnifiedTabsContent value="検査">
          <div className={`${LAYOUT.fullHeight} ${activeTab === "検査" ? "" : "hidden"}`}>
            <MedicalRecordExamination isNewRecord={isNewRecord} petId={selectedPet.id} medicalRecordId={recordId} />
          </div>
        </UnifiedTabsContent>
      ) : null}
      {mountedTabs.has("画像") ? (
        <UnifiedTabsContent value="画像">
          <div className={`${LAYOUT.fullHeight} ${activeTab === "画像" ? "" : "hidden"}`}>
            <MedicalRecordImage isNewRecord={isNewRecord} medicalRecordId={recordId} />
          </div>
        </UnifiedTabsContent>
      ) : null}
      {mountedTabs.has("見積書") ? (
        <UnifiedTabsContent value="見積書">
          <div className={`${LAYOUT.fullHeight} ${activeTab === "見積書" ? "" : "hidden"}`}>
            <MedicalRecordEstimate
              isNewRecord={isNewRecord}
              ownerDiscountRate={ownerDiscountRate}
              medicalRecordId={recordId}
              onRegisterSave={onRegisterEstimateSave}
            />
          </div>
        </UnifiedTabsContent>
      ) : null}
      {mountedTabs.has("会計(医師確認)") ? (
        <UnifiedTabsContent value="会計(医師確認)">
          <div className={`${LAYOUT.fullHeight} ${activeTab === "会計(医師確認)" ? "" : "hidden"}`}>
            <MedicalRecordBillCheck
              isNewRecord={isNewRecord}
              medicalRecordId={recordId}
              ownerDiscountRate={ownerDiscountRate}
            />
          </div>
        </UnifiedTabsContent>
      ) : null}
    </div>
  );
}
