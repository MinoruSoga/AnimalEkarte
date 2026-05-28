import type { Dispatch, SetStateAction } from "react";
import { useMemo } from "react";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { MasterLink } from "@/components/shared/MasterLink";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { PetDeceasedRecordButton } from "@/features/pets";
import { C, STYLE } from "@/lib/design-tokens";
import { isOneOf } from "@/lib/type-utils";
import {
  ACQUISITION_TYPE_VALUES,
  DANGER_LEVEL_VALUES,
  PET_GENDER_VALUES,
  type PetFormData,
} from "../types";

const LABEL_CLS = `text-sm ${C.text60}`;
const INPUT_CLS = STYLE.formInput;

const BREED_SUGGESTIONS: Record<string, string[]> = {
  "犬": [
    "柴犬", "トイプードル", "チワワ", "ダックスフンド", "フレンチブルドッグ",
    "ゴールデンレトリバー", "ラブラドールレトリバー", "ポメラニアン", "ビーグル",
    "シバイヌ(赤)", "シバイヌ(黒)", "ミニチュアシュナウザー", "マルチーズ",
    "ヨークシャーテリア", "シーズー", "ボーダーコリー", "コーギー", "ハスキー",
    "サモエド", "柴ミックス", "雑種",
  ],
  "猫": [
    "アメリカンショートヘア", "スコティッシュフォールド", "ロシアンブルー",
    "メインクーン", "ペルシャ", "ノルウェージャンフォレストキャット", "ラグドール",
    "ベンガル", "マンチカン", "ヒマラヤン", "アビシニアン", "バーマン",
    "ブリティッシュショートヘア", "日本猫", "雑種",
  ],
  "鳥": ["セキセイインコ", "オカメインコ", "コザクラインコ", "文鳥", "カナリア", "その他"],
  "ウサギ": ["ネザーランドドワーフ", "ホーランドロップ", "ミニレッキス", "その他"],
  "ハムスター": ["ゴールデンハムスター", "ジャンガリアン", "キャンベル", "その他"],
  "フェレット": ["フェレット"],
};

const GENDER_SELECT_ITEMS = PET_GENDER_VALUES.map((g) => (
  <SelectItem key={g} value={g}>{g}</SelectItem>
));

const ACQUISITION_SELECT_ITEMS = ACQUISITION_TYPE_VALUES.map((t) => (
  <SelectItem key={t} value={t}>{t}</SelectItem>
));

const DANGER_SELECT_ITEMS = DANGER_LEVEL_VALUES.map((d) => (
  <SelectItem key={d} value={d}>{d}</SelectItem>
));

interface AnimalSpeciesOption {
  id: string | number;
  name: string;
  label?: string;
  isInactive?: boolean;
}

interface InsuranceOption {
  id: string | number;
  name: string;
  coverage_rate?: number | null;
}

export function createPetFormData(petData?: PetFormData): PetFormData {
  return {
    id: petData?.id || "",
    petNumber: petData?.petNumber || "",
    petName: petData?.petName || "",
    petNameKana: petData?.petNameKana || "",
    species: petData?.species || "",
    animalSpeciesId: petData?.animalSpeciesId || "",
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
    insuranceId: petData?.insuranceId || "",
    insuranceName: petData?.insuranceName,
    insuranceDetails: petData?.insuranceDetails,
  };
}

interface PetEditModalFieldsProps {
  formData: PetFormData;
  setFormData: Dispatch<SetStateAction<PetFormData>>;
  fieldErrors: Record<string, string>;
  clearFieldError: (field: string) => void;
  animalSpeciesList: AnimalSpeciesOption[];
  isLoadingSpecies: boolean;
  insuranceList: InsuranceOption[];
  isLoadingInsurances: boolean;
  canEdit: boolean;
  isEdit: boolean;
}

export function PetEditModalFields({
  formData,
  setFormData,
  fieldErrors,
  clearFieldError,
  animalSpeciesList,
  isLoadingSpecies,
  insuranceList,
  isLoadingInsurances,
  canEdit,
  isEdit,
}: PetEditModalFieldsProps) {
  const animalSpeciesSelectItems = useMemo(() =>
    animalSpeciesList.map((s) => (
      <SelectItem key={s.id} value={String(s.id)} disabled={s.isInactive}>
        {s.label || s.name}
      </SelectItem>
    )),
    [animalSpeciesList]
  );

  const insuranceSelectItems = useMemo(() =>
    insuranceList.map((ins) => (
      <SelectItem key={ins.id} value={String(ins.id)}>
        {ins.name}{ins.coverage_rate != null ? ` (${ins.coverage_rate}%補償)` : ""}
      </SelectItem>
    )),
    [insuranceList]
  );

  const handleAnimalSpeciesChange = (value: string) => {
    const selected = animalSpeciesList.find((s) => String(s.id) === value);
    setFormData((prev) => ({
      ...prev,
      animalSpeciesId: value,
      species: selected?.name ?? prev.species,
      breed: "",
    }));
    clearFieldError("animalSpeciesId");
  };

  const handleInsuranceChange = (value: string) => {
    const actualValue = value === "none" ? "" : value;
    const selected = insuranceList.find((ins) => String(ins.id) === actualValue);
    setFormData((prev) => ({
      ...prev,
      insuranceId: actualValue,
      insuranceName: selected?.name as PetFormData["insuranceName"],
      insuranceDetails: actualValue === ""
        ? undefined
        : selected?.coverage_rate != null
          ? `${selected.coverage_rate}%補償`
          : prev.insuranceDetails,
    }));
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
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
          <Select
            value={formData.animalSpeciesId || ""}
            onValueChange={handleAnimalSpeciesChange}
            disabled={isLoadingSpecies}
          >
            <SelectTrigger className={`${INPUT_CLS} ${fieldErrors.animalSpeciesId ? STYLE.formInputError : ""}`}>
              <SelectValue placeholder={isLoadingSpecies ? "読み込み中..." : "選択してください"} />
            </SelectTrigger>
            <SelectContent>{animalSpeciesSelectItems}</SelectContent>
          </Select>
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
          <NotionDatePicker
            id="birthDate"
            value={formData.birthDate}
            onChange={(val) => setFormData((prev) => ({ ...prev, birthDate: val }))}
            placeholder="生年月日を選択…"
            disabledDays={{ after: new Date() }}
          />
        </div>
      </div>

      <div className="space-y-2">
        <div className="space-y-1">
          <Label htmlFor="breed" className={LABEL_CLS}>品種</Label>
          <Input
            id="breed"
            list={BREED_SUGGESTIONS[formData.species] ? "breed-suggestions" : undefined}
            value={formData.breed || ""}
            onChange={(e) => setFormData((prev) => ({ ...prev, breed: e.target.value }))}
            placeholder={BREED_SUGGESTIONS[formData.species] ? "品種を選択または入力..." : undefined}
            className={INPUT_CLS}
          />
          {BREED_SUGGESTIONS[formData.species] ? (
            <datalist id="breed-suggestions">
              {BREED_SUGGESTIONS[formData.species].map((breed) => (
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
          <NotionDatePicker
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
            onValueChange={handleInsuranceChange}
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
    </div>
  );
}
