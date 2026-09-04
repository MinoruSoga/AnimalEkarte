import { useCallback, useDeferredValue, useMemo, useState } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import {
  DataTable,
  DESIGN_TABLE_HEADER_ROW,
  DESIGN_TABLE_HEADER_CELL,
} from "@/components/shared/DataTable/DataTable";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { useSortableList } from "@/hooks/use-sortable-list";
import type { TreatmentItem } from "@/lib/transforms/treatment";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { TreatmentPlanRow } from "./TreatmentPlanRows";
import {
  TREATMENT_COLUMNS,
  type TreatmentTabConfig,
  buildTreatmentRows,
  buildTreatmentTree,
  filterTreatmentRoots,
} from "../lib/treatment-plan-tab-content-model";

interface TreatmentPlanTabContentProps extends TreatmentTabConfig {
  onEditTargetChange: (value: TreatmentItem | "new" | null) => void;
  canEdit: boolean;
}

export function TreatmentPlanTabContent({
  data: rawData,
  emptyMessage,
  searchPlaceholder,
  onReorder,
  onEditTargetChange,
  canEdit,
}: TreatmentPlanTabContentProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());

  const treeItems = useMemo(() => buildTreatmentTree(rawData ?? []), [rawData]);

  const {
    orderedItems: orderedRoots,
    sensors,
    handleDragEnd,
  } = useSortableList({
    items: treeItems,
    onReorder: (newIds) => {
      if (!canEdit) return;
      onReorder(newIds.map(Number));
    },
  });

  const filteredRoots = useMemo(
    () =>
      filterTreatmentRoots({
        roots: orderedRoots,
        activeFilters,
        searchTerm: deferredSearch,
      }),
    [orderedRoots, activeFilters, deferredSearch],
  );

  const toggleExpanded = useCallback((id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const flatRows = useMemo(
    () => buildTreatmentRows({ roots: filteredRoots, expandedIds }),
    [filteredRoots, expandedIds],
  );

  const totalCount = (rawData ?? []).length;

  const handleEdit = useCallback(
    (item: TreatmentItem) => {
      onEditTargetChange(item);
    },
    [onEditTargetChange],
  );

  return (
    <div className="flex flex-col gap-4">
      <PropertyFilter
        properties={[MASTER_STATUS_FILTER]}
        activeFilters={activeFilters}
        onFilterChange={setActiveFilters}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder={searchPlaceholder}
        count={totalCount}
      />

      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <SortableContext
          items={filteredRoots.map((root) => root.id)}
          strategy={verticalListSortingStrategy}
        >
          <DataTable
            headerRowClassName={DESIGN_TABLE_HEADER_ROW}
            headerCellClassName={DESIGN_TABLE_HEADER_CELL}
            columns={TREATMENT_COLUMNS}
            data={flatRows}
            emptyMessage={emptyMessage}
            renderRow={(row) => (
              <TreatmentPlanRow
                key={row.item.id}
                row={row}
                canEdit={canEdit}
                onEdit={handleEdit}
                onToggleExpanded={toggleExpanded}
              />
            )}
          />
        </SortableContext>
      </DndContext>
    </div>
  );
}
