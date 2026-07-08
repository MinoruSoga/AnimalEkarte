import { memo, useCallback, useEffect, useState } from "react";
import { CreditCard } from "lucide-react";

import { MasterSidePanel, StatusToggleButton } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";

import type { PaymentMethod } from "../api/payment-method-master";
import {
  paymentMethodToFormData,
  type PaymentMethodFormData,
} from "./payment-method-side-panel-model";

interface PaymentMethodSidePanelProps {
  item: PaymentMethod | null;
  onClose: () => void;
  onSave: (data: PaymentMethodFormData) => void;
  onDeleteRequest?: (item: PaymentMethod) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const PaymentMethodSidePanel = memo(function PaymentMethodSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: PaymentMethodSidePanelProps) {
  const [formData, setFormData] = useState<PaymentMethodFormData>(() => paymentMethodToFormData(item));
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
      icon={<CreditCard className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
    </MasterSidePanel>
  );
});
