// React/Framework
import { useState, useDeferredValue, useCallback, useEffect, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";

// External
import { Plus, Package } from "lucide-react";

// Internal
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { useSortableData } from "@/hooks/use-sortable-data";
import { usePermission } from "@/hooks/use-permission";
import { paths } from "@/config/paths";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { ResourceInventory } from "@/types/generated/models";

// Relative
import { useInventoryList } from "../hooks/use-inventory";
import { InventoryListContent } from "./InventoryListPanels";
import {
  INVENTORY_PAGE_SIZE,
  buildServerPagePagination,
  excludeInventoryItems,
  nextListSearchParamsWithPage,
  nextListSearchParamsWithoutPage,
  resolveInventoryListFilters,
} from "./inventory-list-model";

// Types
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";

export function InventoryList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission("inventory");
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const isFiltering = searchTerm !== deferredSearch;

  // FE-144 / BUG-412: URLクエリパラメータからページ番号を読み取る（サーバフェッチのキーそのものになる）
  const urlPage = Math.max(1, Number(searchParams.get("page") ?? 1) || 1);

  const { category, excludeCategory, statusFilter, excludeStatus } =
    resolveInventoryListFilters(activeFilters);

  const {
    data: filteredItems,
    summary,
    total: serverTotal,
    limit: serverLimit,
    page: serverPage,
    isLoading,
    isError,
  } = useInventoryList({
    searchTerm: deferredSearch,
    category,
    statusFilter,
    page: urlPage,
    limit: INVENTORY_PAGE_SIZE,
  });

  const excludedFilteredItems = useMemo(
    () => excludeInventoryItems(filteredItems, excludeCategory, excludeStatus),
    [filteredItems, excludeCategory, excludeStatus],
  );

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(excludedFilteredItems, { numericKeys: ["quantity"] });

  const pagination = buildServerPagePagination({
    rows: sortedData,
    total: serverTotal,
    page: serverPage,
    limit: serverLimit,
  });

  // BUG-412: URLの page がサーバ total から導いた totalPages を超えている場合はクランプする
  // （フィルタ変更等で母集団が縮んだ場合に空ページへ迷い込むのを防ぐ。BUG-411踏襲）。
  useEffect(() => {
    if (isLoading) return;
    const clampedPage = Math.max(1, Math.min(urlPage, pagination.totalPages));
    if (clampedPage !== urlPage) {
      setSearchParams(
        (prev) => nextListSearchParamsWithPage(prev, clampedPage),
        { replace: true },
      );
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps -- setSearchParams は安定参照。urlPage/totalPages/isLoading の変化時のみ再評価する設計（FE-144/BUG-411踏襲）。
  }, [urlPage, pagination.totalPages, isLoading]);

  const handlePageChange = useCallback((page: number) => {
    setSearchParams(
      (prev) => nextListSearchParamsWithPage(prev, page),
      { replace: true },
    );
  }, [setSearchParams]);

  const handleSearchChange = useCallback((value: string) => {
    setSearchTerm(value);
    setSearchParams(
      (prev) => nextListSearchParamsWithoutPage(prev),
      { replace: true },
    );
  }, [setSearchParams]);

  const handleFilterChange = useCallback((next: ActiveFilter[]) => {
    setActiveFilters(next);
    setSearchParams(
      (prev) => nextListSearchParamsWithoutPage(prev),
      { replace: true },
    );
  }, [setSearchParams]);

  const hasPageScopedFilter =
    deferredSearch !== "" || activeSorts.length > 0 || excludeCategory !== null || excludeStatus !== null;

  const handleCreate = useCallback(() => {
    navigate(paths.inventory.new.getHref());
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(paths.inventory.detail.getHref(id));
  }, [navigate]);

  if (isLoading) return <LoadingFallback />;
  if (isError) return <ErrorFallback />;

  return (
    <PageLayout
      title="在庫管理"
      resource={ResourceInventory}
      icon={<Package className={`${ICON.page} ${C.text}`} />}
      headerAction={
        <div className="flex items-center gap-2">
          {canCreate ? (
            <PrimaryButton colorVariant="primary" onClick={handleCreate}>
              <Plus className={`mr-1.5 ${ICON.action}`} />
              新規登録
            </PrimaryButton>
          ) : null}
        </div>
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <InventoryListContent
        summary={summary}
        activeFilters={activeFilters}
        searchTerm={searchTerm}
        onSearchChange={handleSearchChange}
        onFilterChange={handleFilterChange}
        count={excludedFilteredItems.length}
        activeSorts={activeSorts}
        onSortChange={setActiveSorts}
        directionFor={directionFor}
        toggleSort={toggleSort}
        isFiltering={isFiltering}
        hasPageScopedFilter={hasPageScopedFilter}
        pagination={pagination}
        canEdit={canEdit}
        onEdit={handleEdit}
        onPageChange={handlePageChange}
      />
    </PageLayout>
  );
}
