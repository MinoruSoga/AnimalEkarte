import { memo, useCallback, useEffect, useRef, useState } from "react";
import FolderTree from "lucide-react/dist/esm/icons/folder-tree";

import { MasterSidePanel, PropertyInput, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";

import type { DiagnosisType } from "../api/diagnosis";
import {
  diagnosisTypeToFormData,
  type DiagnosisTypeFormData,
} from "./diagnosis-side-panel-model";

interface DiagnosisTypeSidePanelProps {
  item: DiagnosisType | null;
  onClose: () => void;
  onSave: (data: DiagnosisTypeFormData) => void;
  onDeleteRequest?: (item: DiagnosisType) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const DiagnosisTypeSidePanel = memo(function DiagnosisTypeSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: DiagnosisTypeSidePanelProps) {
  const [formData, setFormData] = useState<DiagnosisTypeFormData>(() => diagnosisTypeToFormData(item));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const formDataRef = useRef(formData);
  useEffect(() => {
    formDataRef.current = formData;
  }, [formData]);

  const handleTitleChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: value }));
    if (value.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleDescriptionChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, description: value }));
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleAction = useCallback(() => {
    const current = formDataRef.current;
    if (!current.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(current);
    setIsDirty(false);
  }, [onSave]);

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
      icon={<FolderTree className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton
        isActive={formData.isActive}
        onToggle={handleToggleActive}
      />
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
