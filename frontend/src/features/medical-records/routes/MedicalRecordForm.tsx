// React/Framework
import { lazy, memo, Suspense, useCallback, useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router";

// External
import { HeartPulse, Trash2 } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { Button } from "@/components/ui/button";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { C, STYLE } from "@/lib/design-tokens";

// Relative
import { MedicalRecordInterview } from "../components/MedicalRecordInterview";
import { MedicalRecordDiagnosisPlan } from "../components/MedicalRecordDiagnosisPlan";
import { MedicalRecordTreatment } from "../components/MedicalRecordTreatment";
import { MedicalRecordVaccination } from "../components/MedicalRecordVaccination";
import { MedicalRecordImage } from "../components/MedicalRecordImage";
import { MedicalRecordEstimate } from "../components/MedicalRecordEstimate";
import { MedicalRecordBillCheck } from "../components/MedicalRecordBillCheck";
import { MedicalRecordExamination } from "../components/MedicalRecordExamination";
import { CheckupsTab } from "../components/CheckupsTab";
import { StaffSelectionModal } from "../components/StaffSelectionModal";
const VitalsModal = lazy(() =>
  import("../components/VitalsModal").then((m) => ({ default: m.VitalsModal }))
);
const OwnerSearchModal = lazy(() =>
  import("@/components/shared/OwnerSearchModal/OwnerSearchModal").then((m) => ({ default: m.OwnerSearchModal }))
);
import { useMedicalRecordForm } from "../hooks/use-medical-record-form";
import { useAuth } from "@/features/auth";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";

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
    handleSave,
    isSaving,
    treatmentPlanItems,
    setTreatmentPlanItems,
    treatmentCompletedItems,
    chiefComplaint,
    setChiefComplaint,
    chiefComplaintCategoryId,
    setChiefComplaintCategoryId,
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
    handleChangeOwner,
  } = useMedicalRecordForm(recordId);

  const { user } = useAuth();

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  // ローカル状態: 担当者・診療種別（hookに追加するまでの暫定）
  const [staffName, setStaffName] = useState(() => user?.displayName ?? "");
  const [serviceType, setServiceType] = useState("診療");
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [isVitalsOpen, setIsVitalsOpen] = useState(false);
  const [isStaffModalOpen, setIsStaffModalOpen] = useState(false);
  const [isOwnerSearchOpen, setIsOwnerSearchOpen] = useState(false);
  // 一度マウントしたタブを記録してhide/show方式で管理
  const [mountedTabs, setMountedTabs] = useState<Set<string>>(() => new Set(["問診"]));

  useEffect(() => {
    if (shouldRedirectToSelectPet) {
      navigate(paths.medicalRecords.selectPet.getHref());
    }
  }, [shouldRedirectToSelectPet, navigate]);

  // Tab definitions
  const tabs = [
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

  // タブ切り替え: 一度開いたタブはhide/showで状態を維持する
  const handleTabChange = useCallback((tab: string) => {
    setActiveTab(tab);
    setMountedTabs((prev) => {
      if (prev.has(tab)) return prev;
      const next = new Set(prev);
      next.add(tab);
      return next;
    });
  }, [setActiveTab]);

  const handleSetChiefComplaint = useCallback((val: string) => {
    markDirty();
    setChiefComplaint(val);
  }, [markDirty, setChiefComplaint]);

  const handleSetChiefComplaintCategoryId = useCallback((id: number | null) => {
    markDirty();
    setChiefComplaintCategoryId(id);
  }, [markDirty, setChiefComplaintCategoryId]);

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

  const handleSetTreatmentPlanItems = useCallback<typeof setTreatmentPlanItems>((items) => {
    markDirty();
    setTreatmentPlanItems(items);
  }, [markDirty, setTreatmentPlanItems]);

  const handleSaveClick = useCallback(() => {
    markClean();
    handleSave();
  }, [markClean, handleSave]);

  const handleSelectStaff = useCallback((newStaffName: string) => {
    setStaffName(newStaffName);
  }, []);

  const handleStaffModalOpenChange = useCallback((open: boolean) => {
    setIsStaffModalOpen(open);
  }, []);

  if (isPetLoading) {
    return null;
  }

  if (!selectedPet) {
    return null;
  }

  return (
    <PageLayout
      title={recordId ? "カルテ編集" : "カルテ入力"}
      onBack={handleBack}
      maxWidth="max-w-[1440px]"
    >
      <NavigationBlocker when={isDirty} />
      {/* Sticky Header: Patient Info + Tabs */}
      <div className={`sticky top-0 z-10 ${C.bgPage}`}>
        {/* Patient Info Card */}
        <PatientInfoCard
          ownerName={selectedPet.ownerName}
          petName={`${selectedPet.name}${selectedPet.species ? `(${selectedPet.species})` : ""}`}
          petNumber={selectedPet.petNumber || selectedPet.id}
          weight={selectedPet.weight || "-"}
          staffName={staffName}
          serviceType={serviceType}
          serviceTypeLabel="診療種別"
          onServiceTypeClick={() => setServiceType(serviceType)}
          onStaffClick={() => setIsStaffModalOpen(true)}
          onOwnerClick={!isNewRecord ? () => setIsOwnerSearchOpen(true) : undefined}
          petDetails={`${selectedPet.birthDate ? `${selectedPet.birthDate}生` : ""} / ${selectedPet.species}`}
          insuranceName={selectedPet.insuranceName || "保険情報未登録"}
          insuranceDetails={selectedPet.insuranceDetails || "-"}
          nextVisitDate="-"
          nextVisitContent="-"
          sticky={false}
        />

        {/* Tabs */}
        <div className={`flex shrink-0 border-b ${C.borderMedium} overflow-x-auto ${C.bgPage}`}>
          {tabs.map((tab) => (
            <button
              key={tab}
              onClick={() => handleTabChange(tab)}
              className={`flex items-center px-4 h-11 text-base font-medium transition-colors relative whitespace-nowrap ${
                activeTab === tab
                  ? C.text
                  : `${C.text60} ${C.hoverText}`
              }`}
            >
              {tab}
              {activeTab === tab ? (
                <div className={`absolute bottom-0 left-0 w-full h-[2px] ${C.bgPrimary}`} />
              ) : null}
            </button>
          ))}
        </div>
      </div>

      {/* Content Area */}
      <div className="mt-4 flex-1 flex flex-col">
        {mountedTabs.has("問診") ? (
          <div className={`flex-1 min-h-0 flex flex-col${activeTab === "問診" ? "" : " hidden"}`}>
            <MedicalRecordInterview
              chiefComplaint={chiefComplaint}
              setChiefComplaint={handleSetChiefComplaint}
              chiefComplaintCategoryId={chiefComplaintCategoryId}
              setChiefComplaintCategoryId={handleSetChiefComplaintCategoryId}
              treatmentPolicy={treatmentPolicy}
              setTreatmentPolicy={handleSetTreatmentPolicy}
            />
          </div>
        ) : null}
        {mountedTabs.has("診察/治療プラン") ? (
          <div style={{ display: activeTab === "診察/治療プラン" ? "block" : "none" }}>
            <MedicalRecordDiagnosisPlan
              isNewRecord={isNewRecord}
              items={treatmentPlanItems}
              setItems={handleSetTreatmentPlanItems}
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
            />
          </div>
        ) : null}
        {mountedTabs.has("治療") ? (
          <div style={{ display: activeTab === "治療" ? "block" : "none" }}>
            <MedicalRecordTreatment
              medicalRecordId={recordId ?? ""}
              isNewRecord={isNewRecord}
            />
          </div>
        ) : null}
        {mountedTabs.has("予防接種") ? (
          <div style={{ display: activeTab === "予防接種" ? "block" : "none" }}>
            <MedicalRecordVaccination petId={selectedPet?.id} />
          </div>
        ) : null}
        {mountedTabs.has("定期健診") ? (
          <div style={{ display: activeTab === "定期健診" ? "block" : "none" }}>
            {isNewRecord || !recordId ? (
              <div className={`flex items-center justify-center h-48 text-sm ${C.text40}`}>
                カルテを保存してから使用できます
              </div>
            ) : (
              <CheckupsTab medicalRecordId={recordId} />
            )}
          </div>
        ) : null}
        {mountedTabs.has("検査") ? (
          <div style={{ display: activeTab === "検査" ? "block" : "none" }}>
            <MedicalRecordExamination isNewRecord={isNewRecord} petId={selectedPet?.id} />
          </div>
        ) : null}
        {mountedTabs.has("画像") ? (
          <div style={{ display: activeTab === "画像" ? "block" : "none" }}>
            <MedicalRecordImage isNewRecord={isNewRecord} medicalRecordId={recordId} />
          </div>
        ) : null}
        {mountedTabs.has("見積書") ? (
          <div style={{ display: activeTab === "見積書" ? "block" : "none" }}>
            <MedicalRecordEstimate isNewRecord={isNewRecord} ownerDiscountRate={ownerDiscountRate} medicalRecordId={recordId} />
          </div>
        ) : null}
        {mountedTabs.has("会計(医師確認)") ? (
          <div style={{ display: activeTab === "会計(医師確認)" ? "block" : "none" }}>
            <MedicalRecordBillCheck
              isNewRecord={isNewRecord}
              medicalRecordId={recordId}
              ownerDiscountRate={ownerDiscountRate}
            />
          </div>
        ) : null}
      </div>

      {/* Floating Save / Delete Buttons */}
      {activeTab !== "会計(医師確認)" ? (
        <div className="fixed bottom-6 right-6 z-50 flex gap-2">
          {!isNewRecord && activeTab === "問診" ? (
            <Button
              variant="ghost-danger"
              onClick={() => setIsDeleteConfirmOpen(true)}
              className={`border ${C.borderDanger} h-10 text-sm px-4`}
            >
              <Trash2 className="h-4 w-4" />
              削除
            </Button>
          ) : null}
          {activeTab !== "見積書" ? (
            <Button
              variant="outline"
              onClick={() => setIsVitalsOpen(true)}
              disabled={isNewRecord}
              title={isNewRecord ? "カルテを保存してから利用できます" : undefined}
              className="h-10 text-sm px-4"
            >
              <HeartPulse className="h-4 w-4" />
              バイタル記録
            </Button>
          ) : null}
          <Button
            onClick={handleSaveClick}
            disabled={isSaving}
            className={`${STYLE.btnPrimary} px-5`}
          >
            {isSaving ? "保存中..." : "保存"}
          </Button>
        </div>
      ) : null}

      {/* Delete confirm placeholder — connects to isDeleteConfirmOpen state */}
      {isDeleteConfirmOpen ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          onClick={() => setIsDeleteConfirmOpen(false)}
        >
          <div
            className="bg-white rounded-lg shadow-lg p-6 w-[400px]"
            onClick={(e) => e.stopPropagation()}
          >
            <p className={`text-base font-medium ${C.text} mb-2`}>カルテを削除しますか？</p>
            <p className={`text-sm ${C.text60} mb-6`}>
              {selectedPet.name}のカルテデータを削除します。この操作は元に戻せません。
            </p>
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={() => setIsDeleteConfirmOpen(false)}>
                キャンセル
              </Button>
              <Button
                className={`${STYLE.btnDanger}`}
                onClick={() => setIsDeleteConfirmOpen(false)}
              >
                削除
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
      <StaffSelectionModal
        open={isStaffModalOpen}
        selectedStaffName={staffName}
        onSelect={handleSelectStaff}
        onOpenChange={handleStaffModalOpenChange}
      />

      {/* Owner Search Modal (edit mode only) */}
      {!isNewRecord && recordId ? (
        <Suspense fallback={null}>
          <OwnerSearchModal
            open={isOwnerSearchOpen}
            onOpenChange={setIsOwnerSearchOpen}
            currentOwnerName={selectedPet?.ownerName}
            onSelect={handleChangeOwner}
          />
        </Suspense>
      ) : null}
    </PageLayout>
  );
});
