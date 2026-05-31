import { memo, useCallback, useEffect, useState } from "react";
import type { ChangeEvent } from "react";
import { MessageSquareText } from "lucide-react";

import {
  MasterSidePanel,
  PropertyRow,
  StatusToggleButton,
} from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";

import type { ChiefComplaintType } from "../api/chief-complaint-types";
import { MASTER_INPUT_CLASS } from "../constants/styles";
import {
  chiefComplaintToFormData,
  type ChiefComplaintFormData,
} from "./ChiefComplaintSidePanelModel";

interface ChiefComplaintSidePanelProps {
  item: ChiefComplaintType | null;
  onClose: () => void;
  onSave: (data: ChiefComplaintFormData) => void;
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
  const [formData, setFormData] = useState<ChiefComplaintFormData>(() =>
    chiefComplaintToFormData(item),
  );
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

  const handleDescriptionChange = useCallback((event: ChangeEvent<HTMLTextAreaElement>) => {
    setFormDataDirty((prev) => ({ ...prev, description: event.target.value }));
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
