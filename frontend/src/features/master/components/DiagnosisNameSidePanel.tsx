import { memo, useCallback, useMemo, useState } from "react";
import ClipboardList from "lucide-react/dist/esm/icons/clipboard-list";

import { FormFieldError } from "@/components/shared/FormFieldError";
import { MasterSidePanel, PropertyInput, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { LAYOUT, STYLE } from "@/lib/design-tokens";

import type { DiagnosisName, DiagnosisType } from "../api/diagnosis";
import { useMasterSidePanelForm } from "../hooks/use-master-side-panel-form";
import {
  diagnosisNameToFormData,
  type DiagnosisNameFormData,
} from "./diagnosis-side-panel-model";

interface DiagnosisNameSidePanelProps {
  item: DiagnosisName | null;
  categories: DiagnosisType[];
  onClose: () => void;
  onSave: (data: DiagnosisNameFormData) => Promise<boolean> | boolean;
  onDeleteRequest?: (item: DiagnosisName) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const DiagnosisNameSidePanel = memo(function DiagnosisNameSidePanel({
  item,
  categories,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: DiagnosisNameSidePanelProps) {
  const [nameError, setNameError] = useState("");
  const [categoryError, setCategoryError] = useState("");

  const { formData, setFormData: setFormDataDirty, isDirty, setIsDirty, handleAction } =
    useMasterSidePanelForm<DiagnosisNameFormData>({
      initialFormData: diagnosisNameToFormData(item, categories),
      onSave,
      onDirtyChange,
      validate: (data) => {
        if (!data.name.trim()) {
          setNameError("診断病名を入力してください");
          return false;
        }
        if (!data.diagnosisTypeId) {
          setCategoryError("カテゴリを選択してください");
          return false;
        }
        setNameError("");
        setCategoryError("");
        return true;
      },
    });

  const handleTitleChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: value }));
    if (value.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleCategoryChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, diagnosisTypeId: value }));
    if (value) setCategoryError("");
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

  const categorySelectItems = useMemo(
    () => categories.map((category) => (
      <SelectItem key={category.id} value={String(category.id)}>{category.name}</SelectItem>
    )),
    [categories],
  );

  return (
    <MasterSidePanel
      isNew={item === null}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<ClipboardList className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton
        isActive={formData.isActive}
        onToggle={handleToggleActive}
      />
      <PropertyRow label="カテゴリ">
        <Select
          value={formData.diagnosisTypeId}
          onValueChange={handleCategoryChange}
        >
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="カテゴリを選択" />
          </SelectTrigger>
          <SelectContent>
            {categorySelectItems}
          </SelectContent>
        </Select>
      </PropertyRow>
      <FormFieldError message={categoryError} />
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
