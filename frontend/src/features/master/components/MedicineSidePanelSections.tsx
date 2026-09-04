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
    <PropertyRow label="親カテゴリ">
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
      <PropertyRow label="単価(税込)">
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

      <PropertyRow label="課税区分">
        <TaxTypeSelector
          value={formData.taxType}
          onChange={(value) => setFormDataDirty((prev) => ({ ...prev, taxType: value }))}
          disabled={isCategory}
        />
      </PropertyRow>

      <PropertyRow label="税率">
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
      <PropertyRow label="ステータス">
        <button
          type="button"
          onClick={() => setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }))}
          className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <StatusPill isActive={formData.isActive} />
        </button>
      </PropertyRow>

      <PropertyRow label="保険対象外">
        <button
          type="button"
          onClick={() =>
            setFormDataDirty((prev) => ({ ...prev, isNonInsurance: !prev.isNonInsurance }))
          }
          aria-label="保険対象外を切り替え"
          className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-1.5 cursor-pointer text-sm ${formData.isNonInsurance ? C.textBrand : C.text50}`}
        >
          {formData.isNonInsurance ? "対象外" : "対象"}
        </button>
      </PropertyRow>

      <PropertyRow label="備考">
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

        <PropertyRow label="剤形">
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

        <PropertyRow label="単位">
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

        <PropertyRow label="計算方式">
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
            <PropertyRow label="製品含量(mg/単位)">
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

            <PropertyRow label="1日投与回数(任意)">
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

            <PropertyRow label="既定投与日数(任意)">
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
