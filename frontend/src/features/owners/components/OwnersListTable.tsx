import { FileText, Pencil, Trash2 } from "lucide-react";

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { Pagination } from "@/components/shared/Pagination";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import type { ActiveFilter, FilterProperty } from "@/components/shared/PropertyFilter/types";
import { paths } from "@/config/paths";
import { C, STYLE } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import { formatWeight } from "@/lib/format/number";
import { getPetStatusColor } from "@/lib/status-helpers";
import type { Pet } from "@/types";

// DESIGN.md ex-data-table-cell: header は canvas-soft 背景 + eyebrow 相当タイポグラフィ（12px/600/tracking）。
// STYLE.tableHeaderRow/tableHeaderCell（他画面と共有）を直接変更すると影響範囲が広がるため、
// DataTable の headerRowClassName/headerCellClassName で本テーブルのみ上書きする。
const OWNERS_TABLE_HEADER_ROW = `border-b ${C.borderLight} ${C.bgPage} h-11`;
const OWNERS_TABLE_HEADER_CELL = `${STYLE.sectionLabel} h-11`;

interface OwnersPaginationView {
  totalPages: number;
  totalCount: number;
  startIndex: number;
  endIndex: number;
  currentPage: number;
}

interface OwnersListTableProps {
  /** #266: サーバサイドページネーション後の現在ページ分の飼主ペット行（フラット化済み） */
  pets: Pet[];
  pagination: OwnersPaginationView;
  searchTerm: string;
  activeFilters: ActiveFilter[];
  /** #266: species は動物種マスタ由来のため呼び出し側 (OwnersList.tsx) が動的に組み立てて渡す */
  filterProperties: FilterProperty[];
  isFiltering: boolean;
  canEdit: boolean;
  canDelete: boolean;
  /** #158: 飼主レポート閲覧権限（medical-records:view）。行操作に「レポート」を出す。 */
  canReport: boolean;
  /** #86: 拠点横断表示時に医院列を出す */
  showClinicColumn?: boolean;
  /** #86: clinicId → 医院名（所属医院由来） */
  clinicNameById?: Map<string, string>;
  /** #86: 現在の医院ID。別医院の行は編集・削除を抑止（閲覧のみ） */
  currentClinicId?: string | null;
  onSearchChange: (value: string) => void;
  onFilterChange: (filters: ActiveFilter[]) => void;
  onEdit: (ownerId: string) => void;
  onDeleteRequest: (ownerId: string, ownerName: string) => void;
  /** #158: 飼主レポートを別ウィンドウで開く（ownerId + 初期 petId）。 */
  onReport: (ownerId: string, petId: string) => void;
  onPageChange: (page: number) => void;
}

export function OwnersListTable({
  pets,
  pagination,
  searchTerm,
  activeFilters,
  filterProperties,
  isFiltering,
  canEdit,
  canDelete,
  canReport,
  showClinicColumn = false,
  clinicNameById,
  currentClinicId,
  onSearchChange,
  onFilterChange,
  onEdit,
  onDeleteRequest,
  onReport,
  onPageChange,
}: OwnersListTableProps) {
  // #266: サーバサイドページネーション化に伴い列ソートは撤去（コメント参照）。プレーンなラベルに変更。
  const columns = [
    { header: "飼主No", className: "w-[100px] hidden lg:table-cell" },
    { header: "飼主名", className: "w-[180px]" },
    // #86: 拠点横断表示時のみ医院列を表示
    ...(showClinicColumn ? [{ header: "医院", className: "w-[110px]" }] : []),
    { header: "ペット番号", className: "w-[100px] hidden lg:table-cell" },
    { header: "ペット名", className: "w-[120px]" },
    { header: "生死", className: "w-[60px]" },
    { header: "種", className: "w-[60px]" },
    { header: "生年月日", className: "w-[100px] hidden lg:table-cell" },
    { header: "体重", className: "w-[80px] hidden lg:table-cell" },
    { header: "環境", className: "w-[120px] hidden lg:table-cell" },
    { header: "前回来院", className: "w-[100px] hidden lg:table-cell" },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ];

  return (
    <div className="flex flex-col gap-4 flex-1 min-h-0">
      <PropertyFilter
        properties={filterProperties}
        activeFilters={activeFilters}
        onFilterChange={onFilterChange}
        searchTerm={searchTerm}
        onSearchChange={onSearchChange}
        searchPlaceholder="飼主名、ペット名、電話番号、飼主No、ペット番号..."
        count={pagination.totalCount}
      />

      <FilteringIndicator isFiltering={isFiltering}>
        <DataTable
          columns={columns}
          data={pets}
          emptyMessage="データが見つかりません"
          headerRowClassName={OWNERS_TABLE_HEADER_ROW}
          headerCellClassName={OWNERS_TABLE_HEADER_CELL}
          renderRow={(pet) => {
            // #86: 別医院の行は詳細リンク・編集・削除を抑止（閲覧のみ）。
            const isOtherClinic = !!(
              currentClinicId &&
              pet.clinicId &&
              pet.clinicId !== currentClinicId
            );
            return (
              <OwnersListRow
                key={pet.id}
                pet={pet}
                canViewDetail={!isOtherClinic}
                canEdit={canEdit ? !isOtherClinic : false}
                canDelete={canDelete ? !isOtherClinic : false}
                canReport={canReport ? !isOtherClinic : false}
                showClinicColumn={showClinicColumn}
                clinicNameById={clinicNameById}
                onEdit={onEdit}
                onDeleteRequest={onDeleteRequest}
                onReport={onReport}
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
  canViewDetail: boolean;
  canEdit: boolean;
  canDelete: boolean;
  canReport: boolean;
  showClinicColumn?: boolean;
  clinicNameById?: Map<string, string>;
  onEdit: (ownerId: string) => void;
  onDeleteRequest: (ownerId: string, ownerName: string) => void;
  onReport: (ownerId: string, petId: string) => void;
}

function OwnersListRow({
  pet,
  canViewDetail,
  canEdit,
  canDelete,
  canReport,
  showClinicColumn = false,
  clinicNameById,
  onEdit,
  onDeleteRequest,
  onReport,
}: OwnersListRowProps) {
  return (
    <DataTableRow>
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap hidden lg:table-cell`}>
        {pet.ownerNumber ?? "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>
        <span className="flex items-center gap-1.5">
          {canViewDetail ? (
            <DataTableRowLink
              to={paths.owners.detail.getHref(pet.ownerId)}
              aria-label={`飼主「${pet.ownerName}」(ID: ${pet.ownerId}) をペット「${pet.name}」(ID: ${pet.id}) の行から開く`}
            >
              {pet.ownerName}
            </DataTableRowLink>
          ) : (
            pet.ownerName
          )}
          {pet.dangerLevel === "高" ? (
            <Popover>
              <PopoverTrigger asChild>
                <button
                  type="button"
                  aria-label={`${pet.name}の危険理由を表示`}
                  className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-semibold ${C.bgDanger10} ${C.danger} ${C.borderDanger20} outline-none focus-visible:ring-2 ${C.focusRingAccent40}`}
                >
                  ⚠ 危険
                </button>
              </PopoverTrigger>
              <PopoverContent
                align="start"
                aria-label={`${pet.name}の危険理由`}
                onOpenAutoFocus={(event) => event.preventDefault()}
                className="w-64"
              >
                <p className={`text-sm font-semibold ${C.danger}`}>危険理由</p>
                <p className={`mt-1 whitespace-pre-wrap break-words text-sm ${C.textInkSecondary}`}>
                  {pet.dangerReason?.trim() || "理由未登録"}
                </p>
              </PopoverContent>
            </Popover>
          ) : null}
        </span>
      </TableCell>
      {/* #86: 拠点横断表示時のみ医院列を表示 */}
      {showClinicColumn ? (
        <TableCell
          className={`${STYLE.tableCell} whitespace-nowrap`}
          data-testid="owner-row-clinic"
        >
          {clinicNameById?.get(pet.clinicId ?? "") ?? "-"}
        </TableCell>
      ) : null}
      <TableCell className={`${STYLE.tableCell} font-mono whitespace-nowrap hidden lg:table-cell`}>
        {pet.petNumber || "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} whitespace-nowrap`}>{pet.name}</TableCell>
      <TableCell className="whitespace-nowrap">
        {pet.status ? (
          <StatusBadge colorClass={getPetStatusColor(pet.status)}>{pet.status}</StatusBadge>
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
      <TableCell className="whitespace-nowrap text-right">
        {canEdit || canDelete || canReport ? (
          <RowActionDropdown
            ariaLabel={`飼主「${pet.ownerName}」(ID: ${pet.ownerId})・ペット「${pet.name}」(ID: ${pet.id}) の操作`}
            actions={[
              ...(canEdit
                ? [
                    {
                      label: "編集",
                      icon: Pencil,
                      onClick: () => onEdit(pet.ownerId),
                    },
                  ]
                : []),
              ...(canReport
                ? [
                    {
                      label: "レポート",
                      icon: FileText,
                      onClick: () => onReport(pet.ownerId, pet.id),
                    },
                  ]
                : []),
              ...(canDelete
                ? [
                    {
                      label: "削除",
                      icon: Trash2,
                      variant: "destructive" as const,
                      onClick: () => onDeleteRequest(pet.ownerId, pet.ownerName),
                    },
                  ]
                : []),
            ]}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  );
}
