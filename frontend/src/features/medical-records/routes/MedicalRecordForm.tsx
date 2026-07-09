// React/Framework
import { memo, Suspense, useCallback, useEffect, useRef, useState } from "react";
import { useParams, useNavigate } from "react-router";

// External
import { HeartPulse } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { UnifiedTabsRoot } from "@/components/shared/UnifiedTabs";

// Relative
import { MedicalRecordAddenda } from "../components/MedicalRecordAddenda";
import { MedicalRecordStickyHeader, MedicalRecordTabsArea } from "../components/MedicalRecordFormPanels";
import {
  MedicalRecordDeleteDialog,
  MedicalRecordFloatingActions,
  MedicalRecordPrintArea,
} from "../components/MedicalRecordFormActions";
import { OwnerSearchModal, StaffSelectionModal, VitalsModal } from "./MedicalRecordLazyModals";
import { MEDICAL_RECORD_TAB_ITEMS } from "./medical-record-form-model";
import { useMedicalRecordDirtyFields } from "../hooks/use-medical-record-dirty-fields";
import { useMedicalRecordPostSave } from "../hooks/use-medical-record-post-save";
import { useMedicalRecordForm } from "../hooks/use-medical-record-form";
import { useGetMedicalRecord } from "../api/get-medical-record";
import { useGetPetMedicalHistory } from "../api/get-medical-records";
import { useGetClinicalPlan } from "../api/clinical-plan";
import { useGetTreatments } from "../api/treatments";
import { useDeleteMedicalRecord } from "../api/delete-medical-record";
import { handleApiError } from "@/lib/handle-api-error";
import { toast } from "sonner";
import { useAuth } from "@/hooks/use-auth";
import { usePermission } from "@/hooks/use-permission";
import { useGetOwnerLineTags } from "@/hooks/use-owner-line-tags";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { useTitle } from "@/hooks/use-title";
import { ResourceMedicalRecords } from "@/types/generated/models";

export const MedicalRecordForm = memo(function MedicalRecordForm() {
  const { id: recordId } = useParams();
  const navigate = useNavigate();
  const {
    isNewRecord,
    activeTab,
    setActiveTab,
    selectedPet,
    isPetLoading,
    shouldRedirectToSelectPet,
    notFound,
    handleBack,
    formAction,
    formState,
    isCreating,
    treatmentPlanItems: _treatmentPlanItems,
    setTreatmentPlanItems: _setTreatmentPlanItems,
    chiefComplaint,
    setChiefComplaint,
    chiefComplaintTypeId,
    setChiefComplaintTypeId,
    treatmentPolicy,
    setTreatmentPolicy,
    plan,
    setPlan,
    assessment,
    setAssessment,
    diagnosis1CategoryId,
    setDiagnosis1CategoryId,
    diagnosis1NameId,
    setDiagnosis1NameId,
    diagnosis2CategoryId,
    setDiagnosis2CategoryId,
    diagnosis2NameId,
    setDiagnosis2NameId,
    ownerDiscountRate,
    visitType,
    visitCount,
    handleChangeDoctor,
    handleVisitTypeChange,
    pendingOwnerChange,
    requestOwnerChange,
    confirmOwnerChange,
    cancelOwnerChange,
    nextVisitDate,
    handleNextVisitDateChange,
    handleNextVisitDatePatch,
    isNextVisitDateValid: _isNextVisitDateValid,
    handleNextVisitDateValidChange,
    recommendationReason,
    setRecommendationReason,
    recordDate,
    handleChangeDate,
    } = useMedicalRecordForm(recordId);

    useTitle(recordId ? `カルテ編集 (#${recordId})` : "カルテ入力");

    // FEAT-003: ペットの過去カルテ履歴を実APIから取得（現在のレコードは除外）
    const { historyItems } = useGetPetMedicalHistory(
      selectedPet?.id,
      recordId,
    );

    const { isDirty, markDirty, markClean } = useUnsavedChanges();

  const { data: ownerLineData } = useGetOwnerLineTags(selectedPet?.ownerId ?? "");
  const hasLineIntegration = (ownerLineData?.is_linked && !ownerLineData?.lstep_opt_out) ?? false;
  const lstepStatus = ownerLineData === undefined
    ? undefined
    : ownerLineData.lstep_opt_out
      ? ("opt-out" as const)
      : ownerLineData.is_linked
        ? ("synced" as const)
        : ("not-linked" as const);

  // BUG-MEDI-005: タブ切替時にスクロール位置をトップにリセットするための ref
  const scrollContainerRef = useRef<HTMLDivElement>(null);

  const {
    handleRegisterClinicalPlanSave,
    handleRegisterEstimateSave,
  } = useMedicalRecordPostSave({
    activeTab,
    formState,
    markClean,
  });

  const { user } = useAuth();
  const { canEdit, canCreate, canDelete } = usePermission("medical-records");
  const canSubmit = isNewRecord ? canCreate : canEdit;

  // 印刷用データ（React Query キャッシュ共有 — 子コンポーネントが既にフェッチ済み）
  const { data: clinicalPlan } = useGetClinicalPlan(recordId ?? "");
  const { data: treatments = [] } = useGetTreatments(recordId ?? "");
  // addenda セクション用: キャッシュ共有のため追加ネットワーク要求なし
  const { data: currentRecord } = useGetMedicalRecord(recordId ?? "");

  // ローカル状態: 担当者（hookに追加するまでの暫定）
  const [staffName, setStaffName] = useState(() => user?.displayName ?? "");
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [isVitalsOpen, setIsVitalsOpen] = useState(false);
  const [isPrinting, setIsPrinting] = useState(false);
  const [isStaffModalOpen, setIsStaffModalOpen] = useState(false);
  const [isOwnerSearchOpen, setIsOwnerSearchOpen] = useState(false);
  // 一度マウントしたタブを記録してhide/show方式で管理
  const [mountedTabs, setMountedTabs] = useState<Set<string>>(() => new Set(["問診"]));

  const { mutate: deleteRecord, isPending: isDeleting } = useDeleteMedicalRecord();

  const handleDeleteConfirm = useCallback(() => {
    if (!recordId) return;
    deleteRecord(recordId, {
      onSuccess: () => {
        toast.success("カルテを削除しました");
        setIsDeleteConfirmOpen(false);
        navigate(paths.medicalRecords.getHref());
      },
      onError: (error) => {
        handleApiError(error, "カルテ削除");
      },
    });
  }, [recordId, deleteRecord, navigate]);

  useEffect(() => {
    if (shouldRedirectToSelectPet) {
      navigate(paths.medicalRecords.selectPet.getHref());
    }
  }, [shouldRedirectToSelectPet, navigate]);

  // タブ切り替え: 一度開いたタブはhide/showで状態を維持する
  // BUG-MEDI-005: タブ切替時にスクロール位置をトップにリセット
  const handleTabChange = useCallback((tab: string) => {
    setActiveTab(tab);
    setMountedTabs((prev) => {
      if (prev.has(tab)) return prev;
      const next = new Set(prev);
      next.add(tab);
      return next;
    });
    if (scrollContainerRef.current) {
      scrollContainerRef.current.scrollTop = 0;
    }
  }, [setActiveTab]);

  const {
    handleSetAssessment,
    handleSetChiefComplaint,
    handleSetChiefComplaintTypeId,
    handleSetDiagnosis1CategoryId,
    handleSetDiagnosis1NameId,
    handleSetDiagnosis2CategoryId,
    handleSetDiagnosis2NameId,
    handleSetPlan,
    handleSetTreatmentPolicy,
  } = useMedicalRecordDirtyFields({
    markDirty,
    setAssessment,
    setChiefComplaint,
    setChiefComplaintTypeId,
    setDiagnosis1CategoryId,
    setDiagnosis1NameId,
    setDiagnosis2CategoryId,
    setDiagnosis2NameId,
    setPlan,
    setTreatmentPolicy,
  });

  const handleSelectStaff = useCallback((newStaffId: string, newStaffName: string) => {
    setStaffName(newStaffName);
    if (recordId) {
      handleChangeDoctor(newStaffId, newStaffName);
    }
  }, [recordId, handleChangeDoctor, setStaffName]);

  const handleStaffModalOpenChange = useCallback((open: boolean) => {
    setIsStaffModalOpen(open);
  }, []);

  const handleOpenStaffModal = useCallback(() => {
    setIsStaffModalOpen(true);
  }, []);

  const handleOpenOwnerSearch = useCallback(() => {
    setIsOwnerSearchOpen(true);
  }, []);

  const handlePrintClick = useCallback(() => {
    setIsPrinting(true);
    setTimeout(() => {
      window.print();
      setIsPrinting(false);
    }, 100);
  }, []);

  if (notFound) {
    return (
      <PageLayout
        title="カルテ"
        onBack={handleBack}
        icon={<HeartPulse className={`${ICON.page} ${C.text}`} />}
        maxWidth={LAYOUT.pageContentMaxWidth.full}
      >
        <ErrorFallback message="カルテが見つかりません" />
      </PageLayout>
    );
  }

  if (isPetLoading) {
    return <LoadingFallback />;
  }

  if (!selectedPet) {
    return null;
  }

  return (
    <form action={formAction} className={LAYOUT.fullHeight}>
    <PageLayout
      title={recordId ? "カルテ編集" : "カルテ入力"}
      onBack={handleBack}
      resource={ResourceMedicalRecords}
      maxWidth={LAYOUT.pageContentMaxWidth.full}
      scrollContainerRef={scrollContainerRef}
    >
      <NavigationBlocker when={isDirty} />
      <UnifiedTabsRoot value={activeTab} onValueChange={handleTabChange}>
      <MedicalRecordStickyHeader
        selectedPet={selectedPet}
        staffName={staffName}
        visitType={visitType}
        visitCount={visitCount ?? 0}
        canEdit={canEdit}
        isNewRecord={isNewRecord}
        tabs={MEDICAL_RECORD_TAB_ITEMS}
        recordDate={recordDate}
        recordStatus={currentRecord?.status}
        nextVisitDate={nextVisitDate}
        onVisitTypeChange={handleVisitTypeChange}
        onStaffClick={handleOpenStaffModal}
        onOwnerClick={handleOpenOwnerSearch}
        onDateChange={handleChangeDate}
        onNextVisitDatePatch={handleNextVisitDatePatch}
        onNextVisitDateValidChange={handleNextVisitDateValidChange}
        hasLineIntegration={hasLineIntegration}
      />
      <MedicalRecordTabsArea
        activeTab={activeTab}
        mountedTabs={mountedTabs}
        isNewRecord={isNewRecord}
        recordId={recordId}
        selectedPet={selectedPet}
        chiefComplaint={chiefComplaint}
        chiefComplaintTypeId={chiefComplaintTypeId}
        treatmentPolicy={treatmentPolicy}
        historyItems={historyItems}
        plan={plan}
        assessment={assessment}
        diagnosis1CategoryId={diagnosis1CategoryId}
        diagnosis1NameId={diagnosis1NameId}
        diagnosis2CategoryId={diagnosis2CategoryId}
        diagnosis2NameId={diagnosis2NameId}
        ownerDiscountRate={ownerDiscountRate ?? 0}
        nextVisitDate={nextVisitDate}
        hasLineIntegration={hasLineIntegration}
        recommendationReason={recommendationReason}
        lstepStatus={lstepStatus}
        recordStatus={currentRecord?.status ?? ""}
        diagnosis1NameIdError={formState.fieldErrors?.diagnosis1_name_id}
        onChiefComplaintChange={handleSetChiefComplaint}
        onChiefComplaintTypeIdChange={handleSetChiefComplaintTypeId}
        onTreatmentPolicyChange={handleSetTreatmentPolicy}
        onPlanChange={handleSetPlan}
        onAssessmentChange={handleSetAssessment}
        onDiagnosis1CategoryIdChange={handleSetDiagnosis1CategoryId}
        onDiagnosis1NameIdChange={handleSetDiagnosis1NameId}
        onDiagnosis2CategoryIdChange={handleSetDiagnosis2CategoryId}
        onDiagnosis2NameIdChange={handleSetDiagnosis2NameId}
        onNextVisitDateChange={handleNextVisitDateChange}
        onNextVisitDateValidChange={handleNextVisitDateValidChange}
        onRecommendationReasonChange={setRecommendationReason}
        onRegisterClinicalPlanSave={handleRegisterClinicalPlanSave}
        onRegisterEstimateSave={handleRegisterEstimateSave}
      />
      </UnifiedTabsRoot>

      {!isNewRecord ? (
        <MedicalRecordAddenda
          medicalRecordId={recordId ?? ""}
          canEdit={canEdit}
          recordStatus={currentRecord?.status ?? ""}
        />
      ) : null}

      <MedicalRecordFloatingActions
        activeTab={activeTab}
        canDelete={canDelete}
        canEdit={canEdit}
        canSubmit={canSubmit}
        isNewRecord={isNewRecord}
        isCreating={isCreating}
        onDeleteClick={() => setIsDeleteConfirmOpen(true)}
        onVitalsClick={() => setIsVitalsOpen(true)}
        onPrintClick={handlePrintClick}
      />

      <MedicalRecordDeleteDialog
        open={isDeleteConfirmOpen}
        petName={selectedPet.name}
        isDeleting={isDeleting}
        onClose={() => setIsDeleteConfirmOpen(false)}
        onConfirm={handleDeleteConfirm}
      />
      {/* Vitals Modal */}
      <Suspense fallback={null}>
        {!isNewRecord && recordId ? (
          <VitalsModal
            open={isVitalsOpen}
            onOpenChange={setIsVitalsOpen}
            medicalRecordId={recordId}
          />
        ) : null}
      </Suspense>

      {/* Staff Selection Modal */}
      <Suspense fallback={null}>
        <StaffSelectionModal
          open={isStaffModalOpen}
          selectedStaffName={staffName}
          onSelect={handleSelectStaff}
          onOpenChange={handleStaffModalOpenChange}
        />
      </Suspense>

      {/* Owner Search Modal (edit mode only) */}
      {!isNewRecord && recordId ? (
        <Suspense fallback={null}>
          <OwnerSearchModal
            open={isOwnerSearchOpen}
            onOpenChange={setIsOwnerSearchOpen}
            currentOwnerName={selectedPet?.ownerName}
            onSelect={requestOwnerChange}
          />
        </Suspense>
      ) : null}

      {/* BUG-373: 飼主変更 確認ダイアログ (discount_rate/membership_type 不一致時のみ表示) */}
      <ConfirmDialog
        open={!!pendingOwnerChange}
        onClose={cancelOwnerChange}
        onConfirm={confirmOwnerChange}
        title="飼主変更の確認"
        description="飼主によって値引率や会員区分が異なるため、今後の会計金額が変動する可能性があります。変更を続行してよろしいですか?"
        confirmLabel="続行"
        cancelLabel="キャンセル"
      />
    </PageLayout>

    <MedicalRecordPrintArea
      isPrinting={isPrinting}
      isNewRecord={isNewRecord}
      recordId={recordId}
      doctorName={staffName}
      recordDate={recordDate}
      pet={{
        name: selectedPet.name,
        species: selectedPet.species,
        ownerName: selectedPet.ownerName,
      }}
      clinic={user?.clinic ?? undefined}
      chiefComplaint={chiefComplaint}
      treatmentPolicy={treatmentPolicy}
      physicalExam={clinicalPlan?.physical_exam}
      diagnosisDetails={clinicalPlan?.diagnosis_details}
      treatments={treatments}
    />
    </form>
  );
});
