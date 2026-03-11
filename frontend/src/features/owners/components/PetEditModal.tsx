import { useState, useCallback } from "react";
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
import { FormFieldError } from "@/components/shared/FormFieldError";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import {
  PET_GENDER_VALUES,
  ACQUISITION_TYPE_VALUES,
  DANGER_LEVEL_VALUES,
  INSURANCE_COMPANY_VALUES,
  PET_INSURANCE_RATIO_VALUES,
  PET_SPECIES_VALUES,
  PetFormData,
} from "../types";
import { isOneOf } from "@/lib/type-utils";

const LABEL_CLS = `text-sm ${C.text60}`;
const INPUT_CLS = STYLE.formInput;

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
  const [formData, setFormData] = useState<PetFormData>(() => ({
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
  }));
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const clearFieldError = useCallback((field: string) => {
    setFieldErrors((prev) => {
      const next = { ...prev };
      delete next[field];
      return next;
    });
  }, []);

  const handleSave = () => {
    const errors: Record<string, string> = {};
    if (!formData.petName.trim()) errors.petName = "ペット名を入力してください";
    if (!formData.species) errors.species = "種別を選択してください";
    if (!formData.gender) errors.gender = "性別を選択してください";
    if (!formData.birthDate) errors.birthDate = "生年月日を選択してください";

    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      toast.error("必須項目が未入力です", {
        description: Object.values(errors).join("、"),
      });
      return;
    }

    setFieldErrors({});
    onSave(formData);
    onOpenChange(false);
    toast.success(petData?.id ? "ペット情報を更新しました" : "ペットを追加しました");
  };

  const isEdit = !!petData?.id;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={`${LAYOUT.modal.xl} overflow-y-auto`}>
        <DialogHeader>
          <DialogTitle className={`text-sm font-bold ${C.text}`}>
            {isEdit ? `${ownerName}のペット情報編集` : `${ownerName}のペット新規登録`}
          </DialogTitle>
          <DialogDescription className={`text-sm ${C.text60}`}>
            {isEdit
              ? "ペットの情報を編集してください。"
              : "ペットの基本情報を入力してください。"}
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {/* Column 1 */}
          <div className="space-y-2">
            <div className="space-y-1">
              <Label htmlFor="petNumber" className={LABEL_CLS}>
                ペットNo
              </Label>
              <Input
                id="petNumber"
                value={formData.petNumber}
                onChange={(e) =>
                  setFormData(prev => ({ ...prev, petNumber: e.target.value }))
                }
                className={INPUT_CLS}
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="petName" className={LABEL_CLS}>
                ペット名 <span className={C.textRequired}>*</span>
              </Label>
              <Input
                id="petName"
                value={formData.petName}
                aria-invalid={!!fieldErrors.petName}
                aria-describedby={fieldErrors.petName ? "petName-error" : undefined}
                onChange={(e) => {
                  setFormData(prev => ({ ...prev, petName: e.target.value }));
                  clearFieldError("petName");
                }}
                className={`${INPUT_CLS} ${fieldErrors.petName ? STYLE.formInputError : ""}`}
              />
              <FormFieldError id="petName-error" message={fieldErrors.petName} />
            </div>

            <div className="space-y-1">
              <Label htmlFor="petNameKana" className={LABEL_CLS}>
                ペット名(カナ)
              </Label>
              <Input
                id="petNameKana"
                placeholder="例: イリス"
                value={formData.petNameKana || ""}
                onChange={(e) =>
                  setFormData(prev => ({ ...prev, petNameKana: e.target.value }))
                }
                className={INPUT_CLS}
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="species" className={LABEL_CLS}>
                種 <span className={C.textRequired}>*</span>
              </Label>
              <Select
                value={formData.species}
                onValueChange={(value) => {
                  if (isOneOf(value, PET_SPECIES_VALUES)) {
                    setFormData(prev => ({ ...prev, species: value }));
                    clearFieldError("species");
                  }
                }}
              >
                <SelectTrigger
                  className={`${INPUT_CLS} ${fieldErrors.species ? STYLE.formInputError : ""}`}
                >
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {PET_SPECIES_VALUES.map((s) => (
                    <SelectItem key={s} value={s}>
                      {s}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormFieldError message={fieldErrors.species} />
            </div>

            <div className="space-y-1">
              <Label htmlFor="gender" className={LABEL_CLS}>
                性別 <span className={C.textRequired}>*</span>
              </Label>
              <Select
                value={formData.gender}
                onValueChange={(value) => {
                  if (isOneOf(value, PET_GENDER_VALUES)) {
                    setFormData(prev => ({ ...prev, gender: value }));
                    clearFieldError("gender");
                  }
                }}
              >
                <SelectTrigger
                  className={`${INPUT_CLS} ${fieldErrors.gender ? STYLE.formInputError : ""}`}
                >
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {PET_GENDER_VALUES.map((g) => (
                    <SelectItem key={g} value={g}>
                      {g}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FormFieldError message={fieldErrors.gender} />
            </div>

            <div className="space-y-1">
              <Label htmlFor="birthDate" className={LABEL_CLS}>
                生年月日 <span className={C.textRequired}>*</span>
              </Label>
              <NotionDatePicker
                id="birthDate"
                value={formData.birthDate}
                onChange={(val) => {
                  setFormData(prev => ({ ...prev, birthDate: val }));
                  clearFieldError("birthDate");
                }}
                placeholder="生年月日を選択…"
              />
              <FormFieldError message={fieldErrors.birthDate} />
            </div>
          </div>

          {/* Column 2 */}
          <div className="space-y-2">
            <div className="space-y-1">
              <Label htmlFor="breed" className={LABEL_CLS}>
                品種
              </Label>
              <Input
                id="breed"
                value={formData.breed || ""}
                onChange={(e) =>
                  setFormData(prev => ({ ...prev, breed: e.target.value }))
                }
                className={INPUT_CLS}
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="color" className={LABEL_CLS}>
                毛色
              </Label>
              <Input
                id="color"
                value={formData.color || ""}
                onChange={(e) =>
                  setFormData(prev => ({ ...prev, color: e.target.value }))
                }
                className={INPUT_CLS}
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="weight" className={LABEL_CLS}>
                体重(kg)
              </Label>
              <Input
                id="weight"
                value={formData.weight || ""}
                onChange={(e) =>
                  setFormData(prev => ({ ...prev, weight: e.target.value }))
                }
                className={INPUT_CLS}
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="neuteredDate" className={LABEL_CLS}>
                去勢・避妊手術日
              </Label>
              <NotionDatePicker
                id="neuteredDate"
                value={formData.neuteredDate || ""}
                onChange={(val) =>
                  setFormData(prev => ({ ...prev, neuteredDate: val }))
                }
                placeholder="手術日を選択…"
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="acquisitionType" className={LABEL_CLS}>
                入手区分
              </Label>
              <Select
                value={formData.acquisitionType || ""}
                onValueChange={(value) => {
                  if (isOneOf(value, ACQUISITION_TYPE_VALUES)) {
                    setFormData(prev => ({ ...prev, acquisitionType: value }));
                  }
                }}
              >
                <SelectTrigger className={INPUT_CLS}>
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {ACQUISITION_TYPE_VALUES.map((t) => (
                    <SelectItem key={t} value={t}>
                      {t}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1">
              <Label htmlFor="dangerLevel" className={LABEL_CLS}>
                ペットの危険度
              </Label>
              <Select
                value={formData.dangerLevel || ""}
                onValueChange={(value) => {
                  if (isOneOf(value, DANGER_LEVEL_VALUES)) {
                    setFormData(prev => ({ ...prev, dangerLevel: value }));
                  }
                }}
              >
                <SelectTrigger className={INPUT_CLS}>
                  <SelectValue placeholder="選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {DANGER_LEVEL_VALUES.map((d) => (
                    <SelectItem key={d} value={d}>
                      {d}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Column 3 */}
          <div className="space-y-2">
            <div className="space-y-1">
              <Label htmlFor="food" className={LABEL_CLS}>
                食べ物
              </Label>
              <Input
                id="food"
                value={formData.food || ""}
                onChange={(e) =>
                  setFormData(prev => ({ ...prev, food: e.target.value }))
                }
                className={INPUT_CLS}
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="environment" className={LABEL_CLS}>
                飼育環境
              </Label>
              <Input
                id="environment"
                value={formData.environment || ""}
                onChange={(e) =>
                  setFormData(prev => ({ ...prev, environment: e.target.value }))
                }
                className={INPUT_CLS}
              />
            </div>

            <div className="space-y-1">
              <Label htmlFor="insuranceName" className={LABEL_CLS}>
                保険名
              </Label>
              <Select
                value={formData.insuranceName || ""}
                onValueChange={(value) => {
                  if (isOneOf(value, INSURANCE_COMPANY_VALUES)) {
                    setFormData(prev => ({ ...prev, insuranceName: value }));
                  }
                }}
              >
                <SelectTrigger className={INPUT_CLS}>
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
              <Label htmlFor="insuranceDetails" className={LABEL_CLS}>
                保険詳細(負担割合など)
              </Label>
              <Select
                value={formData.insuranceDetails || ""}
                onValueChange={(value) => {
                  if (isOneOf(value, PET_INSURANCE_RATIO_VALUES)) {
                    setFormData(prev => ({ ...prev, insuranceDetails: value }));
                  }
                }}
              >
                <SelectTrigger className={INPUT_CLS}>
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
              <Label htmlFor="remarks" className={LABEL_CLS}>
                備考・特記事項
              </Label>
              <Textarea
                id="remarks"
                rows={3}
                value={formData.remarks || ""}
                onChange={(e) =>
                  setFormData(prev => ({ ...prev, remarks: e.target.value }))
                }
                className={`text-sm ${C.text} ${C.borderMedium} min-h-[80px] resize-none`}
              />
            </div>
          </div>
        </div>

        <div className={`flex justify-end gap-2 mt-4 pt-4 border-t ${C.borderDivider}`}>
          <Button
            variant="outline"
            className={`h-10 text-sm ${C.borderMedium}`}
            onClick={() => onOpenChange(false)}
          >
            キャンセル
          </Button>
          <Button
            onClick={handleSave}
            className={`${STYLE.confirmPrimary} h-10 text-sm px-4`}
          >
            {isEdit ? "更新" : "登録"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
