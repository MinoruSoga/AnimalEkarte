import { memo, useCallback, useEffect, useState } from "react";
import { FlaskConical } from "lucide-react";

import { MasterSidePanel, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { C, LAYOUT } from "@/lib/design-tokens";

import type { LabDeviceItemMaster } from "../api/lab-device-item-masters";
import type { ExaminationTypeMaster } from "../api/exam-types-master";
import {
  LAB_DEVICE_UNMAPPED_FIELD,
  countDraftsUnmappedByExamChange,
  examTypeSelectOptions,
  examTypeSelectValue,
  itemToLabDeviceDraft,
  labDeviceFieldUnit,
  labDeviceItemSelectOptions,
  labDeviceSourceLabel,
  labDeviceToFormData,
  labDeviceUnitMismatch,
  labDeviceValueShapeLabel,
  parseExamFieldSelectValue,
  parseExamTypeSelectValue,
  restrictDraftsToExamType,
  type LabDeviceFormData,
  type LabDeviceItemDraft,
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
  onSave: (data: LabDeviceFormData, drafts: LabDeviceItemDraft[]) => void;
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
  const [drafts, setDrafts] = useState<LabDeviceItemDraft[]>(() => items.map(itemToLabDeviceDraft));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");
  const [unmapNotice, setUnmapNotice] = useState("");

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  useEffect(() => {
    setDrafts((prev) => (isDirty ? prev : items.map(itemToLabDeviceDraft)));
    // 編集中（dirty）の background refetch で入力を潰さない
  }, [isDirty, items]);

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
    const nextExamTypeId = parseExamTypeSelectValue(value);
    const unmapped = countDraftsUnmappedByExamChange(drafts, nextExamTypeId, examTypes);
    setUnmapNotice(
      unmapped > 0 ? `検査の変更で ${unmapped} 件の項目対応が外れます。保存で確定します` : "",
    );
    if (nextExamTypeId !== null) {
      setDrafts((prev) => restrictDraftsToExamType(prev, nextExamTypeId, examTypes));
    }
    setFormDataDirty((prev) => ({ ...prev, examTypeId: nextExamTypeId }));
  }, [drafts, examTypes, setFormDataDirty]);

  const handleItemFieldChange = useCallback((draftId: string, value: string) => {
    const fieldId = parseExamFieldSelectValue(value);
    setDrafts((prev) => prev.map((draft) => (
      draft.id === draftId ? { ...draft, examTypeFieldId: fieldId } : draft
    )));
    setUnmapNotice("");
    setIsDirty(true);
  }, []);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleAction = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("機器名を入力してください");
      return;
    }
    setNameError("");
    // dirty 解除は保存成功側（ページの useSidePeekDirty / パネル unmount）に任せる。
    // ここで消すと保存失敗時に NavigationBlocker が外れ、編集が黙って失われる。
    onSave(formData, drafts);
  }, [drafts, formData, onSave]);

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
      {unmapNotice === "" ? null : (
        <p className={`text-sm ${C.textWarning}`}>{unmapNotice}</p>
      )}
      <div className="py-1">
        {items.length === 0 ? (
          <p className={`text-sm ${C.text40}`}>
            まだ項目がありません。一覧の「既定項目を用意」で投入します
          </p>
        ) : (
          items.map((item) => {
            const draft = drafts.find((candidate) => candidate.id === item.id);
            const fieldId = draft?.examTypeFieldId ?? null;
            const fieldUnit = labDeviceFieldUnit(fieldId, examTypes);
            const unitMismatch = labDeviceUnitMismatch(item.unit, fieldUnit);
            return (
              <section
                key={item.id}
                className={`py-3 border-b ${C.borderLight} last:border-b-0`}
              >
                <div className="flex items-center justify-between gap-3 px-2">
                  <span className={`text-sm font-medium whitespace-nowrap ${C.text}`}>
                    {item.deviceItemCode}
                  </span>
                  <SearchableSelect
                    value={fieldId ?? LAB_DEVICE_UNMAPPED_FIELD}
                    onValueChange={(value) => handleItemFieldChange(item.id, value)}
                    options={labDeviceItemSelectOptions(examTypes, formData.examTypeId, fieldId)}
                    placeholder="検査項目を選ぶ"
                    searchPlaceholder="項目名で検索"
                    ariaLabel={`${item.deviceItemCode} の検査項目`}
                    disabled={readOnly}
                    className="min-w-0 flex-1"
                    contentClassName={EXAM_SELECT_CONTENT}
                  />
                  <span className={`text-sm whitespace-nowrap ${C.text55}`}>
                    {labDeviceValueShapeLabel(item.valueShape)}
                    {item.unit ? ` · ${item.unit}` : ""}
                  </span>
                </div>
                {unitMismatch ? (
                  <p className={`px-2 pt-1 text-sm ${C.textWarning}`}>
                    単位不一致: 電文 {item.unit} / 検査項目 {fieldUnit}
                  </p>
                ) : null}
              </section>
            );
          })
        )}
      </div>
    </MasterSidePanel>
  );
});
