import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "../../../components/ui/dialog";
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import { Label } from "../../../components/ui/label";
import { Textarea } from "../../../components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../components/ui/select";

interface PetEditModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  ownerName?: string;
  petData?: {
    petNumber: string;
    petName: string;
    petNameKana: string;
    species: string;
    gender: string;
    birthDate: string;
    breed: string;
    color: string;
    neuteredDate: string;
    acquisitionType: string;
    dangerLevel: string;
    food: string;
    insuranceName?: string;
    insuranceDetails?: string;
    remarks: string;
  };
  onSave: (data: any) => void;
}

const INSURANCE_COMPANIES = [
  "アニコム",
  "アイペット",
  "ペット＆ファミリー",
  "楽天ペット保険",
  "アクサダイレクト",
  "SBIいきいき少短",
  "FPC",
  "その他"
];

const INSURANCE_RATIOS = [
  "50%",
  "70%",
  "90%",
  "100%",
  "その他"
];

export default function PetEditModal({
  open,
  onOpenChange,
  ownerName = "飼主名",
  petData,
  onSave,
}: PetEditModalProps) {
  const [formData, setFormData] = useState({
    petNumber: petData?.petNumber || "",
    petName: petData?.petName || "",
    petNameKana: petData?.petNameKana || "",
    species: petData?.species || "",
    gender: petData?.gender || "",
    birthDate: petData?.birthDate || "",
    breed: petData?.breed || "",
    color: petData?.color || "",
    neuteredDate: petData?.neuteredDate || "",
    acquisitionType: petData?.acquisitionType || "",
    dangerLevel: petData?.dangerLevel || "",
    food: petData?.food || "",
    insuranceName: petData?.insuranceName || "",
    insuranceDetails: petData?.insuranceDetails || "",
    remarks: petData?.remarks || "",
  });

  useEffect(() => {
    if (open) {
      setFormData({
        petNumber: petData?.petNumber || "",
        petName: petData?.petName || "",
        petNameKana: petData?.petNameKana || "",
        species: petData?.species || "",
        gender: petData?.gender || "",
        birthDate: petData?.birthDate || "",
        breed: petData?.breed || "",
        color: petData?.color || "",
        neuteredDate: petData?.neuteredDate || "",
        acquisitionType: petData?.acquisitionType || "",
        dangerLevel: petData?.dangerLevel || "",
        food: petData?.food || "",
        insuranceName: petData?.insuranceName || "",
        insuranceDetails: petData?.insuranceDetails || "",
        remarks: petData?.remarks || "",
      });
    }
  }, [open, petData]);

  const handleSave = () => {
    onSave(formData);
    onOpenChange(false);
  };

  const isEdit = !!petData;

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
                ペット名(カナ) <span className="text-[#E03E3E]">*</span>
              </Label>
              <Input
                id="petNameKana"
                placeholder="例: イリス"
                value={formData.petNameKana}
                onChange={(e) =>
                  setFormData({ ...formData, petNameKana: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="species" className="text-sm text-[#37352F]/60">
                CFBE(種) <span className="text-[#E03E3E]">*</span>
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
              <Label htmlFor="gender1" className="text-sm text-[#37352F]/60">
                性別 <span className="text-[#E03E3E]">*</span>
              </Label>
              <Select
                value={formData.gender}
                onValueChange={(value) =>
                  setFormData({ ...formData, gender: value })
                }
              >
                <SelectTrigger className="h-10 text-sm bg-white text-[#37352F]">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="雄">雄</SelectItem>
                  <SelectItem value="雌">雌</SelectItem>
                  <SelectItem value="不明">不明</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1">
              <Label htmlFor="birthDate" className="text-sm text-[#37352F]/60">
                生年月日 <span className="text-[#E03E3E]">*</span>
              </Label>
              <Input
                id="birthDate"
                type="date"
                value={formData.birthDate}
                onChange={(e) =>
                  setFormData({ ...formData, birthDate: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>
          </div>

          {/* Column 2 */}
          <div className="space-y-2">
            <div className="space-y-1">
              <Label htmlFor="breed" className="text-sm text-[#37352F]/60">
                Breed
              </Label>
              <Input
                id="breed"
                value={formData.breed}
                onChange={(e) =>
                  setFormData({ ...formData, breed: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="color" className="text-sm text-[#37352F]/60">
                Color
              </Label>
              <Input
                id="color"
                value={formData.color}
                onChange={(e) =>
                  setFormData({ ...formData, color: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="neuteredDate" className="text-sm text-[#37352F]/60">
                去勢・避妊手術日
              </Label>
              <Input
                id="neuteredDate"
                type="date"
                value={formData.neuteredDate}
                onChange={(e) =>
                  setFormData({ ...formData, neuteredDate: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>
          </div>

          {/* Column 3 */}
          <div className="space-y-2">
            <div className="space-y-1">
              <Label htmlFor="acquisitionType" className="text-sm text-[#37352F]/60">
                入手区分
              </Label>
              <Select
                value={formData.acquisitionType}
                onValueChange={(value) =>
                  setFormData({ ...formData, acquisitionType: value })
                }
              >
                <SelectTrigger className="h-10 text-sm bg-white text-[#37352F]">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="購入">購入</SelectItem>
                  <SelectItem value="譲渡">譲渡</SelectItem>
                  <SelectItem value="保護">保護</SelectItem>
                  <SelectItem value="その他">その他</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1">
              <Label htmlFor="dangerLevel" className="text-sm text-[#37352F]/60">
                ペットの危険度
              </Label>
              <Select
                value={formData.dangerLevel}
                onValueChange={(value) =>
                  setFormData({ ...formData, dangerLevel: value })
                }
              >
                <SelectTrigger className="h-10 text-sm bg-white text-[#37352F]">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="低">低</SelectItem>
                  <SelectItem value="中">中</SelectItem>
                  <SelectItem value="高">高</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1">
              <Label htmlFor="food" className="text-sm text-[#37352F]/60">
                Food
              </Label>
              <Input
                id="food"
                value={formData.food}
                onChange={(e) =>
                  setFormData({ ...formData, food: e.target.value })
                }
                className="h-10 text-sm bg-white text-[#37352F]"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="insuranceName" className="text-sm text-[#37352F]/60">
                保険名
              </Label>
              <Select
                value={formData.insuranceName}
                onValueChange={(value) =>
                  setFormData({ ...formData, insuranceName: value })
                }
              >
                <SelectTrigger className="h-10 text-sm bg-white text-[#37352F]">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {INSURANCE_COMPANIES.map((company) => (
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
                value={formData.insuranceDetails}
                onValueChange={(value) =>
                  setFormData({ ...formData, insuranceDetails: value })
                }
              >
                <SelectTrigger className="h-10 text-sm bg-white text-[#37352F]">
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {INSURANCE_RATIOS.map((ratio) => (
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
                value={formData.remarks}
                onChange={(e) =>
                  setFormData({ ...formData, remarks: e.target.value })
                }
                className="text-sm text-[#37352F] border-[rgba(55,53,47,0.16)] min-h-[80px] resize-none"
              />
            </div>
          </div>
        </div>

        <div className="flex justify-end mt-4 pt-4 border-t border-[rgba(55,53,47,0.09)]">
           <Button variant="outline" className="mr-2 h-10 text-sm" onClick={() => onOpenChange(false)}>
             キャンセル
           </Button>
          <Button onClick={handleSave} className="bg-[#37352F] hover:bg-[#37352F]/90 text-white h-10 text-sm px-4">
            {isEdit ? "更新" : "登録"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
