// React/Framework
import { lazy, memo, Suspense, useCallback, useEffect, useRef, useState } from "react";
import { useParams, useNavigate } from "react-router";

// External
import { HeartPulse, Printer, Trash2 } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { Button } from "@/components/ui/button";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { C, STYLE, ICON, LAYOUT } from "@/lib/design-tokens";
import { LoadingFallback } from "@/components/shared/DataStates";
import { UnifiedTabsRoot, UnifiedTabsList, UnifiedTabsContent } from "@/components/shared/UnifiedTabs";

// Relative
import { MedicalRecordInterview } from "../components/MedicalRecordInterview";
import { NextVisitDateField } from "../components/NextVisitDateField";
import { RecommendationReasonSelect } from "../components/RecommendationReasonSelect";
import { MedicalRecordDiagnosisPlan } from "../components/MedicalRecordDiagnosisPlan";
import { MedicalRecordTreatment } from "../components/MedicalRecordTreatment";
import { MedicalRecordVaccination } from "../components/MedicalRecordVaccination";
import { MedicalRecordImage } from "../components/MedicalRecordImage";
import { MedicalRecordEstimate } from "../components/MedicalRecordEstimate";
import { MedicalRecordBillCheck } from "../components/MedicalRecordBillCheck";
import { MedicalRecordExamination } from "../components/MedicalRecordExamination";
import { CheckupsTab } from "../components/CheckupsTab/CheckupsTab";
const StaffSelectionModal = lazy(() =>
  import("../components/StaffSelectionModal").then((m) => ({ default: m.StaffSelectionModal }))
);
const VitalsModal = lazy(() =>
  import("../components/VitalsModal").then((m) => ({ default: m.VitalsModal }))
);
const OwnerSearchModal = lazy(() =>
  import("@/components/shared/OwnerSearchModal/OwnerSearchModal").then((m) => ({ default: m.OwnerSearchModal }))
);
import { useMedicalRecordForm } from "../hooks/use-medical-record-form";
import { useGetPetMedicalHistory } from "../api/get-medical-records";
import { useGetClinicalPlan } from "../api/clinical-plan";
import { useGetTreatments } from "../api/treatments";
import { useDeleteMedicalRecord } from "../api/delete-medical-record";
import { handleApiError } from "@/lib/handle-api-error";
import { toast } from "sonner";
import { MedicalRecordPrintView } from "../components/MedicalRecordPrintView";
import { useAuth } from "@/hooks/use-auth";
import { usePermission } from "@/hooks/use-permission";
import { useGetOwnerLineTags } from "@/features/owners";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { useTitle } from "@/hooks/use-title";
import { ResourceMedicalRecords } from "@/types/generated/models";

const VISIT_TYPE_OPTIONS = ["初診", "再診", "緊急", "往診"] as const;

const MEDICAL_RECORD_TABS = [
  "問診",
  "診察/治療プラン",
  "治療",
  "予防接種",
  "定期健診",
  "検査",
  "画像",
  "見積書",
  "会計(医師確認)",
];

const TABS = MEDICAL_RECORD_TABS.map((tab) => ({
  value: tab,
  label: tab,
}));

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
    setVisitType,
    visitCount,
    handleChangeDoctor,
    pendingOwnerChange,
    requestOwnerChange,
    confirmOwnerChange,
    cancelOwnerChange,
    nextVisitDate,
    handleNextVisitDateChange,
    isNextVisitDateValid: _isNextVisitDateValid,
    handleNextVisitDateValidChange,
    recommendationReason,
    setRecommendationReason,
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

  // activeTab を保存時に正確に参照するための ref
  const activeTabRef = useRef(activeTab);
  useEffect(() => {
    activeTabRef.current = activeTab;
  }, [activeTab]);

  const clinicalPlanSaveRef = useRef<(() => Promise<void>) | null>(null);
  const estimateSaveRef = useRef<(() => Promise<void>) | null>(null);

  const handleRegisterClinicalPlanSave = useCallback((fn: () => Promise<void>) => {
    clinicalPlanSaveRef.current = fn;
  }, []);

  const handleRegisterEstimateSave = useCallback((fn: () => Promise<void>) => {
    estimateSaveRef.current = fn;
  }, []);

  // React 19 Action の成功を検知してタブ別サブ保存を実行
  useEffect(() => {
    if (!formState.success) return;

    const currentTab = activeTabRef.current;

    const doPostSave = async () => {
      try {
        if (currentTab === "診察/治療プラン") {
          await (clinicalPlanSaveRef.current?.() ?? Promise.resolve());
        } else if (currentTab === "見積書") {
          await (estimateSaveRef.current?.() ?? Promise.resolve());
        }
      } catch (error) {
        // サブ保存失敗: メインカルテは保存済み。ユーザーに通知して再保存を促す
        handleApiError(error, "データの保存");
      }
      markClean();
      // ナビゲーションなし: タブに留まる
    };

    doPostSave();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [formState.success, formState.timestamp]);

  const { user } = useAuth();
  const { canEdit, canCreate, canDelete } = usePermission("medical-records");
  const canSubmit = isNewRecord ? canCreate : canEdit;

  // 印刷用データ（React Query キャッシュ共有 — 子コンポーネントが既にフェッチ済み）
  const { data: clinicalPlan } = useGetClinicalPlan(recordId ?? "");
  const { data: treatments = [] } = useGetTreatments(recordId ?? "");

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

  const handleSetChiefComplaint = useCallback((val: string) => {
    markDirty();
    setChiefComplaint(val);
  }, [markDirty, setChiefComplaint]);

  const handleSetChiefComplaintTypeId = useCallback((id: number | null) => {
    markDirty();
    setChiefComplaintTypeId(id);
  }, [markDirty, setChiefComplaintTypeId]);

  const handleSetTreatmentPolicy = useCallback((val: string) => {
    markDirty();
    setTreatmentPolicy(val);
  }, [markDirty, setTreatmentPolicy]);

  const handleSetPlan = useCallback((val: string) => {
    markDirty();
    setPlan(val);
  }, [markDirty, setPlan]);

  const handleSetAssessment = useCallback((val: string) => {
    markDirty();
    setAssessment(val);
  }, [markDirty, setAssessment]);

  const handleSetDiagnosis1CategoryId = useCallback((id: number | null) => {
    markDirty();
    setDiagnosis1CategoryId(id);
  }, [markDirty, setDiagnosis1CategoryId]);

  const handleSetDiagnosis1NameId = useCallback((id: number | null) => {
    markDirty();
    setDiagnosis1NameId(id);
  }, [markDirty, setDiagnosis1NameId]);

  const handleSetDiagnosis2CategoryId = useCallback((id: number | null) => {
    markDirty();
    setDiagnosis2CategoryId(id);
  }, [markDirty, setDiagnosis2CategoryId]);

  const handleSetDiagnosis2NameId = useCallback((id: number | null) => {
    markDirty();
    setDiagnosis2NameId(id);
  }, [markDirty, setDiagnosis2NameId]);

  const handleSelectStaff = useCallback((newStaffId: string, newStaffName: string) => {
    setStaffName(newStaffName);
    if (recordId) {
      handleChangeDoctor(newStaffId, newStaffName);
    }
  }, [recordId, handleChangeDoctor, setStaffName]);

  const handleStaffModalOpenChange = useCallback((open: boolean) => {
    setIsStaffModalOpen(open);
  }, []);

  const handleVisitTypeCycle = useCallback(() => {
    setVisitType((prev) => {
      const idx = VISIT_TYPE_OPTIONS.indexOf(prev as typeof VISIT_TYPE_OPTIONS[number]);
      return VISIT_TYPE_OPTIONS[(idx + 1) % VISIT_TYPE_OPTIONS.length];
    });
  }, [setVisitType]);

  const handleOpenStaffModal = useCallback(() => {
    setIsStaffModalOpen(true);
  }, []);

  const handleOpenOwnerSearch = useCallback(() => {
    setIsOwnerSearchOpen(true);
  }, []);

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
      maxWidth="max-w-[1440px]"
      scrollContainerRef={scrollContainerRef}
    >
      <NavigationBlocker when={isDirty} />
      <UnifiedTabsRoot value={activeTab} onValueChange={handleTabChange}>
      {/* Sticky Header: Patient Info + Tabs */}
      <div className={`sticky top-0 z-10 ${C.bgPage}`}>
        {/* Patient Info Card */}
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
          onReservationTypeClick={handleVisitTypeCycle}
          onStaffClick={canEdit ? handleOpenStaffModal : undefined}
          onOwnerClick={!isNewRecord ? handleOpenOwnerSearch : undefined}
          petDetails={`${selectedPet.birthDate ? `${selectedPet.birthDate}生` : ""} / ${selectedPet.species}`}
          insuranceName={selectedPet.insuranceName || "保険情報未登録"}
          insuranceDetails={selectedPet.insuranceDetails || "-"}
          nextVisitDate="-"
          nextVisitContent="-"
          visitCount={visitCount}
          sticky={false}
        />

        {/* Tabs */}
        <div className={`flex shrink-0 overflow-x-auto ${C.bgPage}`}>
          <UnifiedTabsList items={TABS} />
        </div>
      </div>

      {/* Content Area */}
      <div className={`mt-4 ${LAYOUT.fullHeight}`}>
        {mountedTabs.has("問診") ? (
          <UnifiedTabsContent value="問診">
            <div className={`${LAYOUT.fullHeight} ${activeTab === "問診" ? "" : "hidden"}`}>
              <MedicalRecordInterview
                chiefComplaint={chiefComplaint}
                setChiefComplaint={handleSetChiefComplaint}
                chiefComplaintTypeId={chiefComplaintTypeId}
                setChiefComplaintTypeId={handleSetChiefComplaintTypeId}
                treatmentPolicy={treatmentPolicy}
                setTreatmentPolicy={handleSetTreatmentPolicy}
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
                setPlan={handleSetPlan}
                assessment={assessment}
                setAssessment={handleSetAssessment}
                diagnosis1CategoryId={diagnosis1CategoryId}
                setDiagnosis1CategoryId={handleSetDiagnosis1CategoryId}
                diagnosis1NameId={diagnosis1NameId}
                setDiagnosis1NameId={handleSetDiagnosis1NameId}
                diagnosis2CategoryId={diagnosis2CategoryId}
                setDiagnosis2CategoryId={handleSetDiagnosis2CategoryId}
                diagnosis2NameId={diagnosis2NameId}
                setDiagnosis2NameId={handleSetDiagnosis2NameId}
                medicalRecordId={recordId}
                ownerDiscountRate={ownerDiscountRate}
                onRegisterClinicalPlanSave={handleRegisterClinicalPlanSave}
                diagnosis1NameIdError={formState.fieldErrors?.diagnosis1_name_id}
              />
              <div className="px-4 pb-4 mt-4 flex flex-col gap-6">
                <NextVisitDateField
                  value={nextVisitDate}
                  onChange={handleNextVisitDateChange}
                  onValidationChange={handleNextVisitDateValidChange}
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
                    onChange={setRecommendationReason}
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
              <MedicalRecordTreatment
                medicalRecordId={recordId ?? ""}
                isNewRecord={isNewRecord}
              />
            </div>
          </UnifiedTabsContent>
        ) : null}
        {mountedTabs.has("予防接種") ? (
          <UnifiedTabsContent value="予防接種">
            <div className={`${LAYOUT.fullHeight} ${activeTab === "予防接種" ? "" : "hidden"}`}>
              <MedicalRecordVaccination petId={selectedPet?.id} lstepStatus={lstepStatus} />
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
                <CheckupsTab medicalRecordId={recordId} lstepStatus={lstepStatus} />
              )}
            </div>
          </UnifiedTabsContent>
        ) : null}
        {mountedTabs.has("検査") ? (
          <UnifiedTabsContent value="検査">
            <div className={`${LAYOUT.fullHeight} ${activeTab === "検査" ? "" : "hidden"}`}>
              <MedicalRecordExamination isNewRecord={isNewRecord} petId={selectedPet?.id} medicalRecordId={recordId} />
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
              <MedicalRecordEstimate isNewRecord={isNewRecord} ownerDiscountRate={ownerDiscountRate} medicalRecordId={recordId} onRegisterSave={handleRegisterEstimateSave} />
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
      </UnifiedTabsRoot>

      {/* Floating Save / Delete Buttons */}
      {activeTab !== "会計(医師確認)" ? (
        <div className="fixed bottom-6 right-6 z-50 flex gap-2">
          {canDelete && !isNewRecord && activeTab === "問診" ? (
            <Button
              type="button"
              variant="ghost-danger"
              onClick={() => setIsDeleteConfirmOpen(true)}
              className={`border ${C.borderDanger} h-10 text-sm px-4`}
            >
              <Trash2 className={ICON.action} />
              削除
            </Button>
          ) : null}
          {activeTab !== "見積書" && canEdit ? (
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsVitalsOpen(true)}
              disabled={isNewRecord}
              title={isNewRecord ? "カルテを保存してから利用できます" : undefined}
              className="h-10 text-sm px-4"
            >
              <HeartPulse className={ICON.action} />
              バイタル記録
            </Button>
          ) : null}
          {!isNewRecord ? (
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                setIsPrinting(true);
                // ブラウザ印刷ダイアログを非同期で開く（DOM更新を待つ）
                setTimeout(() => {
                  window.print();
                  setIsPrinting(false);
                }, 100);
              }}
              className="h-10 text-sm px-4"
            >
              <Printer className={ICON.action} />
              印刷
            </Button>
          ) : null}
          {canSubmit ? (
            <SubmitButton
              className={`${STYLE.btnPrimary} px-5`}
              disabled={isCreating}
            >
              {isCreating ? "カルテ作成中..." : "保存"}
            </SubmitButton>
          ) : null}
        </div>
      ) : null}

      {/* Delete confirm placeholder — connects to isDeleteConfirmOpen state */}
      {isDeleteConfirmOpen ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          onClick={() => setIsDeleteConfirmOpen(false)}
        >
          <div
            className={`${C.bgWhite} rounded-lg shadow-lg p-6 w-[400px]`}
            onClick={(e) => e.stopPropagation()}
          >
            <p className={`text-base font-medium ${C.text} mb-2`}>カルテを削除しますか？</p>
            <p className={`text-sm ${C.text60} mb-6`}>
              {selectedPet.name}のカルテデータを削除します。この操作は元に戻せません。
            </p>
            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setIsDeleteConfirmOpen(false)}
                disabled={isDeleting}
              >
                キャンセル
              </Button>
              <Button
                type="button"
                className={`${STYLE.btnDanger}`}
                onClick={handleDeleteConfirm}
                disabled={isDeleting}
              >
                {isDeleting ? "削除中..." : "削除する"}
              </Button>
            </div>
          </div>
        </div>
      ) : null}
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

    {/* Print area — hidden on screen, visible on print */}
    {isPrinting && !isNewRecord && selectedPet ? (
      <div className={`hidden print:block fixed inset-0 ${C.bgWhite} z-[9999]`}>
        <style type="text/css" media="print">
          {`@page { size: A4 portrait; margin: 15mm; } body { margin: 0; -webkit-print-color-adjust: exact; }`}
        </style>
        <MedicalRecordPrintView
          recordNo={recordId}
          date={new Date().toLocaleDateString("ja-JP")}
          doctorName={staffName}
          pet={{
            name: selectedPet.name,
            species: selectedPet.species,
            ownerName: selectedPet.ownerName,
          }}
          clinic={{
            name: user?.clinic?.name,
            address: user?.clinic?.address,
            phone: user?.clinic?.phoneNumber,
          }}
          chiefComplaint={chiefComplaint}
          treatmentPolicy={treatmentPolicy}
          physicalExam={clinicalPlan?.physical_exam}
          diagnosisDetails={clinicalPlan?.diagnosis_details}
          treatments={treatments}
        />
      </div>
    ) : null}
    </form>
  );
});
