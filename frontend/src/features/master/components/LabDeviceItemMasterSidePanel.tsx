import { memo, useCallback, useEffect, useMemo, useState } from "react";
import { FlaskConical } from "lucide-react";

import { NavigationBlocker } from "@/components/shared/NavigationBlocker/NavigationBlocker";
import { PropertyInput, PropertyRow } from "@/components/shared/SidePeek";
import { SidePeekBody } from "@/components/shared/SidePeek/SidePeekBody";
import { SidePeekFooter } from "@/components/shared/SidePeek/SidePeekFooter";
import { SidePeekPanel } from "@/components/shared/SidePeek/SidePeekPanel";
import { SidePeekToolbar } from "@/components/shared/SidePeek/SidePeekToolbar";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { C, LAYOUT, STYLE } from "@/lib/design-tokens";

import type { LabDeviceItemMaster } from "../api/lab-device-item-masters";
import type { ExaminationTypeMaster } from "../api/exam-types-master";
import {
  LAB_DEVICE_UNMAPPED_FIELD,
  buildExamFieldOptions,
  examFieldOptionsForItem,
  examFieldSelectValue,
  itemToLabDeviceDraft,
  labDeviceValueShapeLabel,
  parseExamFieldSelectValue,
  type LabDeviceItemDraft,
  type LabDeviceRow,
} from "../routes/lab-device-item-master-settings-model";

interface LabDeviceItemMasterSidePanelProps {
  device: LabDeviceRow;
  items: LabDeviceItemMaster[];
  examTypes: ExaminationTypeMaster[];
  onClose: () => void;
  onSave: (drafts: LabDeviceItemDraft[]) => Promise<boolean> | boolean | void;
  readOnly?: boolean;
  isPending?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const LabDeviceItemMasterSidePanel = memo(function LabDeviceItemMasterSidePanel({
  device,
  items,
  examTypes,
  onClose,
  onSave,
  readOnly,
  isPending,
  onDirtyChange,
}: LabDeviceItemMasterSidePanelProps) {
  const [drafts, setDrafts] = useState<LabDeviceItemDraft[]>(() => items.map(itemToLabDeviceDraft));
  const [isDirty, setIsDirty] = useState(false);
  const fieldOptions = useMemo(() => buildExamFieldOptions(examTypes), [examTypes]);

  useEffect(() => {
    setDrafts(items.map(itemToLabDeviceDraft));
    setIsDirty(false);
  }, [items]);

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const patchDraft = useCallback((id: string, patch: Partial<LabDeviceItemDraft>) => {
    setDrafts((prev) => prev.map((draft) => (draft.id === id ? { ...draft, ...patch } : draft)));
    setIsDirty(true);
  }, []);

  const handleSave = useCallback(async () => {
    const result = await onSave(drafts);
    if (result !== false) {
      setIsDirty(false);
    }
  }, [drafts, onSave]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <SidePeekPanel>
      <NavigationBlocker
        when={isDirty}
        title="変更が保存されていません"
        description="変更が保存されていません。ページを離れますか？"
      />
      <SidePeekToolbar isNew={false} onClose={handleClose} readOnly={readOnly} />
      <SidePeekBody>
        <div className="pt-4 pb-2">
          <div className={STYLE.pageIcon}>
            <FlaskConical className={LAYOUT.pageIcon.innerIcon} />
          </div>
        </div>
        <h2
          className={`pb-1 mb-4 ${C.text}`}
          style={{
            fontSize: LAYOUT.pageTitle.fontSize,
            fontWeight: LAYOUT.pageTitle.fontWeight,
            lineHeight: LAYOUT.pageTitle.lineHeight,
            letterSpacing: LAYOUT.pageTitle.letterSpacing,
          }}
        >
          {device.name}
        </h2>
        <div className={`${STYLE.sectionDivider} mb-1`} />
        <div className="py-1">
          {drafts.length === 0 ? (
            <p className={`text-sm ${C.text40}`}>
              まだ項目がありません。一覧の「既定項目を用意」で投入します
            </p>
          ) : (
            drafts.map((draft) => {
              const item = items.find((candidate) => candidate.id === draft.id);
              if (item === undefined) {
                return null;
              }
              return (
                <LabDeviceItemDraftFields
                  key={draft.id}
                  item={item}
                  draft={draft}
                  fieldOptions={fieldOptions}
                  readOnly={readOnly === true}
                  onPatch={patchDraft}
                />
              );
            })
          )}
        </div>
      </SidePeekBody>
      <SidePeekFooter
        onCancel={handleClose}
        onSave={readOnly ? undefined : handleSave}
        isPending={isPending}
        readOnly={readOnly}
      />
    </SidePeekPanel>
  );
});

function LabDeviceItemDraftFields({
  item,
  draft,
  fieldOptions,
  readOnly,
  onPatch,
}: {
  item: LabDeviceItemMaster;
  draft: LabDeviceItemDraft;
  fieldOptions: ReturnType<typeof buildExamFieldOptions>;
  readOnly: boolean;
  onPatch: (id: string, patch: Partial<LabDeviceItemDraft>) => void;
}) {
  return (
    <section className={`py-3 border-b ${C.borderLight} last:border-b-0`}>
      <div className="flex items-baseline justify-between gap-3 px-2 mb-1">
        <span className={`text-sm font-medium ${C.text}`}>{item.deviceItemCode}</span>
        <span className={`text-sm ${C.text55}`}>
          {labDeviceValueShapeLabel(item.valueShape)}
          {item.unit ? ` · ${item.unit}` : ""}
        </span>
      </div>
      <PropertyRow label="表示名">
        {readOnly ? (
          <span className={`text-sm ${C.text}`}>{draft.displayName}</span>
        ) : (
          <PropertyInput
            value={draft.displayName}
            onChange={(value) => onPatch(draft.id, { displayName: value })}
            ariaLabel={`${item.deviceItemCode}の表示名`}
          />
        )}
      </PropertyRow>
      <PropertyRow label="載せる先">
        {readOnly ? (
          <span className={`text-sm ${C.text}`}>
            {draft.examTypeFieldId === null
              ? "未設定"
              : (examFieldOptionsForItem(fieldOptions, draft.examTypeFieldId).find(
                  (option) => option.id === draft.examTypeFieldId,
                )?.label ?? draft.examTypeFieldId)}
          </span>
        ) : (
          <Select
            value={examFieldSelectValue(draft.examTypeFieldId)}
            onValueChange={(value) => onPatch(draft.id, { examTypeFieldId: parseExamFieldSelectValue(value) })}
          >
            <SelectTrigger className={STYLE.selectCompact} aria-label={`${item.deviceItemCode}の載せる先`}>
              <SelectValue placeholder="未設定" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={LAB_DEVICE_UNMAPPED_FIELD}>未設定</SelectItem>
              {examFieldOptionsForItem(fieldOptions, draft.examTypeFieldId).map((option) => (
                <SelectItem key={option.id} value={option.id}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </PropertyRow>
      <PropertyRow label="ステータス">
        {readOnly ? (
          <StatusPill isActive={draft.isActive} />
        ) : (
          <button
            type="button"
            onClick={() => onPatch(draft.id, { isActive: !draft.isActive })}
            aria-label={`${item.deviceItemCode}の有効を切り替え`}
            className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
          >
            <StatusPill isActive={draft.isActive} />
          </button>
        )}
      </PropertyRow>
    </section>
  );
}
