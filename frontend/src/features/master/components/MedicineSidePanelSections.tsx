import type { Dispatch, SetStateAction } from "react";
import Pill from "lucide-react/dist/esm/icons/pill";

import { PropertyInput, PropertyRow } from "@/components/shared/SidePeek";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { TaxRateSelector } from "@/components/shared/TaxRateSelector/TaxRateSelector";
import { TaxTypeSelector } from "@/components/shared/TaxTypeSelector/TaxTypeSelector";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { FIELD_DESCRIPTIONS } from "@/constants/field-descriptions";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import type { Medicine } from "@/types";
import {
  MedicineCalculationTypeNone,
  MedicineCalculationTypePerWeight,
} from "@/types/generated/models";

import type { MedicineFormData } from "../lib/medicine-side-panel-model";

const DOSAGE_FORM_SELECT_ITEMS = (
  <>
    <SelectItem value="tablet">錠剤</SelectItem>
    <SelectItem value="liquid">液剤</SelectItem>
    <SelectItem value="injection">注射剤</SelectItem>
    <SelectItem value="topical">外用剤</SelectItem>
    <SelectItem value="powder">散剤</SelectItem>
  </>
);

const MEDICINE_UNIT_SELECT_ITEMS = (
  <>
    <SelectItem value="per_tablet">1錠あたり</SelectItem>
    <SelectItem value="per_ml">1mlあたり</SelectItem>
    <SelectItem value="per_dose">1回あたり</SelectItem>
    <SelectItem value="per_gram">1gあたり</SelectItem>
  </>
);

// ts-review-201 MEDIUM: MedicineDoseParamsEditor.tsx と共有するため export する（重複定義の解消）。
export const SELECT_TRIGGER_FULL = `h-[30px] text-base bg-transparent ${C.text} border-0 ${C.hoverBgLight} px-1.5 shadow-none rounded-xxs w-full`;

type SetMedicineFormDataDirty = Dispatch<SetStateAction<MedicineFormData>>;

interface MedicineParentCategorySectionProps {
  formData: MedicineFormData;
  isCategory: boolean;
  categoryMedicines: Medicine[];
  setFormDataDirty: SetMedicineFormDataDirty;
}

export function MedicineParentCategorySection({
  formData,
  isCategory,
  categoryMedicines,
  setFormDataDirty,
}: MedicineParentCategorySectionProps) {
  return (
    <PropertyRow
      label="親カテゴリ"
      description="この薬品が属する親カテゴリです。薬品マスタを階層構造で整理するために使われます。"
    >
      {isCategory ? (
        <span className={`text-base ${C.text}`}>なし（ルート）</span>
      ) : (
        <SearchableSelect
          value={formData.parentId || "__none__"}
          onValueChange={(value) =>
            setFormDataDirty((prev) => ({
              ...prev,
              parentId: value === "__none__" ? "" : value,
            }))
          }
          options={[
            { value: "__none__", label: "なし（未分類）" },
            ...categoryMedicines.map((category) => ({ value: category.id, label: category.name })),
          ]}
          searchPlaceholder="親カテゴリを検索..."
          className={SELECT_TRIGGER_FULL}
        />
      )}
    </PropertyRow>
  );
}

interface MedicinePriceTaxSectionProps {
  formData: MedicineFormData;
  isCategory: boolean;
  setFormDataDirty: SetMedicineFormDataDirty;
}

export function MedicinePriceTaxSection({
  formData,
  isCategory,
  setFormDataDirty,
}: MedicinePriceTaxSectionProps) {
  return (
    <>
      <PropertyRow label="単価(税込)" description={FIELD_DESCRIPTIONS.unitPrice}>
        {isCategory ? (
          <span className={`text-base ${C.text35} select-none`}>子項目に金額を設定</span>
        ) : (
          <div className="flex items-center gap-1">
            <span className={`text-base ${C.text40}`}>¥</span>
            <input
              type="number"
              min={0}
              aria-label="単価(税込)"
              value={formData.price}
              onChange={(event) =>
                setFormDataDirty((prev) => ({ ...prev, price: Number(event.target.value) }))
              }
              placeholder="0"
              className={`${STYLE.propertyInput} w-28`}
            />
          </div>
        )}
      </PropertyRow>

      <PropertyRow label="課税区分" description={FIELD_DESCRIPTIONS.taxCategory}>
        <TaxTypeSelector
          value={formData.taxType}
          onChange={(value) => setFormDataDirty((prev) => ({ ...prev, taxType: value }))}
          disabled={isCategory}
        />
      </PropertyRow>

      <PropertyRow label="税率" description={FIELD_DESCRIPTIONS.taxRate}>
        <TaxRateSelector
          value={formData.taxRate}
          onChange={(value) => setFormDataDirty((prev) => ({ ...prev, taxRate: value }))}
          disabled={isCategory}
        />
      </PropertyRow>
    </>
  );
}

interface MedicineBasicFlagsSectionProps {
  formData: MedicineFormData;
  setFormDataDirty: SetMedicineFormDataDirty;
}

export function MedicineBasicFlagsSection({
  formData,
  setFormDataDirty,
}: MedicineBasicFlagsSectionProps) {
  return (
    <>
      <PropertyRow label="ステータス" description={FIELD_DESCRIPTIONS.status}>
        <button
          type="button"
          onClick={() => setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }))}
          className={`inline-flex min-h-11 items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <StatusPill isActive={formData.isActive} />
        </button>
      </PropertyRow>

      <PropertyRow
        label="保険対象外"
        description="この薬品が保険適用の対象外かどうかです。対象外にすると保険計算から除外されます。"
      >
        <button
          type="button"
          onClick={() =>
            setFormDataDirty((prev) => ({ ...prev, isNonInsurance: !prev.isNonInsurance }))
          }
          aria-label="保険対象外を切り替え"
          className={`inline-flex min-h-11 items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-1.5 cursor-pointer text-sm ${formData.isNonInsurance ? C.textBrand : C.text50}`}
        >
          {formData.isNonInsurance ? "対象外" : "対象"}
        </button>
      </PropertyRow>

      <PropertyRow label="備考" description={FIELD_DESCRIPTIONS.note}>
        <PropertyInput
          value={formData.description}
          onChange={(value) => setFormDataDirty((prev) => ({ ...prev, description: value }))}
          placeholder="空"
        />
      </PropertyRow>
    </>
  );
}

interface MedicineDetailSectionProps {
  formData: MedicineFormData;
  setFormDataDirty: SetMedicineFormDataDirty;
}

export function MedicineDetailSection({ formData, setFormDataDirty }: MedicineDetailSectionProps) {
  return (
    <>
      <div className={`${STYLE.sectionDivider} mt-3 mb-1`} />
      <div className="py-1">
        <div className="flex items-center gap-1.5 py-2 mb-1">
          <Pill className={`${ICON.xs} ${C.text40}`} />
          <span className={`${STYLE.sectionLabel}`}>薬剤詳細</span>
        </div>

        <PropertyRow label="剤形" description="薬の剤形（錠剤・液剤・注射剤・外用剤・散剤）です。">
          <Select
            value={formData.dosageForm}
            onValueChange={(value) => setFormDataDirty((prev) => ({ ...prev, dosageForm: value }))}
          >
            <SelectTrigger className={SELECT_TRIGGER_FULL}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{DOSAGE_FORM_SELECT_ITEMS}</SelectContent>
          </Select>
        </PropertyRow>

        <PropertyRow
          label="単位"
          description="投与量を数える基準単位（1錠あたり・1mlあたり等）です。製品含量や投与量計算の基準になります。"
        >
          <Select
            value={formData.medicineUnit}
            onValueChange={(value) =>
              setFormDataDirty((prev) => ({ ...prev, medicineUnit: value }))
            }
          >
            <SelectTrigger className={SELECT_TRIGGER_FULL}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{MEDICINE_UNIT_SELECT_ITEMS}</SelectContent>
          </Select>
        </PropertyRow>
      </div>
    </>
  );
}

interface MedicineDoseCalculationSectionProps {
  formData: MedicineFormData;
  setFormDataDirty: SetMedicineFormDataDirty;
}

/**
 * #201 投与量自動計算（製品軸）。calculation_type=none（既定・手動）/ per_weight（mg/kg 自動計算）。
 * per_weight 選択時のみ strength/frequencyPerDay/defaultDurationDays を表示する。
 * 種別（犬・猫）パラメータは別 API（MedicineDoseParamsEditor）で編集する — ここでは製品軸のみ。
 */
export function MedicineDoseCalculationSection({
  formData,
  setFormDataDirty,
}: MedicineDoseCalculationSectionProps) {
  const isPerWeight = formData.calculationType === MedicineCalculationTypePerWeight;

  return (
    <>
      <div className={`${STYLE.sectionDivider} mt-3 mb-1`} />
      <div className="py-1">
        <div className="flex items-center gap-1.5 py-2 mb-1">
          <Pill className={`${ICON.xs} ${C.text40}`} />
          <span className={`${STYLE.sectionLabel}`}>投与量自動計算</span>
        </div>

        <PropertyRow
          label="計算方式"
          description="投与量の計算方法です。「体重換算(mg/kg)」を選ぶと体重から投与量を自動計算し、以下のパラメータ欄が表示されます。"
        >
          <Select
            value={formData.calculationType}
            onValueChange={(value) =>
              setFormDataDirty((prev) => ({
                ...prev,
                calculationType: value as MedicineFormData["calculationType"],
              }))
            }
          >
            <SelectTrigger className={SELECT_TRIGGER_FULL}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={MedicineCalculationTypeNone}>手動</SelectItem>
              <SelectItem value={MedicineCalculationTypePerWeight}>体重換算(mg/kg)</SelectItem>
            </SelectContent>
          </Select>
        </PropertyRow>

        {isPerWeight ? (
          <>
            <PropertyRow
              label="製品含量(mg/単位)"
              description="この製品1単位（1錠・1ml等）に含まれる有効成分量（mg）です。体重換算の投与量を実際の単位数へ変換する際に使われます。"
            >
              <input
                type="number"
                min={0}
                step="any"
                aria-label="製品含量(mg/単位)"
                value={formData.strength}
                onChange={(event) =>
                  setFormDataDirty((prev) => ({ ...prev, strength: event.target.value }))
                }
                placeholder="未設定"
                className={`${STYLE.propertyInput} w-28`}
              />
            </PropertyRow>

            <PropertyRow
              label="1日投与回数(任意)"
              description="1日に投与する回数の既定値です。処方時のデフォルトとして使われます。"
            >
              <input
                type="number"
                min={1}
                step={1}
                aria-label="1日投与回数"
                value={formData.frequencyPerDay}
                onChange={(event) =>
                  setFormDataDirty((prev) => ({ ...prev, frequencyPerDay: event.target.value }))
                }
                placeholder="未設定"
                className={`${STYLE.propertyInput} w-28`}
              />
            </PropertyRow>

            <PropertyRow
              label="既定投与日数(任意)"
              description="標準的な投与日数の既定値です。処方時のデフォルトとして使われます。"
            >
              <input
                type="number"
                min={1}
                step={1}
                aria-label="既定投与日数"
                value={formData.defaultDurationDays}
                onChange={(event) =>
                  setFormDataDirty((prev) => ({ ...prev, defaultDurationDays: event.target.value }))
                }
                placeholder="未設定"
                className={`${STYLE.propertyInput} w-28`}
              />
            </PropertyRow>
          </>
        ) : null}
      </div>
    </>
  );
}
