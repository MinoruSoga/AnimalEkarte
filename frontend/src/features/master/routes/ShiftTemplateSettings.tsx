// React/Framework
import { useState, useMemo, useCallback, memo } from "react";
import { useNavigate } from "react-router";

// DnD
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

// Shared hooks
import { useSortableList } from "@/hooks/use-sortable-list";

// External
import Calendar from "lucide-react/dist/esm/icons/calendar";
import Plus from "lucide-react/dist/esm/icons/plus";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import X from "lucide-react/dist/esm/icons/x";
import GripVertical from "lucide-react/dist/esm/icons/grip-vertical";
import { toast } from "sonner";

// Internal
import { TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { C, STYLE, LAYOUT, ICON } from "@/lib/design-tokens";

// Shifts feature API (cross-feature via index)
import {
  useGetShiftTemplates,
  useCreateShiftTemplate,
  useUpdateShiftTemplate,
  useDeleteShiftTemplate,
  useReorderShiftTemplates,
} from "@/features/shifts";
import type { ShiftTemplate } from "@/features/shifts";
import { SHIFT_TYPE_LABELS } from "@/features/shifts/types";
import { ShiftTypeOff, ShiftTypePaidLeave } from "@/types/generated/models";
import type { ShiftType } from "@/features/shifts/types";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const COLUMNS = [
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

// ─────────────────────────────────────────────────
// Form state
// ─────────────────────────────────────────────────

interface BreakInput {
  break_start: string;
  break_end: string;
}

interface TemplateFormData {
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

function templateToFormData(t: ShiftTemplate): TemplateFormData {
  return {
    name: t.name,
    shift_type: t.shift_type,
    start_time: t.start_time,
    end_time: t.end_time,
    notes: t.notes,
    is_active: t.is_active,
    breaks: t.breaks.map((b) => ({ break_start: b.break_start, break_end: b.break_end })),
  };
}

// ─────────────────────────────────────────────────
// PropertyRow (Notion style)
// ─────────────────────────────────────────────────

function PropertyRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div
      className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px]`}
    >
      <div className="w-[120px] shrink-0 text-sm text-[#37352F]/65 select-none truncate flex items-center">
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

// ─────────────────────────────────────────────────
// SortableRow
// ─────────────────────────────────────────────────

interface SortableRowProps {
  item: ShiftTemplate;
  onEdit: () => void;
}

const SortableRow = memo(function SortableRow({ item, onEdit }: SortableRowProps) {
  const isTimeHidden = item.shift_type === ShiftTypeOff || item.shift_type === ShiftTypePaidLeave;
  const timeLabel = isTimeHidden
    ? "-"
    : item.start_time && item.end_time
      ? `${item.start_time}〜${item.end_time}`
      : "-";

  return (
    <SortableDataTableRow id={item.id} onClick={onEdit}>
      {({ listeners }) => (
        <>
          <TableCell className="w-[32px] py-2.5 text-[#37352F]/20 cursor-grab" {...listeners}>
            <GripVertical className="size-4" />
          </TableCell>
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
        </>
      )}
    </SortableDataTableRow>
  );
});

// ─────────────────────────────────────────────────
// SidePanel
// ─────────────────────────────────────────────────

interface SidePanelProps {
  item: ShiftTemplate | null;
  formData: TemplateFormData;
  onChange: (data: TemplateFormData) => void;
  onClose: () => void;
  onSave: () => void;
  onDeleteRequest: () => void;
  isSaving: boolean;
}

const SidePanel = memo(function SidePanel({
  item,
  formData,
  onChange,
  onClose,
  onSave,
  onDeleteRequest,
  isSaving,
}: SidePanelProps) {
  const isTimeHidden =
    formData.shift_type === ShiftTypeOff || formData.shift_type === ShiftTypePaidLeave;

  const handleField = useCallback(
    <K extends keyof TemplateFormData>(key: K, value: TemplateFormData[K]) => {
      onChange({ ...formData, [key]: value });
    },
    [formData, onChange],
  );

  const handleBreakChange = useCallback(
    (index: number, field: "break_start" | "break_end", value: string) => {
      const newBreaks = formData.breaks.map((b, i) =>
        i === index ? { ...b, [field]: value } : b,
      );
      onChange({ ...formData, breaks: newBreaks });
    },
    [formData, onChange],
  );

  const handleAddBreak = useCallback(() => {
    onChange({
      ...formData,
      breaks: [...formData.breaks, { break_start: "12:00", break_end: "13:00" }],
    });
  }, [formData, onChange]);

  const handleRemoveBreak = useCallback(
    (index: number) => {
      onChange({ ...formData, breaks: formData.breaks.filter((_, i) => i !== index) });
    },
    [formData, onChange],
  );

  return (
    <div className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0`}>
      {/* Toolbar */}
      <div className={STYLE.sidePeekToolbar}>
        <span className={`text-xs ${C.text35} pl-1 select-none`}>
          {item !== null ? "編集" : "新規作成"}
        </span>
        <div className="flex items-center gap-1">
          {item !== null ? (
            <button
              type="button"
              onClick={onDeleteRequest}
              className={`${STYLE.sidePeekToolbarBtn} cursor-pointer text-[#EB5757] hover:bg-[#EB5757]/10`}
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

      {/* Body */}
      <div className={STYLE.sidePeekBody}>
        <div className="px-10 pb-8">
          {/* Icon */}
          <div className="pt-4 pb-2">
            <div className={STYLE.pageIcon}>
              <Calendar className={LAYOUT.pageIcon.innerIcon} />
            </div>
          </div>

          {/* Title */}
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

          {/* Properties */}
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

          {/* Breaks */}
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

      {/* Footer */}
      <div className={STYLE.sidePeekFooter}>
        <button type="button" onClick={onClose} className={STYLE.sidePeekCancelBtn}>
          キャンセル
        </button>
        <button
          type="button"
          onClick={onSave}
          disabled={isSaving || !formData.name.trim()}
          className={STYLE.sidePeekSaveBtn}
        >
          {isSaving ? "保存中..." : "保存"}
        </button>
      </div>
    </div>
  );
});

// ─────────────────────────────────────────────────
// Main component
// ─────────────────────────────────────────────────

export function ShiftTemplateSettings() {
  const navigate = useNavigate();
  const { data: templates = [], isLoading } = useGetShiftTemplates();
  const createMutation = useCreateShiftTemplate();
  const updateMutation = useUpdateShiftTemplate();
  const deleteMutation = useDeleteShiftTemplate();
  const reorderMutation = useReorderShiftTemplates();

  const [selectedItem, setSelectedItem] = useState<ShiftTemplate | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState<TemplateFormData>(DEFAULT_FORM);
  const [pendingDelete, setPendingDelete] = useState<ShiftTemplate | null>(null);

  // D&D reorder
  const { orderedItems, sensors, handleDragEnd } = useSortableList({
    items: templates,
    onReorder: (newIds) =>
      reorderMutation.mutate(newIds.map(Number)),
  });

  const handleCreate = useCallback(() => {
    setSelectedItem(null);
    setFormData(DEFAULT_FORM);
    setIsEditing(true);
  }, []);

  const handleEdit = useCallback((item: ShiftTemplate) => {
    setSelectedItem(item);
    setFormData(templateToFormData(item));
    setIsEditing(true);
  }, []);

  const handleClose = useCallback(() => {
    setIsEditing(false);
    setSelectedItem(null);
  }, []);

  const handleSave = useCallback(() => {
    const breaks = formData.breaks.filter((b) => b.break_start && b.break_end);
    const isTimeHidden =
      formData.shift_type === ShiftTypeOff || formData.shift_type === ShiftTypePaidLeave;

    if (selectedItem !== null) {
      updateMutation.mutate(
        {
          id: selectedItem.id,
          input: {
            name: formData.name,
            shift_type: formData.shift_type,
            start_time: isTimeHidden ? null : formData.start_time || undefined,
            end_time: isTimeHidden ? null : formData.end_time || undefined,
            notes: formData.notes,
            is_active: formData.is_active,
            breaks: isTimeHidden ? [] : breaks,
          },
        },
        {
          onSuccess: () => {
            toast.success("テンプレートを更新しました");
            handleClose();
          },
        },
      );
    } else {
      createMutation.mutate(
        {
          name: formData.name,
          shift_type: formData.shift_type,
          start_time: isTimeHidden ? undefined : formData.start_time || undefined,
          end_time: isTimeHidden ? undefined : formData.end_time || undefined,
          notes: formData.notes,
          is_active: formData.is_active,
          breaks: isTimeHidden ? [] : breaks,
        },
        {
          onSuccess: () => {
            toast.success("テンプレートを作成しました");
            handleClose();
          },
        },
      );
    }
  }, [formData, selectedItem, createMutation, updateMutation, handleClose]);

  const handleDeleteConfirm = useCallback(() => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        toast.success("テンプレートを削除しました");
        setPendingDelete(null);
        if (selectedItem?.id === pendingDelete.id) handleClose();
      },
    });
  }, [pendingDelete, deleteMutation, selectedItem, handleClose]);

  const isSaving = createMutation.isPending || updateMutation.isPending;

  const emptyContent = useMemo(
    () =>
      !isLoading && orderedItems.length === 0 ? (
        <div className={`flex items-center justify-center py-12 text-sm ${C.text40}`}>
          テンプレートがありません
        </div>
      ) : null,
    [isLoading, orderedItems.length],
  );

  return (
    <div className="flex h-full overflow-hidden">
      <div className="flex-1 min-w-0 overflow-auto">
        <PageLayout
          title="シフトテンプレートマスタ"
          icon={<Calendar className="size-5" />}
          onBack={() => navigate("/settings")}
          maxWidth="max-w-full"
        >
          {/* Toolbar */}
          <div className="flex items-center justify-end mb-4">
            <button
              type="button"
              onClick={handleCreate}
              className="inline-flex items-center gap-1 text-sm font-medium text-[#2383E2] hover:text-[#1B6EC2] cursor-pointer transition-colors"
            >
              <Plus className="size-4" />
              新規登録
            </button>
          </div>

          {/* Table */}
          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={handleDragEnd}
          >
            <SortableContext
              items={orderedItems.map((i) => i.id)}
              strategy={verticalListSortingStrategy}
            >
              <DataTable
                columns={COLUMNS}
                data={orderedItems}
                renderRow={(item) => (
                  <SortableRow key={item.id} item={item} onEdit={() => handleEdit(item)} />
                )}
              />
            </SortableContext>
          </DndContext>

          {emptyContent}
        </PageLayout>
      </div>

      {/* Side peek */}
      {isEditing ? (
        <SidePanel
          key={selectedItem ? selectedItem.id : "new"}
          item={selectedItem}
          formData={formData}
          onChange={setFormData}
          onClose={handleClose}
          onSave={handleSave}
          onDeleteRequest={() => {
            if (selectedItem) setPendingDelete(selectedItem);
          }}
          isSaving={isSaving}
        />
      ) : null}

      {/* Delete confirm */}
      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        title="このテンプレートを削除しますか？"
        description="この操作は取り消せません。"
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </div>
  );
}
