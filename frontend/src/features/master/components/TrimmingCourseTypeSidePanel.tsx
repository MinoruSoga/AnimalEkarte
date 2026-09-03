import { memo, useCallback, useState } from "react";
import { Scissors } from "lucide-react";

import { MasterSidePanel, StatusToggleButton } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";

import type { TrimmingCourseType } from "../api/trimming-course-type";
import { useMasterSidePanelForm } from "../hooks/use-master-side-panel-form";
import {
  trimmingCourseTypeToFormData,
  type TrimmingCourseTypeFormData,
} from "./trimming-course-type-side-panel-model";

interface TrimmingCourseTypeSidePanelProps {
  item: TrimmingCourseType | null;
  onClose: () => void;
  onSave: (data: TrimmingCourseTypeFormData) => Promise<boolean> | boolean;
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
  const [nameError, setNameError] = useState("");

  const { formData, setFormData: setFormDataDirty, isDirty, setIsDirty, handleAction } =
    useMasterSidePanelForm<TrimmingCourseTypeFormData>({
      initialFormData: trimmingCourseTypeToFormData(item),
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
