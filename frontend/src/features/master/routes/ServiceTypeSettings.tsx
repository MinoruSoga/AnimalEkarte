// React/Framework
import type { ReactNode } from "react";
import { useState, useMemo, useCallback, useRef, useEffect } from "react";
import { useNavigate } from "react-router";

// DnD
import {
  DndContext,
  closestCenter,
} from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
  useSortable,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";

// Shared hooks
import { useSortableList } from "@/hooks/useSortableList";

// External
import {
  Plus,
  Activity,
  X,
  GripVertical,
  Trash2,
} from "lucide-react";
import { toast } from "sonner";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import {
  useListServiceTypes,
  useCreateServiceType,
  useUpdateServiceType,
  useDeleteServiceType,
  useReorderServiceTypes,
} from "@/features/master/api/service-types";

// Types
import type { ServiceType } from "@/features/master/api/service-types";
import type {
  CreateServiceTypeRequest,
  UpdateServiceTypeRequest,
} from "@/types/service-type";

// ─────────────────────────────────────────────────
// Columns
// ─────────────────────────────────────────────────

const COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "名称" },
  { header: "備考", className: "w-[240px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

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
// フォームデータ型
// ─────────────────────────────────────────────────

interface ServiceTypeFormData {
  name: string;
  description: string;
  color: string;
  isActive: boolean;
}

// ─────────────────────────────────────────────────
// ServiceTypeSidePanel
// ─────────────────────────────────────────────────

function ServiceTypeSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: {
  item: ServiceType | null;
  onClose: () => void;
  onSave: (data: ServiceTypeFormData) => void;
  onDeleteRequest: (item: ServiceType) => void;
}) {
  const [formData, setFormData] = useState<ServiceTypeFormData>(() => ({
    name: item?.name ?? "",
    description: item?.description ?? "",
    color: item?.color ?? "#3B82F6",
    isActive: item?.isActive ?? true,
  }));

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
              onClick={() => onDeleteRequest(item)}
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
              <Activity className={LAYOUT.pageIcon.innerIcon} />
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
              onChange={(e) => setFormData((prev) => ({ ...prev, name: e.target.value }))}
              placeholder="無題"
              autoFocus
            />
          </div>
          <div className={`${STYLE.sectionDivider} mb-1`} />
          <div className="py-1">
            <PropertyRow label="ステータス">
              <button
                type="button"
                onClick={() => setFormData((prev) => ({ ...prev, isActive: !prev.isActive }))}
                className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
              >
                <NotionStatusPill isActive={formData.isActive} />
              </button>
            </PropertyRow>
            <PropertyRow label="カラー">
              <div className="flex items-center gap-2">
                <input
                  type="color"
                  value={formData.color}
                  onChange={(e) => setFormData((prev) => ({ ...prev, color: e.target.value }))}
                  className="w-7 h-7 rounded cursor-pointer border-0 bg-transparent p-0"
                />
                <PropInput
                  value={formData.color}
                  onChange={(v) => setFormData((prev) => ({ ...prev, color: v }))}
                  placeholder="#3B82F6"
                />
              </div>
            </PropertyRow>
            <PropertyRow label="備考">
              <PropInput
                value={formData.description}
                onChange={(v) => setFormData((prev) => ({ ...prev, description: v }))}
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
// SortableServiceTypeRow
// ─────────────────────────────────────────────────

function SortableServiceTypeRow({
  item,
  onEdit,
}: {
  item: ServiceType;
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
        className="w-[32px] text-[#37352F]/20 cursor-grab"
        {...listeners}
      >
        <GripVertical className="size-4" />
      </TableCell>
      <TableCell className={`font-medium text-sm ${C.text}`}>
        <div className="flex items-center gap-2">
          <span
            className="size-3 rounded-full shrink-0"
            style={{ backgroundColor: item.color }}
          />
          {item.name}
        </div>
      </TableCell>
      <TableCell className={`text-sm ${C.text70} truncate max-w-[240px]`}>
        {item.description || "-"}
      </TableCell>
      <TableCell className="text-center">
        <NotionStatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="text-right">
        <RowActionButton onClick={onEdit} />
      </TableCell>
    </DataTableRow>
  );
}

// ─────────────────────────────────────────────────
// ServiceTypeSettings (main page)
// ─────────────────────────────────────────────────

export function ServiceTypeSettings() {
  const navigate = useNavigate();

  const [selectedItem, setSelectedItem] = useState<ServiceType | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<ServiceType | null>(null);

  const { data: rawData } = useListServiceTypes();
  const createMutation = useCreateServiceType();
  const updateMutation = useUpdateServiceType();
  const deleteMutation = useDeleteServiceType();
  const reorderMutation = useReorderServiceTypes();

  // resetOrder は useSortableList が返す安定関数。ref 経由で onReorder に渡す。
  const resetOrderRef = useRef<() => void>(() => {});

  const handleReorder = useCallback(
    (newIds: string[]) => {
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        {
          onError: () => {
            resetOrderRef.current();
            toast.error("並び替えに失敗しました");
          },
        },
      );
    },
    [reorderMutation],
  );

  const { orderedItems, sensors, handleDragStart, handleDragEnd, handleDragCancel, resetOrder } =
    useSortableList({
      items: rawData ?? [],
      onReorder: handleReorder,
    });
  // resetOrder は useSortableList が返す安定関数（deps: []）だが effect 経由で同期する
  useEffect(() => {
    resetOrderRef.current = resetOrder;
  }, [resetOrder]);

  const filteredItems = useMemo(() => {
    if (!searchTerm) return orderedItems;
    const lower = searchTerm.toLowerCase();
    return orderedItems.filter((i) => i.name.toLowerCase().includes(lower));
  }, [orderedItems, searchTerm]);

  const handleEdit = useCallback((item: ServiceType) => {
    setSelectedItem(item);
    setIsEditing(true);
  }, []);

  const handleCreate = useCallback(() => {
    setSelectedItem(null);
    setIsEditing(true);
  }, []);

  const handleClose = useCallback(() => {
    setIsEditing(false);
    setSelectedItem(null);
  }, []);

  const handleSave = useCallback(
    (data: ServiceTypeFormData) => {
      if (!data.name.trim()) {
        toast.error("名称は必須です");
        return;
      }

      if (selectedItem) {
        const req: UpdateServiceTypeRequest = {
          name: data.name,
          description: data.description || undefined,
          color: data.color || undefined,
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
        const req: CreateServiceTypeRequest = {
          name: data.name,
          description: data.description || undefined,
          color: data.color || undefined,
          is_active: true,
        };
        createMutation.mutate(req, {
          onSuccess: () => { toast.success("登録しました"); handleClose(); },
          onError: () => toast.error("登録に失敗しました"),
        });
      }
    },
    [selectedItem, updateMutation, createMutation, handleClose],
  );

  const handleDeleteRequest = useCallback((item: ServiceType) => {
    setPendingDelete(item);
  }, []);

  const handleDeleteConfirm = useCallback(() => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        handleClose();
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  }, [pendingDelete, deleteMutation, handleClose]);

  return (
    <PageLayout
      title="予約区分マスタ"
      icon={<Activity className="size-5 text-[#37352F]" />}
      onBack={() => navigate("/settings")}
      maxWidth="max-w-full"
    >
      <div className="flex h-full">
        {/* Table area */}
        <div className="flex flex-col gap-4 flex-1 min-w-0">
          <div className="flex items-center gap-3">
            <div className="flex-1 min-w-0">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="予約区分名で検索..."
                count={filteredItems.length}
              />
            </div>
            <button
              type="button"
              onClick={handleCreate}
              className="inline-flex items-center gap-1 text-sm font-medium text-[#2383E2] hover:text-[#1B6EC2] cursor-pointer transition-colors"
            >
              <Plus className="size-4" />
              新規登録
            </button>
          </div>

          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
            onDragCancel={handleDragCancel}
          >
            <SortableContext
              items={filteredItems.map((i) => i.id)}
              strategy={verticalListSortingStrategy}
            >
              <DataTable
                columns={COLUMNS}
                data={filteredItems}
                emptyMessage="予約区分が登録されていません"
                renderRow={(item) => (
                  <SortableServiceTypeRow
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
          <ServiceTypeSidePanel
            key={selectedItem ? String(selectedItem.id) : "new-service-type"}
            item={selectedItem}
            onClose={handleClose}
            onSave={handleSave}
            onDeleteRequest={handleDeleteRequest}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        title="予約区分を削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </PageLayout>
  );
}
