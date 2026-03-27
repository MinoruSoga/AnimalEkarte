// React/Framework
import { ICON } from "@/lib/design-tokens";
import { memo, type ReactNode } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

// External
import { Plus } from "lucide-react";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
  ActiveSort,
} from "@/components/shared/NotionFilter/types";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

interface MasterListPageProps {
  /** Page title (e.g. "ケージマスタ") */
  title: string;
  /** Page icon element */
  icon: ReactNode;

  /** Search state */
  searchTerm: string;
  onSearchChange: (term: string) => void;
  searchPlaceholder: string;
  /** Filtered item count displayed in NotionFilter */
  count: number;
  /** Handler for "新規登録" button */
  onNew: () => void;

  /** NotionFilter filter properties */
  filterProperties?: FilterProperty[];
  /** Active filters state */
  activeFilters?: ActiveFilter[];
  /** Filter change handler */
  onFilterChange?: (filters: ActiveFilter[]) => void;

  /** NotionFilter sort properties */
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
  const navigate = useNavigate();

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title={title}
            icon={icon}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth="max-w-full"
            headerAction={
              <PrimaryButton onClick={onNew}>
                <Plus className={`mr-1.5 ${ICON.action}`} />
                新規登録
              </PrimaryButton>
            }
          >
            <div className="flex flex-col gap-4">
              <NotionFilter
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
          </PageLayout>
        </div>
        {sidePanel}
      </div>

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
