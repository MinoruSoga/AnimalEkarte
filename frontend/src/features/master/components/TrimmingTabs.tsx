import { useDeferredValue, useMemo, useState } from "react";
import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { C } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import {
  TARGET_SIZE_LABELS,
  useGetTrimmingCourses,
  useGetTrimmingOptions,
  type TrimmingCourse,
  type TrimmingOption,
} from "../api/trimming";

const COURSE_COLUMNS = [
  { header: "コース名" },
  { header: "対象サイズ", className: "w-[120px]" },
  { header: "所要時間", className: "w-[100px]" },
  { header: "単価(税込)", className: "w-[110px]", align: "right" as const },
  { header: "ステータス", className: "w-[90px]", align: "right" as const },
];

const OPTION_COLUMNS = [
  { header: "オプション名" },
  { header: "所要時間", className: "w-[100px]" },
  { header: "組合せ可否", className: "w-[110px]", align: "center" as const },
  { header: "単価(税込)", className: "w-[110px]", align: "right" as const },
  { header: "ステータス", className: "w-[90px]", align: "right" as const },
];

export function CombinablePill({ combinable }: { combinable: boolean }) {
  if (combinable) {
    return (
      <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-sm text-base ${C.bgStatusGreen} ${C.textStatusGreen}`}>
        可
      </span>
    );
  }
  return (
    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-sm text-base ${C.bgInactive} ${C.text60}`}>
      不可
    </span>
  );
}

interface TrimmingCourseTabProps {
  onEditTargetChange: (value: TrimmingCourse | "new" | null) => void;
  canEdit: boolean;
}

export function TrimmingCourseTab({ onEditTargetChange, canEdit }: TrimmingCourseTabProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  const { data: rawCourses } = useGetTrimmingCourses();
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    let items = rawCourses ?? [];
    for (const filter of activeFilters) {
      if (filter.key === "status" && typeof filter.value === "string") {
        const want = filter.value === "active";
        items = items.filter((course) =>
          filter.condition === "is" ? course.isActive === want : course.isActive !== want,
        );
      }
    }
    if (deferredSearch) {
      const lower = deferredSearch.toLowerCase();
      items = items.filter((course) => course.name.toLowerCase().includes(lower));
    }
    return items;
  }, [rawCourses, activeFilters, deferredSearch]);

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
        columns={COURSE_COLUMNS}
        data={filteredItems}
        emptyMessage="トリミングコースが登録されていません"
        renderRow={(item) => (
          <DataTableRow key={item.id} onClick={canEdit ? () => onEditTargetChange(item) : undefined}>
            <TableCell className={`font-medium text-base ${C.text}`}>
              {item.name}
            </TableCell>
            <TableCell className={`text-base ${C.text70}`}>
              {item.targetSize ? TARGET_SIZE_LABELS[item.targetSize] : "-"}
            </TableCell>
            <TableCell className={`text-base ${C.text70}`}>
              {item.duration != null ? `${item.duration}分` : "-"}
            </TableCell>
            <TableCell className={`text-right font-mono text-base ${C.text}`}>
              {item.price != null ? `¥${item.price.toLocaleString()}` : "-"}
            </TableCell>
            <TableCell className="text-right">
              <NotionStatusPill isActive={item.isActive} />
            </TableCell>
          </DataTableRow>
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

  const filteredItems = useMemo(() => {
    let items = rawOptions ?? [];
    for (const filter of activeFilters) {
      if (filter.key === "status" && typeof filter.value === "string") {
        const want = filter.value === "active";
        items = items.filter((option) =>
          filter.condition === "is" ? option.isActive === want : option.isActive !== want,
        );
      }
    }
    if (deferredSearch) {
      const lower = deferredSearch.toLowerCase();
      items = items.filter((option) => option.name.toLowerCase().includes(lower));
    }
    return items;
  }, [rawOptions, activeFilters, deferredSearch]);

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
        columns={OPTION_COLUMNS}
        data={filteredItems}
        emptyMessage="トリミングオプションが登録されていません"
        renderRow={(item) => (
          <DataTableRow key={item.id} onClick={canEdit ? () => onEditTargetChange(item) : undefined}>
            <TableCell className={`font-medium text-base ${C.text}`}>
              {item.name}
            </TableCell>
            <TableCell className={`text-base ${C.text70}`}>
              {item.duration != null ? `${item.duration}分` : "-"}
            </TableCell>
            <TableCell className="text-center">
              <CombinablePill combinable={item.combinable} />
            </TableCell>
            <TableCell className={`text-right font-mono text-base ${C.text}`}>
              {item.price != null ? `¥${item.price.toLocaleString()}` : "-"}
            </TableCell>
            <TableCell className="text-right">
              <NotionStatusPill isActive={item.isActive} />
            </TableCell>
          </DataTableRow>
        )}
      />
    </div>
  );
}
