import { memo, useCallback, useEffect, useState } from "react";
import type { ChangeEvent } from "react";
import { Shield } from "lucide-react";

import {
  MasterSidePanel,
  PropertyInput,
  PropertyRow,
  StatusToggleButton,
} from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";

import type { Insurance } from "../api/insurances";
import { MASTER_INPUT_CLASS } from "../constants/styles";
import {
  insuranceToFormData,
  type InsuranceFormData,
} from "./insurance-side-panel-model";

interface InsuranceSidePanelProps {
  item: Insurance | null;
  onClose: () => void;
  onSave: (data: InsuranceFormData) => void;
  onDeleteRequest?: (item: Insurance) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const InsuranceSidePanel = memo(function InsuranceSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: InsuranceSidePanelProps) {
  const [formData, setFormData] = useState<InsuranceFormData>(() => insuranceToFormData(item));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleTitleChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: value }));
    if (value.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleCoverageRateChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, coverageRate: event.target.value }));
  }, [setFormDataDirty]);

  const handleContactPhoneChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, contactPhone: event.target.value }));
  }, [setFormDataDirty]);

  const handleDescriptionChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, description: value }));
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleAction = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(formData);
    setIsDirty(false);
  }, [formData, onSave]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel
      isNew={item === null}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Shield className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="補償率(%)">
        <input
          type="number"
          min={0}
          max={100}
          aria-label="補償率(%)"
          className={MASTER_INPUT_CLASS}
          value={formData.coverageRate}
          onChange={handleCoverageRateChange}
          placeholder="0"
        />
      </PropertyRow>
      <PropertyRow label="連絡先">
        <input
          type="tel"
          aria-label="連絡先"
          className={MASTER_INPUT_CLASS}
          value={formData.contactPhone}
          onChange={handleContactPhoneChange}
          placeholder="電話番号"
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
