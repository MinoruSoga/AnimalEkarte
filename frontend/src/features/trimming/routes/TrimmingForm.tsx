// React/Framework
import { useEffect } from "react";
import { useNavigate, useParams, useLocation, useSearchParams } from "react-router";

// External
import {
  Scissors,
  Upload,
  X,
  Search,
  Calendar,
  Trash2,
} from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Checkbox } from "@/components/ui/checkbox";
import { PatientInfoCard } from "@/components/shared/PatientInfoCard";
import { PageLayout } from "@/components/shared/PageLayout";

// Relative
import { useTrimmingForm } from "../hooks/useTrimmingForm";
import { useMasterItems } from "@/hooks/use-master-items";

export const TrimmingForm = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { id } = useParams();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  
  const { data: courses } = useMasterItems("trimming_course");
  const { data: options } = useMasterItems("trimming_option");

  const {
      mode,
      formData,
      setFormData,
      styleImagePreview,
      completedImagePreview,
      petSelection,
      handleStyleImageChange,
      handleCompletedImageChange,
      removeStyleImage,
      removeCompletedImage,
      handleSave,
  } = useTrimmingForm(id);

  const { selectedPets } = petSelection;
  const selectedPet = selectedPets[0];

  useEffect(() => {
    // Redirect only if:
    // 1. No pet is selected
    // 2. Not in edit mode
    // 3. No petId in URL (if petId exists, we wait for it to load)
    if (!selectedPet && mode === "new" && !petId) {
        navigate("/trimming/select-pet");
    }
  }, [selectedPet, mode, navigate, petId]);

  if (!selectedPet && mode === "new" && petId) return null;
  if (!selectedPet && mode === "new") return null;

  const handleBack = () => {
      if (location.state?.from) {
          navigate(location.state.from);
      } else {
          navigate("/trimming");
      }
  };

  return (
    <PageLayout
      title={mode === "new" ? "トリミング登録" : "トリミング編集"}
      onBack={handleBack}
      icon={<Scissors className="h-4 w-4 text-[#37352F]" />}
      maxWidth="max-w-[1400px]"
      headerAction={
          <div className="flex gap-2">
            {mode === "edit" && (
                <Button
                    variant="ghost"
                    className="h-10 text-red-600 hover:text-red-700 hover:bg-red-50 rounded-[6px] text-sm px-4"
                >
                    <Trash2 className="mr-1.5 size-4" />
                    削除
                </Button>
            )}
            <Button
                onClick={handleSave}
                size="sm"
                className="h-10 bg-[#37352F] hover:bg-[#37352F]/90 text-white rounded-[6px] text-sm px-4"
            >
                保存
            </Button>
          </div>
      }
    >
      {/* Patient Info Card */}
      {selectedPet && (
          <PatientInfoCard
            ownerName={selectedPet.ownerName}
            petName={`${selectedPet.name}${selectedPet.species ? `(${selectedPet.species})` : ""}`}
            petNumber={selectedPet.petNumber || selectedPet.id}
            weight={selectedPet.weight || "-"}
            staffName="トリマーA"
            serviceType="トリミング"
            petDetails={`${selectedPet.birthDate ? `${selectedPet.birthDate}生` : ""} / ${selectedPet.species}`}
            insuranceName={selectedPet.insuranceName || "保険情報未登録"}
            insuranceDetails={selectedPet.insuranceDetails || "-"}
            nextVisitDate="-"
            nextVisitContent="-"
            className="mb-4"
          />
      )}

      {/* 3 Column Layout */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {/* Left Column */}
          <div className="bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] p-3">
            <div className="space-y-3">
              {/* Course Selection */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">コース選択</Label>
                <div className="grid grid-cols-1 gap-2">
                  {courses.map(course => (
                    <div 
                      key={course.id}
                      onClick={() => setFormData({...formData, courseId: course.id, charge: course.price.toLocaleString()})}
                      className={`
                        p-2 border rounded-md cursor-pointer transition-colors flex justify-between items-center
                        ${formData.courseId === course.id 
                          ? 'bg-[#E3F2FD] border-[#2EAADC]' 
                          : 'bg-white border-[rgba(55,53,47,0.16)] hover:bg-gray-50'}
                      `}
                    >
                      <div className="text-sm text-[#37352F]">{course.name}</div>
                      <div className="text-xs text-[#37352F]/60">¥{course.price.toLocaleString()}</div>
                    </div>
                  ))}
                </div>
              </div>

              {/* スタイルの希望 */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">
                  スタイルの希望
                </Label>
                <Textarea
                  value={formData.styleRequest}
                  onChange={(e) =>
                    setFormData({ ...formData, styleRequest: e.target.value })
                  }
                  placeholder="スタイルの希望を入力..."
                  className="min-h-[80px] text-sm resize-none bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                />
              </div>

              {/* メモ */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">メモ</Label>
                <Textarea
                  value={formData.memo}
                  onChange={(e) =>
                    setFormData({ ...formData, memo: e.target.value })
                  }
                  placeholder="メモを入力..."
                  className="min-h-[80px] text-sm resize-none bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                />
              </div>

              {/* 虫卵 */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">虫卵</Label>
                <Input
                  value={formData.eggs}
                  onChange={(e) =>
                    setFormData({ ...formData, eggs: e.target.value })
                  }
                  placeholder="虫卵の有無を入力..."
                  className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                />
              </div>

              {/* 部位 */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">部位</Label>
                <div className="space-y-1.5">
                  <div className="flex items-center gap-2 h-10">
                    <Switch
                      checked={formData.parts.nail}
                      onCheckedChange={(checked) =>
                        setFormData({
                          ...formData,
                          parts: { ...formData.parts, nail: checked },
                        })
                      }
                      className="mr-2"
                    />
                    <span className="text-sm text-[#37352F]">爪</span>
                  </div>
                  <div className="flex items-center gap-2 h-10">
                    <Switch
                      checked={formData.parts.analGland}
                      onCheckedChange={(checked) =>
                        setFormData({
                          ...formData,
                          parts: { ...formData.parts, analGland: checked },
                        })
                      }
                      className="mr-2"
                    />
                    <span className="text-sm text-[#37352F]">肛門腺</span>
                  </div>
                  <div className="flex items-center gap-2 h-10">
                    <Switch
                      checked={formData.parts.eye}
                      onCheckedChange={(checked) =>
                        setFormData({
                          ...formData,
                          parts: { ...formData.parts, eye: checked },
                        })
                      }
                      className="mr-2"
                    />
                    <span className="text-sm text-[#37352F]">目</span>
                  </div>
                  <div className="flex items-center gap-2 h-10">
                    <Switch
                      checked={formData.parts.ear}
                      onCheckedChange={(checked) =>
                        setFormData({
                          ...formData,
                          parts: { ...formData.parts, ear: checked },
                        })
                      }
                      className="mr-2"
                    />
                    <span className="text-sm text-[#37352F]">耳</span>
                  </div>
                  <div className="flex items-center gap-2 h-10">
                    <Switch
                      checked={formData.parts.skin}
                      onCheckedChange={(checked) =>
                        setFormData({
                          ...formData,
                          parts: { ...formData.parts, skin: checked },
                        })
                      }
                      className="mr-2"
                    />
                    <span className="text-sm text-[#37352F]">皮膚</span>
                  </div>
                  <div className="flex items-center gap-2 h-10">
                    <Switch
                      checked={formData.parts.oral}
                      onCheckedChange={(checked) =>
                        setFormData({
                          ...formData,
                          parts: { ...formData.parts, oral: checked },
                        })
                      }
                      className="mr-2"
                    />
                    <span className="text-sm text-[#37352F]">口腔</span>
                  </div>
                </div>
              </div>

              {/* Options */}
              {options.length > 0 && (
                <div>
                    <Label className="text-sm text-[#37352F]/60 mb-1.5 block">オプション</Label>
                    <div className="grid grid-cols-2 gap-2">
                        {options.map(option => (
                            <div key={option.id} className="flex items-center space-x-2">
                                <Checkbox 
                                    id={option.id} 
                                    checked={formData.optionIds?.includes(option.id)}
                                    onCheckedChange={(checked) => {
                                        if (checked) {
                                            setFormData({ ...formData, optionIds: [...(formData.optionIds || []), option.id] });
                                        } else {
                                            setFormData({ ...formData, optionIds: (formData.optionIds || []).filter(id => id !== option.id) });
                                        }
                                    }}
                                />
                                <Label 
                                    htmlFor={option.id} 
                                    className="text-sm font-normal text-[#37352F] cursor-pointer"
                                >
                                    {option.name} (+¥{option.price})
                                </Label>
                            </div>
                        ))}
                    </div>
                </div>
              )}

              {/* 希望スタイル画像 */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">
                  希望スタイル画像
                </Label>
                {styleImagePreview ? (
                  <div className="relative">
                    <img
                      src={styleImagePreview}
                      alt="希望スタイル"
                      className="w-full h-[180px] object-cover rounded-md border border-[rgba(55,53,47,0.09)]"
                    />
                    <button
                      onClick={removeStyleImage}
                      className="absolute top-2 right-2 bg-white rounded-full p-1 shadow-md hover:bg-slate-100"
                    >
                      <X className="h-3.5 w-3.5 text-[#37352F]" />
                    </button>
                  </div>
                ) : (
                  <label className="block cursor-pointer">
                    <input
                      type="file"
                      accept="image/*"
                      onChange={handleStyleImageChange}
                      className="hidden"
                    />
                    <div className="w-full h-[180px] bg-[#F7F6F3] rounded-md flex flex-col items-center justify-center hover:bg-[rgba(55,53,47,0.08)] transition-colors border-2 border-dashed border-[rgba(55,53,47,0.16)]">
                      <Upload className="h-6 w-6 text-[#37352F]/40 mb-2" />
                      <p className="text-sm text-[#37352F]/60 text-center">
                        希望スタイルの画像
                        <br />
                        アップロード
                      </p>
                    </div>
                  </label>
                )}
              </div>
            </div>
          </div>

          {/* Middle Column */}
          <div className="bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] p-3">
            <div className="space-y-3">
              {/* BW & BT */}
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <Label className="text-sm text-[#37352F]/60 mb-1.5 block">BW</Label>
                  <div className="flex gap-2 items-center">
                    <Input
                      value={formData.bw}
                      onChange={(e) =>
                        setFormData({ ...formData, bw: e.target.value })
                      }
                      placeholder="体重"
                      className="flex-1 h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                    />
                    <div className="flex gap-2 items-center">
                      <label className="flex items-center gap-1 cursor-pointer">
                        <input
                          type="radio"
                          checked={formData.bwUnit === "Kg"}
                          onChange={() =>
                            setFormData({ ...formData, bwUnit: "Kg" })
                          }
                          className="w-4 h-4 text-[#37352F] focus:ring-[#37352F]"
                        />
                        <span className="text-sm text-[#37352F]">Kg</span>
                      </label>
                      <label className="flex items-center gap-1 cursor-pointer">
                        <input
                          type="radio"
                          checked={formData.bwUnit === "g"}
                          onChange={() =>
                            setFormData({ ...formData, bwUnit: "g" })
                          }
                          className="w-4 h-4 text-[#37352F] focus:ring-[#37352F]"
                        />
                        <span className="text-sm text-[#37352F]">g</span>
                      </label>
                    </div>
                  </div>
                </div>
                <div>
                  <Label className="text-sm text-[#37352F]/60 mb-1.5 block">BT</Label>
                  <Input
                    value={formData.bt}
                    onChange={(e) =>
                      setFormData({ ...formData, bt: e.target.value })
                    }
                    placeholder="体温"
                    className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                  />
                </div>
              </div>

              {/* USED SHAMPOO */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">
                  USED SHAMPOO
                </Label>
                <Input
                  value={formData.usedShampoo}
                  onChange={(e) =>
                    setFormData({ ...formData, usedShampoo: e.target.value })
                  }
                  placeholder="使用したシャンプーを入力..."
                  className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                />
              </div>

              {/* USED RIBBON */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">
                  USED RIBBON
                </Label>
                <Input
                  value={formData.usedRibbon}
                  onChange={(e) =>
                    setFormData({ ...formData, usedRibbon: e.target.value })
                  }
                  placeholder="使用したリボンを入力..."
                  className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                />
              </div>

              {/* TREATMENT */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">
                  TREATMENT
                </Label>
                <Input
                  value={formData.treatment}
                  onChange={(e) =>
                    setFormData({ ...formData, treatment: e.target.value })
                  }
                  placeholder="処置内容を入力..."
                  className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                />
              </div>

              {/* MEDICINE */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">
                  MEDICINE
                </Label>
                <Input
                  value={formData.medicine}
                  onChange={(e) =>
                    setFormData({ ...formData, medicine: e.target.value })
                  }
                  placeholder="使用した薬を入力..."
                  className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                />
              </div>

              {/* CHARGE */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">CHARGE</Label>
                <Input
                  value={formData.charge}
                  onChange={(e) =>
                    setFormData({ ...formData, charge: e.target.value })
                  }
                  placeholder="料金を入力..."
                  className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                />
              </div>

              {/* FINAL CHECK */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">
                  FINAL CHECK
                </Label>
                <Input
                  value={formData.finalCheck}
                  onChange={(e) =>
                    setFormData({ ...formData, finalCheck: e.target.value })
                  }
                  placeholder="最終確認内容を入力..."
                  className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                />
              </div>

              {/* 備考 */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">備考</Label>
                <Input
                  value={formData.remarks}
                  onChange={(e) =>
                    setFormData({ ...formData, remarks: e.target.value })
                  }
                  placeholder="備考を入力..."
                  className="h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                />
              </div>

              {/* 完成画像 */}
              <div>
                <Label className="text-sm text-[#37352F]/60 mb-1.5 block">
                  完成画像
                </Label>
                {completedImagePreview ? (
                  <div className="relative">
                    <img
                      src={completedImagePreview}
                      alt="完成"
                      className="w-full h-[180px] object-cover rounded-md border border-[rgba(55,53,47,0.09)]"
                    />
                    <button
                      onClick={removeCompletedImage}
                      className="absolute top-2 right-2 bg-white rounded-full p-1 shadow-md hover:bg-slate-100"
                    >
                      <X className="h-3.5 w-3.5 text-[#37352F]" />
                    </button>
                  </div>
                ) : (
                  <label className="block cursor-pointer">
                    <input
                      type="file"
                      accept="image/*"
                      onChange={handleCompletedImageChange}
                      className="hidden"
                    />
                    <div className="w-full h-[180px] bg-[#F7F6F3] rounded-md flex flex-col items-center justify-center hover:bg-[rgba(55,53,47,0.08)] transition-colors border-2 border-dashed border-[rgba(55,53,47,0.16)]">
                      <Upload className="h-6 w-6 text-[#37352F]/40 mb-2" />
                      <p className="text-sm text-[#37352F]/60 text-center">
                        完成画像
                        <br />
                        アップロード
                      </p>
                    </div>
                  </label>
                )}
              </div>
            </div>
          </div>

          {/* Right Column - トリミング履歴 */}
          <div className="bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] p-3">
            <h2 className="text-sm font-bold mb-3 text-[#37352F]">トリミング履歴</h2>
            
            {/* 診療日フィルター */}
            <div className="mb-3">
              <Label className="text-sm text-[#37352F]/60 mb-1.5 block">診療日</Label>
              <div className="flex items-center gap-2">
                <div className="relative flex-1">
                    <Calendar className="absolute left-2.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
                    <Input
                    type="date"
                    className="pl-9 flex-1 h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                    />
                </div>
                <span className="text-[#37352F]/60 text-sm">〜</span>
                <div className="relative flex-1">
                    <Calendar className="absolute left-2.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
                    <Input
                    type="date"
                    className="pl-9 flex-1 h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                    />
                </div>
              </div>
            </div>

            {/* 検索 */}
            <div className="mb-3 flex gap-2">
              <div className="relative flex-1">
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
                <Input
                    placeholder="検索単語"
                    className="pl-9 flex-1 h-10 text-sm bg-white border-[rgba(55,53,47,0.16)] focus-visible:ring-[#2EAADC]"
                />
              </div>
              <Button
                variant="outline"
                size="sm"
                className="h-10 text-sm border-[rgba(55,53,47,0.16)] text-[#37352F]"
              >
                クリア
              </Button>
            </div>

            {/* 履歴リスト */}
            <div className="space-y-2 max-h-[600px] overflow-y-auto">
              {[1, 2].map((i) => (
                <div
                  key={i}
                  className="border border-[rgba(55,53,47,0.16)] rounded-md p-3 text-sm space-y-1 hover:bg-[#F7F6F3] transition-colors"
                >
                  <p className="text-[#37352F]">
                    診療日: 2022/10/10 (月)
                  </p>
                  <p className="text-[#37352F]/80 text-sm">
                    作成者: スタッフA 2022/10/10 10:10
                  </p>
                  <p className="text-[#37352F]/80 text-sm">
                    更新者: スタッフB 2022/10/10 10:10
                  </p>
                  <div className="border-t border-[rgba(55,53,47,0.09)] my-2 pt-2">
                    <p className="font-bold mb-0.5 text-[#37352F]"># スタイルの希望</p>
                    <p className="text-[#37352F]/80 mb-1">
                      Dummy Text Dummy Text
                    </p>
                    <p className="font-bold mb-0.5 text-[#37352F]"># メモ</p>
                    <p className="text-[#37352F]/80 mb-1">
                      Dummy Text Dummy Text
                    </p>
                    <p className="font-bold text-[#37352F]"># 虫卵</p>
                    <p className="font-bold text-[#37352F]"># 部位</p>
                    <p className="font-bold text-[#37352F] mt-1"># 画像</p>
                    <p className="text-[#37352F]/80 mb-1">Image01.jpg</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
      </div>
    </PageLayout>
  );
}
