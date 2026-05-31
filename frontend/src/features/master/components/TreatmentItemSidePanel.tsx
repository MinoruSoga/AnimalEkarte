import { memo, useCallback, useEffect, useState } from "react";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import { MasterSidePanel, MoneyInput, PropertyInput, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { TaxRateSelector } from "@/components/shared/TaxRateSelector/TaxRateSelector";
import { TaxTypeSelector } from "@/components/shared/TaxTypeSelector/TaxTypeSelector";
import { C, LAYOUT } from "@/lib/design-tokens";
import type { TreatmentItem } from "@/lib/transforms/treatment";
import type { TaxType } from "@/types/generated/models";

export type TreatmentFormData = {
  name: string;
  price: number;
  description: string;
  isActive: boolean;
  taxType: TaxType;
  taxRate: number;
  isNonInsurance: boolean;
};

interface TreatmentItemSidePanelProps {
  item: TreatmentItem | null;
  onClose: () => void;
  onSave: (data: TreatmentFormData) => void;
  onDeleteRequest?: () => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const TreatmentItemSidePanel = memo(function TreatmentItemSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: TreatmentItemSidePanelProps) {
  const [formData, setFormData] = useState<TreatmentFormData>(() => ({
    name: item?.name ?? "",
    price: item?.price ?? 0,
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
    taxType: (item?.taxType ?? "excluded") as TaxType,
    taxRate: item?.taxRate ?? 0.1,
    isNonInsurance: item?.isNonInsurance ?? false,
  }));
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
      isNew={item === null}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={onClose}
      action={readOnly ? undefined : handleAction}
      onDelete={!readOnly && item !== null && onDeleteRequest ? onDeleteRequest : undefined}
      icon={<Stethoscope className={LAYOUT.pageIcon.innerIcon} />}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton
        isActive={formData.isActive}
        onToggle={() => setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }))}
      />
      <MoneyInput
        value={formData.price}
        onChange={(value) => setFormDataDirty((prev) => ({ ...prev, price: value }))}
      />
      <PropertyRow label="課税区分">
        <TaxTypeSelector
          value={formData.taxType}
          onChange={(value) => setFormDataDirty((prev) => ({ ...prev, taxType: value }))}
        />
      </PropertyRow>
      <PropertyRow label="税率">
        <TaxRateSelector
          value={formData.taxRate}
          onChange={(value) => setFormDataDirty((prev) => ({ ...prev, taxRate: value }))}
        />
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
          placeholder="補足情報など"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});
