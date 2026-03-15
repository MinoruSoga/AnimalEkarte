// React/Framework
import { useEffect, useState, useCallback } from "react";
import { useNavigate, useParams, useLocation, useSearchParams } from "react-router";

// External
import { FileText, Trash2, MessageSquare, AlertCircle } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { useAuth } from "@/features/auth";

// Relative
import { useHospitalizationForm } from "../hooks/useHospitalizationForm";
import { useDeleteHospitalization } from "../api/delete-hospitalization";
import { paths } from "@/config/paths";
import { useMasterItems } from "@/hooks/use-master-items";
import { HospitalizationBasicInfo } from "../components/HospitalizationBasicInfo";
import { HospitalizationNoteCard } from "../components/HospitalizationNoteCard";
import { HospitalizationTreatmentTable } from "../components/HospitalizationTreatmentTable";
import { HospitalizationCostSummary } from "../components/HospitalizationCostSummary";

export function HospitalizationForm() {
  const navigate = useNavigate();
  const location = useLocation();
  const { id: hospitalizationId } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  
  const { data: cageItems } = useMasterItems("cage");

  const { user } = useAuth();
  const deleteMutation = useDeleteHospitalization();
  const [isDeleteConfirmOpen, setIsDeleteConfirmOpen] = useState(false);

  const {
      isEdit,
      formData,
      handleFormDataChange,
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
      handleSave
  } = useHospitalizationForm(hospitalizationId);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];
  const totals = calculateTotals();

  const handleBack = useCallback(() => {
    if (location.state?.from) {
        navigate(location.state.from as string);
    } else {
        navigate(paths.hospitalization.getHref());
    }
  }, [location.state, navigate]);

  const handleDelete = useCallback(() => {
    if (!hospitalizationId) return;
    deleteMutation.mutate(hospitalizationId, {
      onSuccess: () => {
        navigate(paths.hospitalization.getHref());
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  }, [hospitalizationId, deleteMutation, navigate]);

  useEffect(() => {
    if (!selectedPet && !isEdit && !petId) {
        navigate(paths.hospitalization.selectPet.getHref());
    }
  }, [selectedPet, isEdit, navigate, petId]);

  if (!selectedPet && !isEdit && petId) return null;
  if (!selectedPet && !isEdit) return null;

  return (
    <>
    <PageLayout
      title={hospitalizationId ? "入院編集" : "入院登録"}
      onBack={handleBack}
      icon={<FileText className="h-4 w-4 text-[#37352F]/60" />}
      maxWidth="max-w-[1400px]"
      headerAction={
        <div className="flex gap-2">
            {hospitalizationId ? (
                <>
                  <Button
                    variant="outline"
                    className="gap-2 h-10 text-sm px-4 text-[#37352F]"
                    onClick={() => navigate(`/hospitalization/${hospitalizationId}`)}
                  >
                    <FileText className="h-4 w-4" />
                    デイリーカルテ
                  </Button>
                  <Button
                    variant="ghost"
                    className="text-red-600 hover:text-red-700 hover:bg-red-50 h-10 text-sm px-4"
                    onClick={() => setIsDeleteConfirmOpen(true)}
                  >
                    <Trash2 className="mr-1.5 size-4" />
                    削除
                  </Button>
                </>
            ) : null}
            <Button
            className="bg-[#2383E2] hover:bg-[#1B6EC2] text-white rounded-[6px] h-10 text-sm px-4"
            onClick={handleSave}
            >
            {hospitalizationId ? "更新" : "登録"}
            </Button>
        </div>
      }
    >
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
            onChange={handleFormDataChange} 
            cageItems={cageItems} 
          />

          {/* Middle Column - 飼主からのリクエスト */}
          <HospitalizationNoteCard 
            title="飼主からのリクエスト"
            icon={MessageSquare}
            value={formData.ownerRequest}
            onChange={(val) => handleFormDataChange({ ownerRequest: val })}
            placeholder="リクエストを入力..."
          />

          {/* Right Column - スタッフへの連絡事項 */}
          <HospitalizationNoteCard 
            title="スタッフへの連絡事項"
            icon={AlertCircle}
            value={formData.staffNotes}
            onChange={(val) => handleFormDataChange({ staffNotes: val })}
            placeholder="連絡事項を入力..."
          />
        </div>

        {/* 治療プラン */}
        <HospitalizationTreatmentTable 
            treatmentPlans={treatmentPlans}
            onAdd={addTreatmentPlan}
            onUpdate={updateTreatmentPlan}
            onRemove={removeTreatmentPlan}
        />

        {/* 診療費計算 */}
        <HospitalizationCostSummary 
            totals={totals}
            globalDiscount={globalDiscount}
            setGlobalDiscount={setGlobalDiscount}
            globalDiscountAmount={globalDiscountAmount}
            setGlobalDiscountAmount={setGlobalDiscountAmount}
        />
    </PageLayout>
      <ConfirmDialog
        open={isDeleteConfirmOpen}
        onClose={() => setIsDeleteConfirmOpen(false)}
        title="入院を削除しますか？"
        description="この操作は取り消せません。"
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDelete}
      />
    </>
  );
}
