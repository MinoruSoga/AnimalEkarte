import { memo, useCallback, useEffect, useState } from "react";
import Pill from "lucide-react/dist/esm/icons/pill";

import { MasterSidePanel, PropertyInput, PropertyRow } from "@/components/shared/SidePeek";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { TaxRateSelector } from "@/components/shared/TaxRateSelector/TaxRateSelector";
import { TaxTypeSelector } from "@/components/shared/TaxTypeSelector/TaxTypeSelector";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { C, ICON, LAYOUT, STYLE } from "@/lib/design-tokens";
import type { Medicine } from "@/types";

import { medicineToFormData, type MedicineFormData } from "./MedicineSidePanelModel";

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

const SELECT_TRIGGER_FULL = `h-[30px] text-base bg-transparent ${C.text} border-0 ${C.hoverBgLight} px-1.5 shadow-none rounded-[3px] w-full`;

interface MedicineSidePanelBodyProps {
  selectedMedicine: Medicine | null;
  isCategory: boolean;
  defaultParentId?: string;
  categoryMedicines: Medicine[];
  onCloseEdit: () => void;
  onSave: (data: MedicineFormData) => void;
  onDeleteRequest: () => void;
  readOnly?: boolean;
  canDelete?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const MedicineSidePanelBody = memo(function MedicineSidePanelBody({
  selectedMedicine,
  isCategory,
  defaultParentId,
  categoryMedicines,
  onCloseEdit,
  onSave,
  onDeleteRequest,
  readOnly,
  canDelete,
  onDirtyChange,
}: MedicineSidePanelBodyProps) {
  const [formData, setFormData] = useState<MedicineFormData>(() =>
    medicineToFormData(selectedMedicine, defaultParentId),
  );
  const [nameError, setNameError] = useState("");
  const [isDirty, setIsDirty] = useState(false);

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(formData);
    setIsDirty(false);
  }, [formData, onSave]);

  const handleTitleChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: value }));
    if (value.trim()) setNameError("");
  }, [setFormDataDirty]);

  return (
    <MasterSidePanel
      isNew={!selectedMedicine}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={onCloseEdit}
      action={handleAction}
      onDelete={selectedMedicine && canDelete ? onDeleteRequest : undefined}
      icon={<Pill className={LAYOUT.pageIcon.innerIcon} />}
      titlePlaceholder="薬品名"
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <PropertyRow label="親カテゴリ">
        {isCategory ? (
          <span className={`text-base ${C.text}`}>なし（ルート）</span>
        ) : (
          <Select
            value={formData.parentId || "__none__"}
            onValueChange={(value) => setFormDataDirty((prev) => ({
              ...prev,
              parentId: value === "__none__" ? "" : value,
            }))}
          >
            <SelectTrigger className={SELECT_TRIGGER_FULL}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="__none__">なし（未分類）</SelectItem>
              {categoryMedicines.map((category) => (
                <SelectItem key={category.id} value={category.id}>
                  {category.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </PropertyRow>

      <PropertyRow label="単価(税込)">
        {isCategory ? (
          <span className={`text-base ${C.text35} select-none`}>子項目に金額を設定</span>
        ) : (
          <div className="flex items-center gap-1">
            <span className={`text-base ${C.text40}`}>¥</span>
            <input
              type="number"
              min={0}
              value={formData.price}
              onChange={(event) => setFormDataDirty((prev) => ({ ...prev, price: Number(event.target.value) }))}
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

      <PropertyRow label="ステータス">
        <button
          type="button"
          onClick={() => setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }))}
          className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <NotionStatusPill isActive={formData.isActive} />
        </button>
      </PropertyRow>

      <PropertyRow label="保険対象外">
        <button
          type="button"
          onClick={() => setFormDataDirty((prev) => ({ ...prev, isNonInsurance: !prev.isNonInsurance }))}
          aria-label="保険対象外を切り替え"
          className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-1.5 cursor-pointer text-sm ${formData.isNonInsurance ? C.accent : C.text50}`}
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

      <div className={`${STYLE.sectionDivider} mt-3 mb-1`} />
      <div className="py-1">
        <div className="flex items-center gap-1.5 py-2 mb-1">
          <Pill className={`${ICON.xs} ${C.text40}`} />
          <span className={`text-base font-medium ${C.text50} uppercase tracking-wide select-none`}>
            薬剤詳細
          </span>
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
            onValueChange={(value) => setFormDataDirty((prev) => ({ ...prev, medicineUnit: value }))}
          >
            <SelectTrigger className={SELECT_TRIGGER_FULL}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{MEDICINE_UNIT_SELECT_ITEMS}</SelectContent>
          </Select>
        </PropertyRow>
      </div>
    </MasterSidePanel>
  );
});
