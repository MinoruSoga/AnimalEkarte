import { useCallback, useDeferredValue, useMemo, useState } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import ChevronDown from "lucide-react/dist/esm/icons/chevron-down";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import { toast } from "sonner";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { TableCell } from "@/components/ui/table";
import { handleApiError } from "@/lib/handle-api-error";
import { C, ICON } from "@/lib/design-tokens";
import { useSortableList } from "@/hooks/use-sortable-list";
import type { TreatmentItem } from "@/lib/transforms/treatment";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import type { TreatmentFormData } from "./TreatmentItemSidePanel";

type TreeItem = TreatmentItem & { children: TreatmentItem[] };

export type MutateCallbacks = {
  onSuccess: () => void;
  onError: (error?: unknown) => void;
};

export interface TreatmentTabConfig {
  data: TreatmentItem[] | undefined;
  entityLabel: string;
  emptyMessage: string;
  searchPlaceholder: string;
  onCreate?: (data: TreatmentFormData, callbacks: MutateCallbacks) => void;
  onUpdate?: (id: string, data: TreatmentFormData, callbacks: MutateCallbacks) => void;
  onDelete?: (id: string, callbacks: MutateCallbacks) => void;
  onReorder: (ids: number[]) => void;
}

type VirtualRow =
  | { type: "root"; item: TreeItem; isExpanded: boolean }
  | { type: "child"; item: TreatmentItem };

const TREATMENT_COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "名称" },
  { header: "単価(税込)", className: "w-[120px]", align: "right" as const },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

function buildTree(items: TreatmentItem[]): TreeItem[] {
  const map = new Map<string, TreeItem>(
    items.map((item) => [item.id, { ...item, children: [] }]),
  );
  const roots: TreeItem[] = [];
  for (const item of map.values()) {
    if (item.parentId) {
      map.get(item.parentId)?.children.push(item);
    } else {
      roots.push(item);
    }
  }
  return roots;
}

interface ChildTreatmentRowProps {
  item: TreatmentItem;
  onEdit: () => void;
  canEdit: boolean;
}

function ChildTreatmentRow({ item, onEdit, canEdit }: ChildTreatmentRowProps) {
  return (
    <DataTableRow onClick={onEdit}>
      <TableCell className="w-[32px]" />
      <TableCell>
        <div className="flex items-center gap-1 pl-[22px]">
          <span className="size-[22px] shrink-0" />
          <span className={`text-base ${C.text}`}>{item.name}</span>
        </div>
      </TableCell>
      <TableCell className="text-right">
        <span className={`text-base ${C.text70} font-mono`}>
          {item.price > 0 ? `¥${item.price.toLocaleString()}` : "-"}
        </span>
      </TableCell>
      <TableCell className="text-center">
        <NotionStatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="p-0 text-right">
        {canEdit ? <RowActionButton onClick={onEdit} /> : null}
      </TableCell>
    </DataTableRow>
  );
}

interface TreatmentPlanTabContentProps extends TreatmentTabConfig {
  onEditTargetChange: (value: TreatmentItem | "new" | null) => void;
  pendingDelete: TreatmentItem | null;
  onPendingDeleteChange: (item: TreatmentItem | null) => void;
  canEdit: boolean;
}

export function TreatmentPlanTabContent({
  data: rawData,
  entityLabel,
  emptyMessage,
  searchPlaceholder,
  onDelete,
  onReorder,
  onEditTargetChange,
  pendingDelete,
  onPendingDeleteChange,
  canEdit,
}: TreatmentPlanTabContentProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());

  const treeItems = useMemo(() => buildTree(rawData ?? []), [rawData]);

  const { orderedItems: orderedRoots, sensors, handleDragEnd } = useSortableList({
    items: treeItems,
    onReorder: (newIds) => {
      onReorder(newIds.map(Number));
    },
  });

  const filteredRoots = useMemo(() => {
    let items = orderedRoots;
    for (const filter of activeFilters) {
      if (filter.key === "status" && typeof filter.value === "string") {
        const want = filter.value === "active";
        items = items.filter((root) =>
          filter.condition === "is" ? root.isActive === want : root.isActive !== want,
        );
      }
    }
    if (deferredSearch) {
      const lower = deferredSearch.toLowerCase();
      items = items.filter(
        (root) =>
          root.name.toLowerCase().includes(lower) ||
          root.children.some((child) => child.name.toLowerCase().includes(lower)),
      );
    }
    return items;
  }, [orderedRoots, activeFilters, deferredSearch]);

  const toggleExpanded = useCallback((id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const flatRows = useMemo<VirtualRow[]>(
    () =>
      filteredRoots.flatMap((root) => {
        const isExpanded = expandedIds.has(root.id);
        const rows: VirtualRow[] = [{ type: "root", item: root, isExpanded }];
        if (isExpanded && root.children.length > 0) {
          for (const child of root.children) {
            rows.push({ type: "child", item: child });
          }
        }
        return rows;
      }),
    [filteredRoots, expandedIds],
  );

  const totalCount = (rawData ?? []).length;

  const handleEdit = useCallback((item: TreatmentItem) => {
    onEditTargetChange(item);
  }, [onEditTargetChange]);

  const handleClose = useCallback(() => {
    onEditTargetChange(null);
  }, [onEditTargetChange]);

  const handleDeleteConfirm = useCallback(() => {
    if (!pendingDelete || !onDelete) return;
    onDelete(pendingDelete.id, {
      onSuccess: () => {
        onPendingDeleteChange(null);
        handleClose();
        toast.success("削除しました");
      },
      onError: (error) => handleApiError(error, "削除"),
    });
  }, [pendingDelete, onDelete, handleClose, onPendingDeleteChange]);

  return (
    <>
      <div className="flex flex-col gap-4">
        <NotionFilter
          properties={[MASTER_STATUS_FILTER]}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder={searchPlaceholder}
          count={totalCount}
        />

        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragEnd={handleDragEnd}
        >
          <SortableContext
            items={filteredRoots.map((root) => root.id)}
            strategy={verticalListSortingStrategy}
          >
            <DataTable
              columns={TREATMENT_COLUMNS}
              data={flatRows}
              emptyMessage={emptyMessage}
              renderRow={(row) => {
                if (row.type === "root") {
                  return (
                    <SortableDataTableRow
                      key={row.item.id}
                      id={row.item.id}
                      onClick={() => handleEdit(row.item)}
                    >
                      <TableCell>
                        <div className="flex items-center gap-1">
                          {row.item.children.length > 0 ? (
                            <button
                              type="button"
                              onClick={(event) => {
                                event.stopPropagation();
                                toggleExpanded(row.item.id);
                              }}
                              className={`size-[22px] flex items-center justify-center rounded-[3px] ${C.text40} ${C.hoverBgMedium} transition-colors shrink-0`}
                            >
                              {row.isExpanded ? (
                                <ChevronDown className={ICON.xs} />
                              ) : (
                                <ChevronRight className={ICON.xs} />
                              )}
                            </button>
                          ) : (
                            <span className="size-[22px] shrink-0" />
                          )}
                          <span className={`text-base font-medium ${C.text}`}>{row.item.name}</span>
                          {row.item.children.length > 0 ? (
                            <span className={`text-base ${C.text25} ml-0.5`}>{row.item.children.length}</span>
                          ) : null}
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className={`text-base ${C.text70} font-mono`}>
                          {row.item.price > 0 ? `¥${row.item.price.toLocaleString()}` : "-"}
                        </span>
                      </TableCell>
                      <TableCell className="text-center">
                        <NotionStatusPill isActive={row.item.isActive} />
                      </TableCell>
                      <TableCell className="p-0 text-right">
                        {canEdit ? <RowActionButton onClick={() => handleEdit(row.item)} /> : null}
                      </TableCell>
                    </SortableDataTableRow>
                  );
                }
                return (
                  <ChildTreatmentRow
                    key={row.item.id}
                    item={row.item}
                    onEdit={() => handleEdit(row.item)}
                    canEdit={canEdit}
                  />
                );
              }}
            />
          </SortableContext>
        </DndContext>
      </div>

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => onPendingDeleteChange(null)}
        title={`${entityLabel}を削除しますか？`}
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}
