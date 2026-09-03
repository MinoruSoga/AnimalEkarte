import { FormFieldError } from "@/components/shared/FormFieldError";
import { DatePicker } from "@/components/shared/DatePicker";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { C, STYLE } from "@/lib/design-tokens";
import { isOneOf } from "@/lib/type-utils";

import { ACQUISITION_TYPE_VALUES, DANGER_LEVEL_VALUES } from "../types";
import {
  LABEL_CLS,
  INPUT_CLS,
  BREED_SUGGESTIONS,
  ACQUISITION_SELECT_ITEMS,
  DANGER_SELECT_ITEMS,
  type PetFieldSectionProps,
} from "./PetEditFieldShared";

export function PetPhysicalSection({
  formData,
  setFormData,
  fieldErrors,
  clearFieldError,
}: PetFieldSectionProps) {
  const breedSuggestions = BREED_SUGGESTIONS[formData.species];

  return (
    <div className="space-y-2">
      <div className="space-y-1">
        <Label htmlFor="breed" className={LABEL_CLS}>
          品種
        </Label>
        <Input
          id="breed"
          list={breedSuggestions ? "breed-suggestions" : undefined}
          value={formData.breed || ""}
          onChange={(e) => setFormData((prev) => ({ ...prev, breed: e.target.value }))}
          placeholder={breedSuggestions ? "品種を選択または入力..." : undefined}
          className={INPUT_CLS}
        />
        {breedSuggestions ? (
          <datalist id="breed-suggestions">
            {breedSuggestions.map((breed) => (
              <option key={breed} value={breed} />
            ))}
          </datalist>
        ) : null}
      </div>

      <div className="space-y-1">
        <Label htmlFor="color" className={LABEL_CLS}>
          毛色
        </Label>
        <Input
          id="color"
          value={formData.color || ""}
          onChange={(e) => setFormData((prev) => ({ ...prev, color: e.target.value }))}
          className={INPUT_CLS}
        />
      </div>

      <div className="space-y-1">
        <Label htmlFor="bloodType" className={LABEL_CLS}>
          血液型
        </Label>
        <Input
          id="bloodType"
          value={formData.bloodType || ""}
          maxLength={32}
          onChange={(e) => setFormData((prev) => ({ ...prev, bloodType: e.target.value }))}
          className={INPUT_CLS}
        />
      </div>

      <div className="space-y-1">
        <Label htmlFor="microchipNumber" className={LABEL_CLS}>
          マイクロチップ番号
        </Label>
        <Input
          id="microchipNumber"
          value={formData.microchipNumber || ""}
          maxLength={64}
          onChange={(e) => setFormData((prev) => ({ ...prev, microchipNumber: e.target.value }))}
          className={INPUT_CLS}
        />
      </div>

      <div className="space-y-1">
        <Label htmlFor="weight" className={LABEL_CLS}>
          体重(kg)
        </Label>
        <NumberInput
          id="weight"
          min={0}
          step={0.1}
          value={formData.weight || ""}
          aria-invalid={!!fieldErrors.weight}
          aria-describedby={fieldErrors.weight ? "weight-error" : undefined}
          onChange={(v) => {
            setFormData((prev) => ({ ...prev, weight: v }));
            clearFieldError("weight");
          }}
          suffix="kg"
          className={`${INPUT_CLS} ${fieldErrors.weight ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="weight-error" message={fieldErrors.weight} />
      </div>

      <div className="space-y-1">
        <Label htmlFor="neuteredDate" className={LABEL_CLS}>
          去勢・避妊手術日
        </Label>
        <DatePicker
          id="neuteredDate"
          value={formData.neuteredDate || ""}
          onChange={(val) => setFormData((prev) => ({ ...prev, neuteredDate: val }))}
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
              setFormData((prev) => ({ ...prev, acquisitionType: value }));
            }
          }}
        >
          <SelectTrigger className={INPUT_CLS}>
            <SelectValue placeholder="選択してください" />
          </SelectTrigger>
          <SelectContent>{ACQUISITION_SELECT_ITEMS}</SelectContent>
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
              setFormData((prev) => ({ ...prev, dangerLevel: value }));
              if (value !== "高") {
                clearFieldError("dangerReason");
              }
            }
          }}
        >
          <SelectTrigger id="dangerLevel" className={INPUT_CLS}>
            <SelectValue placeholder="選択してください" />
          </SelectTrigger>
          <SelectContent>{DANGER_SELECT_ITEMS}</SelectContent>
        </Select>
      </div>

      <div className="space-y-1">
        <Label htmlFor="dangerReason" className={LABEL_CLS}>
          危険と判断した理由
          {formData.dangerLevel === "高" ? (
            <span className={C.textRequired} aria-hidden="true">
              *
            </span>
          ) : null}
        </Label>
        <Textarea
          id="dangerReason"
          value={formData.dangerReason || ""}
          placeholder="危険と判断した理由を入力してください（500文字以内）"
          aria-label="危険と判断した理由"
          aria-required={formData.dangerLevel === "高"}
          aria-invalid={!!fieldErrors.dangerReason}
          aria-describedby={fieldErrors.dangerReason ? "dangerReason-error" : undefined}
          onChange={(e) => {
            setFormData((prev) => ({ ...prev, dangerReason: e.target.value }));
            clearFieldError("dangerReason");
          }}
          className={`${STYLE.textarea} ${fieldErrors.dangerReason ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="dangerReason-error" message={fieldErrors.dangerReason} />
      </div>
    </div>
  );
}
