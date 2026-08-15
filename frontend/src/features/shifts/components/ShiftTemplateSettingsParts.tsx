import { memo, useCallback, useEffect, useState } from "react";
import Calendar from "lucide-react/dist/esm/icons/calendar";
import Plus from "lucide-react/dist/esm/icons/plus";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import X from "lucide-react/dist/esm/icons/x";
import { toast } from "sonner";

import { TableCell } from "@/components/ui/table";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { C, LAYOUT, STYLE } from "@/lib/design-tokens";
import { SHIFT_TYPE_LABELS, type ShiftTemplate } from "../types";
import {
  DEFAULT_SHIFT_TEMPLATE_FORM,
  templateToFormData,
  type TemplateFormData,
} from "./shift-template-form-model";
import { isShiftTemplateTimeHidden } from "./shift-template-form-utils";
import { ShiftTemplateProperties } from "./ShiftTemplateSidePanelFields";

interface ShiftTemplateRowProps {
  item: ShiftTemplate;
  canEdit: boolean;
  onEdit: () => void;
}

export const ShiftTemplateRow = memo(function ShiftTemplateRow({
  item,
  canEdit,
  onEdit,
}: ShiftTemplateRowProps) {
  const isTimeHidden = isShiftTemplateTimeHidden(item.shift_type);
  const timeLabel = isTimeHidden
    ? "-"
    : item.start_time && item.end_time
      ? `${item.start_time}〜${item.end_time}`
      : "-";

  return (
    <SortableDataTableRow
      id={item.id}
      dragLabel={`並べ替え: シフトテンプレート ${item.name} (ID ${item.id})`}
      dragDisabled={!canEdit}
    >
      <TableCell className={`font-medium text-sm ${C.text}`}>
        <DataTableRowButton
          aria-label={`詳細: シフトテンプレート ${item.name} (ID ${item.id})`}
          onClick={onEdit}
        >
          {item.name}
        </DataTableRowButton>
      </TableCell>
      <TableCell className={`text-sm ${C.text70}`}>
        {SHIFT_TYPE_LABELS[item.shift_type] ?? item.shift_type}
      </TableCell>
      <TableCell className={`text-sm ${C.text70}`}>{timeLabel}</TableCell>
      <TableCell className="text-center">
        <StatusPill isActive={item.is_active} />
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            aria-label={`編集: シフトテンプレート ${item.name} (ID ${item.id})`}
            onClick={onEdit}
          />
        ) : null}
      </TableCell>
    </SortableDataTableRow>
  );
});

interface ShiftTemplateToolbarProps {
  count: number;
  onCreate?: () => void;
}

export function ShiftTemplateToolbar({ count, onCreate }: ShiftTemplateToolbarProps) {
  return (
    <div className="flex items-center justify-between mb-4">
      <span className={`text-sm ${C.text50}`}>{count} 件</span>
      {onCreate !== undefined ? (
        <button
          type="button"
          onClick={onCreate}
          className={`inline-flex min-h-11 min-w-11 items-center gap-1 text-sm font-medium ${C.textBrand} ${C.hoverTextBrand} cursor-pointer transition-colors`}
        >
          <Plus className="size-4" />
          新規登録
        </button>
      ) : null}
    </div>
  );
}

interface ShiftTemplateSidePanelProps {
  item: ShiftTemplate | null;
  onClose: () => void;
  onSave: (data: TemplateFormData) => void;
  onDeleteRequest?: () => void;
  isSaving: boolean;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const ShiftTemplateSidePanel = memo(function ShiftTemplateSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  isSaving,
  readOnly = false,
  onDirtyChange,
}: ShiftTemplateSidePanelProps) {
  const [formData, setFormData] = useState<TemplateFormData>(() =>
    item ? templateToFormData(item) : DEFAULT_SHIFT_TEMPLATE_FORM,
  );
  const [isDirty, setIsDirty] = useState(false);

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    if (readOnly) return;
    setFormData(updater);
    setIsDirty(true);
  }, [readOnly]);

  const isTimeHidden = isShiftTemplateTimeHidden(formData.shift_type);

  const handleField = useCallback(
    <K extends keyof TemplateFormData>(key: K, value: TemplateFormData[K]) => {
      setFormDataDirty((prev) => ({ ...prev, [key]: value }));
    },
    [setFormDataDirty],
  );

  const handleBreakChange = useCallback(
    (index: number, field: "break_start" | "break_end", value: string) => {
      setFormDataDirty((prev) => ({
        ...prev,
        breaks: prev.breaks.map((b, i) => (i === index ? { ...b, [field]: value } : b)),
      }));
    },
    [setFormDataDirty],
  );

  const handleAddBreak = useCallback(() => {
    setFormDataDirty((prev) => ({
      ...prev,
      breaks: [...prev.breaks, { break_start: "12:00", break_end: "13:00" }],
    }));
  }, [setFormDataDirty]);

  const handleRemoveBreak = useCallback(
    (index: number) => {
      setFormDataDirty((prev) => ({
        ...prev,
        breaks: prev.breaks.filter((_, i) => i !== index),
      }));
    },
    [setFormDataDirty],
  );

  const handleAction = useCallback(() => {
    if (readOnly) return;
    if (!isTimeHidden && (!formData.start_time || !formData.end_time)) {
      toast.error("勤務種別では開始時刻と終了時刻を入力してください");
      return;
    }
    onSave(formData);
    setIsDirty(false);
  }, [formData, isTimeHidden, onSave, readOnly]);

  return (
    <div className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0`}>
      <div className={STYLE.sidePeekToolbar}>
        <span className={`text-xs ${C.text35} pl-1 select-none`}>
          {item !== null ? (readOnly ? "詳細" : "編集") : "新規作成"}
        </span>
        <div className="flex items-center gap-1">
          {item !== null && onDeleteRequest !== undefined ? (
            <button
              type="button"
              onClick={onDeleteRequest}
              className={`${STYLE.sidePeekToolbarBtn} cursor-pointer ${STYLE.btnDangerGhost}`}
              aria-label={`削除: シフトテンプレート ${item.name} (ID ${item.id})`}
            >
              <Trash2 className="size-4" aria-hidden="true" />
            </button>
          ) : null}
          <button
            type="button"
            onClick={onClose}
            className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
            aria-label="閉じる"
            >
              <X className="size-4" aria-hidden="true" />
            </button>
        </div>
      </div>

      <div className={STYLE.sidePeekBody}>
        <div className="px-10 pb-8">
          <div className="pt-4 pb-2">
            <div className={STYLE.pageIcon}>
              <Calendar className={LAYOUT.pageIcon.innerIcon} />
            </div>
          </div>

          <div className="pb-1 mb-4">
            <input
              type="text"
              aria-label="テンプレート名"
              className={`w-full bg-transparent ${C.text} ${C.textPlaceholderFaint} outline-none border-none p-0`}
              style={{
                fontSize: LAYOUT.pageTitle.fontSize,
                fontWeight: LAYOUT.pageTitle.fontWeight,
                lineHeight: LAYOUT.pageTitle.lineHeight,
                letterSpacing: LAYOUT.pageTitle.letterSpacing,
              }}
              value={formData.name}
              onChange={(e) => handleField("name", e.target.value)}
              placeholder="テンプレート名"
              readOnly={readOnly}
              autoFocus={!readOnly}
            />
          </div>

          <div className={`${STYLE.sectionDivider} mb-1`} />

          <ShiftTemplateProperties
            formData={formData}
            isTimeHidden={isTimeHidden}
            readOnly={readOnly}
            onField={handleField}
            onBreakChange={handleBreakChange}
            onAddBreak={handleAddBreak}
            onRemoveBreak={handleRemoveBreak}
          />
        </div>
      </div>

      <div className={STYLE.sidePeekFooter}>
        <button type="button" onClick={onClose} className={STYLE.sidePeekCancelBtn}>
          {readOnly ? "閉じる" : "キャンセル"}
        </button>
        {!readOnly ? (
          <button
            type="button"
            onClick={handleAction}
            disabled={isSaving || !formData.name.trim()}
            className={`px-4 py-[7px] text-base ${C.textOnBrand} ${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} rounded-full transition-colors cursor-pointer ${STYLE.pillShadow}`}
          >
            {isSaving ? "保存中..." : "保存"}
          </button>
        ) : null}
      </div>
    </div>
  );
});

interface ShiftTemplateDeleteDialogProps {
  pendingDelete: ShiftTemplate | null;
  onClose: () => void;
  onConfirm: () => void;
}

export function ShiftTemplateDeleteDialog({
  pendingDelete,
  onClose,
  onConfirm,
}: ShiftTemplateDeleteDialogProps) {
  return (
    <ConfirmDialog
      open={pendingDelete !== null}
      onClose={onClose}
      title="テンプレートを削除しますか？"
      description={`「${pendingDelete?.name ?? ""}」を削除します。この操作は取り消せません。`}
      confirmLabel="削除"
      variant="destructive"
      onConfirm={onConfirm}
    />
  );
}
