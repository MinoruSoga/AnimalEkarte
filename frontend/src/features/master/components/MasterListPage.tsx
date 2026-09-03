// React/Framework
import { memo, type ReactNode } from "react";

// Internal
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { MasterPageShell } from "./MasterPageShell";
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
  ActiveSort,
} from "@/components/shared/PropertyFilter/types";
import type { Resource } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

interface MasterListPageProps {
  /** Page title (e.g. "ケージマスタ") */
  title: string;
  /** Page icon element */
  icon: ReactNode;
  /** 権限バッジ表示用リソース */
  resource?: Resource;

  /** Search state */
  searchTerm: string;
  onSearchChange: (term: string) => void;
  searchPlaceholder: string;
  /** Filtered item count displayed in PropertyFilter */
  count: number;
  /** Handler for "新規登録" button */
  onNew: () => void;

  /** PropertyFilter filter properties */
  filterProperties?: FilterProperty[];
  /** Active filters state */
  activeFilters?: ActiveFilter[];
  /** Filter change handler */
  onFilterChange?: (filters: ActiveFilter[]) => void;

  /** PropertyFilter sort properties */
  sortProperties?: SortProperty[];
  /** Active sorts state */
  activeSorts?: ActiveSort[];
  /** Sort change handler */
  onSortChange?: (sorts: ActiveSort[]) => void;

  /** SidePanel rendered next to main content. null when closed. */
  sidePanel: ReactNode | null;

  /** Whether delete dialog is open */
  deleteOpen: boolean;
  /** Delete dialog title (e.g. "ケージを削除しますか？") */
  deleteTitle: string;
  /** Delete dialog description (e.g. "「XXX」を削除します。...") */
  deleteDescription: string;
  onDeleteConfirm: () => void;
  onDeleteCancel: () => void;

  /** Table content */
  children: ReactNode;
}

// ─────────────────────────────────────────────────
// Component
// ─────────────────────────────────────────────────

export const MasterListPage = memo(function MasterListPage({
  title,
  icon,
  resource,
  searchTerm,
  onSearchChange,
  searchPlaceholder,
  count,
  onNew,
  filterProperties,
  activeFilters,
  onFilterChange,
  sortProperties,
  activeSorts,
  onSortChange,
  sidePanel,
  deleteOpen,
  deleteTitle,
  deleteDescription,
  onDeleteConfirm,
  onDeleteCancel,
  children,
}: MasterListPageProps) {
  return (
    <>
      <MasterPageShell
        title={title}
        icon={icon}
        resource={resource}
        onNew={onNew}
        sidePanel={sidePanel}
      >
        <div className="flex flex-col gap-4">
          <PropertyFilter
            properties={filterProperties ?? []}
            activeFilters={activeFilters ?? []}
            onFilterChange={onFilterChange ?? (() => {})}
            searchTerm={searchTerm}
            onSearchChange={onSearchChange}
            searchPlaceholder={searchPlaceholder}
            count={count}
            sortProperties={sortProperties}
            activeSorts={activeSorts}
            onSortChange={onSortChange}
          />
          {children}
        </div>
      </MasterPageShell>

      <ConfirmDialog
        open={deleteOpen}
        onClose={onDeleteCancel}
        title={deleteTitle}
        description={deleteDescription}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={onDeleteConfirm}
      />
    </>
  );
});
