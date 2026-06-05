import { memo, useCallback, useEffect, useState } from "react";
import { Scissors } from "lucide-react";

import { MasterSidePanel, StatusToggleButton } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";

import type { TrimmingCourseType } from "../api/trimming-course-type";
import {
  trimmingCourseTypeToFormData,
  type TrimmingCourseTypeFormData,
} from "./TrimmingCourseTypeSidePanelModel";

interface TrimmingCourseTypeSidePanelProps {
  item: TrimmingCourseType | null;
  onClose: () => void;
  onSave: (data: TrimmingCourseTypeFormData) => void;
  onDeleteRequest?: (item: TrimmingCourseType) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const TrimmingCourseTypeSidePanel = memo(function TrimmingCourseTypeSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: TrimmingCourseTypeSidePanelProps) {
  const [formData, setFormData] = useState<TrimmingCourseTypeFormData>(() =>
    trimmingCourseTypeToFormData(item),
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
      icon={<Scissors className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={50}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
    </MasterSidePanel>
  );
});
