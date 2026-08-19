import { memo, useCallback, useEffect, useState } from "react";
import { FlaskConical } from "lucide-react";

import { MasterSidePanel, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { C, LAYOUT } from "@/lib/design-tokens";

import type { LabDeviceItemMaster } from "../api/lab-device-item-masters";
import type { ExaminationTypeMaster } from "../api/exam-types-master";
import {
  examTypeSelectOptions,
  examTypeSelectValue,
  labDeviceFieldName,
  labDeviceSourceLabel,
  labDeviceToFormData,
  labDeviceValueShapeLabel,
  parseExamTypeSelectValue,
  type LabDeviceFormData,
  type LabDeviceRow,
} from "../routes/lab-device-item-master-settings-model";

interface LabDeviceItemMasterSidePanelProps {
  device: LabDeviceRow | null;
  items: LabDeviceItemMaster[];
  examTypes: ExaminationTypeMaster[];
  unusedSourceTypes: string[];
  readOnly?: boolean;
  isPending?: boolean;
  onClose: () => void;
  onSave: (data: LabDeviceFormData) => void;
  onDirtyChange?: (dirty: boolean) => void;
}

const EXAM_SELECT_CONTENT = "min-w-[28rem] w-max max-w-[min(40rem,90vw)]";

export const LabDeviceItemMasterSidePanel = memo(function LabDeviceItemMasterSidePanel({
  device,
  items,
  examTypes,
  unusedSourceTypes,
  readOnly,
  isPending,
  onClose,
  onSave,
  onDirtyChange,
}: LabDeviceItemMasterSidePanelProps) {
  const isNew = device === null;
  const [formData, setFormData] = useState<LabDeviceFormData>(() =>
    labDeviceToFormData(device, unusedSourceTypes),
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

  const handleSourceTypeChange = useCallback((value: string) => {
    setFormDataDirty((prev) => {
      const previousDefault = prev.sourceType === "" ? "" : labDeviceSourceLabel(prev.sourceType);
      const nextDefault = labDeviceSourceLabel(value);
      const name = prev.name.trim() === "" || prev.name === previousDefault ? nextDefault : prev.name;
      return { ...prev, sourceType: value, name };
    });
  }, [setFormDataDirty]);

  const handleExamTypeChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, examTypeId: parseExamTypeSelectValue(value) }));
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleAction = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("機器名を入力してください");
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
      isNew={isNew}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      onSave={readOnly ? undefined : handleAction}
      icon={<FlaskConical className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      isPending={isPending}
      titleError={nameError}
      titlePlaceholder="機器名"
      titleMaxLength={100}
      readOnly={readOnly}
      className="min-w-160"
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      {isNew ? (
        <PropertyRow label="プロトコル">
          <SearchableSelect
            value={formData.sourceType}
            onValueChange={handleSourceTypeChange}
            options={unusedSourceTypes.map((sourceType) => ({
              value: sourceType,
              label: labDeviceSourceLabel(sourceType),
            }))}
            placeholder="プロトコルを選ぶ"
            searchPlaceholder="プロトコルを検索"
            ariaLabel="プロトコル"
            disabled={readOnly}
            className="min-w-0 w-full"
            contentClassName={EXAM_SELECT_CONTENT}
          />
        </PropertyRow>
      ) : (
        <PropertyRow label="プロトコル">
          <span className={`text-sm whitespace-nowrap ${C.text}`}>
            {labDeviceSourceLabel(formData.sourceType)}
          </span>
        </PropertyRow>
      )}
      <PropertyRow label="検査">
        <SearchableSelect
          value={examTypeSelectValue(formData.examTypeId)}
          onValueChange={handleExamTypeChange}
          options={examTypeSelectOptions(examTypes)}
          placeholder="検査を選ぶ"
          searchPlaceholder="検査名で検索"
          ariaLabel="検査"
          disabled={readOnly}
          className="min-w-0 w-full"
          contentClassName={EXAM_SELECT_CONTENT}
        />
      </PropertyRow>
      <div className="py-1">
        {items.length === 0 ? (
          <p className={`text-sm ${C.text40}`}>
            まだ項目がありません。一覧の「既定項目を用意」で投入します
          </p>
        ) : (
          items.map((item) => {
            const fieldName = labDeviceFieldName(item.examTypeFieldId, examTypes);
            return (
              <section
                key={item.id}
                className={`py-3 border-b ${C.borderLight} last:border-b-0`}
              >
                <div className="flex items-baseline justify-between gap-3 px-2">
                  <span className={`text-sm font-medium whitespace-nowrap ${C.text}`}>
                    {item.deviceItemCode}
                  </span>
                  {fieldName === "" ? null : (
                    <span className={`text-sm whitespace-nowrap ${C.text}`}>{fieldName}</span>
                  )}
                  <span className={`text-sm whitespace-nowrap ${C.text55}`}>
                    {labDeviceValueShapeLabel(item.valueShape)}
                    {item.unit ? ` · ${item.unit}` : ""}
                  </span>
                </div>
              </section>
            );
          })
        )}
      </div>
    </MasterSidePanel>
  );
});
