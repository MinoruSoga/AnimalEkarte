// React/Framework
import { useCallback, useEffect } from "react";
import { useNavigate, useParams, useLocation, useSearchParams } from "react-router";

// External
import { Trash2 } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker";
import { useUnsavedChanges } from "@/hooks/useUnsavedChanges";
import { C, STYLE } from "@/lib/design-tokens";

// Relative
import { useExaminationForm } from "../hooks/useExaminationForm";
import { useMasterItems } from "@/hooks/use-master-items";
import { paths } from "@/config/paths";

export function ExaminationForm() {
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
      isEdit,
      isSaving,
  } = useExaminationForm(id);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];

  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  const handleBack = useCallback(() => {
    if (location.state?.from) {
      navigate(location.state.from);
    } else {
      navigate(paths.examinations.getHref());
    }
  }, [location.state, navigate]);

  const handleSetFormData = useCallback((next: Parameters<typeof setFormData>[0]) => {
    markDirty();
    setFormData(next);
  }, [markDirty, setFormData]);

  const handleSaveClick = useCallback(() => {
    markClean();
    handleSave();
  }, [markClean, handleSave]);

  useEffect(() => {
    if (!selectedPet && !isEdit && !petId) {
        navigate(paths.examinations.selectPet.getHref());
    }
  }, [selectedPet, isEdit, navigate, petId]);

  if (!selectedPet && !isEdit && petId) return null;
  if (!selectedPet && !isEdit) return null;

  return (
    <PageLayout
      title={isEdit ? "検査詳細・編集" : "新規検査登録"}
      onBack={handleBack}
      maxWidth="max-w-3xl"
      align="left"
    >
      <NavigationBlocker when={isDirty && !isSaving} />
      <div className="flex flex-col gap-4">
          {/* Patient Info Card */}
          {selectedPet ? (
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
          ) : null}

          <div className={`bg-white p-4 rounded-lg border ${C.borderMedium} space-y-4 shadow-sm`}>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label className={`text-sm ${C.text60}`}>検査種別</Label>
                <Select 
                    value={formData.testType} 
                    onValueChange={(v) => handleSetFormData({ testType: v })}
                >
                  <SelectTrigger className={`h-10 text-sm ${C.text} bg-white ${C.borderMedium}`}>
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
                <Label className={`text-sm ${C.text60}`}>担当医</Label>
                <Select 
                    value={formData.doctor} 
                    onValueChange={(v) => handleSetFormData({ doctor: v })}
                >
                  <SelectTrigger className={`h-10 text-sm ${C.text} bg-white ${C.borderMedium}`}>
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
              <Label className={`text-sm ${C.text60}`}>ステータス</Label>
              <Select 
                value={formData.status} 
                onValueChange={(v: "依頼中" | "検査中" | "完了") => handleSetFormData({ status: v })}
              >
                <SelectTrigger className={`h-10 text-sm ${C.text} bg-white ${C.borderMedium}`}>
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
              <Label className={`text-sm ${C.text60}`}>備考・所見</Label>
              <Textarea 
                className={`h-24 text-sm ${C.text} bg-white ${C.borderMedium} resize-none`} 
                placeholder="検査結果や備考を入力" 
                value={formData.resultSummary || ""} 
                onChange={(e) => handleSetFormData({ resultSummary: e.target.value })}
              />
            </div>

            <div className="flex justify-end gap-2 pt-2">
              {isEdit ? (
                <Button variant="ghost" className={`h-10 text-sm ${STYLE.btnDangerGhost} mr-auto`}>
                    <Trash2 className="mr-1.5 size-4" />
                    削除
                </Button>
              ) : null}
              <Button variant="outline" onClick={handleBack} className="h-10 text-sm">キャンセル</Button>
              <Button className={`${C.bgAccent} ${C.bgAccentHover} text-white h-10 text-sm`} onClick={handleSaveClick} disabled={isSaving}>{isSaving ? "保存中..." : "保存"}</Button>
            </div>
          </div>
      </div>
    </PageLayout>
  );
}