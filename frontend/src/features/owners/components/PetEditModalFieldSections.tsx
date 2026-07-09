import type { Dispatch, ReactNode, SetStateAction } from "react";

import { FormFieldError } from "@/components/shared/FormFieldError";
import { MasterLink } from "@/components/shared/MasterLink";
import { DatePicker } from "@/components/shared/DatePicker";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { PetDeceasedRecordButton } from "@/components/shared/PetDeceasedRecordButton/PetDeceasedRecordButton";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { SearchableSelect, type SearchableSelectOption } from "@/components/ui/searchable-select";
import { Textarea } from "@/components/ui/textarea";
import { C, STYLE } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";
import { isOneOf } from "@/lib/type-utils";

import {
  ACQUISITION_TYPE_VALUES,
  DANGER_LEVEL_VALUES,
  PET_GENDER_VALUES,
  type PetFormData,
} from "../types";

export interface AnimalSpeciesOption {
  id: string | number;
  name: string;
  label?: string;
  isInactive?: boolean;
}

export interface InsuranceOption {
  id: string | number;
  name: string;
  coverage_rate?: number | null;
}

const LABEL_CLS = `text-sm ${C.text60}`;
const INPUT_CLS = STYLE.formInput;

const BREED_SUGGESTIONS: Record<string, string[]> = {
  "犬": [
    "柴犬",
    "トイプードル",
    "チワワ",
    "ダックスフンド",
    "フレンチブルドッグ",
    "ゴールデンレトリバー",
    "ラブラドールレトリバー",
    "ポメラニアン",
    "ビーグル",
    "シバイヌ(赤)",
    "シバイヌ(黒)",
    "ミニチュアシュナウザー",
    "マルチーズ",
    "ヨークシャーテリア",
    "シーズー",
    "ボーダーコリー",
    "コーギー",
    "ハスキー",
    "サモエド",
    "柴ミックス",
    "雑種",
  ],
  "猫": [
    "アメリカンショートヘア",
    "スコティッシュフォールド",
    "ロシアンブルー",
    "メインクーン",
    "ペルシャ",
    "ノルウェージャンフォレストキャット",
    "ラグドール",
    "ベンガル",
    "マンチカン",
    "ヒマラヤン",
    "アビシニアン",
    "バーマン",
    "ブリティッシュショートヘア",
    "日本猫",
    "雑種",
  ],
  "鳥": ["セキセイインコ", "オカメインコ", "コザクラインコ", "文鳥", "カナリア", "その他"],
  "ウサギ": ["ネザーランドドワーフ", "ホーランドロップ", "ミニレッキス", "その他"],
  "ハムスター": ["ゴールデンハムスター", "ジャンガリアン", "キャンベル", "その他"],
  "フェレット": ["フェレット"],
};

const GENDER_SELECT_ITEMS = PET_GENDER_VALUES.map((g) => (
  <SelectItem key={g} value={g}>
    {g}
  </SelectItem>
));

const ACQUISITION_SELECT_ITEMS = ACQUISITION_TYPE_VALUES.map((t) => (
  <SelectItem key={t} value={t}>
    {t}
  </SelectItem>
));

const DANGER_SELECT_ITEMS = DANGER_LEVEL_VALUES.map((d) => (
  <SelectItem key={d} value={d}>
    {d}
  </SelectItem>
));

interface PetFieldSectionProps {
  formData: PetFormData;
  setFormData: Dispatch<SetStateAction<PetFormData>>;
  fieldErrors: Record<string, string>;
  clearFieldError: (field: string) => void;
}

interface PetIdentitySectionProps extends PetFieldSectionProps {
  animalSpeciesOptions: SearchableSelectOption[];
  isLoadingSpecies: boolean;
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
  isEdit,
  onAnimalSpeciesChange,
}: PetIdentitySectionProps) {
  return (
    <div className="space-y-2">
      <div className="space-y-1">
        <Label htmlFor="petNumber" className={LABEL_CLS}>ペットNo</Label>
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
        <Label htmlFor="petNameKana" className={LABEL_CLS}>ペット名よみ</Label>
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
          placeholder={isLoadingSpecies ? "読み込み中..." : "選択してください"}
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
          <SelectTrigger className={`${INPUT_CLS} ${fieldErrors.gender ? STYLE.formInputError : ""}`}>
            <SelectValue placeholder="選択してください" />
          </SelectTrigger>
          <SelectContent>{GENDER_SELECT_ITEMS}</SelectContent>
        </Select>
        <FormFieldError id="gender-error" message={fieldErrors.gender} />
      </div>

      <div className="space-y-1">
        <Label htmlFor="birthDate" className={LABEL_CLS}>生年月日</Label>
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
        <Label htmlFor="breed" className={LABEL_CLS}>品種</Label>
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
        <Label htmlFor="color" className={LABEL_CLS}>毛色</Label>
        <Input
          id="color"
          value={formData.color || ""}
          onChange={(e) => setFormData((prev) => ({ ...prev, color: e.target.value }))}
          className={INPUT_CLS}
        />
      </div>

      <div className="space-y-1">
        <Label htmlFor="bloodType" className={LABEL_CLS}>血液型</Label>
        <Input
          id="bloodType"
          value={formData.bloodType || ""}
          maxLength={32}
          onChange={(e) => setFormData((prev) => ({ ...prev, bloodType: e.target.value }))}
          className={INPUT_CLS}
        />
      </div>

      <div className="space-y-1">
        <Label htmlFor="microchipNumber" className={LABEL_CLS}>マイクロチップ番号</Label>
        <Input
          id="microchipNumber"
          value={formData.microchipNumber || ""}
          maxLength={64}
          onChange={(e) => setFormData((prev) => ({ ...prev, microchipNumber: e.target.value }))}
          className={INPUT_CLS}
        />
      </div>

      <div className="space-y-1">
        <Label htmlFor="weight" className={LABEL_CLS}>体重(kg)</Label>
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
        <Label htmlFor="neuteredDate" className={LABEL_CLS}>去勢・避妊手術日</Label>
        <DatePicker
          id="neuteredDate"
          value={formData.neuteredDate || ""}
          onChange={(val) => setFormData((prev) => ({ ...prev, neuteredDate: val }))}
          placeholder="手術日を選択…"
        />
      </div>

      <div className="space-y-1">
        <Label htmlFor="acquisitionType" className={LABEL_CLS}>入手区分</Label>
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
        <Label htmlFor="dangerLevel" className={LABEL_CLS}>ペットの危険度</Label>
        <Select
          value={formData.dangerLevel || ""}
          onValueChange={(value) => {
            if (isOneOf(value, DANGER_LEVEL_VALUES)) {
              setFormData((prev) => ({ ...prev, dangerLevel: value }));
            }
          }}
        >
          <SelectTrigger className={INPUT_CLS}>
            <SelectValue placeholder="選択してください" />
          </SelectTrigger>
          <SelectContent>{DANGER_SELECT_ITEMS}</SelectContent>
        </Select>
      </div>
    </div>
  );
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
