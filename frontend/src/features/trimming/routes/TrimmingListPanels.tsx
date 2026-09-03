import { ICON, C, LAYOUT } from "@/lib/design-tokens";
import { Plus, Scissors } from "lucide-react";
import type {
  ActiveFilter,
  ActiveSort,
  FilterProperty,
} from "@/components/shared/PropertyFilter/types";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { ResourceTrimming } from "@/types/generated/models";
import type { TrimmingUI } from "@/types";
import { TrimmingListTable } from "../components/TrimmingListTable";

interface TrimmingListContentProps {
  canCreate: boolean;
  onNew: () => void;
  isTruncated: boolean;
  paginatedData: TrimmingUI[];
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
  isValidStaff: (name: string) => boolean;
  directionFor: (key: string) => "ascending" | "descending" | "none";
  onSearchChange: (value: string) => void;
  onFilterChange: (next: ActiveFilter[]) => void;
  onSortChange: (sorts: ActiveSort[]) => void;
  onToggleSort: (key: string) => void;
  onEdit: (id: string) => void;
  onDeleteClick: (record: TrimmingUI) => void;
  onPageChange: (page: number) => void;
  deleteOpen: boolean;
  deleteLabel: string | undefined;
  onDeleteClose: () => void;
  onDeleteConfirm: () => void;
}

export function TrimmingListContent({
  canCreate,
  onNew,
  isTruncated,
  paginatedData,
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
  deleteOpen,
  deleteLabel,
  onDeleteClose,
  onDeleteConfirm,
}: TrimmingListContentProps) {
  return (
    <PageLayout
      title="トリミング管理"
      icon={<Scissors className={`${ICON.page} ${C.text}`} />}
      resource={ResourceTrimming}
      headerAction={
        canCreate ? (
          <PrimaryButton onClick={onNew}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規登録
          </PrimaryButton>
        ) : null
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <div className="flex flex-col gap-4">
        {isTruncated ? (
          <p className={`text-xs ${C.text50}`} role="status">
            取得上限の{HISTORY_FETCH_LIMIT}件を対象に検索・絞り込みしています
          </p>
        ) : null}
        <TrimmingListTable
          records={paginatedData}
          filteredCount={filteredCount}
          currentPage={currentPage}
          totalPages={totalPages}
          startIndex={startIndex}
          endIndex={endIndex}
          searchKeyword={searchKeyword}
          activeFilters={activeFilters}
          activeSorts={activeSorts}
          filterProperties={filterProperties}
          isFiltering={isFiltering}
          canEdit={canEdit}
          canDelete={canDelete}
          isValidStaff={isValidStaff}
          directionFor={directionFor}
          onSearchChange={onSearchChange}
          onFilterChange={onFilterChange}
          onSortChange={onSortChange}
          onToggleSort={onToggleSort}
          onEdit={onEdit}
          onDeleteClick={onDeleteClick}
          onPageChange={onPageChange}
        />
      </div>

      <ConfirmDialog
        open={deleteOpen}
        onClose={onDeleteClose}
        title="削除確認"
        description={`${deleteLabel} を削除してもよろしいですか？`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={onDeleteConfirm}
      />
    </PageLayout>
  );
}
