import type { Dispatch, ReactNode, SetStateAction } from "react";

import { MasterLink } from "@/components/shared/MasterLink";
import { PetDeceasedRecordButton } from "@/components/shared/PetDeceasedRecordButton/PetDeceasedRecordButton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { C } from "@/lib/design-tokens";

import { isPersistedPetId } from "@/lib/pet-id";

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
  /** BUG-002: 専用 lifecycle 成功後に外側 pets 一覧を同期する（失敗時は呼ばない） */
  onPetLifecycleChange?: (result: {
    petId: string;
    status: "死亡" | "生存";
    deceasedAt: string | null;
    deceasedReason?: string | null;
  }) => void;
}

export function PetCareSection({
  formData,
  setFormData,
  insuranceSelectItems,
  isLoadingInsurances,
  canEdit,
  onInsuranceChange,
  onPetLifecycleChange,
}: PetCareSectionProps) {
  // BUG-022: pending ペットはサーバ未登録のため死亡記録 API を出さない
  const targetPetId =
    formData.isPending === true || !isPersistedPetId(formData.id)
      ? undefined
      : formData.id;

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
          <MasterLink category="insurance" label="編集" className="text-2xs" />
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
        {/* BUG-409: 生死は deceased_at・監査ログと一体の事実であり、下の PetDeceasedRecordButton
            （死亡登録・記録解除）経由の一箇所でのみ変更できる。ここは現在値の表示専用（disabled）。
            直接クリックで status だけを書き換えられると deceased_at と食い違う不整合を再導入する。 */}
        <div className="flex gap-4 h-9 items-center">
          <label className={`flex items-center gap-1.5 text-sm ${C.text}`}>
            <input
              type="radio"
              name="petStatus"
              value="生存"
              checked={formData.status === "生存"}
              disabled
              className="accent-current"
            />
            生存
          </label>
          <label className={`flex items-center gap-1.5 text-sm ${C.text}`}>
            <input
              type="radio"
              name="petStatus"
              value="死亡"
              checked={formData.status === "死亡"}
              disabled
              className="accent-current"
            />
            死亡
          </label>
        </div>
        <p className={`text-xs ${C.textMuted}`}>
          生死の変更は下記のボタンから行ってください
        </p>
        {targetPetId ? (
          <div className="pt-1">
            <PetDeceasedRecordButton
              petId={targetPetId}
              petName={formData.petName}
              petBreed={formData.breed}
              petGender={formData.gender}
              birthDate={formData.birthDate}
              deceasedAt={formData.deceasedAt ?? null}
              deceasedReason={formData.deceasedReason}
              petStatus={formData.status}
              canEdit={canEdit}
              onRecorded={({ deceasedAt, deceasedReason }) => {
                setFormData((prev) =>
                  prev.id === targetPetId
                    ? {
                        ...prev,
                        status: "死亡",
                        deceasedAt,
                        deceasedReason: deceasedReason ?? null,
                      }
                    : prev,
                );
                // BUG-002: mutation 成功後のみ外側一覧へ通知（API は 204・本文なし）
                onPetLifecycleChange?.({
                  petId: targetPetId,
                  status: "死亡",
                  deceasedAt,
                  deceasedReason: deceasedReason ?? null,
                });
              }}
              onRevoked={() => {
                setFormData((prev) =>
                  prev.id === targetPetId
                    ? { ...prev, status: "生存", deceasedAt: null, deceasedReason: null }
                    : prev,
                );
                onPetLifecycleChange?.({
                  petId: targetPetId,
                  status: "生存",
                  deceasedAt: null,
                  deceasedReason: null,
                });
              }}
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
