import { memo, useCallback, useState } from "react";
import FolderTree from "lucide-react/dist/esm/icons/folder-tree";

import { MasterSidePanel, PropertyInput, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";

import type { DiagnosisType } from "../api/diagnosis";
import { useMasterSidePanelForm } from "../hooks/use-master-side-panel-form";
import {
  diagnosisTypeToFormData,
  type DiagnosisTypeFormData,
} from "./diagnosis-side-panel-model";

interface DiagnosisTypeSidePanelProps {
  item: DiagnosisType | null;
  onClose: () => void;
  onSave: (data: DiagnosisTypeFormData) => Promise<boolean> | boolean;
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
  const [nameError, setNameError] = useState("");

  const { formData, setFormData: setFormDataDirty, isDirty, setIsDirty, handleAction } =
    useMasterSidePanelForm<DiagnosisTypeFormData>({
      initialFormData: diagnosisTypeToFormData(item),
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
