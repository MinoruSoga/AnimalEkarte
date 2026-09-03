import {
  useCallback,
  useRef,
  useState,
  useTransition,
} from "react";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { C, LAYOUT } from "@/lib/design-tokens";
import { UnifiedTabsRoot } from "@/components/shared/UnifiedTabs";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { useAuth } from "@/hooks/use-auth";
import { useGetOwnerLineTags } from "@/hooks/use-owner-line-tags";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { ResourceMedicalRecords } from "@/types/generated/models";
import type { Pet } from "@/types";
import { MedicalRecordAddenda } from "../components/MedicalRecordAddenda";
import { MedicalRecordAutoCreateFailure } from "../components/MedicalRecordAutoCreateFailure";
import { MedicalRecordStickyHeader, MedicalRecordTabsArea } from "../components/MedicalRecordFormPanels";
import {
  MedicalRecordFinalizeDialog,
  MedicalRecordFloatingActions,
  MedicalRecordPrintArea,
} from "../components/MedicalRecordFormActions";
import { MedicalRecordFormModals } from "../components/MedicalRecordFormModals";
import { isMedicalRecordFinalizedStatus } from "../lib/medical-record-lock";
import { MEDICAL_RECORD_TAB_ITEMS } from "./medical-record-form-model";
import { useMedicalRecordDirtyFields } from "../hooks/use-medical-record-dirty-fields";
import { useMedicalRecordFormModals } from "../hooks/use-medical-record-form-modals";
import { useMedicalRecordPostSave } from "../hooks/use-medical-record-post-save";
import { useMedicalRecordForm } from "../hooks/use-medical-record-form";
import { useGetMedicalRecord } from "../api/get-medical-record";
import { useGetPetMedicalHistory } from "../api/get-medical-records";
import { useGetClinicalPlan } from "../api/clinical-plan";
import { useGetTreatments } from "../api/treatments";
import { useGetBillingConfirmation } from "../api/billing-confirmation";

type MedicalRecordFormModel = ReturnType<typeof useMedicalRecordForm>;

function useMedicalRecordFormReadyState(input: {
  recordId: string | undefined;
  selectedPet: Pet;
  form: MedicalRecordFormModel;
}) {
  const { recordId, selectedPet, form } = input;
  const { historyItems } = useGetPetMedicalHistory(selectedPet.id, recordId);
  const { isDirty, markDirty, markClean } = useUnsavedChanges();
  const { data: ownerLineData } = useGetOwnerLineTags(selectedPet.ownerId ?? "");
  const hasLineIntegration = (ownerLineData?.is_linked && !ownerLineData?.lstep_opt_out) ?? false;
  const lstepStatus = ownerLineData === undefined
    ? undefined
    : ownerLineData.lstep_opt_out
      ? ("opt-out" as const)
      : ownerLineData.is_linked
        ? ("synced" as const)
        : ("not-linked" as const);
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const { handleRegisterEstimateSave } = useMedicalRecordPostSave({
    activeTab: form.activeTab,
    formState: form.formState,
    markClean,
  });
  const { user } = useAuth();
  const { data: currentRecord } = useGetMedicalRecord(recordId ?? "");
  const recordFinalized =
    form.isFinalized || isMedicalRecordFinalizedStatus(currentRecord?.status);
  const recordClinicId = currentRecord?.clinicId;
  const { data: clinicalPlan } = useGetClinicalPlan(recordId ?? "", recordClinicId);
  const { data: treatments = [] } = useGetTreatments(recordId ?? "", recordClinicId);
  const {
    data: billingConfirmation,
    isLoading: isBillingConfirmationLoading,
    isError: isBillingConfirmationError,
  } = useGetBillingConfirmation(recordId ?? "");
  const [staffName, setStaffName] = useState(() => user?.displayName ?? "");
  const modals = useMedicalRecordFormModals();
  const [mountedTabs, setMountedTabs] = useState<Set<string>>(() => new Set(["問診", form.activeTab]));
  const [isTabPending, startTabTransition] = useTransition();
  const dirtyFields = useMedicalRecordDirtyFields({
    markDirty,
    setAssessment: form.setAssessment,
    setChiefComplaint: form.setChiefComplaint,
    setChiefComplaintTypeId: form.setChiefComplaintTypeId,
    setDiagnosis1CategoryId: form.setDiagnosis1CategoryId,
    setDiagnosis1NameId: form.setDiagnosis1NameId,
    setDiagnosis2CategoryId: form.setDiagnosis2CategoryId,
    setDiagnosis2NameId: form.setDiagnosis2NameId,
    setPhysicalExam: form.setPhysicalExam,
    setPlan: form.setPlan,
    setTreatmentPolicy: form.setTreatmentPolicy,
  });

  const handleChangeDoctor = form.handleChangeDoctor;
  const handleFinalize = form.handleFinalize;
  const setActiveTab = form.setActiveTab;

  const handleTabChange = useCallback((tab: string) => {
    startTabTransition(() => {
      setActiveTab(tab);
      setMountedTabs((prev) => {
        if (prev.has(tab)) return prev;
        const next = new Set(prev);
        next.add(tab);
        return next;
      });
    });
    if (scrollContainerRef.current) {
      scrollContainerRef.current.scrollTop = 0;
    }
  }, [setActiveTab]);

  const handleSelectStaff = useCallback((newStaffId: string, newStaffName: string) => {
    setStaffName(newStaffName);
    if (recordId) {
      handleChangeDoctor(newStaffId, newStaffName);
    }
  }, [recordId, handleChangeDoctor]);

  const handleFinalizeConfirm = useCallback(() => {
    handleFinalize();
    modals.setIsFinalizeConfirmOpen(false);
  }, [handleFinalize, modals]);

  return {
    historyItems,
    isDirty,
    hasLineIntegration,
    lstepStatus,
    scrollContainerRef,
    handleRegisterEstimateSave,
    user,
    currentRecord,
    recordFinalized,
    recordClinicId,
    clinicalPlan,
    treatments,
    billingConfirmation,
    isBillingConfirmationLoading,
    isBillingConfirmationError,
    staffName,
    modals,
    mountedTabs,
    isTabPending,
    dirtyFields,
    handleTabChange,
    handleSelectStaff,
    handleFinalizeConfirm,
  };
}

interface MedicalRecordFormReadyPageProps {
  recordId: string | undefined;
  selectedPet: Pet;
  form: MedicalRecordFormModel;
  canEdit: boolean;
  canSubmit: boolean;
  canDelete: boolean | undefined;
  isDeleting: boolean;
  onDeleteConfirm: () => void;
}

export function MedicalRecordFormReadyPage({
  recordId,
  selectedPet,
  form,
  canEdit,
  canSubmit,
  canDelete,
  isDeleting,
  onDeleteConfirm,
}: MedicalRecordFormReadyPageProps) {
  const ready = useMedicalRecordFormReadyState({ recordId, selectedPet, form });
  const formState = form.formState;

  return (
    <form action={form.formAction} className={LAYOUT.fullHeight}>
    <PageLayout
      title={recordId ? "カルテ編集" : "カルテ入力"}
      onBack={form.handleBack}
      resource={ResourceMedicalRecords}
      maxWidth={LAYOUT.pageContentMaxWidth.full}
      scrollContainerRef={ready.scrollContainerRef}
    >
      <NavigationBlocker when={ready.isDirty} />
      {form.autoCreateFailurePhase !== null ? (
        <MedicalRecordAutoCreateFailure
          failurePhase={form.autoCreateFailurePhase}
          isRetrying={form.isCreating}
          onRetry={form.retryAutoCreate}
        />
      ) : null}
      <UnifiedTabsRoot
        value={form.activeTab}
        onValueChange={ready.handleTabChange}
        className={form.activeTab === "問診" ? "flex-1 min-h-0" : undefined}
        ariaBusy={ready.isTabPending}
      >
      <MedicalRecordStickyHeader
        selectedPet={selectedPet}
        cohabitingPets={form.cohabitingPets}
        staffName={ready.staffName}
        visitType={form.visitType}
        visitCount={form.visitCount ?? 0}
        canEdit={canEdit}
        isNewRecord={form.isNewRecord}
        tabs={MEDICAL_RECORD_TAB_ITEMS}
        recordDate={form.recordDate}
        recordStatus={ready.currentRecord?.status}
        nextVisitDate={form.nextVisitDate}
        onVisitTypeChange={form.handleVisitTypeChange}
        onStaffClick={ready.modals.handleOpenStaffModal}
        onOwnerClick={ready.modals.handleOpenOwnerSearch}
        onDateChange={form.handleChangeDate}
        onNextVisitDatePatch={form.handleNextVisitDatePatch}
        onNextVisitDateValidChange={form.handleNextVisitDateValidChange}
        hasLineIntegration={ready.hasLineIntegration}
      />
      <fieldset
        disabled={ready.recordFinalized || !canSubmit}
        className="border-0 p-0 m-0 min-w-0"
        data-testid="medical-record-edit-lock"
      >
        {ready.recordFinalized ? (
          <div className={`mx-4 mt-3 rounded border ${C.borderMedium} ${C.bgPage} px-3 py-2 text-sm ${C.text60}`}>
            このカルテは確定済みのため編集できません。修正が必要な場合は下部の訂正追記（addendum）をご利用ください。
          </div>
        ) : null}
        <MedicalRecordTabsArea
          activeTab={form.activeTab}
          mountedTabs={ready.mountedTabs}
          isNewRecord={form.isNewRecord}
          recordId={recordId}
          selectedPet={selectedPet}
          chiefComplaint={form.chiefComplaint}
          chiefComplaintTypeId={form.chiefComplaintTypeId}
          treatmentPolicy={form.treatmentPolicy}
          historyItems={ready.historyItems}
          physicalExam={form.physicalExam}
          plan={form.plan}
          assessment={form.assessment}
          diagnosis1CategoryId={form.diagnosis1CategoryId}
          diagnosis1NameId={form.diagnosis1NameId}
          diagnosis2CategoryId={form.diagnosis2CategoryId}
          diagnosis2NameId={form.diagnosis2NameId}
          ownerDiscountRate={form.ownerDiscountRate ?? 0}
          nextVisitDate={form.nextVisitDate}
          hasLineIntegration={ready.hasLineIntegration}
          recommendationReason={form.recommendationReason}
          lstepStatus={ready.lstepStatus}
          recordStatus={ready.currentRecord?.status ?? ""}
          diagnosis1NameIdError={formState?.fieldErrors?.diagnosis1_name_id}
          onChiefComplaintChange={ready.dirtyFields.handleSetChiefComplaint}
          onChiefComplaintTypeIdChange={ready.dirtyFields.handleSetChiefComplaintTypeId}
          onTreatmentPolicyChange={ready.dirtyFields.handleSetTreatmentPolicy}
          onPhysicalExamChange={ready.dirtyFields.handleSetPhysicalExam}
          onPlanChange={ready.dirtyFields.handleSetPlan}
          onAssessmentChange={ready.dirtyFields.handleSetAssessment}
          onDiagnosis1CategoryIdChange={ready.dirtyFields.handleSetDiagnosis1CategoryId}
          onDiagnosis1NameIdChange={ready.dirtyFields.handleSetDiagnosis1NameId}
          onDiagnosis2CategoryIdChange={ready.dirtyFields.handleSetDiagnosis2CategoryId}
          onDiagnosis2NameIdChange={ready.dirtyFields.handleSetDiagnosis2NameId}
          onNextVisitDateChange={form.handleNextVisitDateChange}
          onNextVisitDateValidChange={form.handleNextVisitDateValidChange}
          onRecommendationReasonChange={form.setRecommendationReason}
          onRegisterEstimateSave={ready.handleRegisterEstimateSave}
          recordClinicId={ready.recordClinicId}
        />
      </fieldset>
      </UnifiedTabsRoot>

      <MedicalRecordFormReadyOverlays
        recordId={recordId}
        selectedPet={selectedPet}
        form={form}
        canEdit={canEdit}
        canSubmit={canSubmit}
        canDelete={canDelete}
        isDeleting={isDeleting}
        onDeleteConfirm={onDeleteConfirm}
        ready={ready}
      />
    </PageLayout>

    <MedicalRecordPrintArea
      isPrinting={ready.modals.isPrinting}
      isNewRecord={form.isNewRecord}
      recordId={recordId}
      doctorName={ready.staffName}
      recordDate={form.recordDate}
      pet={{
        name: selectedPet.name,
        species: selectedPet.species,
        ownerName: selectedPet.ownerName,
      }}
      clinic={ready.user?.clinic ?? undefined}
      chiefComplaint={form.chiefComplaint}
      treatmentPolicy={form.treatmentPolicy}
      physicalExam={ready.clinicalPlan?.physical_exam}
      diagnosisDetails={ready.clinicalPlan?.diagnosis_details}
      treatments={ready.treatments}
    />
    </form>
  );
}

function MedicalRecordFormReadyOverlays({
  recordId,
  selectedPet,
  form,
  canEdit,
  canSubmit,
  canDelete,
  isDeleting,
  onDeleteConfirm,
  ready,
}: MedicalRecordFormReadyPageProps & {
  ready: ReturnType<typeof useMedicalRecordFormReadyState>;
}) {
  return (
    <>
      {!form.isNewRecord ? (
        <MedicalRecordAddenda
          medicalRecordId={recordId ?? ""}
          canEdit={canEdit}
          recordStatus={ready.currentRecord?.status ?? ""}
        />
      ) : null}

      <MedicalRecordFloatingActions
        activeTab={form.activeTab}
        canDelete={canDelete ? !ready.recordFinalized : false}
        canEdit={canEdit}
        canSubmit={canSubmit}
        isNewRecord={form.isNewRecord}
        isCreating={form.isCreating}
        isSaving={form.isSaving}
        isFinalized={ready.recordFinalized}
        billingConfirmationStatus={ready.billingConfirmation?.status}
        isBillingConfirmationLoading={ready.isBillingConfirmationLoading}
        isBillingConfirmationError={ready.isBillingConfirmationError}
        onDeleteClick={() => ready.modals.setIsDeleteConfirmOpen(true)}
        onVitalsClick={() => ready.modals.setIsVitalsOpen(true)}
        onPrintClick={ready.modals.handlePrintClick}
        onFinalizeClick={() => ready.modals.setIsFinalizeConfirmOpen(true)}
      />

      <MedicalRecordFinalizeDialog
        open={ready.modals.isFinalizeConfirmOpen}
        isFinalizing={form.isFinalizeSaving}
        onClose={() => ready.modals.setIsFinalizeConfirmOpen(false)}
        onConfirm={ready.handleFinalizeConfirm}
      />

      <MedicalRecordFormModals
        isDeleteConfirmOpen={ready.modals.isDeleteConfirmOpen}
        selectedPetName={selectedPet.name}
        isDeleting={isDeleting}
        onCloseDeleteConfirm={() => ready.modals.setIsDeleteConfirmOpen(false)}
        onConfirmDelete={onDeleteConfirm}
        isNewRecord={form.isNewRecord}
        recordId={recordId}
        isVitalsOpen={ready.modals.isVitalsOpen}
        onVitalsOpenChange={ready.modals.setIsVitalsOpen}
        recordClinicId={ready.recordClinicId}
        isStaffModalOpen={ready.modals.isStaffModalOpen}
        staffName={ready.staffName}
        onSelectStaff={ready.handleSelectStaff}
        onStaffModalOpenChange={ready.modals.handleStaffModalOpenChange}
        isOwnerSearchOpen={ready.modals.isOwnerSearchOpen}
        onOwnerSearchOpenChange={ready.modals.setIsOwnerSearchOpen}
        selectedPetOwnerName={selectedPet.ownerName}
        onSelectOwner={form.requestOwnerChange}
        pendingOwnerChange={form.pendingOwnerChange}
        onCancelOwnerChange={form.cancelOwnerChange}
        onConfirmOwnerChange={form.confirmOwnerChange}
      />
    </>
  );
}
