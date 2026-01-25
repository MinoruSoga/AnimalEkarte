// React/Framework
import { useParams } from "react-router-dom";

// Internal
import { Button } from "../../../components/ui/button";
import { PatientInfoCard } from "../../../components/shared/PatientInfoCard";
import { PageLayout } from "../../../components/shared/PageLayout";

// Relative
import { MedicalRecordInterview } from "../components/MedicalRecordInterview";
import { MedicalRecordDiagnosisPlan } from "../components/MedicalRecordDiagnosisPlan";
import { MedicalRecordTreatment } from "../components/MedicalRecordTreatment";
import { MedicalRecordVaccination } from "../components/MedicalRecordVaccination";
import { MedicalRecordImage } from "../components/MedicalRecordImage";
import { MedicalRecordEstimate } from "../components/MedicalRecordEstimate";
import { MedicalRecordBillCheck } from "../components/MedicalRecordBillCheck";
import { MedicalRecordExamination } from "../components/MedicalRecordExamination";
import { useMedicalRecordForm } from "../hooks/useMedicalRecordForm";

export const MedicalRecordForm = () => {
  const { id: recordId } = useParams();
  const {
    isNewRecord,
    activeTab,
    setActiveTab,
    selectedPet,
    handleBack,
    handleSave,
    treatmentPlanItems,
    setTreatmentPlanItems,
    treatmentCompletedItems,
    setTreatmentCompletedItems
  } = useMedicalRecordForm(recordId);

  // Tab definitions
  const tabs = [
    "問診",
    "診察/治療プラン",
    "治療",
    "予防接種",
    "検査",
    "画像",
    "見積書",
    "会計(医師確認)",
  ];

  if (!selectedPet) {
      return null; 
  }

  return (
    <PageLayout
      title={recordId ? "カルテ編集" : "カルテ入力"}
      onBack={handleBack}
      maxWidth="max-w-[1440px]"
    >
        {/* Patient Info Card */}
        {selectedPet && (
            <PatientInfoCard
              ownerName={selectedPet.ownerName}
              petName={`${selectedPet.name}${selectedPet.species ? `(${selectedPet.species})` : ""}`}
              petNumber={selectedPet.petNumber || selectedPet.id}
              weight={selectedPet.weight || "-"}
              staffName="医師A"
              serviceType="診療"
              petDetails={`${selectedPet.birthDate ? `${selectedPet.birthDate}生` : ""} / ${selectedPet.species}`}
              insuranceName={selectedPet.insuranceName || "保険情報未登録"}
              insuranceDetails={selectedPet.insuranceDetails || "-"}
              nextVisitDate="-"
              nextVisitContent="-"
            />
        )}

        {/* Tabs */}
        <div className="flex border-b border-[rgba(55,53,47,0.16)] mb-4 overflow-x-auto">
          {tabs.map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-4 py-2.5 text-sm font-medium transition-colors relative whitespace-nowrap ${
                activeTab === tab
                  ? "text-[#37352F]"
                  : "text-[#37352F]/60 hover:text-[#37352F]"
              }`}
            >
              {tab}
              {activeTab === tab && (
                <div className="absolute bottom-0 left-0 w-full h-[2px] bg-[#37352F]" />
              )}
            </button>
          ))}
        </div>

        {/* Content Area */}
        <div className={activeTab === "問診" ? "block" : "hidden"}>
          <MedicalRecordInterview />
        </div>
        <div className={activeTab === "診察/治療プラン" ? "block" : "hidden"}>
          <MedicalRecordDiagnosisPlan 
            isNewRecord={isNewRecord} 
            items={treatmentPlanItems}
            setItems={setTreatmentPlanItems}
          />
        </div>
        <div className={activeTab === "治療" ? "block" : "hidden"}>
          <MedicalRecordTreatment 
            isNewRecord={isNewRecord} 
            planItems={treatmentPlanItems}
            setPlanItems={setTreatmentPlanItems}
            completedItems={treatmentCompletedItems}
            setCompletedItems={setTreatmentCompletedItems}
          />
        </div>
        <div className={activeTab === "予防接種" ? "block" : "hidden"}>
          <MedicalRecordVaccination />
        </div>
        <div className={activeTab === "検査" ? "block" : "hidden"}>
          <MedicalRecordExamination isNewRecord={isNewRecord} />
        </div>
        <div className={activeTab === "画像" ? "block" : "hidden"}>
          <MedicalRecordImage isNewRecord={isNewRecord} />
        </div>
        <div className={activeTab === "見積書" ? "block" : "hidden"}>
          <MedicalRecordEstimate isNewRecord={isNewRecord} />
        </div>
        <div className={activeTab === "会計(医師確認)" ? "block" : "hidden"}>
          <MedicalRecordBillCheck 
            isNewRecord={isNewRecord} 
            petId={selectedPet.id} 
            completedItems={treatmentCompletedItems}
          />
        </div>

        {/* Floating Save Button */}
        {activeTab !== "会計(医師確認)" && (
          <div className="fixed bottom-6 right-6 z-50">
            <Button 
                onClick={handleSave}
                className="bg-[#37352F] hover:bg-[#37352F]/90 text-white shadow-lg px-5 h-10 text-sm rounded-md"
            >
              保存
            </Button>
          </div>
        )}
    </PageLayout>
  );
}
