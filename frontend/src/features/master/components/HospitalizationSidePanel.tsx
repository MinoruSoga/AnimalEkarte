import { memo, useCallback, useState } from "react";
import { Bed } from "lucide-react";

import { MoneyInput, MasterSidePanel, PropertyInput, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { TaxRateSelector } from "@/components/shared/TaxRateSelector/TaxRateSelector";
import { TaxTypeSelector } from "@/components/shared/TaxTypeSelector/TaxTypeSelector";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { LAYOUT, STYLE } from "@/lib/design-tokens";
import type { BodySize, BillingUnit, TaxType } from "@/types/generated/models";

import {
  BILLING_UNIT_OPTIONS,
  BODY_SIZE_OPTIONS,
  type HospitalizationPlan,
} from "../api/hospitalization-plans";
import { useMasterSidePanelForm } from "../hooks/use-master-side-panel-form";
import {
  hospitalizationToFormData,
  type HospitalizationFormData,
} from "./hospitalization-side-panel-model";

interface HospitalizationSidePanelProps {
  item: HospitalizationPlan | null;
  onClose: () => void;
  onSave: (data: HospitalizationFormData) => Promise<boolean> | boolean;
  onDeleteRequest?: (item: HospitalizationPlan) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const HospitalizationSidePanel = memo(function HospitalizationSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: HospitalizationSidePanelProps) {
  const [nameError, setNameError] = useState("");

  const { formData, setFormData: setFormDataDirty, isDirty, setIsDirty, handleAction } =
    useMasterSidePanelForm<HospitalizationFormData>({
      initialFormData: hospitalizationToFormData(item),
      onSave,
      onDirtyChange,
      validate: (data) => {
        if (!data.name.trim()) {
          setNameError("名称を入力してください");
          return false;
        }
        setNameError("");
        return true;
      },
    });

  const handleTitleChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: value }));
    if (value.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleBodySizeChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, bodySize: value as BodySize }));
  }, [setFormDataDirty]);

  const handleBillingUnitChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, billingUnit: value as BillingUnit }));
  }, [setFormDataDirty]);

  const handlePriceChange = useCallback((value: number) => {
    setFormDataDirty((prev) => ({ ...prev, price: value }));
  }, [setFormDataDirty]);

  const handleTaxTypeChange = useCallback((value: TaxType) => {
    setFormDataDirty((prev) => ({ ...prev, taxType: value }));
  }, [setFormDataDirty]);

  const handleTaxRateChange = useCallback((value: number) => {
    setFormDataDirty((prev) => ({ ...prev, taxRate: value }));
  }, [setFormDataDirty]);

  const handleDescriptionChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, description: value }));
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose, setIsDirty]);

  return (
    <MasterSidePanel
      isNew={item === null}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Bed className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="対象体格">
        <Select value={formData.bodySize} onValueChange={handleBodySizeChange}>
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>
            {BODY_SIZE_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </PropertyRow>
      <PropertyRow label="料金単位">
        <Select value={formData.billingUnit} onValueChange={handleBillingUnitChange}>
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>
            {BILLING_UNIT_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </PropertyRow>
      <MoneyInput value={formData.price} onChange={handlePriceChange} />
      <PropertyRow label="課税区分">
        <TaxTypeSelector
          value={formData.taxType}
          onChange={handleTaxTypeChange}
        />
      </PropertyRow>
      <PropertyRow label="税率">
        <TaxRateSelector
          value={formData.taxRate}
          onChange={handleTaxRateChange}
        />
      </PropertyRow>
      <PropertyRow label="備考">
        <PropertyInput
          value={formData.description}
          onChange={handleDescriptionChange}
          placeholder="補足情報など"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});
