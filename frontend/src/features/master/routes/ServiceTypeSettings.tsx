// React/Framework
import { useState, useMemo, useCallback, memo, useDeferredValue, useTransition, useRef, useEffect } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

// DnD
import {
  DndContext,
  closestCenter,
} from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";

// Shared hooks
import { useSortableList } from "@/hooks/useSortableList";

// External
import {
  Plus,
  Activity,
} from "lucide-react";
import { toast } from "sonner";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar/SearchFilterBar";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { PropInput } from "@/components/shared/SidePeek/PropInput";
import { SidePeekPanel } from "@/components/shared/SidePeek/SidePeekPanel";
import { SidePeekToolbar } from "@/components/shared/SidePeek/SidePeekToolbar";
import { SidePeekBody } from "@/components/shared/SidePeek/SidePeekBody";
import { SidePeekTitleInput } from "@/components/shared/SidePeek/SidePeekTitleInput";
import { SidePeekFooter } from "@/components/shared/SidePeek/SidePeekFooter";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import {
  useGetServiceTypes,
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

interface ServiceTypeSidePanelProps {
  item: ServiceType | null;
  onClose: () => void;
  onSave: (data: ServiceTypeFormData) => void;
  onDeleteRequest: (item: ServiceType) => void;
}

const ServiceTypeSidePanel = memo(function ServiceTypeSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: ServiceTypeSidePanelProps) {
  const [formData, setFormData] = useState<ServiceTypeFormData>(() => ({
    name: item?.name ?? "",
    description: item?.description ?? "",
    color: item?.color ?? "#3B82F6",
    isActive: item?.isActive ?? true,
  }));

  return (
    <SidePeekPanel>
      <SidePeekToolbar
        isNew={item === null}
        onClose={onClose}
        onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      />
      <SidePeekBody>
        <div className="pt-4 pb-2">
          <div className={STYLE.pageIcon}>
            <Activity className={LAYOUT.pageIcon.innerIcon} />
          </div>
        </div>
        <SidePeekTitleInput
          value={formData.name}
          onChange={(v) => setFormData((prev) => ({ ...prev, name: v }))}
        />
        <div className={`${STYLE.sectionDivider} mb-1`} />
        <div className="py-1">
          <StatusToggleButton
            isActive={formData.isActive}
            onToggle={() => setFormData((prev) => ({ ...prev, isActive: !prev.isActive }))}
          />
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
      </SidePeekBody>
      <SidePeekFooter onCancel={onClose} onSave={() => onSave(formData)} />
    </SidePeekPanel>
  );
});

// ─────────────────────────────────────────────────
// ServiceTypeSettings (main page)
// ─────────────────────────────────────────────────

export function ServiceTypeSettings() {
  const navigate = useNavigate();

  // null=closed, "new"=create mode, ServiceType=edit mode
  const [editTarget, setEditTarget] = useState<ServiceType | "new" | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<ServiceType | null>(null);
  const [, startSaveTransition] = useTransition();

  const { data: rawData } = useGetServiceTypes();
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

  // rerender-transitions: 検索フィルタを低優先度に遅延
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    if (!deferredSearch) return orderedItems;
    const lower = deferredSearch.toLowerCase();
    return orderedItems.filter((i) => i.name.toLowerCase().includes(lower));
  }, [orderedItems, deferredSearch]);

  const handleClose = useCallback(() => setEditTarget(null), []);

  const handleSave = useCallback(
    (data: ServiceTypeFormData) => {
      if (!data.name.trim()) {
        toast.error("名称は必須です");
        return;
      }

      // rerender-transitions: API書き込みを非緊急マーク
      startSaveTransition(() => {
        if (editTarget !== null && editTarget !== "new") {
          const req: UpdateServiceTypeRequest = {
            name: data.name,
            description: data.description || undefined,
            color: data.color || undefined,
            is_active: data.isActive,
          };
          updateMutation.mutate(
            { id: editTarget.id, req },
            {
              onSuccess: () => {
                toast.success("更新しました");
                handleClose();
              },
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
            onSuccess: () => {
              toast.success("登録しました");
              handleClose();
            },
            onError: () => toast.error("登録に失敗しました"),
          });
        }
      });
    },
    [editTarget, updateMutation, createMutation, handleClose],
  );

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

  const panelItem = editTarget !== null && editTarget !== "new" ? editTarget : null;

  return (
    <PageLayout
      title="予約区分マスタ"
      icon={<Activity className="size-5 text-[#37352F]" />}
      onBack={() => navigate(paths.settings.getHref())}
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
              onClick={() => setEditTarget("new")}
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
                  <SortableDataTableRow
                    key={item.id}
                    id={item.id}
                    onClick={() => setEditTarget(item)}
                  >
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
                    <TableCell className="p-0 text-right">
                      <RowActionButton onClick={() => setEditTarget(item)} />
                    </TableCell>
                  </SortableDataTableRow>
                )}
              />
            </SortableContext>
          </DndContext>
        </div>

        {/* Side peek */}
        {editTarget !== null ? (
          <ServiceTypeSidePanel
            key={panelItem ? String(panelItem.id) : "new-service-type"}
            item={panelItem}
            onClose={handleClose}
            onSave={handleSave}
            onDeleteRequest={setPendingDelete}
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
