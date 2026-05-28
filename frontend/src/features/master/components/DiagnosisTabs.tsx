import { useDeferredValue, useMemo, useState } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { TableCell } from "@/components/ui/table";
import { C } from "@/lib/design-tokens";
import { useSortableList } from "@/hooks/use-sortable-list";
import {
  useGetDiagnosisNames,
  useGetDiagnosisTypes,
  useReorderDiagnosisNames,
  useReorderDiagnosisTypes,
  type DiagnosisName,
  type DiagnosisType,
} from "../api/diagnosis";

const CATEGORY_COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "カテゴリ名" },
  { header: "備考", className: "w-[240px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

const NAME_COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "所属カテゴリ", className: "w-[160px]" },
  { header: "診断病名" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface DiagnosisTypeTabProps {
  onEditTargetChange: (value: DiagnosisType | "new" | null) => void;
  canEdit: boolean;
}

export function DiagnosisTypeTab({ onEditTargetChange, canEdit }: DiagnosisTypeTabProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const { data: rawCategories } = useGetDiagnosisTypes();
  const reorderMutation = useReorderDiagnosisTypes();

  const { orderedItems: orderedCategories, sensors, handleDragEnd: handleCategoryDragEnd } =
    useSortableList({
      items: rawCategories ?? [],
      onReorder: (newIds) => {
        reorderMutation.mutate({ ids: newIds.map(Number) });
      },
    });

  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    if (!deferredSearch) return orderedCategories;
    const lower = deferredSearch.toLowerCase();
    return orderedCategories.filter((category) => category.name.toLowerCase().includes(lower));
  }, [orderedCategories, deferredSearch]);

  return (
    <div className="flex flex-col gap-4">
      <NotionFilter
        properties={[]}
        activeFilters={[]}
        onFilterChange={() => {}}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder="カテゴリ名で検索..."
        count={filteredItems.length}
      />

      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleCategoryDragEnd}
      >
        <SortableContext
          items={filteredItems.map((item) => item.id)}
          strategy={verticalListSortingStrategy}
        >
          <DataTable
            columns={CATEGORY_COLUMNS}
            data={filteredItems}
            emptyMessage="診断カテゴリが登録されていません"
            renderRow={(item) => (
              <SortableDataTableRow
                key={item.id}
                id={item.id}
                onClick={canEdit ? () => onEditTargetChange(item) : undefined}
              >
                <TableCell className={`font-medium text-base ${C.text}`}>
                  {item.name}
                </TableCell>
                <TableCell className={`text-base ${C.text70} truncate max-w-[240px]`}>
                  {item.description || "-"}
                </TableCell>
                <TableCell className="text-center">
                  <NotionStatusPill isActive={item.isActive} />
                </TableCell>
                <TableCell className="p-0 text-right">
                  {canEdit ? <RowActionButton onClick={() => onEditTargetChange(item)} /> : null}
                </TableCell>
              </SortableDataTableRow>
            )}
          />
        </SortableContext>
      </DndContext>
    </div>
  );
}

interface DiagnosisNameTabProps {
  onEditTargetChange: (value: DiagnosisName | "new" | null) => void;
  canEdit: boolean;
}

export function DiagnosisNameTab({ onEditTargetChange, canEdit }: DiagnosisNameTabProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const { data: rawCategories } = useGetDiagnosisTypes();
  const { data: rawNames } = useGetDiagnosisNames();
  const reorderMutation = useReorderDiagnosisNames();

  const { orderedItems: orderedNames, sensors, handleDragEnd: handleNameDragEnd } =
    useSortableList({
      items: rawNames ?? [],
      onReorder: (newIds) => {
        reorderMutation.mutate({ ids: newIds.map(Number) });
      },
    });

  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    if (!deferredSearch) return orderedNames;
    const lower = deferredSearch.toLowerCase();
    return orderedNames.filter((name) => name.name.toLowerCase().includes(lower));
  }, [orderedNames, deferredSearch]);

  const categoryMap = useMemo(
    () => new Map<string, string>((rawCategories ?? []).map((category) => [category.id, category.name])),
    [rawCategories],
  );

  return (
    <div className="flex flex-col gap-4">
      <NotionFilter
        properties={[]}
        activeFilters={[]}
        onFilterChange={() => {}}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder="診断病名で検索..."
        count={filteredItems.length}
      />

      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleNameDragEnd}
      >
        <SortableContext
          items={filteredItems.map((item) => item.id)}
          strategy={verticalListSortingStrategy}
        >
          <DataTable
            columns={NAME_COLUMNS}
            data={filteredItems}
            emptyMessage="診断病名が登録されていません"
            renderRow={(item) => (
              <SortableDataTableRow
                key={item.id}
                id={item.id}
                onClick={canEdit ? () => onEditTargetChange(item) : undefined}
              >
                <TableCell className={`text-base ${C.text70}`}>
                  {categoryMap.get(item.diagnosisTypeId) ?? "-"}
                </TableCell>
                <TableCell className={`font-medium text-base ${C.text}`}>
                  {item.name}
                </TableCell>
                <TableCell className="text-center">
                  <NotionStatusPill isActive={item.isActive} />
                </TableCell>
                <TableCell className="p-0 text-right">
                  {canEdit ? <RowActionButton onClick={() => onEditTargetChange(item)} /> : null}
                </TableCell>
              </SortableDataTableRow>
            )}
          />
        </SortableContext>
      </DndContext>
    </div>
  );
}
