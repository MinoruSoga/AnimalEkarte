import { FormFieldError } from "@/components/shared/FormFieldError";
import { DatePicker } from "@/components/shared/DatePicker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectTrigger, SelectValue } from "@/components/ui/select";
import { SearchableSelect, type SearchableSelectOption } from "@/components/ui/searchable-select";
import { C, STYLE } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";
import { isOneOf } from "@/lib/type-utils";

import { PET_GENDER_VALUES } from "../types";
import {
  LABEL_CLS,
  INPUT_CLS,
  GENDER_SELECT_ITEMS,
  type PetFieldSectionProps,
} from "./PetEditFieldShared";

export interface AnimalSpeciesOption {
  id: string | number;
  name: string;
  label?: string;
  isInactive?: boolean;
}

interface PetIdentitySectionProps extends PetFieldSectionProps {
  animalSpeciesOptions: SearchableSelectOption[];
  isLoadingSpecies: boolean;
  speciesPlaceholder: string;
  isEdit: boolean;
  onAnimalSpeciesChange: (value: string) => void;
}

export function PetIdentitySection({
  formData,
  setFormData,
  fieldErrors,
  clearFieldError,
  animalSpeciesOptions,
  isLoadingSpecies,
  speciesPlaceholder,
  isEdit,
  onAnimalSpeciesChange,
}: PetIdentitySectionProps) {
  return (
    <div className="space-y-2">
      <div className="space-y-1">
        <Label htmlFor="petNumber" className={LABEL_CLS}>
          ペットNo
        </Label>
        {isEdit ? (
          <Input
            id="petNumber"
            value={formData.petNumber}
            disabled
            className={`${INPUT_CLS} disabled:opacity-50`}
          />
        ) : (
          <p className={`flex h-9 items-center px-3 text-sm ${C.text40} italic`}>
            登録時に自動採番されます
          </p>
        )}
      </div>

      <div className="space-y-1">
        <Label htmlFor="petName" className={LABEL_CLS}>
          ペット名 <span className={C.textRequired}>*</span>
        </Label>
        <Input
          id="petName"
          value={formData.petName}
          maxLength={100}
          aria-invalid={!!fieldErrors.petName}
          aria-describedby={fieldErrors.petName ? "petName-error" : undefined}
          onChange={(e) => {
            setFormData((prev) => ({ ...prev, petName: e.target.value }));
            clearFieldError("petName");
          }}
          className={`${INPUT_CLS} ${fieldErrors.petName ? STYLE.formInputError : ""}`}
        />
        <FormFieldError id="petName-error" message={fieldErrors.petName} />
      </div>

      <div className="space-y-1">
        <Label htmlFor="petNameKana" className={LABEL_CLS}>
          ペット名よみ
        </Label>
        <Input
          id="petNameKana"
          placeholder="例: いりす"
          value={formData.petNameKana || ""}
          onChange={(e) => setFormData((prev) => ({ ...prev, petNameKana: e.target.value }))}
          className={INPUT_CLS}
        />
      </div>

      <div className="space-y-1">
        <Label htmlFor="animalSpeciesId" className={LABEL_CLS}>
          動物種 <span className={C.textRequired}>*</span>
        </Label>
        <SearchableSelect
          id="animalSpeciesId"
          value={formData.animalSpeciesId || ""}
          onValueChange={onAnimalSpeciesChange}
          options={animalSpeciesOptions}
          disabled={isLoadingSpecies}
          placeholder={speciesPlaceholder}
          searchPlaceholder="動物種を検索..."
          ariaInvalid={Boolean(fieldErrors.animalSpeciesId)}
          className={INPUT_CLS}
        />
        <FormFieldError id="animalSpeciesId-error" message={fieldErrors.animalSpeciesId} />
      </div>

      <div className="space-y-1">
        <Label htmlFor="gender" className={LABEL_CLS}>
          性別 <span className={C.textRequired}>*</span>
        </Label>
        <Select
          value={formData.gender}
          onValueChange={(value) => {
            if (isOneOf(value, PET_GENDER_VALUES)) {
              setFormData((prev) => ({ ...prev, gender: value }));
              clearFieldError("gender");
            }
          }}
        >
          <SelectTrigger
            className={`${INPUT_CLS} ${fieldErrors.gender ? STYLE.formInputError : ""}`}
          >
            <SelectValue placeholder="選択してください" />
          </SelectTrigger>
          <SelectContent>{GENDER_SELECT_ITEMS}</SelectContent>
        </Select>
        <FormFieldError id="gender-error" message={fieldErrors.gender} />
      </div>

      <div className="space-y-1">
        <Label htmlFor="birthDate" className={LABEL_CLS}>
          生年月日
        </Label>
        <DatePicker
          id="birthDate"
          value={formData.birthDate}
          onChange={(val) => setFormData((prev) => ({ ...prev, birthDate: val }))}
          placeholder="生年月日を選択…"
          disabledDays={{ after: toJSTWallDate(new Date()) }}
        />
      </div>
    </div>
  );
}
