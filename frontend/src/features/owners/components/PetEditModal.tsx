import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "sonner";

import { NotionDatePicker } from "@/components/shared/NotionDatePicker";
import {
  PET_GENDER_VALUES,
  ACQUISITION_TYPE_VALUES,
  DANGER_LEVEL_VALUES,
  INSURANCE_COMPANY_VALUES,
  PET_INSURANCE_RATIO_VALUES,
  PetFormData,
} from "../types";
import { isOneOf } from "@/lib/type-utils";

interface PetEditModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  ownerName?: string;
  petData?: PetFormData;
  onSave: (data: PetFormData) => void;
}

export function PetEditModal({
  open,
  onOpenChange,
  ownerName = "飼主名",
  petData,
  onSave,
}: PetEditModalProps) {
  const [formData, setFormData] = useState<PetFormData>({
    id: "",
    petNumber: "",
    petName: "",
    petNameKana: "",
    species: "",
    gender: "",
    birthDate: "",
    breed: "",
    color: "",
    weight: "",
    neuteredDate: "",
    acquisitionType: "購入",
    dangerLevel: "低",
    food: "",
    environment: "",
    status: "生存",
    remarks: "",
  });

  // Reset form data when modal opens
  useEffect(() => {
    if (open) {
      setFormData({
        id: petData?.id || "",
        petNumber: petData?.petNumber || "",
        petName: petData?.petName || "",
        petNameKana: petData?.petNameKana || "",
        species: petData?.species || "",
        gender: petData?.gender || "",
        birthDate: petData?.birthDate || "",
        breed: petData?.breed || "",
        color: petData?.color || "",
        weight: petData?.weight || "",
        neuteredDate: petData?.neuteredDate || "",
        acquisitionType: (petData?.acquisitionType || "購入") as typeof ACQUISITION_TYPE_VALUES[number],
        dangerLevel: (petData?.dangerLevel || "低") as typeof DANGER_LEVEL_VALUES[number],
        food: petData?.food || "",
        environment: petData?.environment || "",
        status: petData?.status || "生存",
        remarks: petData?.remarks || "",
        insuranceName: petData?.insuranceName,
        insuranceDetails: petData?.insuranceDetails,
      });
    }
  }, [open, petData]);

  const handleSave = () => {
    // Validation
    if (!formData.petName.trim()) {
      toast.error("ペット名を入力してください");
      return;
    }
    if (!formData.species) {
      toast.error("種を選択してください");
      return;
    }
    if (!formData.gender) {
      toast.error("性別を選択してください");
      return;
    }
    if (!formData.birthDate) {
      toast.error("生年月日を選択してください");
      return;
    }

    onSave(formData);
    onOpenChange(false);
    toast.success(petData ? "ペット情報を更新しました" : "ペットを追加しました");
  };

  const isEdit = !!petData?.id;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[1000px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-sm font-bold text-[#37352F]">
            {isEdit ? `${ownerName}のペット情報編集` : `${ownerName}のペット新規登録`}
          </DialogTitle>
          <DialogDescription className="text-sm text-[#37352F]/60">
            {isEdit ? "ペットの情報を編集してください。" : "ペットの基本情報を入力してください。"}
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {/* Column 1 */}
          <div className="space-y-2">
            <div className="space-y-1">
              <Label htmlFor="petNumber" className="text-sm text-[#37352F]/60">
                ペットNo
              </Label>
              <Input
                id="petNumber"
                value={formData.petNumber}
                onChange={(e) =>
                  setFormData({ ...formData, petNumber: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="petName" className="text-sm text-[#37352F]/60">
                ペット名 <span className="text-[#E03E3E]">*</span>
              </Label>
              <Input
                id="petName"
                value={formData.petName}
                onChange={(e) =>
                  setFormData({ ...formData, petName: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="petNameKana" className="text-sm text-[#37352F]/60">
                ペット名(カナ)
              </Label>
              <Input
                id="petNameKana"
                placeholder="例: イリス"
                value={formData.petNameKana || ""}
                onChange={(e) =>
                  setFormData({ ...formData, petNameKana: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="species" className="text-sm text-[#37352F]/60">
                種 <span className="text-[#E03E3E]">*</span>
              </Label>
              <Select
                value={formData.species}
                onValueChange={(value) =>
                  setFormData({ ...formData, species: value })
                }
              >
                <SelectTrigger className="h-10 text-sm bg-white text-[#37352F]">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="犬">犬</SelectItem>
                  <SelectItem value="猫">猫</SelectItem>
                  <SelectItem value="鳥">鳥</SelectItem>
                  <SelectItem value="その他">その他</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1">
              <Label htmlFor="gender" className="text-sm text-[#37352F]/60">
                性別 <span className="text-[#E03E3E]">*</span>
              </Label>
              <Select
                value={formData.gender}
                onValueChange={(value) => {
                  if (isOneOf(value, PET_GENDER_VALUES)) {
                    setFormData({ ...formData, gender: value });
                  }
                }}
              >
                <SelectTrigger className="h-10 text-sm bg-white text-[#37352F]">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {PET_GENDER_VALUES.map((gender) => (
                    <SelectItem key={gender} value={gender}>
                      {gender}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1">
              <Label htmlFor="birthDate" className="text-sm text-[#37352F]/60">
                生年月日 <span className="text-[#E03E3E]">*</span>
              </Label>
              <NotionDatePicker
                id="birthDate"
                value={formData.birthDate}
                onChange={(value) =>
                  setFormData({ ...formData, birthDate: value })
                }
                placeholder="生年月日を選択…"
              />
            </div>
          </div>

          {/* Column 2 */}
          <div className="space-y-2">
            <div className="space-y-1">
              <Label htmlFor="breed" className="text-sm text-[#37352F]/60">
                品種
              </Label>
              <Input
                id="breed"
                value={formData.breed || ""}
                onChange={(e) =>
                  setFormData({ ...formData, breed: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="color" className="text-sm text-[#37352F]/60">
                毛色
              </Label>
              <Input
                id="color"
                value={formData.color || ""}
                onChange={(e) =>
                  setFormData({ ...formData, color: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="weight" className="text-sm text-[#37352F]/60">
                体重(kg)
              </Label>
              <Input
                id="weight"
                value={formData.weight || ""}
                onChange={(e) =>
                  setFormData({ ...formData, weight: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="neuteredDate" className="text-sm text-[#37352F]/60">
                去勢・避妊手術日
              </Label>
              <NotionDatePicker
                id="neuteredDate"
                value={formData.neuteredDate || ""}
                onChange={(value) =>
                  setFormData({ ...formData, neuteredDate: value })
                }
                placeholder="手術日を選択…"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="acquisitionType" className="text-sm text-[#37352F]/60">
                入手区分
              </Label>
              <Select
                value={formData.acquisitionType || ""}
                onValueChange={(value) => {
                  if (isOneOf(value, ACQUISITION_TYPE_VALUES)) {
                    setFormData({ ...formData, acquisitionType: value });
                  }
                }}
              >
                <SelectTrigger className="h-10 text-sm bg-white text-[#37352F]">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {ACQUISITION_TYPE_VALUES.map((type) => (
                    <SelectItem key={type} value={type}>
                      {type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1">
              <Label htmlFor="dangerLevel" className="text-sm text-[#37352F]/60">
                危険度
              </Label>
              <Select
                value={formData.dangerLevel || ""}
                onValueChange={(value) => {
                  if (isOneOf(value, DANGER_LEVEL_VALUES)) {
                    setFormData({ ...formData, dangerLevel: value });
                  }
                }}
              >
                <SelectTrigger className="h-10 text-sm bg-white text-[#37352F]">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {DANGER_LEVEL_VALUES.map((level) => (
                    <SelectItem key={level} value={level}>
                      {level}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Column 3 */}
          <div className="space-y-2">
            <div className="space-y-1">
              <Label htmlFor="food" className="text-sm text-[#37352F]/60">
                食べ物
              </Label>
              <Input
                id="food"
                value={formData.food || ""}
                onChange={(e) =>
                  setFormData({ ...formData, food: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="environment" className="text-sm text-[#37352F]/60">
                飼育環境
              </Label>
              <Input
                id="environment"
                value={formData.environment || ""}
                onChange={(e) =>
                  setFormData({ ...formData, environment: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="insuranceName" className="text-sm text-[#37352F]/60">
                保険名
              </Label>
              <Select
                value={formData.insuranceName || ""}
                onValueChange={(value) => {
                  if (isOneOf(value, INSURANCE_COMPANY_VALUES)) {
                    setFormData({ ...formData, insuranceName: value });
                  }
                }}
              >
                <SelectTrigger className="h-10 text-sm bg-white text-[#37352F]">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {INSURANCE_COMPANY_VALUES.map((company) => (
                    <SelectItem key={company} value={company}>
                      {company}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1">
              <Label htmlFor="insuranceDetails" className="text-sm text-[#37352F]/60">
                保険詳細(負担割合など)
              </Label>
              <Select
                value={formData.insuranceDetails || ""}
                onValueChange={(value) => {
                  if (isOneOf(value, PET_INSURANCE_RATIO_VALUES)) {
                    setFormData({ ...formData, insuranceDetails: value });
                  }
                }}
              >
                <SelectTrigger className="h-10 text-sm bg-white text-[#37352F]">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {PET_INSURANCE_RATIO_VALUES.map((ratio) => (
                    <SelectItem key={ratio} value={ratio}>
                      {ratio}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1">
              <Label htmlFor="remarks" className="text-sm text-[#37352F]/60">
                備考・特記事項
              </Label>
              <Textarea
                id="remarks"
                rows={3}
                value={formData.remarks || ""}
                onChange={(e) =>
                  setFormData({ ...formData, remarks: e.target.value })
                }
                className="text-sm text-[#37352F] border-[rgba(55,53,47,0.16)] min-h-[80px] resize-none"
              />
            </div>
          </div>
        </div>

        <div className="flex justify-end mt-4 pt-4 border-t border-[rgba(55,53,47,0.09)]">
          <Button
            variant="outline"
            className="mr-2 h-10 text-sm"
            onClick={() => onOpenChange(false)}
          >
            キャンセル
          </Button>
          <Button
            onClick={handleSave}
            className="bg-[#37352F] hover:bg-[#37352F]/90 text-white h-10 text-sm px-4"
          >
            {isEdit ? "更新" : "登録"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
