import type { Dispatch, ReactNode, SetStateAction } from "react";

import { MasterLink } from "@/components/shared/MasterLink";
import { PetDeceasedRecordButton } from "@/components/shared/PetDeceasedRecordButton/PetDeceasedRecordButton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { C } from "@/lib/design-tokens";

import type { PetFormData } from "../types";
import { LABEL_CLS, INPUT_CLS } from "./pet-edit-field-shared";

export interface InsuranceOption {
  id: string | number;
  name: string;
  coverage_rate?: number | null;
}

interface PetCareSectionProps {
  formData: PetFormData;
  setFormData: Dispatch<SetStateAction<PetFormData>>;
  insuranceSelectItems: ReactNode;
  isLoadingInsurances: boolean;
  canEdit: boolean;
  onInsuranceChange: (value: string) => void;
}

export function PetCareSection({
  formData,
  setFormData,
  insuranceSelectItems,
  isLoadingInsurances,
  canEdit,
  onInsuranceChange,
}: PetCareSectionProps) {
  return (
    <div className="space-y-2">
      <div className="space-y-1">
        <Label htmlFor="food" className={LABEL_CLS}>食べ物</Label>
        <Input
          id="food"
          value={formData.food || ""}
          onChange={(e) => setFormData((prev) => ({ ...prev, food: e.target.value }))}
          className={INPUT_CLS}
        />
      </div>

      <div className="space-y-1">
        <Label htmlFor="environment" className={LABEL_CLS}>飼育環境</Label>
        <Input
          id="environment"
          value={formData.environment || ""}
          onChange={(e) => setFormData((prev) => ({ ...prev, environment: e.target.value }))}
          className={INPUT_CLS}
        />
      </div>

      <div className="space-y-1">
        <div className="flex items-center justify-between">
          <Label htmlFor="insuranceId" className={LABEL_CLS}>保険</Label>
          <MasterLink category="insurance" label="編集" className="text-[11px]" />
        </div>
        <Select
          value={formData.insuranceId || "none"}
          onValueChange={onInsuranceChange}
          disabled={isLoadingInsurances}
        >
          <SelectTrigger className={INPUT_CLS}>
            <SelectValue placeholder={isLoadingInsurances ? "読み込み中..." : "保険を選択"} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="none">なし</SelectItem>
            {insuranceSelectItems}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-1">
        <Label className={LABEL_CLS}>生死ステータス</Label>
        <div className="flex gap-4 h-9 items-center">
          <label className={`flex items-center gap-1.5 text-sm cursor-pointer ${C.text}`}>
            <input
              type="radio"
              name="petStatus"
              value="生存"
              checked={formData.status === "生存"}
              onChange={() => setFormData((prev) => ({ ...prev, status: "生存" }))}
              className="accent-current"
            />
            生存
          </label>
          <label className={`flex items-center gap-1.5 text-sm cursor-pointer ${C.text}`}>
            <input
              type="radio"
              name="petStatus"
              value="死亡"
              checked={formData.status === "死亡"}
              onChange={() => setFormData((prev) => ({ ...prev, status: "死亡" }))}
              className="accent-current"
            />
            死亡
          </label>
        </div>
        {formData.id && formData.status === "生存" ? (
          <div className="pt-1">
            <PetDeceasedRecordButton
              petId={formData.id}
              petName={formData.petName}
              petBreed={formData.breed}
              petGender={formData.gender}
              birthDate={formData.birthDate}
              deceasedAt={null}
              canEdit={canEdit}
            />
          </div>
        ) : null}
      </div>

      <div className="space-y-1">
        <Label htmlFor="remarks" className={LABEL_CLS}>備考・特記事項</Label>
        <Textarea
          id="remarks"
          rows={3}
          value={formData.remarks || ""}
          onChange={(e) => setFormData((prev) => ({ ...prev, remarks: e.target.value }))}
          className={`text-sm ${C.text} ${C.borderMedium} min-h-[80px] resize-none`}
        />
      </div>
    </div>
  );
}
