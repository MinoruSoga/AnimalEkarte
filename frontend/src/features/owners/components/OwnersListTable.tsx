import { Heart, PawPrint, Pencil, Trash2 } from "lucide-react";

import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { Pagination } from "@/components/shared/Pagination";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { SortableHeader } from "@/components/shared/SortableHeader";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/NotionFilter/types";
import type {
  ActiveFilter,
  ActiveSort,
  FilterProperty,
  SortProperty,
} from "@/components/shared/NotionFilter/types";
import { C, STYLE } from "@/lib/design-tokens";
import { formatDate } from "@/utils/format/date";
import { formatWeight } from "@/utils/format/number";
import { getPetStatusColor } from "@/utils/status-helpers";
import type { Pet } from "@/types";

const OWNER_SORT_PROPERTIES: SortProperty[] = [
  { key: "ownerNumber", label: "飼主No" },
  { key: "ownerName", label: "飼主名" },
  { key: "name", label: "ペット名" },
  { key: "species", label: "種" },
  { key: "birthDate", label: "生年月日" },
  { key: "lastVisit", label: "前回来院" },
];

const OWNER_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "species",
    label: "種",
    type: "select",
    icon: PawPrint,
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "犬", label: "犬" },
      { value: "猫", label: "猫" },
      { value: "鳥", label: "鳥" },
      { value: "うさぎ", label: "うさぎ" },
      { value: "ハムスター", label: "ハムスター" },
      { value: "その他", label: "その他" },
    ],
  },
  {
    key: "status",
    label: "生死",
    type: "select",
    icon: Heart,
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "alive", label: "生存" },
      { value: "deceased", label: "死亡" },
    ],
  },
];

interface OwnersPaginationView {
  paginatedData: Pet[];
  totalPages: number;
  totalCount: number;
  startIndex: number;
  endIndex: number;
  currentPage: number;
}

interface OwnersListTableProps {
  filteredCount: number;
  pagination: OwnersPaginationView;
  searchTerm: string;
  activeFilters: ActiveFilter[];
  activeSorts: ActiveSort[];
  isFiltering: boolean;
  canEdit: boolean;
  canDelete: boolean;
  /** #86: 拠点横断表示時に医院列を出す */
  showClinicColumn?: boolean;
  /** #86: clinicId → 医院名（所属医院由来） */
  clinicNameById?: Map<string, string>;
  /** #86: 現在の医院ID。別医院の行は編集・削除を抑止（閲覧のみ） */
  currentClinicId?: string | null;
  directionFor: (key: string) => "ascending" | "descending" | "none";
  onSearchChange: (value: string) => void;
  onFilterChange: (filters: ActiveFilter[]) => void;
  onSortChange: (sorts: ActiveSort[]) => void;
  onToggleSort: (key: string) => void;
  onRowClick: (pet: Pet) => void;
  onEdit: (ownerId: string) => void;
  onDeleteRequest: (ownerId: string, ownerName: string) => void;
  onPageChange: (page: number) => void;
}

export function OwnersListTable({
  filteredCount,
  pagination,
  searchTerm,
  activeFilters,
  activeSorts,
  isFiltering,
  canEdit,
  canDelete,
  showClinicColumn = false,
  clinicNameById,
  currentClinicId,
  directionFor,
  onSearchChange,
  onFilterChange,
  onSortChange,
  onToggleSort,
  onRowClick,
  onEdit,
  onDeleteRequest,
  onPageChange,
}: OwnersListTableProps) {
  const columns = [
    {
      header: (
        <SortableHeader
          label="飼主No"
          direction={directionFor("ownerNumber")}
          onToggle={() => onToggleSort("ownerNumber")}
        />
      ),
      className: "w-[100px] hidden lg:table-cell",
    },
    {
      header: (
        <SortableHeader
          label="飼主名"
          direction={directionFor("ownerName")}
          onToggle={() => onToggleSort("ownerName")}
        />
      ),
      className: "w-[180px]",
    },
    // #86: 拠点横断表示時のみ医院列を表示
    ...(showClinicColumn ? [{ header: "医院", className: "w-[110px]" }] : []),
    { header: "ペット番号", className: "w-[100px] hidden lg:table-cell" },
    {
      header: (
        <SortableHeader
          label="ペット名"
          direction={directionFor("name")}
          onToggle={() => onToggleSort("name")}
        />
      ),
      className: "w-[120px]",
    },
    { header: "生死", className: "w-[60px] hidden lg:table-cell" },
    {
      header: (
        <SortableHeader
          label="種"
          direction={directionFor("species")}
          onToggle={() => onToggleSort("species")}
        />
      ),
      className: "w-[60px]",
    },
    {
      header: (
        <SortableHeader
          label="生年月日"
          direction={directionFor("birthDate")}
          onToggle={() => onToggleSort("birthDate")}
        />
      ),
      className: "w-[100px] hidden lg:table-cell",
    },
    { header: "体重", className: "w-[80px] hidden lg:table-cell" },
    { header: "環境", className: "w-[120px] hidden lg:table-cell" },
    {
      header: (
        <SortableHeader
          label="前回来院"
          direction={directionFor("lastVisit")}
          onToggle={() => onToggleSort("lastVisit")}
        />
      ),
      className: "w-[100px] hidden lg:table-cell",
    },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ];

  return (
    <div className="flex flex-col gap-4 flex-1 min-h-0">
      <NotionFilter
        properties={OWNER_FILTER_PROPERTIES}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchTerm}
        onSearchChange={onSearchChange}
        searchPlaceholder="飼主名、ペット名、飼主No、種別..."
        count={filteredCount}
        sortProperties={OWNER_SORT_PROPERTIES}
        activeSorts={activeSorts}
        onSortChange={onSortChange}
      />

      <FilteringIndicator isFiltering={isFiltering}>
        <DataTable
          columns={columns}
          data={pagination.paginatedData}
          emptyMessage="データが見つかりません"
          renderRow={(pet) => {
            // #86: 別医院の行は編集・削除を抑止（閲覧のみ）。行クリックの遷移ガードは呼び出し側
            const isOtherClinic = !!(
              currentClinicId && pet.clinicId && pet.clinicId !== currentClinicId
            );
            return (
              <OwnersListRow
                key={pet.id}
                pet={pet}
                canEdit={canEdit && !isOtherClinic}
                canDelete={canDelete && !isOtherClinic}
                showClinicColumn={showClinicColumn}
                clinicNameById={clinicNameById}
                onRowClick={onRowClick}
                onEdit={onEdit}
                onDeleteRequest={onDeleteRequest}
              />
            );
          }}
        />
      </FilteringIndicator>

      {pagination.totalPages > 1 ? (
        <Pagination
          currentPage={pagination.currentPage}
          totalPages={pagination.totalPages}
          totalCount={pagination.totalCount}
          startIndex={pagination.startIndex}
          endIndex={pagination.endIndex}
          onPageChange={onPageChange}
          onPrev={() => onPageChange(pagination.currentPage - 1)}
          onNext={() => onPageChange(pagination.currentPage + 1)}
        />
      ) : null}
    </div>
  );
}

interface OwnersListRowProps {
  pet: Pet;
  canEdit: boolean;
  canDelete: boolean;
  showClinicColumn?: boolean;
  clinicNameById?: Map<string, string>;
  onRowClick: (pet: Pet) => void;
  onEdit: (ownerId: string) => void;
  onDeleteRequest: (ownerId: string, ownerName: string) => void;
}

function OwnersListRow({
  pet,
  canEdit,
  canDelete,
  showClinicColumn = false,
  clinicNameById,
  onRowClick,
  onEdit,
  onDeleteRequest,
}: OwnersListRowProps) {
  return (
    <DataTableRow onClick={canEdit ? () => onRowClick(pet) : undefined}>
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap hidden lg:table-cell`}>
        {pet.ownerNumber ?? "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>
        <span className="flex items-center gap-1.5">
          {pet.ownerName}
          {pet.dangerLevel === "高" ? (
            <span className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-semibold ${C.bgDanger10} ${C.danger} ${C.borderDanger20}`}>
              ⚠ 危険
            </span>
          ) : null}
        </span>
      </TableCell>
      {/* #86: 拠点横断表示時のみ医院列を表示 */}
      {showClinicColumn ? (
        <TableCell className={`${STYLE.tableCell} whitespace-nowrap`} data-testid="owner-row-clinic">
          {clinicNameById?.get(pet.clinicId ?? "") ?? "-"}
        </TableCell>
      ) : null}
      <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap hidden lg:table-cell`}>
        {pet.petNumber || "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>{pet.name}</TableCell>
      <TableCell className="whitespace-nowrap py-2 hidden lg:table-cell">
        {pet.status ? (
          <StatusBadge colorClass={getPetStatusColor(pet.status)}>
            {pet.status}
          </StatusBadge>
        ) : null}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>{pet.species}</TableCell>
      <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap hidden lg:table-cell`}>
        {formatDate(pet.birthDate)}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap hidden lg:table-cell`}>
        {formatWeight(pet.weight)}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap hidden lg:table-cell`}>
        {pet.environment || "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap hidden lg:table-cell`}>
        {formatDate(pet.lastVisit)}
      </TableCell>
      <TableCell className="whitespace-nowrap py-2 text-right">
        {canEdit || canDelete ? (
          <RowActionDropdown
            actions={[
              ...(canEdit ? [{
                label: "編集",
                icon: Pencil,
                onClick: () => onEdit(pet.ownerId),
              }] : []),
              ...(canDelete ? [{
                label: "削除",
                icon: Trash2,
                variant: "destructive" as const,
                onClick: () => onDeleteRequest(pet.ownerId, pet.ownerName),
              }] : []),
            ]}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}
