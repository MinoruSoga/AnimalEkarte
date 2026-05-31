import type { ComponentProps } from "react";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { MedicineTable } from "./MedicineTable";

interface MedicineTableSectionProps {
  searchTerm: string;
  onSearchChange: (term: string) => void;
  activeFilters: ActiveFilter[];
  onFilterChange: (filters: ActiveFilter[]) => void;
  totalCount: number;
  tableProps: ComponentProps<typeof MedicineTable>;
}

export function MedicineTableSection({
  searchTerm,
  onSearchChange,
  activeFilters,
  onFilterChange,
  totalCount,
  tableProps,
}: MedicineTableSectionProps) {
  return (
    <div className="flex flex-col gap-4">
      <NotionFilter
        properties={[MASTER_STATUS_FILTER]}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchTerm}
        onSearchChange={onSearchChange}
        searchPlaceholder="薬品名で検索..."
        count={totalCount}
      />
      <MedicineTable {...tableProps} />
    </div>
  );
}
