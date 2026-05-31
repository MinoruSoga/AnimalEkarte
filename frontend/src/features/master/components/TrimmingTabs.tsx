import { useDeferredValue, useMemo, useState } from "react";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import {
  useGetTrimmingCourses,
  useGetTrimmingOptions,
  type TrimmingCourse,
  type TrimmingOption,
} from "../api/trimming";
import { TrimmingCourseRow, TrimmingOptionRow } from "./TrimmingTabRows";
import {
  TRIMMING_COURSE_COLUMNS,
  TRIMMING_OPTION_COLUMNS,
  filterTrimmingItems,
} from "./TrimmingTabsModel";

interface TrimmingCourseTabProps {
  onEditTargetChange: (value: TrimmingCourse | "new" | null) => void;
  canEdit: boolean;
}

export function TrimmingCourseTab({ onEditTargetChange, canEdit }: TrimmingCourseTabProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  const { data: rawCourses } = useGetTrimmingCourses();
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(
    () => filterTrimmingItems({ items: rawCourses, activeFilters, searchTerm: deferredSearch }),
    [rawCourses, activeFilters, deferredSearch],
  );

  return (
    <div className="flex flex-col gap-4">
      <NotionFilter
        properties={[MASTER_STATUS_FILTER]}
        activeFilters={activeFilters}
        onFilterChange={setActiveFilters}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder="コース名で検索..."
        count={filteredItems.length}
      />

      <DataTable
        columns={TRIMMING_COURSE_COLUMNS}
        data={filteredItems}
        emptyMessage="トリミングコースが登録されていません"
        renderRow={(item) => (
          <TrimmingCourseRow
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

interface TrimmingOptionTabProps {
  onEditTargetChange: (value: TrimmingOption | "new" | null) => void;
  canEdit: boolean;
}

export function TrimmingOptionTab({ onEditTargetChange, canEdit }: TrimmingOptionTabProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  const { data: rawOptions } = useGetTrimmingOptions();
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(
    () => filterTrimmingItems({ items: rawOptions, activeFilters, searchTerm: deferredSearch }),
    [rawOptions, activeFilters, deferredSearch],
  );

  return (
    <div className="flex flex-col gap-4">
      <NotionFilter
        properties={[MASTER_STATUS_FILTER]}
        activeFilters={activeFilters}
        onFilterChange={setActiveFilters}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder="オプション名で検索..."
        count={filteredItems.length}
      />

      <DataTable
        columns={TRIMMING_OPTION_COLUMNS}
        data={filteredItems}
        emptyMessage="トリミングオプションが登録されていません"
        renderRow={(item) => (
          <TrimmingOptionRow
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
