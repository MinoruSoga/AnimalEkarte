import { memo, useCallback } from "react";
import { Link } from "react-router";
import { ChevronDown, FileText } from "lucide-react";
import { PatientContextHeader } from "@/components/shared/PatientContextHeader";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { UnifiedTabsContent, UnifiedTabsList } from "@/components/shared/UnifiedTabs";
import { Button } from "@/components/ui/button";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { cn } from "@/lib/utils";
import { openOwnerReport } from "@/lib/owner-report-window";
import { usePermission } from "@/hooks/use-permission";
import { getMedicalRecordStatusColor } from "@/lib/status-helpers";
import { ResourceMedicalRecords } from "@/types/generated/models";
import { todayJSTISO } from "@/lib/jst-date";
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
import { VisitTypeSelect } from "./VisitTypeSelect";
import { NextVisitButton } from "./NextVisitButton";
import type { RecommendationReason } from "../constants/recommendation-reason";
import type { InterviewHistoryItem } from "../types";
import { isMedicalRecordFinalizedStatus } from "../lib/medical-record-lock";

interface MedicalRecordStickyHeaderProps {
  selectedPet: Pet;
  cohabitingPets: Pet[];
  staffName: string;
  visitType: string;
  visitCount: number;
  canEdit: boolean;
  isNewRecord: boolean;
  tabs: { value: string; label: string }[];
  recordDate?: string;
  recordStatus?: string;
  nextVisitDate: string;
  onVisitTypeChange: (value: string) => void;
  onStaffClick: () => void;
  onOwnerClick: () => void;
  onDateChange?: (date: string) => void;
  onNextVisitDatePatch: (date: string) => void;
  onNextVisitDateValidChange: (valid: boolean) => void;
  hasLineIntegration?: boolean;
}

const CohabitingPetChips = memo(function CohabitingPetChips({
  pets,
}: {
  pets: Pet[];
}) {
  return (
    <section
      aria-label="同居ペット"
      className={cn(
        "flex items-center gap-1.5 overflow-x-auto rounded-md p-2 [&::-webkit-scrollbar]:hidden",
        C.bgPage30,
      )}
      style={{ scrollbarWidth: "none" }}
    >
      <span className={`shrink-0 text-xs ${C.text50}`}>同居ペット</span>
      <div className="flex min-w-max gap-1.5">
        {pets.map((pet) => {
          const label = pet.species ? `${pet.name}（${pet.species}）` : pet.name;
          return (
            <Link
              key={pet.id}
              to={`${paths.medicalRecords.getHref()}?pet_id=${encodeURIComponent(pet.id)}`}
              className={cn(
                "h-8 shrink-0 rounded-md border bg-white px-2.5 text-sm leading-8 whitespace-nowrap transition-colors",
                C.text,
                C.borderMedium,
                C.hoverBgLight,
              )}
            >
              {label}
            </Link>
          );
        })}
      </div>
    </section>
  );
});

export function MedicalRecordStickyHeader({
  selectedPet,
  cohabitingPets,
  staffName,
  visitType,
  visitCount,
  canEdit,
  isNewRecord,
  tabs,
  recordDate,
  recordStatus,
  nextVisitDate,
  onVisitTypeChange,
  onStaffClick,
  onOwnerClick,
  onDateChange,
  onNextVisitDatePatch,
  onNextVisitDateValidChange,
  hasLineIntegration,
}: MedicalRecordStickyHeaderProps) {
  const isFinalized = isMedicalRecordFinalizedStatus(recordStatus);
  const canEditDate = canEdit && !isFinalized && !!onDateChange && !isNewRecord;
  const dateInputValue = recordDate ? recordDate.replace(/\//g, "-") : undefined;

  // #158: 飼主レポートを別ウィンドウで開く。view 権限がない場合はボタンを出さない。
  const { canView: canViewReport } = usePermission(ResourceMedicalRecords);
  const reportOwnerId = selectedPet.ownerId;
  const reportPetId = selectedPet.id;
  const handleOpenReport = useCallback(() => {
    openOwnerReport(reportOwnerId, reportPetId);
  }, [reportOwnerId, reportPetId]);

  const contextControls = (
    <>
      {/* SPEC-GAP: 確定済みバッジ。臨床記録の真正性担保のため、確定状態を常時明示する */}
      {!isNewRecord && isFinalized ? (
        <StatusBadge colorClass={getMedicalRecordStatusColor("確定済")}>
          確定済
        </StatusBadge>
      ) : null}

      {/* 来院種別 */}
      <VisitTypeSelect
        value={visitType}
        onChange={onVisitTypeChange}
        disabled={!canEdit || isFinalized}
      />

      {/* 診察日 */}
      <div className="flex flex-col gap-0 shrink-0 min-w-[110px]">
        <span className={`text-xs ${C.text50}`}>診察日</span>
        {canEditDate ? (
          <input
            key={dateInputValue}
            type="date"
            aria-label="診察日"
            defaultValue={dateInputValue}
            onChange={(e) => {
              if (e.target.value) onDateChange!(e.target.value);
            }}
            className={`h-11 text-sm ${C.text} bg-transparent rounded px-1 cursor-pointer focus:outline-none focus:ring-1 focus:ring-current`}
          />
        ) : (
          <span className={`h-8 flex items-center text-sm ${C.text}`}>
            {isNewRecord ? todayJSTISO() : (recordDate ?? "-")}
          </span>
        )}
      </div>

      {/* 担当医 */}
      {canEdit && !isFinalized ? (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className={`h-8 shrink-0 text-sm gap-1 px-2 max-w-[160px] ${C.hoverBgPage} ${C.text} border-none`}
          onClick={onStaffClick}
          aria-label={`担当医: ${staffName}`}
        >
          <span className={`text-xs ${C.text50} mr-0.5 shrink-0`}>担当医</span>
          <span className="truncate">{staffName}</span>
          <ChevronDown className={`${ICON.sm} ${C.text40} shrink-0`} aria-hidden="true" />
        </Button>
      ) : (
        <div className="flex flex-col gap-0 shrink-0 max-w-[160px]">
          <span className={`text-xs ${C.text50}`}>担当医</span>
          <span className={`h-8 flex items-center text-sm ${C.text} truncate`}>{staffName}</span>
        </div>
      )}

      {/* 次回予定 */}
      {!isNewRecord ? (
        <NextVisitButton
          value={nextVisitDate}
          onChange={onNextVisitDatePatch}
          onValidationChange={onNextVisitDateValidChange}
          hasLineIntegration={hasLineIntegration}
          disabled={!canEdit || isFinalized}
        />
      ) : null}

      {/* #158: 飼主レポート（別ウィンドウ）。当該飼主・当該ペットを初期選択して開く。 */}
      {canViewReport ? (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className={`h-8 shrink-0 text-sm gap-1 px-2 ${C.hoverBgPage} ${C.text} border-none`}
          onClick={handleOpenReport}
          aria-label="飼主レポートを開く"
        >
          <FileText className={`${ICON.sm} ${C.text40}`} aria-hidden="true" />
          レポート
        </Button>
      ) : null}
    </>
  );

  return (
    <div className={`sticky top-0 z-10 ${C.bgPage}`}>
      <PatientContextHeader
        ownerName={selectedPet.ownerName}
        petName={selectedPet.name}
        petNumber={selectedPet.petNumber || selectedPet.id}
        weight={selectedPet.weight ?? undefined}
        status={selectedPet.status === "死亡" ? "deceased" : "alive"}
        birthDate={selectedPet.birthDate ?? undefined}
        species={selectedPet.species}
        gender={selectedPet.gender}
        neuteredDate={selectedPet.neuteredDate}
        breed={selectedPet.breed}
        insuranceName={selectedPet.insuranceName ?? undefined}
        insuranceDetails={selectedPet.insuranceDetails ?? undefined}
        visitCount={visitCount}
        onOwnerClick={!isNewRecord && canEdit && !isFinalized ? onOwnerClick : undefined}
        contextControls={contextControls}
      />
      {!isNewRecord && cohabitingPets.length > 0 ? (
        <CohabitingPetChips pets={cohabitingPets} />
      ) : null}
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
  physicalExam: string;
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
  /** P2-15: 拠点横断で開いたカルテの子リソース操作用。レコード自身の clinicId */
  recordClinicId?: string;
  onChiefComplaintChange: (value: string) => void;
  onChiefComplaintTypeIdChange: (id: number | null) => void;
  onTreatmentPolicyChange: (value: string) => void;
  onPhysicalExamChange: (value: string) => void;
  onPlanChange: (value: string) => void;
  onAssessmentChange: (value: string) => void;
  onDiagnosis1CategoryIdChange: (id: number | null) => void;
  onDiagnosis1NameIdChange: (id: number | null) => void;
  onDiagnosis2CategoryIdChange: (id: number | null) => void;
  onDiagnosis2NameIdChange: (id: number | null) => void;
  onNextVisitDateChange: (value: string) => void;
  onNextVisitDateValidChange: (valid: boolean) => void;
  onRecommendationReasonChange: (value: RecommendationReason | null) => void;
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
  lstepStatus,
  recordStatus,
  diagnosis1NameIdError,
  recordClinicId,
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
  onRegisterEstimateSave,
}: MedicalRecordTabsAreaProps) {
  const isFinalized = isMedicalRecordFinalizedStatus(recordStatus);
  return (
    <div className={`mt-4 ${LAYOUT.fullHeight}`}>
      {mountedTabs.has("問診") ? (
        <UnifiedTabsContent value="問診" className="min-h-0 flex flex-col">
          <div className={`${LAYOUT.fullHeight} ${activeTab === "問診" ? "" : "hidden"}`}>
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
          </div>
        </UnifiedTabsContent>
      ) : null}
      {mountedTabs.has("診察/治療プラン") ? (
        <UnifiedTabsContent value="診察/治療プラン">
          <div className={`${LAYOUT.fullHeight} ${activeTab === "診察/治療プラン" ? "" : "hidden"}`}>
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
          </div>
        </UnifiedTabsContent>
      ) : null}
      {mountedTabs.has("治療") ? (
        <UnifiedTabsContent value="治療">
          <div className={`${LAYOUT.fullHeight} ${activeTab === "治療" ? "" : "hidden"}`}>
            <MedicalRecordTreatment
              medicalRecordId={recordId ?? ""}
              isNewRecord={isNewRecord}
              petSpecies={selectedPet.species}
              recordClinicId={recordClinicId}
            />
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
              <CheckupsTab medicalRecordId={recordId} lstepStatus={lstepStatus} isFinalized={isFinalized} />
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
            <MedicalRecordImage
              isNewRecord={isNewRecord}
              medicalRecordId={recordId}
              recordClinicId={recordClinicId}
              isPetDeceased={selectedPet.status === "死亡"}
            />
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
              recordClinicId={recordClinicId}
            />
          </div>
        </UnifiedTabsContent>
      ) : null}
    </div>
  );
}
