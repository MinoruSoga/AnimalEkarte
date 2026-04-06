// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { useState, useDeferredValue, useCallback, useEffect, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";

// External
import { Plus, Package, FileSpreadsheet, AlertTriangle, CircleDot, FolderOpen } from "lucide-react";

// Types
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
} from "@/components/shared/NotionFilter/types";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/NotionFilter/types";

// Internal
import { paths } from "@/config/paths";
import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { getInventoryStatusColor, getInventoryStatusLabel } from "@/utils/status-helpers";
import { usePagination } from "@/hooks/use-pagination";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";

// Relative
import { useInventory } from "../hooks/use-inventory";
import { usePermission } from "@/features/auth";

// Types
import type { InventoryItem } from "@/types";
import { ResourceInventory } from "@/types/generated/models";

type CategoryFilter = InventoryItem["category"] | "all";
type StatusFilter = InventoryItem["status"] | "all";
const CATEGORY_LABELS: Record<InventoryItem["category"], string> = {
  medicine: "医薬品",
  consumable: "消耗品",
  food: "フード",
  other: "その他",
};

const INVENTORY_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "category",
    label: "カテゴリ",
    type: "select",
    icon: FolderOpen,
    // inventory_items.category NOT NULL — 空値は存在しない
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "medicine", label: "医薬品" },
      { value: "consumable", label: "消耗品" },
      { value: "food", label: "フード" },
      { value: "other", label: "その他" },
    ],
  },
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    // inventory_items.status DEFAULT 'sufficient' — 空値は存在しない
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "sufficient", label: "十分" },
      { value: "low", label: "残少" },
      { value: "out_of_stock", label: "在庫切れ" },
    ],
  },
];

// rendering-hoist-jsx: 静的ソートプロパティ定義
const INVENTORY_SORT_PROPERTIES: SortProperty[] = [
  { key: "name", label: "品名" },
  { key: "category", label: "カテゴリ" },
  { key: "quantity", label: "在庫数" },
  { key: "status", label: "ステータス" },
];

export function InventoryList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission("inventory");
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const isFiltering = searchTerm !== deferredSearch;

  const categoryFilter = activeFilters.find((f) => f.key === "category");
  // "is" 条件のみサーバーサイド、"is_not" はクライアントサイドで処理
  const category: CategoryFilter = categoryFilter?.condition === "is"
    ? (categoryFilter.value as CategoryFilter)
    : "all";
  const statusFilterEntry = activeFilters.find((f) => f.key === "status");
  const statusFilter: StatusFilter = statusFilterEntry?.condition === "is"
    ? (statusFilterEntry.value as StatusFilter)
    : "all";

  const { data: filteredItems, summary } = useInventory({
    searchTerm: deferredSearch,
    category,
    statusFilter,
  });

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredItems, { numericKeys: ["quantity"] });

  const pagination = usePagination(sortedData, {
    pageSize: 20,
    resetKey: deferredSearch,
  });

  // FE-144: URLクエリパラメータからページ番号を読み取る
  const urlPage = Number(searchParams.get("page") ?? 1);

  // FE-144: URLのページ番号とローカル状態を同期（URLが変わったときのみ）
  useEffect(() => {
    const clampedPage = Math.max(1, Math.min(urlPage, pagination.totalPages));
    if (clampedPage !== pagination.currentPage) {
      pagination.goToPage(clampedPage);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlPage, pagination.totalPages]);

  // FE-144: ページ変更時にURLクエリパラメータを更新
  const handlePageChange = useCallback((page: number) => {
    pagination.goToPage(page);
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (page === 1) {
        next.delete("page");
      } else {
        next.set("page", String(page));
      }
      return next;
    }, { replace: true });
  }, [pagination, setSearchParams]);

  const handleCreate = useCallback(() => {
    navigate(paths.inventory.new.getHref());
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(paths.inventory.detail.getHref(id));
  }, [navigate]);

  const columns = useMemo(() => [
    {
      header: (
        <SortableHeader
          label="品名"
          direction={directionFor("name")}
          onToggle={() => toggleSort("name")}
        />
      ),
      className: "min-w-[200px]",
    },
    {
      header: (
        <SortableHeader
          label="カテゴリ"
          direction={directionFor("category")}
          onToggle={() => toggleSort("category")}
        />
      ),
      className: "w-[100px]",
    },
    {
      header: (
        <SortableHeader
          label="在庫数"
          direction={directionFor("quantity")}
          onToggle={() => toggleSort("quantity")}
        />
      ),
      className: "w-[100px]",
      align: "right" as const,
    },
    { header: "最低在庫", className: "w-[100px] hidden lg:table-cell", align: "right" as const },
    { header: "保管場所", className: "w-[120px]" },
    { header: "有効期限", className: "w-[120px] hidden lg:table-cell" },
    {
      header: (
        <SortableHeader
          label="ステータス"
          direction={directionFor("status")}
          onToggle={() => toggleSort("status")}
        />
      ),
      className: "w-[100px]",
    },
    { header: "操作", className: "w-[80px]", align: "right" as const },
  ], [directionFor, toggleSort]);

  // rerender-memo: renderRow を useCallback でメモ化（DataTable への参照を安定化）
  const renderRow = useCallback((item: InventoryItem) => (
    <DataTableRow key={item.id} onClick={() => handleEdit(item.id)}>
      <TableCell className={`text-base font-medium ${C.text} py-2`}>
        {item.name}
      </TableCell>
      <TableCell className={`text-base ${C.text} py-2`}>
        {CATEGORY_LABELS[item.category]}
      </TableCell>
      <TableCell className={`text-base ${C.text} py-2 text-right font-mono`}>
        {item.quantity} {item.unit}
      </TableCell>
      <TableCell className={`text-base ${C.text60} py-2 text-right font-mono hidden lg:table-cell`}>
        {item.minStockLevel} {item.unit}
      </TableCell>
      <TableCell className={`text-base ${C.text} py-2`}>
        {item.location ?? "-"}
      </TableCell>
      <TableCell className={`text-base ${C.text} py-2 font-mono hidden lg:table-cell`}>
        {item.expiryDate ?? "-"}
      </TableCell>
      <TableCell className="py-2">
        <StatusBadge colorClass={getInventoryStatusColor(item.status)}>
          {getInventoryStatusLabel(item.status)}
        </StatusBadge>
      </TableCell>
      <TableCell className="text-right py-2">
        {canEdit ? <RowActionButton onClick={() => handleEdit(item.id)} /> : null}
      </TableCell>
    </DataTableRow>
  ), [handleEdit, canEdit]);

  return (
    <PageLayout
      title="在庫管理"
      resource={ResourceInventory}
      icon={<Package className={`${ICON.page} ${C.text}`} />}
      headerAction={
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            className="h-10 text-base gap-2 bg-white"
            onClick={() => {}}
          >
            <FileSpreadsheet className={ICON.action} />
            データ取込
          </Button>
          {canCreate ? (
            <PrimaryButton onClick={handleCreate}>
              <Plus className={`mr-1.5 ${ICON.action}`} />
              新規登録
            </PrimaryButton>
          ) : null}
        </div>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        {/* Alert summary */}
        {(summary.lowStock > 0 || summary.outOfStock > 0) ? (
          <div className="flex items-center gap-4 p-3 bg-amber-50 border border-amber-200 rounded-lg">
            <AlertTriangle className={`${ICON.page} text-amber-600`} />
            <div className="flex gap-4 text-base">
              {summary.outOfStock > 0 ? (
                <span className="text-red-600 font-medium">
                  在庫切れ: {summary.outOfStock}件
                </span>
              ) : null}
              {summary.lowStock > 0 ? (
                <span className="text-amber-600 font-medium">
                  残少: {summary.lowStock}件
                </span>
              ) : null}
            </div>
          </div>
        ) : null}

        {/* Search & Filters */}
        <NotionFilter
          properties={INVENTORY_FILTER_PROPERTIES}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder="品名、保管場所、仕入先..."
          count={filteredItems.length}
          sortProperties={INVENTORY_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={setActiveSorts}
        />

        {/* Table */}
        <FilteringIndicator isFiltering={isFiltering}>
          <DataTable
            columns={columns}
            data={pagination.paginatedData}
            emptyMessage="在庫データが見つかりません"
            renderRow={renderRow}
          />
        </FilteringIndicator>

        {pagination.totalPages > 1 ? (
          <Pagination
            currentPage={pagination.currentPage}
            totalPages={pagination.totalPages}
            totalCount={pagination.totalCount}
            startIndex={pagination.startIndex}
            endIndex={pagination.endIndex}
            onPageChange={handlePageChange}
            onPrev={() => handlePageChange(pagination.currentPage - 1)}
            onNext={() => handlePageChange(pagination.currentPage + 1)}
          />
        ) : null}
      </div>
    </PageLayout>
  );
}
