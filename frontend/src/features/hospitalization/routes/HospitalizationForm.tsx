// React/Framework
import { useEffect, useState, useCallback, useTransition } from "react";
import { useNavigate, useParams, useLocation, useSearchParams } from "react-router";

// External
import { FileText, Trash2, MessageSquare, AlertCircle } from "lucide-react";
import { handleApiError } from "@/lib/handle-api-error";

// Internal
import { Button } from "@/components/ui/button";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { useAuth, usePermission } from "@/features/auth";

// Relative
import { useHospitalizationForm } from "../hooks/use-hospitalization-form";
import { useDeleteHospitalization } from "../api/delete-hospitalization";
import { paths } from "@/config/paths";
import { useMasterItems } from "@/hooks/use-master-items";
import { HospitalizationBasicInfo } from "../components/HospitalizationBasicInfo";
import { HospitalizationNoteCard } from "../components/HospitalizationNoteCard";
import { HospitalizationTreatmentTable } from "../components/HospitalizationTreatmentTable";
import { HospitalizationCostSummary } from "../components/HospitalizationCostSummary";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { LoadingFallback } from "@/components/shared/DataStates/DataStates";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { ResourceHospitalization } from "@/types/generated/models";

export function HospitalizationForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const { id: hospitalizationId } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  
  const { data: cageItems } = useMasterItems("cage");

  const { user } = useAuth();
  const { canEdit, canCreate, canDelete } = usePermission("hospitalization");
  const canSubmit = hospitalizationId ? canEdit : canCreate;
  const deleteMutation = useDeleteHospitalization();
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);
  const [isDeletePending, startDeleteTransition] = useTransition();

  const {
      isEdit,
      formData,
      handleFormDataChange: handleFormDataChangeRaw,
      treatmentPlans,
      addTreatmentPlan,
      removeTreatmentPlan,
      updateTreatmentPlan,
      globalDiscount,
      setGlobalDiscount,
      globalDiscountAmount,
      setGlobalDiscountAmount,
      calculateTotals,
      petSelection,
      formAction,
      formState,
  } = useHospitalizationForm(hospitalizationId);

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  // rerender-dependencies: location.state（オブジェクト）から primitive を抽出して deps に使用
  const locationFrom = location.state?.from as string | undefined;

  // React 19 Action の成功を検知して遷移
  useEffect(() => {
    if (formState.success) {
      markClean();
      if (locationFrom) {
        navigate(locationFrom);
      } else {
        navigate(paths.hospitalization.getHref());
      }
    }
  }, [formState.success, formState.timestamp, navigate, markClean, locationFrom]);

  // エラー発生時に最初のエラーフィールドにフォーカス
  useEffect(() => {
    if (formState.fieldErrors && Object.keys(formState.fieldErrors).length > 0) {
      const firstErrorKey = Object.keys(formState.fieldErrors)[0];
      const element = document.getElementById(firstErrorKey);
      if (element) {
        element.focus();
        element.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }
  }, [formState.fieldErrors, formState.timestamp]);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];
  const totals = calculateTotals();

  const handleBack = useCallback(() => {
    if (locationFrom) {
        navigate(locationFrom);
    } else {
        navigate(paths.hospitalization.getHref());
    }
  }, [locationFrom, navigate]);

  const handleDelete = useCallback(() => {
    if (!hospitalizationId) return;
    startDeleteTransition(() => {
      deleteMutation.mutate(hospitalizationId, {
        onSuccess: () => {
          navigate(paths.hospitalization.getHref());
        },
        onError: (error) => handleApiError(error, "削除"),
      });
    });
  }, [hospitalizationId, deleteMutation, navigate]);

  const handleFormChange = useCallback((updates: Parameters<typeof handleFormDataChangeRaw>[0]) => {
    markDirty();
    handleFormDataChangeRaw(updates);
  }, [markDirty, handleFormDataChangeRaw]);

  const handleGlobalDiscountChange = useCallback((val: number) => {
    markDirty();
    setGlobalDiscount(val);
  }, [markDirty, setGlobalDiscount]);

  const handleGlobalDiscountAmountChange = useCallback((val: number) => {
    markDirty();
    setGlobalDiscountAmount(val);
  }, [markDirty, setGlobalDiscountAmount]);

  useEffect(() => {
    if (!selectedPet && !isEdit && !petId) {
        navigate(paths.hospitalization.selectPet.getHref());
    }
  }, [selectedPet, isEdit, navigate, petId]);

  if (!selectedPet && !isEdit && petId) return <LoadingFallback />;
  if (!selectedPet && !isEdit) return null;

  return (
    <>
    <form action={formAction}>
    <PageLayout
      title={hospitalizationId ? "入院編集" : "入院登録"}
      onBack={handleBack}
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      resource={ResourceHospitalization}
      maxWidth="max-w-[1400px]"
      headerAction={
        <div className="flex gap-2">
            {hospitalizationId ? (
                <>
                  <Button
                    variant="outline"
                    type="button"
                    className={`gap-2 h-10 text-sm px-4 ${C.text}`}
                    onClick={() => navigate(`/hospitalization/${hospitalizationId}`)}
                  >
                    <FileText className={ICON.action} />
                    デイリーカルテ
                  </Button>
                  {canDelete ? (
                    <Button
                      variant="ghost"
                      type="button"
                      className={`${STYLE.btnDangerGhost} h-10 text-sm px-4`}
                      onClick={() => setIsDeleteConfirmOpen(true)}
                    >
                      <Trash2 className={`mr-1.5 ${ICON.action}`} />
                      削除
                    </Button>
                  ) : null}
                </>
            ) : null}
            {canSubmit ? (
              <SubmitButton
              className={`${C.bgAccent} ${C.bgAccentHover} ${C.textWhite} rounded-[6px] h-10 text-sm px-4`}
              >
              {hospitalizationId ? "更新" : "登録"}
              </SubmitButton>
            ) : null}
        </div>
      }
    >
        <NavigationBlocker when={isDirty} />
        <fieldset disabled={!canSubmit} className="border-0 p-0 m-0 min-w-0">
        {/* Patient Info Card */}
        {selectedPet ? (
            <PatientInfoCard
              ownerName={selectedPet.ownerName}
              petName={`${selectedPet.name}${selectedPet.species ? `(${selectedPet.species})` : ""}`}
              petNumber={selectedPet.petNumber || selectedPet.id}
              weight={selectedPet.weight || "-"}
              staffName={user?.displayName ?? ""}
              serviceType={formData.hospitalizationType}
              petDetails={`${selectedPet.birthDate ? `${selectedPet.birthDate}生` : ""} / ${selectedPet.species}`}
              insuranceName={selectedPet.insuranceName || "保険情報未登録"}
              insuranceDetails={selectedPet.insuranceDetails || "-"}
              nextVisitDate="-"
              nextVisitContent="-"
            />
        ) : null}

        {/* Main Form Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 mb-3">
          {/* Left Column - Basic Info */}
          <HospitalizationBasicInfo
            formData={formData}
            onChange={handleFormChange}
            cageItems={cageItems}
          />

          {/* Middle Column - 飼主からのリクエスト */}
          <HospitalizationNoteCard 
            id="owner_request"
            title="飼主からのリクエスト"
            icon={MessageSquare}
            value={formData.ownerRequest}
            onChange={(val) => handleFormChange({ ownerRequest: val })}
            placeholder="リクエストを入力..."
          />

          {/* Right Column - スタッフへの連絡事項 */}
          <HospitalizationNoteCard 
            id="staff_notes"
            title="スタッフへの連絡事項"
            icon={AlertCircle}
            value={formData.staffNotes}
            onChange={(val) => handleFormChange({ staffNotes: val })}
            placeholder="連絡事項を入力..."
          />
        </div>

        {/* 治療プラン */}
        <HospitalizationTreatmentTable
            treatmentPlans={treatmentPlans}
            onAdd={addTreatmentPlan}
            onUpdate={updateTreatmentPlan}
            onRemove={canDelete ? removeTreatmentPlan : undefined}
        />

        {/* 診療費計算 */}
        <HospitalizationCostSummary
            totals={totals}
            globalDiscount={globalDiscount}
            setGlobalDiscount={handleGlobalDiscountChange}
            globalDiscountAmount={globalDiscountAmount}
            setGlobalDiscountAmount={handleGlobalDiscountAmountChange}
        />
        </fieldset>
    </PageLayout>
    </form>
    <ConfirmDialog
        open={isDeleteConfirmOpen}
        onClose={() => setIsDeleteConfirmOpen(false)}
        title="入院を削除しますか？"
        description="この操作は取り消せません。"
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDelete}
        isPending={isDeletePending}
      />
    </>
  );
}
