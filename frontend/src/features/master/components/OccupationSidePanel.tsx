import { memo, useCallback, useState } from "react";
import { Briefcase } from "lucide-react";

import { MasterSidePanel, PropertyInput, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";

import type { Occupation } from "../api/occupations";
import { useMasterSidePanelForm } from "../hooks/use-master-side-panel-form";
import {
  occupationToFormData,
  type OccupationFormData,
} from "./occupation-side-panel-model";

interface OccupationSidePanelProps {
  item: Occupation | null;
  onClose: () => void;
  onSave: (data: OccupationFormData) => Promise<boolean> | boolean;
  onDeleteRequest?: (item: Occupation) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const OccupationSidePanel = memo(function OccupationSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: OccupationSidePanelProps) {
  const [nameError, setNameError] = useState("");

  const { formData, setFormData: setFormDataDirty, isDirty, setIsDirty, handleAction } =
    useMasterSidePanelForm<OccupationFormData>({
      initialFormData: occupationToFormData(item),
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
      icon={<Briefcase className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="説明">
        <PropertyInput
          value={formData.description}
          onChange={handleDescriptionChange}
          placeholder="説明を入力"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});
