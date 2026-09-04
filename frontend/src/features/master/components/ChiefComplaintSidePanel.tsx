import { memo, useCallback, useState } from "react";
import type { ChangeEvent } from "react";
import { MessageSquareText } from "lucide-react";

import { MasterSidePanel, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";

import type { ChiefComplaintType } from "../api/chief-complaint-types";
import { MASTER_INPUT_CLASS } from "../constants/styles";
import { useMasterSidePanelForm } from "../hooks/use-master-side-panel-form";
import {
  chiefComplaintToFormData,
  type ChiefComplaintFormData,
} from "../lib/chief-complaint-side-panel-model";

interface ChiefComplaintSidePanelProps {
  item: ChiefComplaintType | null;
  onClose: () => void;
  onSave: (data: ChiefComplaintFormData) => Promise<boolean> | boolean;
  onDeleteRequest?: (item: ChiefComplaintType) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const ChiefComplaintSidePanel = memo(function ChiefComplaintSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: ChiefComplaintSidePanelProps) {
  const [nameError, setNameError] = useState("");

  const {
    formData,
    setFormData: setFormDataDirty,
    isDirty,
    setIsDirty,
    handleAction,
  } = useMasterSidePanelForm<ChiefComplaintFormData>({
    initialFormData: chiefComplaintToFormData(item),
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

  const handleTitleChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, name: value }));
      if (value.trim()) setNameError("");
    },
    [setFormDataDirty],
  );

  const handleDescriptionChange = useCallback(
    (event: ChangeEvent<HTMLTextAreaElement>) => {
      setFormDataDirty((prev) => ({ ...prev, description: event.target.value }));
    },
    [setFormDataDirty],
  );

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
      icon={<MessageSquareText className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="説明">
        <textarea
          className={`${MASTER_INPUT_CLASS} min-h-[80px] resize-none`}
          value={formData.description}
          onChange={handleDescriptionChange}
          placeholder="説明を入力"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});
