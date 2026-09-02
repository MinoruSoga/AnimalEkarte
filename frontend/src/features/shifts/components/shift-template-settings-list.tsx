import { DndContext, closestCenter, type DragEndEvent } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { useSortableList } from "@/hooks/use-sortable-list";

import { ShiftTemplateRow } from "./ShiftTemplateSettingsParts";
import { SHIFT_STATUS_FILTER, SHIFT_TEMPLATE_COLUMNS } from "./shift-template-table-model";
import type { ShiftTemplate } from "../types";

type SortableSensors = ReturnType<typeof useSortableList<ShiftTemplate>>["sensors"];

interface ShiftTemplateSettingsListProps {
  searchTerm: string;
  onSearchChange: (value: string) => void;
  activeFilters: ActiveFilter[];
  onFilterChange: (filters: ActiveFilter[]) => void;
  filteredItems: ShiftTemplate[];
  canEdit: boolean;
  sensors: SortableSensors;
  onDragEnd: (event: DragEndEvent) => void;
  onEdit: (item: ShiftTemplate) => void;
}

export function ShiftTemplateSettingsList({
  searchTerm,
  onSearchChange,
  activeFilters,
  onFilterChange,
  filteredItems,
  canEdit,
  sensors,
  onDragEnd,
  onEdit,
}: ShiftTemplateSettingsListProps) {
  return (
    <div className="flex flex-col gap-4">
      <PropertyFilter
        properties={[SHIFT_STATUS_FILTER]}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchTerm}
        onSearchChange={onSearchChange}
        searchPlaceholder="テンプレート名で検索..."
        count={filteredItems.length}
      />
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={onDragEnd}
      >
        <SortableContext
          items={filteredItems.map((item) => item.id)}
          strategy={verticalListSortingStrategy}
        >
          <DataTable
            headerRowClassName={DESIGN_TABLE_HEADER_ROW}
            headerCellClassName={DESIGN_TABLE_HEADER_CELL}
            columns={SHIFT_TEMPLATE_COLUMNS}
            data={filteredItems}
            emptyMessage="テンプレートがありません"
            renderRow={(item) => (
              <ShiftTemplateRow
                key={item.id}
                item={item}
                canEdit={canEdit}
                onEdit={() => onEdit(item)}
              />
            )}
          />
        </SortableContext>
      </DndContext>
    </div>
  );
}
