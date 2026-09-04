import { useDeferredValue, useMemo, useState } from "react";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { useSortableList } from "@/hooks/use-sortable-list";
import {
  useGetDiagnosisNames,
  useGetDiagnosisTypes,
  useReorderDiagnosisNames,
  useReorderDiagnosisTypes,
  type DiagnosisName,
  type DiagnosisType,
} from "../api/diagnosis";
import { DiagnosisNameRow, DiagnosisTypeRow } from "./DiagnosisTabRows";
import {
  DIAGNOSIS_NAME_COLUMNS,
  DIAGNOSIS_TYPE_COLUMNS,
  buildDiagnosisTypeNameMap,
  filterDiagnosisNamesBySearch,
  filterDiagnosisTypesBySearch,
} from "../lib/diagnosis-tabs-model";
import { DiagnosisSortableTable } from "./DiagnosisSortableTable";

interface DiagnosisTypeTabProps {
  onEditTargetChange: (value: DiagnosisType | "new" | null) => void;
  canEdit: boolean;
}

export function DiagnosisTypeTab({ onEditTargetChange, canEdit }: DiagnosisTypeTabProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const { data: rawCategories } = useGetDiagnosisTypes();
  const reorderMutation = useReorderDiagnosisTypes();

  const {
    orderedItems: orderedCategories,
    sensors,
    handleDragEnd: handleCategoryDragEnd,
  } = useSortableList({
    items: rawCategories ?? [],
    onReorder: (newIds) => {
      if (!canEdit) return;
      reorderMutation.mutate({ ids: newIds.map(Number) });
    },
  });

  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    return filterDiagnosisTypesBySearch(orderedCategories, deferredSearch);
  }, [orderedCategories, deferredSearch]);

  return (
    <div className="flex flex-col gap-4">
      <PropertyFilter
        properties={[]}
        activeFilters={[]}
        onFilterChange={() => {}}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder="カテゴリ名で検索..."
        count={filteredItems.length}
      />

      <DiagnosisSortableTable
        items={filteredItems}
        columns={DIAGNOSIS_TYPE_COLUMNS}
        emptyMessage="診断カテゴリが登録されていません"
        sensors={sensors}
        onDragEnd={handleCategoryDragEnd}
        renderRow={(item) => (
          <DiagnosisTypeRow
            key={item.id}
            item={item}
            canEdit={canEdit}
            onEdit={onEditTargetChange}
          />
        )}
      />
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

  const {
    orderedItems: orderedNames,
    sensors,
    handleDragEnd: handleNameDragEnd,
  } = useSortableList({
    items: rawNames ?? [],
    onReorder: (newIds) => {
      if (!canEdit) return;
      reorderMutation.mutate({ ids: newIds.map(Number) });
    },
  });

  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    return filterDiagnosisNamesBySearch(orderedNames, deferredSearch);
  }, [orderedNames, deferredSearch]);

  const categoryMap = useMemo(() => buildDiagnosisTypeNameMap(rawCategories), [rawCategories]);

  return (
    <div className="flex flex-col gap-4">
      <PropertyFilter
        properties={[]}
        activeFilters={[]}
        onFilterChange={() => {}}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder="診断病名で検索..."
        count={filteredItems.length}
      />

      <DiagnosisSortableTable
        items={filteredItems}
        columns={DIAGNOSIS_NAME_COLUMNS}
        emptyMessage="診断病名が登録されていません"
        sensors={sensors}
        onDragEnd={handleNameDragEnd}
        renderRow={(item) => (
          <DiagnosisNameRow
            key={item.id}
            item={item}
            categoryMap={categoryMap}
            canEdit={canEdit}
            onEdit={onEditTargetChange}
          />
        )}
      />
    </div>
  );
}
