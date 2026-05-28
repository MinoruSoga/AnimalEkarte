import { memo, useCallback, useEffect, useState } from "react";
import type { ReactNode } from "react";
import Calendar from "lucide-react/dist/esm/icons/calendar";
import Plus from "lucide-react/dist/esm/icons/plus";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import X from "lucide-react/dist/esm/icons/x";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { TableCell } from "@/components/ui/table";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { C, ICON, LAYOUT, STYLE } from "@/lib/design-tokens";
import { ShiftTypeOff, ShiftTypePaidLeave } from "@/types/generated/models";
import { SHIFT_TYPE_LABELS, type ShiftTemplate, type ShiftType } from "../types";

export const SHIFT_TEMPLATE_COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "テンプレート名" },
  { header: "種別", className: "w-[80px]" },
  { header: "時間", className: "w-[140px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

const SHIFT_TYPE_OPTIONS = (
  Object.entries(SHIFT_TYPE_LABELS) as [ShiftType, string][]
).map(([value, label]) => (
  <SelectItem key={value} value={value}>
    {label}
  </SelectItem>
));

export interface BreakInput {
  break_start: string;
  break_end: string;
}

export interface TemplateFormData {
  name: string;
  shift_type: ShiftType;
  start_time: string;
  end_time: string;
  notes: string;
  is_active: boolean;
  breaks: BreakInput[];
}

const DEFAULT_FORM: TemplateFormData = {
  name: "",
  shift_type: "full",
  start_time: "",
  end_time: "",
  notes: "",
  is_active: true,
  breaks: [],
};

function templateToFormData(template: ShiftTemplate): TemplateFormData {
  return {
    name: template.name,
    shift_type: template.shift_type,
    start_time: template.start_time,
    end_time: template.end_time,
    notes: template.notes,
    is_active: template.is_active,
    breaks: template.breaks.map((b) => ({
      break_start: b.break_start,
      break_end: b.break_end,
    })),
  };
}

interface ShiftTemplateRowProps {
  item: ShiftTemplate;
  onEdit: () => void;
}

export const ShiftTemplateRow = memo(function ShiftTemplateRow({
  item,
  onEdit,
}: ShiftTemplateRowProps) {
  const isTimeHidden = item.shift_type === ShiftTypeOff || item.shift_type === ShiftTypePaidLeave;
  const timeLabel = isTimeHidden
    ? "-"
    : item.start_time && item.end_time
      ? `${item.start_time}〜${item.end_time}`
      : "-";

  return (
    <SortableDataTableRow id={item.id} onClick={onEdit}>
      <TableCell className={`font-medium text-sm ${C.text} py-2.5`}>{item.name}</TableCell>
      <TableCell className={`text-sm ${C.text70} py-2.5`}>
        {SHIFT_TYPE_LABELS[item.shift_type] ?? item.shift_type}
      </TableCell>
      <TableCell className={`text-sm ${C.text70} py-2.5`}>{timeLabel}</TableCell>
      <TableCell className="text-center py-2.5">
        <NotionStatusPill isActive={item.is_active} />
      </TableCell>
      <TableCell className="text-right py-2.5">
        <RowActionButton onClick={onEdit} />
      </TableCell>
    </SortableDataTableRow>
  );
});

interface ShiftTemplateToolbarProps {
  count: number;
  onCreate: () => void;
}

export function ShiftTemplateToolbar({ count, onCreate }: ShiftTemplateToolbarProps) {
  return (
    <div className="flex items-center justify-between mb-4">
      <span className={`text-sm ${C.text50}`}>{count} 件</span>
      <button
        type="button"
        onClick={onCreate}
        className={`inline-flex items-center gap-1 text-sm font-medium ${C.accent} ${C.hoverTextAccent} cursor-pointer transition-colors`}
      >
        <Plus className="size-4" />
        新規登録
      </button>
    </div>
  );
}

interface ShiftTemplateSidePanelProps {
  item: ShiftTemplate | null;
  onClose: () => void;
  onSave: (data: TemplateFormData) => void;
  onDeleteRequest: () => void;
  isSaving: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const ShiftTemplateSidePanel = memo(function ShiftTemplateSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  isSaving,
  onDirtyChange,
}: ShiftTemplateSidePanelProps) {
  const [formData, setFormData] = useState<TemplateFormData>(() =>
    item ? templateToFormData(item) : DEFAULT_FORM,
  );
  const [isDirty, setIsDirty] = useState(false);

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const isTimeHidden =
    formData.shift_type === ShiftTypeOff || formData.shift_type === ShiftTypePaidLeave;

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
    if (!isTimeHidden && (!formData.start_time || !formData.end_time)) {
      toast.error("勤務種別では開始時刻と終了時刻を入力してください");
      return;
    }
    onSave(formData);
    setIsDirty(false);
  }, [formData, isTimeHidden, onSave]);

  return (
    <div className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0`}>
      <div className={STYLE.sidePeekToolbar}>
        <span className={`text-xs ${C.text35} pl-1 select-none`}>
          {item !== null ? "編集" : "新規作成"}
        </span>
        <div className="flex items-center gap-1">
          {item !== null ? (
            <button
              type="button"
              onClick={onDeleteRequest}
              className={`${STYLE.sidePeekToolbarBtn} cursor-pointer ${STYLE.btnDangerGhost}`}
            >
              <Trash2 className="size-4" />
            </button>
          ) : null}
          <button
            type="button"
            onClick={onClose}
            className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
            aria-label="閉じる"
          >
            <X className="size-4" />
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
              className={`w-full bg-transparent ${C.text} placeholder:text-[rgba(55,53,47,0.15)] outline-none border-none p-0`}
              style={{
                fontSize: LAYOUT.pageTitle.fontSize,
                fontWeight: LAYOUT.pageTitle.fontWeight,
                lineHeight: LAYOUT.pageTitle.lineHeight,
              }}
              value={formData.name}
              onChange={(e) => handleField("name", e.target.value)}
              placeholder="テンプレート名"
              autoFocus
            />
          </div>

          <div className={`${STYLE.sectionDivider} mb-1`} />

          <div className="py-1">
            <PropertyRow label="ステータス">
              <button
                type="button"
                onClick={() => handleField("is_active", !formData.is_active)}
                className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
              >
                <NotionStatusPill isActive={formData.is_active} />
              </button>
            </PropertyRow>

            <PropertyRow label="シフト種別">
              <Select
                value={formData.shift_type}
                onValueChange={(v) => handleField("shift_type", v as ShiftType)}
              >
                <SelectTrigger className="h-7 text-sm border-0 shadow-none bg-transparent px-1.5">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>{SHIFT_TYPE_OPTIONS}</SelectContent>
              </Select>
            </PropertyRow>

            {!isTimeHidden ? (
              <>
                <PropertyRow label="開始時刻">
                  <PropInput
                    type="time"
                    value={formData.start_time}
                    onChange={(v) => handleField("start_time", v)}
                  />
                </PropertyRow>
                <PropertyRow label="終了時刻">
                  <PropInput
                    type="time"
                    value={formData.end_time}
                    onChange={(v) => handleField("end_time", v)}
                  />
                </PropertyRow>
              </>
            ) : null}

            <PropertyRow label="メモ">
              <PropInput
                value={formData.notes}
                onChange={(v) => handleField("notes", v)}
                placeholder="補足情報など"
              />
            </PropertyRow>
          </div>

          {!isTimeHidden ? (
            <div className="mt-4">
              <div className="flex items-center justify-between mb-2">
                <span className={`text-sm font-medium ${C.text}`}>休憩時間</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-7 px-2 text-xs"
                  onClick={handleAddBreak}
                >
                  <Plus className={`${ICON.xxs} mr-1`} />
                  追加
                </Button>
              </div>
              {formData.breaks.map((b, i) => (
                <div key={i} className="flex items-center gap-2 mb-2">
                  <Input
                    type="time"
                    value={b.break_start}
                    onChange={(e) => handleBreakChange(i, "break_start", e.target.value)}
                    className="flex-1 h-8 text-sm"
                  />
                  <span className={`text-xs ${C.text50}`}>〜</span>
                  <Input
                    type="time"
                    value={b.break_end}
                    onChange={(e) => handleBreakChange(i, "break_end", e.target.value)}
                    className="flex-1 h-8 text-sm"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="h-8 w-8 p-0"
                    onClick={() => handleRemoveBreak(i)}
                  >
                    <X className={ICON.smXs} />
                  </Button>
                </div>
              ))}
              {formData.breaks.length === 0 ? (
                <p className={`text-xs ${C.text40}`}>休憩なし</p>
              ) : null}
            </div>
          ) : null}
        </div>
      </div>

      <div className={STYLE.sidePeekFooter}>
        <button type="button" onClick={onClose} className={STYLE.sidePeekCancelBtn}>
          キャンセル
        </button>
        <button
          type="button"
          onClick={handleAction}
          disabled={isSaving || !formData.name.trim()}
          className={STYLE.sidePeekSaveBtn}
        >
          {isSaving ? "保存中..." : "保存"}
        </button>
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

function PropertyRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div
      className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px]`}
    >
      <div className={`w-[120px] shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}>
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
}

function PropInput({
  value,
  onChange,
  placeholder,
  type = "text",
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  type?: string;
}) {
  return (
    <input
      type={type}
      className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} transition-colors ${C.textPlaceholder}`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder ?? "空"}
    />
  );
}
