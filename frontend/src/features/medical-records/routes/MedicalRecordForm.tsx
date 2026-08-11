// React/Framework
import {
  memo,
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  useTransition,
} from "react";
import { useParams, useNavigate } from "react-router";

// External
import { HeartPulse } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { Button } from "@/components/ui/button";
import { UnifiedTabsRoot } from "@/components/shared/UnifiedTabs";

// Relative
import { MedicalRecordAddenda } from "../components/MedicalRecordAddenda";
import { MedicalRecordAutoCreateFailure } from "../components/MedicalRecordAutoCreateFailure";
import { MedicalRecordStickyHeader, MedicalRecordTabsArea } from "../components/MedicalRecordFormPanels";
import {
  MedicalRecordFinalizeDialog,
  MedicalRecordFloatingActions,
  MedicalRecordPrintArea,
} from "../components/MedicalRecordFormActions";
import { MedicalRecordFormModals } from "../components/MedicalRecordFormModals";
import { MEDICAL_RECORD_TAB_ITEMS } from "./medical-record-form-model";
import { useMedicalRecordDirtyFields } from "../hooks/use-medical-record-dirty-fields";
import { useMedicalRecordFormModals } from "../hooks/use-medical-record-form-modals";
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
    cohabitingPets,
    isPetLoading,
    shouldRedirectToSelectPet,
    notFound,
    isReadLoading,
    isReadNotFound,
    isReadError,
    retryRead,
    handleBack,
    formAction,
    formState,
    isSaving,
    isFinalized,
    isCreating,
    autoCreateFailurePhase,
    retryAutoCreate,
    treatmentPlanItems: _treatmentPlanItems,
    setTreatmentPlanItems: _setTreatmentPlanItems,
    chiefComplaint,
    setChiefComplaint,
    chiefComplaintTypeId,
    setChiefComplaintTypeId,
    treatmentPolicy,
    setTreatmentPolicy,
    physicalExam,
    setPhysicalExam,
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
    handleFinalize,
    isFinalizeSaving,
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
    handleRegisterEstimateSave,
  } = useMedicalRecordPostSave({
    activeTab,
    formState,
    markClean,
  });

  const { user } = useAuth();
  const { canEdit, canCreate, canDelete } = usePermission("medical-records");
  const canSubmit = isNewRecord ? canCreate : canEdit;
  const canDeleteRef = useRef(canDelete);
  const selectedPetStatusRef = useRef(selectedPet?.status);
  useLayoutEffect(() => {
    canDeleteRef.current = canDelete;
  }, [canDelete]);
  useLayoutEffect(() => {
    selectedPetStatusRef.current = selectedPet?.status;
  }, [selectedPet?.status]);

  // addenda セクション用: キャッシュ共有のため追加ネットワーク要求なし
  const { data: currentRecord } = useGetMedicalRecord(recordId ?? "");
  // P2-15: 拠点横断で開いたカルテ（record.clinicId）の子リソースは、グローバル選択クリニックではなく
  // レコード自身の clinicId を X-Clinic-ID として送る必要がある。currentRecord 解決前は undefined —
  // クエリキーに clinicId を含めているため解決後に自動で正しいクリニックへ再フェッチされる。
  const recordClinicId = currentRecord?.clinicId;

  // 印刷用データ（React Query キャッシュ共有 — 子コンポーネントが既にフェッチ済み）
  const { data: clinicalPlan } = useGetClinicalPlan(recordId ?? "", recordClinicId);
  const { data: treatments = [] } = useGetTreatments(recordId ?? "", recordClinicId);

  // ローカル状態: 担当者（hookに追加するまでの暫定）
  const [staffName, setStaffName] = useState(() => user?.displayName ?? "");
  const {
    isDeleteConfirmOpen,
    setIsDeleteConfirmOpen,
    isFinalizeConfirmOpen,
    setIsFinalizeConfirmOpen,
    isVitalsOpen,
    setIsVitalsOpen,
    isPrinting,
    isStaffModalOpen,
    isOwnerSearchOpen,
    setIsOwnerSearchOpen,
    handleStaffModalOpenChange,
    handleOpenStaffModal,
    handleOpenOwnerSearch,
    handlePrintClick,
  } = useMedicalRecordFormModals();
  // 一度マウントしたタブを記録してhide/show方式で管理
  const [mountedTabs, setMountedTabs] = useState<Set<string>>(() => new Set(["問診"]));
  const [isTabPending, startTabTransition] = useTransition();

  const { mutate: deleteRecord, isPending: isDeleting } = useDeleteMedicalRecord(recordClinicId);

  const handleFinalizeConfirm = useCallback(() => {
    handleFinalize();
    setIsFinalizeConfirmOpen(false);
  }, [handleFinalize, setIsFinalizeConfirmOpen]);

  const handleDeleteConfirm = useCallback(() => {
    if (
      !recordId
      || canDeleteRef.current !== true
      || selectedPetStatusRef.current === "死亡"
    ) return;
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
  }, [recordId, deleteRecord, navigate, setIsDeleteConfirmOpen]);

  useEffect(() => {
    if (shouldRedirectToSelectPet) {
      navigate(paths.medicalRecords.selectPet.getHref());
    }
  }, [shouldRedirectToSelectPet, navigate]);

  // タブ切り替え: 一度開いたタブはhide/showで状態を維持する
  // BUG-MEDI-005: タブ切替時にスクロール位置をトップにリセット
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

  const {
    handleSetAssessment,
    handleSetChiefComplaint,
    handleSetChiefComplaintTypeId,
    handleSetDiagnosis1CategoryId,
    handleSetDiagnosis1NameId,
    handleSetDiagnosis2CategoryId,
    handleSetDiagnosis2NameId,
    handleSetPhysicalExam,
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
    setPhysicalExam,
    setPlan,
    setTreatmentPolicy,
  });

  const handleSelectStaff = useCallback((newStaffId: string, newStaffName: string) => {
    setStaffName(newStaffName);
    if (recordId) {
      handleChangeDoctor(newStaffId, newStaffName);
    }
  }, [recordId, handleChangeDoctor, setStaffName]);

  if (isReadLoading) {
    return (
      <PageLayout
        title="カルテ"
        onBack={handleBack}
        icon={<HeartPulse className={`${ICON.page} ${C.text}`} />}
        maxWidth={LAYOUT.pageContentMaxWidth.full}
      >
        <LoadingFallback />
      </PageLayout>
    );
  }

  // BUG-017: missing / other-clinic / forbidden → Not Found (non-disclosure), never blank
  if (notFound || isReadNotFound) {
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

  if (isReadError) {
    return (
      <PageLayout
        title="カルテ"
        onBack={handleBack}
        icon={<HeartPulse className={`${ICON.page} ${C.text}`} />}
        maxWidth={LAYOUT.pageContentMaxWidth.full}
      >
        <div className="space-y-3">
          <ErrorFallback message="カルテの取得に失敗しました" />
          {retryRead ? (
            <Button type="button" variant="outline" size="sm" onClick={retryRead}>
              再試行
            </Button>
          ) : null}
        </div>
      </PageLayout>
    );
  }

  if (isPetLoading) {
    return <LoadingFallback />;
  }

  // Edit route with loaded record but pet unresolved: never blank white page
  if (!selectedPet) {
    if (!isNewRecord) {
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
    return null;
  }

  // BUG-002: 死亡ペットへの /medical-records/new?petId=… 直叩きは編集フォームを出さない。
  // mutation 拒否だけでは UAT で【死亡】バナー付きフルフォームが残るため、UI を hard stop する。
  // BE medicalRecordDeceasedPetMessage と同文言。
  if (isNewRecord && selectedPet.status === "死亡") {
    return (
      <PageLayout
        title="カルテ入力"
        onBack={handleBack}
        icon={<HeartPulse className={`${ICON.page} ${C.text}`} />}
        resource={ResourceMedicalRecords}
        maxWidth={LAYOUT.pageContentMaxWidth.full}
      >
        <ErrorFallback message="死亡したペットは新規カルテを作成できません" />
      </PageLayout>
    );
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
      {autoCreateFailurePhase !== null ? (
        <MedicalRecordAutoCreateFailure
          failurePhase={autoCreateFailurePhase}
          isRetrying={isCreating}
          onRetry={retryAutoCreate}
        />
      ) : null}
      <UnifiedTabsRoot
        value={activeTab}
        onValueChange={handleTabChange}
        className={activeTab === "問診" ? "flex-1 min-h-0" : undefined}
        ariaBusy={isTabPending}
      >
      <MedicalRecordStickyHeader
        selectedPet={selectedPet}
        cohabitingPets={cohabitingPets}
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
      {/* SPEC-GAP (GAP-1): 確定済みカルテと非編集権限のカルテは編集不可 UI。BE は主要な子リソース
          (治療/検査/バイタル/処方/健診/画像) を 409 で拒否するが、UI が押せて
          エラーになるのは不可のため、タブ内の全編集導線をここで一括無効化する。
          個別コンポーネントへ disabled を都度配線すると新規フィールド追加時に
          ガード漏れが起きやすいため、fieldset の cascade を単一の強制点にする。 */}
      <fieldset disabled={isFinalized || !canSubmit} className="contents border-0 m-0 p-0">
        {isFinalized ? (
          <div className={`mx-4 mt-3 rounded border ${C.borderMedium} ${C.bgPage} px-3 py-2 text-sm ${C.text60}`}>
            このカルテは確定済みのため編集できません。修正が必要な場合は下部の訂正追記（addendum）をご利用ください。
          </div>
        ) : null}
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
          physicalExam={physicalExam}
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
          onPhysicalExamChange={handleSetPhysicalExam}
          onPlanChange={handleSetPlan}
          onAssessmentChange={handleSetAssessment}
          onDiagnosis1CategoryIdChange={handleSetDiagnosis1CategoryId}
          onDiagnosis1NameIdChange={handleSetDiagnosis1NameId}
          onDiagnosis2CategoryIdChange={handleSetDiagnosis2CategoryId}
          onDiagnosis2NameIdChange={handleSetDiagnosis2NameId}
          onNextVisitDateChange={handleNextVisitDateChange}
          onNextVisitDateValidChange={handleNextVisitDateValidChange}
          onRecommendationReasonChange={setRecommendationReason}
          onRegisterEstimateSave={handleRegisterEstimateSave}
          recordClinicId={recordClinicId}
        />
      </fieldset>
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
        canDelete={!!canDelete && !isFinalized}
        canEdit={canEdit}
        canSubmit={canSubmit}
        isNewRecord={isNewRecord}
        isCreating={isCreating}
        isSaving={isSaving}
        isFinalized={isFinalized}
        onDeleteClick={() => setIsDeleteConfirmOpen(true)}
        onVitalsClick={() => setIsVitalsOpen(true)}
        onPrintClick={handlePrintClick}
        onFinalizeClick={() => setIsFinalizeConfirmOpen(true)}
      />

      <MedicalRecordFinalizeDialog
        open={isFinalizeConfirmOpen}
        isFinalizing={isFinalizeSaving}
        onClose={() => setIsFinalizeConfirmOpen(false)}
        onConfirm={handleFinalizeConfirm}
      />

      <MedicalRecordFormModals
        isDeleteConfirmOpen={isDeleteConfirmOpen}
        selectedPetName={selectedPet.name}
        isDeleting={isDeleting}
        onCloseDeleteConfirm={() => setIsDeleteConfirmOpen(false)}
        onConfirmDelete={handleDeleteConfirm}
        isNewRecord={isNewRecord}
        recordId={recordId}
        isVitalsOpen={isVitalsOpen}
        onVitalsOpenChange={setIsVitalsOpen}
        recordClinicId={recordClinicId}
        isStaffModalOpen={isStaffModalOpen}
        staffName={staffName}
        onSelectStaff={handleSelectStaff}
        onStaffModalOpenChange={handleStaffModalOpenChange}
        isOwnerSearchOpen={isOwnerSearchOpen}
        onOwnerSearchOpenChange={setIsOwnerSearchOpen}
        selectedPetOwnerName={selectedPet?.ownerName}
        onSelectOwner={requestOwnerChange}
        pendingOwnerChange={pendingOwnerChange}
        onCancelOwnerChange={cancelOwnerChange}
        onConfirmOwnerChange={confirmOwnerChange}
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
