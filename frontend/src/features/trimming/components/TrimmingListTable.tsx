import { AlertTriangle, Edit, Trash2 } from "lucide-react";
import { memo } from "react";
import { TableCell } from "@/components/ui/table";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import type {
  ActiveFilter,
  ActiveSort,
  FilterProperty,
  SortProperty,
} from "@/components/shared/PropertyFilter/types";
import { Pagination } from "@/components/shared/Pagination";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { C, ICON } from "@/lib/design-tokens";
import type { TrimmingUI } from "@/types";
import { getTrimmingStatusColor } from "@/lib/status-helpers";
import { paths } from "@/config/paths";

// フィルタ定義 (TRIMMING_STATIC_FILTER_PROPERTIES / buildTrimmingDynamicFilterProperties) は
// TrimmingListTableModel.ts に分離済み (react-refresh/only-export-components)。

const TRIMMING_SORT_PROPERTIES: SortProperty[] = [
  { key: "date", label: "診療日" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "species", label: "種" },
  { key: "staff", label: "担当" },
  { key: "status", label: "ステータス" },
];

interface TrimmingListTableProps {
  records: TrimmingUI[];
  filteredCount: number;
  currentPage: number;
  totalPages: number;
  startIndex: number;
  endIndex: number;
  searchKeyword: string;
  activeFilters: ActiveFilter[];
  activeSorts: ActiveSort[];
  filterProperties: FilterProperty[];
  isFiltering: boolean;
  canEdit: boolean;
  canDelete: boolean;
  isValidStaff: (staff: string) => boolean;
  directionFor: (key: string) => "ascending" | "descending" | "none";
  onSearchChange: (value: string) => void;
  onFilterChange: (filters: ActiveFilter[]) => void;
  onSortChange: (sorts: ActiveSort[]) => void;
  onToggleSort: (key: string) => void;
  onEdit: (id: string) => void;
  onDeleteClick: (record: TrimmingUI) => void;
  onPageChange: (page: number) => void;
}

export function TrimmingListTable({
  records,
  filteredCount,
  currentPage,
  totalPages,
  startIndex,
  endIndex,
  searchKeyword,
  activeFilters,
  activeSorts,
  filterProperties,
  isFiltering,
  canEdit,
  canDelete,
  isValidStaff,
  directionFor,
  onSearchChange,
  onFilterChange,
  onSortChange,
  onToggleSort,
  onEdit,
  onDeleteClick,
  onPageChange,
}: TrimmingListTableProps) {
  const columns = [
    {
      header: (
        <SortableHeader
          label="診療日"
          direction={directionFor("date")}
          onToggle={() => onToggleSort("date")}
        />
      ),
      className: "w-[120px]",
    },
    {
      header: (
        <SortableHeader
          label="飼主名"
          direction={directionFor("ownerName")}
          onToggle={() => onToggleSort("ownerName")}
        />
      ),
    },
    {
      header: (
        <SortableHeader
          label="ペット名"
          direction={directionFor("petName")}
          onToggle={() => onToggleSort("petName")}
        />
      ),
    },
    {
      header: (
        <SortableHeader
          label="種"
          direction={directionFor("species")}
          onToggle={() => onToggleSort("species")}
        />
      ),
      className: "w-[80px] hidden lg:table-cell",
    },
    { header: "犬種", className: "w-[100px] hidden lg:table-cell" },
    { header: "体重", className: "w-[80px] hidden lg:table-cell" },
    { header: "スタイル希望", className: "hidden lg:table-cell" },
    {
      header: (
        <SortableHeader
          label="担当"
          direction={directionFor("staff")}
          onToggle={() => onToggleSort("staff")}
        />
      ),
      className: "w-[100px]",
    },
    {
      header: (
        <SortableHeader
          label="ステータス"
          direction={directionFor("status")}
          onToggle={() => onToggleSort("status")}
        />
      ),
      className: "w-[100px]",
    },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ];

  return (
    <>
      <PropertyFilter
        properties={filterProperties}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchKeyword}
        onSearchChange={onSearchChange}
        searchPlaceholder="飼主名、ペット名、犬種..."
        count={filteredCount}
        sortProperties={TRIMMING_SORT_PROPERTIES}
        activeSorts={activeSorts}
        onSortChange={onSortChange}
      />

      <FilteringIndicator isFiltering={isFiltering}>
        <DataTable
          headerRowClassName={DESIGN_TABLE_HEADER_ROW}
          headerCellClassName={DESIGN_TABLE_HEADER_CELL}
          columns={columns}
          data={records}
          renderRow={(record) => (
            <TrimmingTableRow
              key={record.id}
              record={record}
              isValidStaff={isValidStaff}
              onEdit={onEdit}
              onDeleteClick={onDeleteClick}
              canEdit={canEdit}
              canDelete={canDelete}
            />
          )}
        />
      </FilteringIndicator>

      <Pagination
        currentPage={currentPage}
        totalPages={totalPages}
        totalCount={filteredCount}
        startIndex={startIndex}
        endIndex={endIndex}
        onPageChange={onPageChange}
        onPrev={() => onPageChange(currentPage - 1)}
        onNext={() => onPageChange(currentPage + 1)}
      />
    </>
  );
}

interface TrimmingTableRowProps {
  record: TrimmingUI;
  isValidStaff: (staff: string) => boolean;
  onEdit: (id: string) => void;
  onDeleteClick: (record: TrimmingUI) => void;
  canEdit: boolean;
  canDelete: boolean;
}

const TrimmingTableRow = memo(function TrimmingTableRow({
  record,
  isValidStaff,
  onEdit,
  onDeleteClick,
  canEdit,
  canDelete,
}: TrimmingTableRowProps) {
  return (
    <DataTableRow>
      <TableCell className={`font-mono ${C.text}`}>
        {record.date}
      </TableCell>
      <TableCell className={C.text}>{record.ownerName}</TableCell>
      <TableCell>
        <div className="flex flex-col">
          <DataTableRowLink
            to={paths.trimming.detail.getHref(record.id)}
            state={{ from: paths.trimming.getHref() }}
            aria-label={`トリミング詳細: ${record.petName} ${record.date} ID ${record.id}`}
          >
            {record.petName}
          </DataTableRowLink>
          <span className={`text-sm ${C.text60}`}>{record.petNumber}</span>
        </div>
      </TableCell>
      <TableCell className={`${C.text} hidden lg:table-cell`}>{record.species}</TableCell>
      <TableCell className={`${C.text} hidden lg:table-cell`}>{record.breed || "-"}</TableCell>
      <TableCell className={`${C.text} hidden lg:table-cell`}>{record.weight}</TableCell>
      <TableCell className={`${C.text} truncate max-w-[200px] hidden lg:table-cell`}>
        {record.styleRequest}
      </TableCell>
      <TableCell className={C.text}>
        <div className="flex items-center gap-1.5">
          {!isValidStaff(record.staff) ? (
            <span
              role="img"
              aria-label={`無効な担当スタッフ: ${record.staff}（退職等）`}
              title="担当スタッフが無効（退職等）に設定されています"
            >
              <AlertTriangle className={`${ICON.action} ${C.textWarningIcon}`} aria-hidden="true" />
            </span>
          ) : null}
          {record.staff}
        </div>
      </TableCell>
      <TableCell>
        <StatusBadge colorClass={getTrimmingStatusColor(record.status)}>
          {record.status}
        </StatusBadge>
      </TableCell>
      <TableCell className="text-right">
        {canEdit || canDelete ? (
          <RowActionDropdown
            ariaLabel={`トリミング操作: ${record.petName} ${record.date} ID ${record.id}`}
            actions={[
              ...(canEdit ? [{
                label: "編集",
                icon: Edit,
                onClick: () => onEdit(record.id),
              }] : []),
              ...(canDelete ? [{
                label: "削除",
                icon: Trash2,
                variant: "destructive" as const,
                onClick: () => onDeleteClick(record),
              }] : []),
            ]}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
});
