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
import { Plus, ClipboardList, X, Trash2, GripVertical } from "lucide-react";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
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
import type {
  DiagnosisCategory,
  DiagnosisName,
  CreateDiagnosisCategoryRequest,
  UpdateDiagnosisCategoryRequest,
  CreateDiagnosisNameRequest,
  UpdateDiagnosisNameRequest,
} from "@/features/master/api/diagnosis";

// ─────────────────────────────────────────────────
// Columns
// ─────────────────────────────────────────────────

const CATEGORY_COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "カテゴリ名" },
  { header: "分類", className: "w-[200px]" },
  { header: "ステータス", className: "w-[90px]", align: "right" as const },
];

const NAME_COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "診断病名" },
  { header: "所属カテゴリ", className: "w-[200px]" },
  { header: "分類", className: "w-[200px]" },
  { header: "ステータス", className: "w-[90px]", align: "right" as const },
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
// Link-style 新規登録 button
// ─────────────────────────────────────────────────

function AddButton({ onClick }: { onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="inline-flex items-center gap-1 text-sm font-medium text-[#2383E2] hover:text-[#1a6bc0] transition-colors cursor-pointer"
    >
      <Plus className="size-4" />
      新規登録
    </button>
  );
}

// ─────────────────────────────────────────────────
// DiagnosisCategoryTab
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
      <TableCell className={`text-sm ${C.text60} py-2.5`}>
        diagnosis_category
      </TableCell>
      <TableCell className="text-right py-2.5">
        <span className="inline-flex items-center gap-1.5">
          <span
            className={`size-[7px] rounded-full ${item.isActive ? "bg-[#2383E2]" : "bg-[#37352F]/20"}`}
          />
          <span
            className={`text-sm ${item.isActive ? "text-[#37352F]/65" : "text-[#37352F]/35"}`}
          >
            {item.isActive ? "有効" : "無効"}
          </span>
        </span>
      </TableCell>
    </DataTableRow>
  );
}

function DiagnosisCategoryTab() {
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<DiagnosisCategory | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<DiagnosisCategory | null>(null);
  const [formData, setFormData] = useState({ name: "", description: "", isActive: true });
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
    const idx = new Map(overrideOrder.map((id, i) => [id, i]));
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
      { ids: newOrder.map((id) => Number(id)) },
      { onSuccess: () => setOverrideOrder([]) },
    );
  };

  const handleEdit = (item: DiagnosisCategory) => {
    setSelectedItem(item);
    setFormData({ name: item.name, description: item.description, isActive: item.isActive });
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setFormData({ name: "", description: "", isActive: true });
    setIsEditing(true);
  };

  const handleCloseEdit = () => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData({ name: "", description: "", isActive: true });
  };

  const handleSave = () => {
    if (!formData.name.trim()) {
      toast.error("カテゴリ名は必須です");
      return;
    }

    if (selectedItem) {
      const req: UpdateDiagnosisCategoryRequest = {
        name: formData.name,
        description: formData.description || undefined,
        is_active: formData.isActive,
      };
      updateMutation.mutate(
        { id: selectedItem.id, req },
        {
          onSuccess: () => {
            toast.success("更新しました");
            setIsEditing(false);
          },
          onError: () => toast.error("更新に失敗しました"),
        },
      );
    } else {
      const req: CreateDiagnosisCategoryRequest = {
        name: formData.name,
        description: formData.description || undefined,
        is_active: true,
      };
      createMutation.mutate(req, {
        onSuccess: () => {
          toast.success("登録しました");
          setIsEditing(false);
        },
        onError: () => toast.error("登録に失敗しました"),
      });
    }
  };

  const handleDeleteConfirm = () => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        setIsEditing(false);
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  };

  return (
    <>
      <div className="flex h-full">
        {/* ── Left: List ── */}
        <div className="flex-1 min-w-0 flex flex-col gap-4">
          <div className="flex items-center gap-3">
            <div className="flex-1 min-w-0">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="カテゴリ名で検索..."
                count={filteredItems.length}
              />
            </div>
            <AddButton onClick={handleCreate} />
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

        {/* ── Right: Side Peek ── */}
        {isEditing ? (
          <div
            className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0 flex flex-col`}
          >
            {/* Toolbar */}
            <div className={STYLE.sidePeekToolbar}>
              <span className="text-xs text-[#37352F]/35 pl-1 select-none">
                {selectedItem ? "編集" : "新規作成"}
              </span>
              <div className="flex items-center gap-1">
                {selectedItem ? (
                  <button
                    type="button"
                    onClick={() => setPendingDelete(selectedItem)}
                    className={`${STYLE.sidePeekToolbarBtn} cursor-pointer text-[#EB5757] hover:bg-[#EB5757]/10`}
                  >
                    <Trash2 className="size-4" />
                  </button>
                ) : null}
                <button
                  type="button"
                  onClick={handleCloseEdit}
                  className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
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
                    <ClipboardList className={LAYOUT.pageIcon.innerIcon} />
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
                  />
                </div>

                <div className={`${STYLE.sectionDivider} mb-1`} />

                <div className="py-1">
                  <PropertyRow label="ステータス">
                    <button
                      type="button"
                      onClick={() =>
                        setFormData({ ...formData, isActive: !formData.isActive })
                      }
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
              <button type="button" onClick={handleCloseEdit} className={STYLE.sidePeekCancelBtn}>
                キャンセル
              </button>
              <button type="button" onClick={handleSave} className={STYLE.sidePeekSaveBtn}>
                保存
              </button>
            </div>
          </div>
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
// DiagnosisNameTab
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
      <TableCell className={`text-sm ${C.text60} py-2.5`}>
        diagnosis_name
      </TableCell>
      <TableCell className="text-right py-2.5">
        <span className="inline-flex items-center gap-1.5">
          <span
            className={`size-[7px] rounded-full ${item.isActive ? "bg-[#2383E2]" : "bg-[#37352F]/20"}`}
          />
          <span
            className={`text-sm ${item.isActive ? "text-[#37352F]/65" : "text-[#37352F]/35"}`}
          >
            {item.isActive ? "有効" : "無効"}
          </span>
        </span>
      </TableCell>
    </DataTableRow>
  );
}

function DiagnosisNameTab() {
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<DiagnosisName | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<DiagnosisName | null>(null);
  const [formData, setFormData] = useState({
    name: "",
    diagnosisCategoryId: "",
    description: "",
    isActive: true,
  });

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

  const [overrideOrder, setOverrideOrder] = useState<string[]>([]);

  const orderedNames = useMemo(() => {
    const base = rawNames ?? [];
    if (overrideOrder.length === 0) return base;
    const idx = new Map(overrideOrder.map((id, i) => [id, i]));
    return [...base].sort((a, b) => (idx.get(a.id) ?? 0) - (idx.get(b.id) ?? 0));
  }, [rawNames, overrideOrder]);

  const filteredItems = useMemo(() => {
    if (!searchTerm) return orderedNames;
    const lower = searchTerm.toLowerCase();
    return orderedNames.filter((n) => n.name.toLowerCase().includes(lower));
  }, [orderedNames, searchTerm]);

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
      { ids: newOrder.map((id) => Number(id)) },
      { onSuccess: () => setOverrideOrder([]) },
    );
  };

  const categoryMap = useMemo(
    () => new Map((rawCategories ?? []).map((c) => [c.id, c.name])),
    [rawCategories],
  );

  const handleEdit = (item: DiagnosisName) => {
    setSelectedItem(item);
    setFormData({
      name: item.name,
      diagnosisCategoryId: item.diagnosisCategoryId,
      description: item.description,
      isActive: item.isActive,
    });
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    const firstCategoryId = rawCategories?.[0]?.id ?? "";
    setFormData({
      name: "",
      diagnosisCategoryId: firstCategoryId,
      description: "",
      isActive: true,
    });
    setIsEditing(true);
  };

  const handleCloseEdit = () => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData({ name: "", diagnosisCategoryId: "", description: "", isActive: true });
  };

  const handleSave = () => {
    if (!formData.name.trim()) {
      toast.error("診断病名は必須です");
      return;
    }
    if (!formData.diagnosisCategoryId) {
      toast.error("カテゴリは必須です");
      return;
    }

    if (selectedItem) {
      const req: UpdateDiagnosisNameRequest = {
        name: formData.name,
        diagnosis_category_id: formData.diagnosisCategoryId,
        description: formData.description || undefined,
        is_active: formData.isActive,
      };
      updateMutation.mutate(
        { id: selectedItem.id, req },
        {
          onSuccess: () => {
            toast.success("更新しました");
            setIsEditing(false);
          },
          onError: () => toast.error("更新に失敗しました"),
        },
      );
    } else {
      const req: CreateDiagnosisNameRequest = {
        name: formData.name,
        diagnosis_category_id: formData.diagnosisCategoryId,
        description: formData.description || undefined,
        is_active: true,
      };
      createMutation.mutate(req, {
        onSuccess: () => {
          toast.success("登録しました");
          setIsEditing(false);
        },
        onError: () => toast.error("登録に失敗しました"),
      });
    }
  };

  const handleDeleteConfirm = () => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        setIsEditing(false);
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  };

  return (
    <>
      <div className="flex h-full">
        {/* ── Left: List ── */}
        <div className="flex-1 min-w-0 flex flex-col gap-4">
          <div className="flex items-center gap-3">
            <div className="flex-1 min-w-0">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="診断病名で検索..."
                count={filteredItems.length}
              />
            </div>
            <AddButton onClick={handleCreate} />
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

        {/* ── Right: Side Peek ── */}
        {isEditing ? (
          <div
            className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0 flex flex-col`}
          >
            {/* Toolbar */}
            <div className={STYLE.sidePeekToolbar}>
              <span className="text-xs text-[#37352F]/35 pl-1 select-none">
                {selectedItem ? "編集" : "新規作成"}
              </span>
              <div className="flex items-center gap-1">
                {selectedItem ? (
                  <button
                    type="button"
                    onClick={() => setPendingDelete(selectedItem)}
                    className={`${STYLE.sidePeekToolbarBtn} cursor-pointer text-[#EB5757] hover:bg-[#EB5757]/10`}
                  >
                    <Trash2 className="size-4" />
                  </button>
                ) : null}
                <button
                  type="button"
                  onClick={handleCloseEdit}
                  className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
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
                    <ClipboardList className={LAYOUT.pageIcon.innerIcon} />
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
                  />
                </div>

                <div className={`${STYLE.sectionDivider} mb-1`} />

                <div className="py-1">
                  <PropertyRow label="ステータス">
                    <button
                      type="button"
                      onClick={() =>
                        setFormData({ ...formData, isActive: !formData.isActive })
                      }
                      className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
                    >
                      <NotionStatusPill isActive={formData.isActive} />
                    </button>
                  </PropertyRow>

                  <PropertyRow label="カテゴリ">
                    <Select
                      value={formData.diagnosisCategoryId}
                      onValueChange={(v) =>
                        setFormData({ ...formData, diagnosisCategoryId: v })
                      }
                    >
                      <SelectTrigger className={STYLE.selectCompact}>
                        <SelectValue placeholder="カテゴリを選択" />
                      </SelectTrigger>
                      <SelectContent>
                        {(rawCategories ?? []).map((cat) => (
                          <SelectItem key={cat.id} value={cat.id}>
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
              <button type="button" onClick={handleCloseEdit} className={STYLE.sidePeekCancelBtn}>
                キャンセル
              </button>
              <button type="button" onClick={handleSave} className={STYLE.sidePeekSaveBtn}>
                保存
              </button>
            </div>
          </div>
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
      title="診断病名マスタ"
      icon={<ClipboardList className="size-5 text-[#37352F]" />}
      onBack={() => navigate("/settings")}
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        <Tabs value={activeTab} onValueChange={handleTabChange}>
          <TabsList
            className={`h-9 bg-transparent border-b ${C.borderLight} rounded-none w-full justify-start gap-0 p-0`}
          >
            {TABS.map((tab) => (
              <TabsTrigger
                key={tab.value}
                value={tab.value}
                className={`h-9 rounded-none border-b-2 border-transparent px-4 text-sm ${C.text60}
                  data-[state=active]:border-[#37352F] data-[state=active]:${C.text}
                  data-[state=active]:shadow-none data-[state=active]:bg-transparent`}
              >
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>
          <TabsContent value="diagnosis_category" className="mt-4">
            <DiagnosisCategoryTab />
          </TabsContent>
          <TabsContent value="diagnosis_name" className="mt-4">
            <DiagnosisNameTab />
          </TabsContent>
        </Tabs>
      </div>
    </PageLayout>
  );
}
