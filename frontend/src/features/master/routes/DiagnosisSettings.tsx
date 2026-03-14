// React/Framework
import type { ReactNode } from "react";
import { useState, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";

// DnD
import {
  DndContext,
  closestCenter,
  type DragEndEvent,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import {
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
  useSortable,
  arrayMove,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";

// External
import {
  Plus,
  FolderTree,
  X,
  GripVertical,
  Trash2,
  ClipboardList,
} from "lucide-react";
import { toast } from "sonner";

// Internal
import { TableCell } from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import * as TabsPrimitive from "@radix-ui/react-tabs";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { StatusBadge } from "@/components/shared/StatusBadge";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { getMasterStatusColor } from "@/utils/status-helpers";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import {
  useListDiagnosisCategories,
  useCreateDiagnosisCategory,
  useUpdateDiagnosisCategory,
  useDeleteDiagnosisCategory,
  useReorderDiagnosisCategories,
  useListDiagnosisNames,
  useCreateDiagnosisName,
  useUpdateDiagnosisName,
  useDeleteDiagnosisName,
  useReorderDiagnosisNames,
} from "@/features/master/api/diagnosis";

// Types
import type { DiagnosisCategory, DiagnosisName } from "@/features/master/api/diagnosis";
import type {
  CreateDiagnosisCategoryRequest,
  UpdateDiagnosisCategoryRequest,
  CreateDiagnosisNameRequest,
  UpdateDiagnosisNameRequest,
} from "@/types/diagnosis";

// ─────────────────────────────────────────────────
// Columns
// ─────────────────────────────────────────────────

const CATEGORY_COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "カテゴリ名" },
  { header: "備考", className: "w-[240px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

const NAME_COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "診断病名" },
  { header: "所属カテゴリ", className: "w-[160px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

const TABS = [
  { value: "diagnosis_category", label: "診断病名カテゴリ" },
  { value: "diagnosis_name", label: "診断病名" },
] as const;

// ─────────────────────────────────────────────────
// Notion Status Pill
// ─────────────────────────────────────────────────

const STATUS_CONFIG = {
  active: {
    dot: "bg-[#2383E2]",
    label: "有効",
    bg: "bg-[#D3E5EF]",
    text: "text-[#183B56]",
  },
  inactive: {
    dot: "bg-[#37352F]/10",
    label: "無効",
    bg: "bg-[#E3E2E0]",
    text: "text-[#37352F]/60",
  },
} as const;

function NotionStatusPill({ isActive }: { isActive: boolean }) {
  const cfg = STATUS_CONFIG[isActive ? "active" : "inactive"];
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-sm text-xs ${cfg.bg} ${cfg.text}`}
    >
      <span className={`size-[7px] rounded-full ${cfg.dot}`} />
      {cfg.label}
    </span>
  );
}

// ─────────────────────────────────────────────────
// Property Row (Notion-style)
// ─────────────────────────────────────────────────

function PropertyRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div
      className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px]`}
    >
      <div className="w-[140px] shrink-0 text-sm text-[#37352F]/65 select-none truncate flex items-center">
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// Inline input for property rows
// ─────────────────────────────────────────────────

function PropInput({
  value,
  onChange,
  placeholder,
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
}) {
  return (
    <input
      type="text"
      className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder ?? "空"}
    />
  );
}

// ─────────────────────────────────────────────────
// DiagnosisCategorySidePanel
// ─────────────────────────────────────────────────

function DiagnosisCategorySidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: {
  item: DiagnosisCategory | null;
  onClose: () => void;
  onSave: (data: { name: string; description: string; isActive: boolean }) => void;
  onDeleteRequest: () => void;
}) {
  const [formData, setFormData] = useState({
    name: item?.name ?? "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  });

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
        <div className="px-16 pb-8">
          <div className="pt-4 pb-2">
            <div className={STYLE.pageIcon}>
              <FolderTree className={LAYOUT.pageIcon.innerIcon} />
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
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="無題"
              autoFocus
            />
          </div>
          <div className={`${STYLE.sectionDivider} mb-1`} />
          <div className="py-1">
            <PropertyRow label="ステータス">
              <button
                type="button"
                onClick={() => setFormData({ ...formData, isActive: !formData.isActive })}
                className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
              >
                <NotionStatusPill isActive={formData.isActive} />
              </button>
            </PropertyRow>
            <PropertyRow label="備考">
              <PropInput
                value={formData.description}
                onChange={(v) => setFormData({ ...formData, description: v })}
                placeholder="補足情報など"
              />
            </PropertyRow>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className={STYLE.sidePeekFooter}>
        <button type="button" onClick={onClose} className={STYLE.sidePeekCancelBtn}>
          キャンセル
        </button>
        <button
          type="button"
          onClick={() => onSave(formData)}
          className={STYLE.sidePeekSaveBtn}
        >
          保存
        </button>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// DiagnosisNameSidePanel
// ─────────────────────────────────────────────────

function DiagnosisNameSidePanel({
  item,
  categories,
  onClose,
  onSave,
  onDeleteRequest,
}: {
  item: DiagnosisName | null;
  categories: DiagnosisCategory[];
  onClose: () => void;
  onSave: (data: { name: string; diagnosisCategoryId: string; description: string; isActive: boolean }) => void;
  onDeleteRequest: () => void;
}) {
  const [formData, setFormData] = useState({
    name: item?.name ?? "",
    diagnosisCategoryId: item
      ? String(item.diagnosisCategoryId)
      : categories[0]
        ? String(categories[0].id)
        : "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  });

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
        <div className="px-16 pb-8">
          <div className="pt-4 pb-2">
            <div className={STYLE.pageIcon}>
              <FolderTree className={LAYOUT.pageIcon.innerIcon} />
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
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="無題"
              autoFocus
            />
          </div>

          <div className={`${STYLE.sectionDivider} mb-1`} />

          <div className="py-1">
            <PropertyRow label="ステータス">
              <button
                type="button"
                onClick={() => setFormData({ ...formData, isActive: !formData.isActive })}
                className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
              >
                <NotionStatusPill isActive={formData.isActive} />
              </button>
            </PropertyRow>

            <PropertyRow label="カテゴリ">
              <Select
                value={formData.diagnosisCategoryId}
                onValueChange={(v) => setFormData({ ...formData, diagnosisCategoryId: v })}
              >
                <SelectTrigger className={STYLE.selectCompact}>
                  <SelectValue placeholder="カテゴリを選択" />
                </SelectTrigger>
                <SelectContent>
                  {categories.map((cat) => (
                    <SelectItem key={cat.id} value={String(cat.id)}>
                      {cat.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </PropertyRow>

            <PropertyRow label="備考">
              <PropInput
                value={formData.description}
                onChange={(v) => setFormData({ ...formData, description: v })}
                placeholder="補足情報など"
              />
            </PropertyRow>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className={STYLE.sidePeekFooter}>
        <button type="button" onClick={onClose} className={STYLE.sidePeekCancelBtn}>
          キャンセル
        </button>
        <button
          type="button"
          onClick={() => onSave(formData)}
          className={STYLE.sidePeekSaveBtn}
        >
          保存
        </button>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// SortableCategoryRow
// ─────────────────────────────────────────────────

function SortableCategoryRow({
  item,
  onEdit,
}: {
  item: DiagnosisCategory;
  onEdit: () => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: item.id });
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };
  return (
    <DataTableRow ref={setNodeRef} style={style} {...attributes} onClick={onEdit}>
      <TableCell
        className="w-[32px] py-2.5 text-[#37352F]/20 cursor-grab"
        {...listeners}
      >
        <GripVertical className="size-4" />
      </TableCell>
      <TableCell className={`font-medium text-sm ${C.text} py-2.5`}>
        {item.name}
      </TableCell>
      <TableCell className={`text-sm ${C.text70} py-2.5 truncate max-w-[240px]`}>
        {item.description || "-"}
      </TableCell>
      <TableCell className="text-center py-2.5">
        <StatusBadge colorClass={getMasterStatusColor(item.isActive ? "active" : "inactive")}>
          {item.isActive ? "有効" : "無効"}
        </StatusBadge>
      </TableCell>
      <TableCell className="text-right py-2.5">
        <RowActionButton onClick={onEdit} />
      </TableCell>
    </DataTableRow>
  );
}

// ─────────────────────────────────────────────────
// DiagnosisCategoryTab
// ─────────────────────────────────────────────────

function DiagnosisCategoryTab() {
  const [selectedItem, setSelectedItem] = useState<DiagnosisCategory | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<DiagnosisCategory | null>(null);
  const [overrideOrder, setOverrideOrder] = useState<string[]>([]);

  const { data: rawCategories } = useListDiagnosisCategories();
  const createMutation = useCreateDiagnosisCategory();
  const updateMutation = useUpdateDiagnosisCategory();
  const deleteMutation = useDeleteDiagnosisCategory();
  const reorderMutation = useReorderDiagnosisCategories();

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  const orderedCategories = useMemo(() => {
    const base = rawCategories ?? [];
    if (overrideOrder.length === 0) return base;
    const idx = new Map<string, number>(overrideOrder.map((id, i) => [id, i]));
    return [...base].sort((a, b) => (idx.get(a.id) ?? 0) - (idx.get(b.id) ?? 0));
  }, [rawCategories, overrideOrder]);

  const filteredItems = useMemo(() => {
    if (!searchTerm) return orderedCategories;
    const lower = searchTerm.toLowerCase();
    return orderedCategories.filter((c) => c.name.toLowerCase().includes(lower));
  }, [orderedCategories, searchTerm]);

  const handleCategoryDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const currentIds = orderedCategories.map((i) => i.id);
    const newOrder = arrayMove(
      currentIds,
      currentIds.indexOf(String(active.id)),
      currentIds.indexOf(String(over.id)),
    );
    setOverrideOrder(newOrder);
    reorderMutation.mutate(
      { ids: newOrder.map(Number) },
      { onSuccess: () => setOverrideOrder([]) },
    );
  };

  const handleEdit = (item: DiagnosisCategory) => {
    setSelectedItem(item);
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setIsEditing(true);
  };

  const handleClose = () => {
    setIsEditing(false);
    setSelectedItem(null);
  };

  const handleSave = (data: { name: string; description: string; isActive: boolean }) => {
    if (!data.name.trim()) {
      toast.error("カテゴリ名は必須です");
      return;
    }
    if (selectedItem) {
      const req: UpdateDiagnosisCategoryRequest = {
        name: data.name,
        description: data.description || undefined,
        is_active: data.isActive,
      };
      updateMutation.mutate(
        { id: selectedItem.id, req },
        {
          onSuccess: () => { toast.success("更新しました"); handleClose(); },
          onError: () => toast.error("更新に失敗しました"),
        },
      );
    } else {
      const req: CreateDiagnosisCategoryRequest = {
        name: data.name,
        description: data.description || undefined,
        is_active: true,
      };
      createMutation.mutate(req, {
        onSuccess: () => { toast.success("登録しました"); handleClose(); },
        onError: () => toast.error("登録に失敗しました"),
      });
    }
  };

  const handleDeleteConfirm = () => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        handleClose();
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  };

  return (
    <>
      <div className="flex h-full">
        {/* Table area */}
        <div className="flex flex-col gap-4 flex-1 min-w-0">
          <div className="flex items-center gap-3">
            <div className="flex-1 min-w-0">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="カテゴリ名で検索..."
                count={filteredItems.length}
              />
            </div>
            <PrimaryButton onClick={handleCreate}>
              <Plus className="mr-1.5 size-4" />
              新規登録
            </PrimaryButton>
          </div>

          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={handleCategoryDragEnd}
          >
            <SortableContext
              items={filteredItems.map((i) => i.id)}
              strategy={verticalListSortingStrategy}
            >
              <DataTable
                columns={CATEGORY_COLUMNS}
                data={filteredItems}
                emptyMessage="診断カテゴリが登録されていません"
                renderRow={(item) => (
                  <SortableCategoryRow
                    key={item.id}
                    item={item}
                    onEdit={() => handleEdit(item)}
                  />
                )}
              />
            </SortableContext>
          </DndContext>
        </div>

        {/* Side peek */}
        {isEditing ? (
          <DiagnosisCategorySidePanel
            key={selectedItem ? String(selectedItem.id) : "new-category"}
            item={selectedItem}
            onClose={handleClose}
            onSave={handleSave}
            onDeleteRequest={() => setPendingDelete(selectedItem)}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        title="診断カテゴリを削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。このカテゴリに属する診断名も影響を受けます。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}

// ─────────────────────────────────────────────────
// SortableNameRow
// ─────────────────────────────────────────────────

function SortableNameRow({
  item,
  categoryMap,
  onEdit,
}: {
  item: DiagnosisName;
  categoryMap: Map<string, string>;
  onEdit: () => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: item.id });
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };
  return (
    <DataTableRow ref={setNodeRef} style={style} {...attributes} onClick={onEdit}>
      <TableCell
        className="w-[32px] py-2.5 text-[#37352F]/20 cursor-grab"
        {...listeners}
      >
        <GripVertical className="size-4" />
      </TableCell>
      <TableCell className={`font-medium text-sm ${C.text} py-2.5`}>
        {item.name}
      </TableCell>
      <TableCell className={`text-sm ${C.text70} py-2.5`}>
        {categoryMap.get(item.diagnosisCategoryId) ?? "-"}
      </TableCell>
      <TableCell className="text-center py-2.5">
        <StatusBadge colorClass={getMasterStatusColor(item.isActive ? "active" : "inactive")}>
          {item.isActive ? "有効" : "無効"}
        </StatusBadge>
      </TableCell>
      <TableCell className="text-right py-2.5">
        <RowActionButton onClick={onEdit} />
      </TableCell>
    </DataTableRow>
  );
}

// ─────────────────────────────────────────────────
// DiagnosisNameTab
// ─────────────────────────────────────────────────

function DiagnosisNameTab() {
  const [selectedItem, setSelectedItem] = useState<DiagnosisName | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<DiagnosisName | null>(null);
  const [overrideOrder, setOverrideOrder] = useState<string[]>([]);

  const { data: rawCategories } = useListDiagnosisCategories();
  const { data: rawNames } = useListDiagnosisNames();

  const createMutation = useCreateDiagnosisName();
  const updateMutation = useUpdateDiagnosisName();
  const deleteMutation = useDeleteDiagnosisName();
  const reorderMutation = useReorderDiagnosisNames();

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  const orderedNames = useMemo(() => {
    const base = rawNames ?? [];
    if (overrideOrder.length === 0) return base;
    const idx = new Map<string, number>(overrideOrder.map((id, i) => [id, i]));
    return [...base].sort((a, b) => (idx.get(a.id) ?? 0) - (idx.get(b.id) ?? 0));
  }, [rawNames, overrideOrder]);

  const filteredItems = useMemo(() => {
    if (!searchTerm) return orderedNames;
    const lower = searchTerm.toLowerCase();
    return orderedNames.filter((n) => n.name.toLowerCase().includes(lower));
  }, [orderedNames, searchTerm]);

  const categoryMap = useMemo(
    () => new Map<string, string>((rawCategories ?? []).map((c) => [c.id, c.name])),
    [rawCategories],
  );

  const handleNameDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const currentIds = orderedNames.map((i) => i.id);
    const newOrder = arrayMove(
      currentIds,
      currentIds.indexOf(String(active.id)),
      currentIds.indexOf(String(over.id)),
    );
    setOverrideOrder(newOrder);
    reorderMutation.mutate(
      { ids: newOrder.map(Number) },
      { onSuccess: () => setOverrideOrder([]) },
    );
  };

  const handleEdit = (item: DiagnosisName) => {
    setSelectedItem(item);
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setIsEditing(true);
  };

  const handleClose = () => {
    setIsEditing(false);
    setSelectedItem(null);
  };

  const handleSave = (data: {
    name: string;
    diagnosisCategoryId: string;
    description: string;
    isActive: boolean;
  }) => {
    if (!data.name.trim()) {
      toast.error("診断病名は必須です");
      return;
    }
    if (!data.diagnosisCategoryId) {
      toast.error("カテゴリは必須です");
      return;
    }

    if (selectedItem) {
      const req: UpdateDiagnosisNameRequest = {
        name: data.name,
        diagnosis_category_id: Number(data.diagnosisCategoryId),
        description: data.description || undefined,
        is_active: data.isActive,
      };
      updateMutation.mutate(
        { id: selectedItem.id, req },
        {
          onSuccess: () => { toast.success("更新しました"); handleClose(); },
          onError: () => toast.error("更新に失敗しました"),
        },
      );
    } else {
      const req: CreateDiagnosisNameRequest = {
        name: data.name,
        diagnosis_category_id: Number(data.diagnosisCategoryId),
        description: data.description || undefined,
        is_active: true,
      };
      createMutation.mutate(req, {
        onSuccess: () => { toast.success("登録しました"); handleClose(); },
        onError: () => toast.error("登録に失敗しました"),
      });
    }
  };

  const handleDeleteConfirm = () => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        handleClose();
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  };

  return (
    <>
      <div className="flex h-full">
        {/* Table area */}
        <div className="flex flex-col gap-4 flex-1 min-w-0">
          <div className="flex items-center gap-3">
            <div className="flex-1 min-w-0">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="診断病名で検索..."
                count={filteredItems.length}
              />
            </div>
            <PrimaryButton onClick={handleCreate}>
              <Plus className="mr-1.5 size-4" />
              新規登録
            </PrimaryButton>
          </div>

          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={handleNameDragEnd}
          >
            <SortableContext
              items={filteredItems.map((i) => i.id)}
              strategy={verticalListSortingStrategy}
            >
              <DataTable
                columns={NAME_COLUMNS}
                data={filteredItems}
                emptyMessage="診断病名が登録されていません"
                renderRow={(item) => (
                  <SortableNameRow
                    key={item.id}
                    item={item}
                    categoryMap={categoryMap}
                    onEdit={() => handleEdit(item)}
                  />
                )}
              />
            </SortableContext>
          </DndContext>
        </div>

        {/* Side peek */}
        {isEditing ? (
          <DiagnosisNameSidePanel
            key={selectedItem ? String(selectedItem.id) : "new-name"}
            item={selectedItem}
            categories={rawCategories ?? []}
            onClose={handleClose}
            onSave={handleSave}
            onDeleteRequest={() => setPendingDelete(selectedItem)}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        title="診断病名を削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}

// ─────────────────────────────────────────────────
// DiagnosisSettings (main page)
// ─────────────────────────────────────────────────

export function DiagnosisSettings() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") ?? "diagnosis_category";

  const handleTabChange = (tab: string) => {
    setSearchParams({ tab });
  };

  return (
    <PageLayout
      title="診断マスタ"
      icon={<ClipboardList className="size-5 text-[#37352F]" />}
      onBack={() => navigate("/settings")}
      maxWidth="max-w-full"
    >
      <TabsPrimitive.Root
        value={activeTab}
        onValueChange={handleTabChange}
        className="flex flex-col gap-4"
      >
        <TabsPrimitive.List
          className={`flex h-9 border-b ${C.borderLight} gap-0`}
        >
          {TABS.map((tab) => (
            <TabsPrimitive.Trigger
              key={tab.value}
              value={tab.value}
              className={`h-9 border-b-2 border-b-transparent px-4 text-sm ${C.text60} outline-none transition-colors cursor-pointer
                data-[state=active]:border-b-[#37352F] data-[state=active]:text-[#37352F] data-[state=active]:font-medium`}
            >
              {tab.label}
            </TabsPrimitive.Trigger>
          ))}
        </TabsPrimitive.List>
        <TabsPrimitive.Content value="diagnosis_category" className="mt-4">
          <DiagnosisCategoryTab />
        </TabsPrimitive.Content>
        <TabsPrimitive.Content value="diagnosis_name" className="mt-4">
          <DiagnosisNameTab />
        </TabsPrimitive.Content>
      </TabsPrimitive.Root>
    </PageLayout>
  );
}
