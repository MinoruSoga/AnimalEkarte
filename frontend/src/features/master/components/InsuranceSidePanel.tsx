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
import { handleApiError } from "@/lib/handle-api-error";

import type { Insurance } from "../api/insurances";
import { MASTER_INPUT_CLASS } from "../constants/styles";
import { validateInsuranceForm } from "../routes/insurance-settings-model";
import {
  insuranceToFormData,
  type InsuranceFormData,
} from "./insurance-side-panel-model";

interface InsuranceSidePanelProps {
  item: Insurance | null;
  onClose: () => void;
  /** Returns true only when mutation succeeded (BUG-026: do not clear dirty on fail). */
  onSave: (data: InsuranceFormData) => Promise<boolean> | boolean;
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
  const [coverageError, setCoverageError] = useState("");

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
    setCoverageError("");
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

  const handleAction = useCallback(async () => {
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    // Surface field-level error; useMasterSave re-validates and toasts (no success path).
    const validationError = validateInsuranceForm(formData);
    setCoverageError(
      validationError && validationError.includes("補償率") ? validationError : "",
    );
    try {
      const saved = await onSave(formData);
      if (saved) {
        setIsDirty(false);
        onDirtyChange?.(false);
      }
    } catch (error) {
      handleApiError(error, "保存");
    }
  }, [formData, onDirtyChange, onSave]);

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
      action={readOnly ? undefined : handleAction}
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
          step={1}
          aria-label="補償率(%)"
          aria-invalid={coverageError ? true : undefined}
          className={MASTER_INPUT_CLASS}
          value={formData.coverageRate}
          onChange={handleCoverageRateChange}
          placeholder="0"
        />
        {coverageError ? (
          <p className="mt-1 text-sm text-red-600" role="alert">
            {coverageError}
          </p>
        ) : null}
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
