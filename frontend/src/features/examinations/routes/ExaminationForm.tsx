// React/Framework
import { useEffect } from "react";
import { useNavigate, useParams, useLocation, useSearchParams } from "react-router-dom";

// External
import { Trash2 } from "lucide-react";

// Internal
import { Button } from "../../../components/ui/button";
import { Label } from "../../../components/ui/label";
import { Textarea } from "../../../components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../../../components/ui/select";
import { PatientInfoCard } from "../../../components/shared/PatientInfoCard";
import { PageLayout } from "../../../components/shared/PageLayout";

// Relative
import { useExaminationForm } from "../hooks/useExaminationForm";
import { useMasterItems } from "../../master/hooks/useMasterItems";

export const ExaminationForm = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  
  const { data: examTypes } = useMasterItems("examination");
  const { data: staffList } = useMasterItems("staff");
  
  const {
      formData,
      setFormData,
      petSelection,
      handleSave,
      isEdit
  } = useExaminationForm(id);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];

  useEffect(() => {
    if (!selectedPet && !isEdit && !petId) {
        navigate("/examinations/select-pet");
    }
  }, [selectedPet, isEdit, navigate, petId]);

  if (!selectedPet && !isEdit && petId) return null;
  if (!selectedPet && !isEdit) return null;

  const handleBack = () => {
    if (location.state?.from) {
        navigate(location.state.from);
    } else {
        navigate("/examinations");
    }
  };

  return (
    <PageLayout
      title={isEdit ? "検査詳細・編集" : "新規検査登録"}
      onBack={handleBack}
      maxWidth="max-w-3xl"
      align="left"
    >
      <div className="flex flex-col gap-4">
          {/* Patient Info Card */}
          {selectedPet && (
              <PatientInfoCard
                ownerName={selectedPet.ownerName}
                petName={`${selectedPet.name}${selectedPet.species ? `(${selectedPet.species})` : ""}`}
                petNumber={selectedPet.petNumber || selectedPet.id}
                weight={selectedPet.weight || "-"}
                staffName="医師A"
                serviceType="検査"
                petDetails={`${selectedPet.birthDate ? `${selectedPet.birthDate}生` : ""} / ${selectedPet.species}`}
                insuranceName={selectedPet.insuranceName || "保険情報未登録"}
                insuranceDetails={selectedPet.insuranceDetails || "-"}
                nextVisitDate="-"
                nextVisitContent="-"
              />
          )}

          <div className="bg-white p-4 rounded-lg border border-[rgba(55,53,47,0.16)] space-y-4 shadow-sm">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label className="text-sm text-[#37352F]/60">検査種別</Label>
                <Select 
                    value={formData.testType} 
                    onValueChange={(v) => setFormData({...formData, testType: v})}
                >
                  <SelectTrigger className="h-10 text-sm text-[#37352F] bg-white border-[rgba(55,53,47,0.16)]">
                    <SelectValue placeholder="選択してください" />
                  </SelectTrigger>
                  <SelectContent>
                    {examTypes.map((item) => (
                        <SelectItem key={item.id} value={item.name}>
                            {item.name}
                        </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label className="text-sm text-[#37352F]/60">担当医</Label>
                <Select 
                    value={formData.doctor} 
                    onValueChange={(v) => setFormData({...formData, doctor: v})}
                >
                  <SelectTrigger className="h-10 text-sm text-[#37352F] bg-white border-[rgba(55,53,47,0.16)]">
                    <SelectValue placeholder="選択してください" />
                  </SelectTrigger>
                  <SelectContent>
                    {staffList.map((staff) => (
                        <SelectItem key={staff.id} value={staff.name}>
                            {staff.name}
                        </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div className="space-y-1.5">
              <Label className="text-sm text-[#37352F]/60">ステータス</Label>
              <Select 
                value={formData.status} 
                onValueChange={(v: "依頼中" | "検査中" | "完了") => setFormData({...formData, status: v})}
              >
                <SelectTrigger className="h-10 text-sm text-[#37352F] bg-white border-[rgba(55,53,47,0.16)]">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="依頼中">依頼中</SelectItem>
                  <SelectItem value="検査中">検査中</SelectItem>
                  <SelectItem value="完了">完了</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1.5">
              <Label className="text-sm text-[#37352F]/60">備考・所見</Label>
              <Textarea 
                className="h-24 text-sm text-[#37352F] bg-white border-[rgba(55,53,47,0.16)] resize-none" 
                placeholder="検査結果や備考を入力" 
                value={formData.resultSummary || ""} 
                onChange={(e) => setFormData({...formData, resultSummary: e.target.value})}
              />
            </div>

            <div className="flex justify-end gap-2 pt-2">
              {isEdit && (
                <Button variant="ghost" className="h-10 text-sm text-red-600 hover:text-red-700 hover:bg-red-50 mr-auto">
                    <Trash2 className="mr-1.5 size-4" />
                    削除
                </Button>
              )}
              <Button variant="outline" onClick={handleBack} className="h-10 text-sm">キャンセル</Button>
              <Button className="bg-[#37352F] hover:bg-[#37352F]/90 text-white h-10 text-sm" onClick={handleSave}>保存</Button>
            </div>
          </div>
      </div>
    </PageLayout>
  );
}
