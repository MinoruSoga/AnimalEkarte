import { memo, useCallback, useEffect, useState } from "react";
import type { ChangeEvent } from "react";
import { FileText } from "lucide-react";

import {
  MasterSidePanel,
  PropertyRow,
  StatusToggleButton,
} from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";

import type { InquiryTemplate } from "../api/inquiry-templates";
import { MASTER_INPUT_CLASS } from "../constants/styles";
import {
  interviewTemplateToFormData,
  type InterviewTemplateFormData,
} from "./interview-template-side-panel-model";

interface InterviewTemplateSidePanelProps {
  item: InquiryTemplate | null;
  onClose: () => void;
  onSave: (data: InterviewTemplateFormData) => void;
  onDeleteRequest?: (item: InquiryTemplate) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const InterviewTemplateSidePanel = memo(function InterviewTemplateSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: InterviewTemplateSidePanelProps) {
  const [formData, setFormData] = useState<InterviewTemplateFormData>(() =>
    interviewTemplateToFormData(item),
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
    setFormDataDirty((prev) => ({ ...prev, title: value }));
    if (value.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleCategoryChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, category: event.target.value }));
  }, [setFormDataDirty]);

  const handleContentChange = useCallback((event: ChangeEvent<HTMLTextAreaElement>) => {
    setFormDataDirty((prev) => ({ ...prev, content: event.target.value }));
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleSave = useCallback(() => {
    if (!formData.title.trim()) {
      setNameError("タイトルを入力してください");
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
      title={formData.title}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={handleSave}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      isDirty={isDirty}
      icon={<FileText className={LAYOUT.pageIcon.innerIcon} />}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="カテゴリ">
        <input
          type="text"
          className={MASTER_INPUT_CLASS}
          value={formData.category}
          onChange={handleCategoryChange}
          placeholder="カテゴリを入力"
        />
      </PropertyRow>
      <PropertyRow label="テンプレート内容">
        <textarea
          className={`${MASTER_INPUT_CLASS} min-h-[150px] resize-none`}
          value={formData.content}
          onChange={handleContentChange}
          placeholder="テンプレート内容を入力"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});
