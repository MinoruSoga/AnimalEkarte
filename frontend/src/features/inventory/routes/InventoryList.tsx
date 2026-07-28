// React/Framework
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { useState, useDeferredValue, useCallback, useEffect, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";

// External
import { Plus, Package, AlertTriangle, CircleDot, FolderOpen, Info } from "lucide-react";

// Types
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
} from "@/components/shared/PropertyFilter/types";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/PropertyFilter/types";

// Internal
import { paths } from "@/config/paths";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { getInventoryStatusColor, getInventoryStatusLabel } from "@/lib/status-helpers";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";

// Relative
import { useInventoryList } from "../hooks/use-inventory";
import { usePermission } from "@/hooks/use-permission";

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

// BUG-412: サーバサイドページネーションのページサイズ。旧 usePagination() の既定値(20件)と
// backend parsePagination の既定 limit(20) を踏襲する（BUG-411/ACCOUNTING_PAGE_SIZEと同型）。
const INVENTORY_PAGE_SIZE = 20;

// BUG-412: 検索・並び替えは backend が非対応（handlerに search param なし、repositoryは
// Order("name ASC")固定）のため client-only のまま残る。サーバページネーション化後は
// 「取得済み1ページの中でだけ」有効になる — BUG-411の残存教訓（advisor指摘）を踏まえ、
// filter(category/status)だけでなく sort も同じ規則で page-scoped と明示する。

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

  const categoryFilter = activeFilters.find((f) => f.key === "category");
  // "is" 条件のみサーバーサイド、"is_not" はクライアントサイドで処理（下記 excludeCategory/excludeStatus）
  const category: CategoryFilter = categoryFilter?.condition === "is"
    ? (categoryFilter.value as CategoryFilter)
    : "all";
  // BUG-414: is_not はサーバーに絞り込み条件を渡さず(=category:"all")、取得後にクライアント側で除外する。
  const excludeCategory: InventoryItem["category"] | null =
    categoryFilter?.condition === "is_not" ? (categoryFilter.value as InventoryItem["category"]) : null;
  const statusFilterEntry = activeFilters.find((f) => f.key === "status");
  const statusFilter: StatusFilter = statusFilterEntry?.condition === "is"
    ? (statusFilterEntry.value as StatusFilter)
    : "all";
  // BUG-414: status も同様に is_not はクライアント側除外。
  const excludeStatus: InventoryItem["status"] | null =
    statusFilterEntry?.condition === "is_not" ? (statusFilterEntry.value as InventoryItem["status"]) : null;

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

  // BUG-414: is_not はサーバーに送らず全件取得した上でここで除外する
  // （PropertyFilter/CONDITIONS_NO_EMPTY は共有定数のため変更しない。除外は本 feature 内で完結させる）。
  const excludedFilteredItems = useMemo(() => {
    if (!excludeCategory && !excludeStatus) return filteredItems;
    return filteredItems.filter(
      (item) =>
        (!excludeCategory || item.category !== excludeCategory) &&
        (!excludeStatus || item.status !== excludeStatus)
    );
  }, [filteredItems, excludeCategory, excludeStatus]);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(excludedFilteredItems, { numericKeys: ["quantity"] });

  // BUG-412: usePagination() のクライアントスライスを撤去し、サーバの total/page/limit をそのまま使う。
  // paginatedData はクライアント検索・ソート済みの「現在ページの行」（サーバは既にページ分だけ返す）。
  const totalPages = Math.max(1, Math.ceil(serverTotal / serverLimit));
  const pagination = {
    paginatedData: sortedData,
    totalPages,
    totalCount: serverTotal,
    startIndex: serverTotal === 0 ? 0 : (serverPage - 1) * serverLimit + 1,
    endIndex: Math.min(serverPage * serverLimit, serverTotal),
    currentPage: serverPage,
  };

  // BUG-412: URLの page がサーバ total から導いた totalPages を超えている場合はクランプする
  // （フィルタ変更等で母集団が縮んだ場合に空ページへ迷い込むのを防ぐ。BUG-411踏襲）。
  useEffect(() => {
    if (isLoading) return;
    const clampedPage = Math.max(1, Math.min(urlPage, totalPages));
    if (clampedPage !== urlPage) {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        if (clampedPage === 1) {
          next.delete("page");
        } else {
          next.set("page", String(clampedPage));
        }
        return next;
      }, { replace: true });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps -- setSearchParams は安定参照。urlPage/totalPages/isLoading の変化時のみ再評価する設計（FE-144/BUG-411踏襲）。
  }, [urlPage, totalPages, isLoading]);

  // FE-144: ページ変更時にURLクエリパラメータを更新
  const handlePageChange = useCallback((page: number) => {
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (page === 1) {
        next.delete("page");
      } else {
        next.set("page", String(page));
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  // BUG-412: 検索変更時は常に page を1へ戻す（旧 usePagination() の resetKey 相当。BUG-411踏襲）。
  const handleSearchChange = useCallback((value: string) => {
    setSearchTerm(value);
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      next.delete("page");
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  // BUG-412: category/status はサーバフィルタのため母集団が変わりうる。変更時は page を1へ戻す。
  const handleFilterChange = useCallback((next: ActiveFilter[]) => {
    setActiveFilters(next);
    setSearchParams((prev) => {
      const params = new URLSearchParams(prev);
      params.delete("page");
      return params;
    }, { replace: true });
  }, [setSearchParams]);

  // BUG-412: 検索・並び替えは backend 非対応で現在ページ内のみが対象（advisor指摘: sortもfilterと同型）。
  // BUG-414: category/status の is_not も同型のクライアント側処理のため現在ページ内のみが対象。
  const hasPageScopedFilter =
    deferredSearch !== "" || activeSorts.length > 0 || excludeCategory !== null || excludeStatus !== null;

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
    <DataTableRow key={item.id}>
      <TableCell className={`font-medium ${C.text}`}>
        {canEdit ? (
          <DataTableRowLink
            to={paths.inventory.detail.getHref(item.id)}
            aria-label={`在庫品「${item.name}」(ID: ${item.id}) の詳細を開く`}
          >
            {item.name}
          </DataTableRowLink>
        ) : item.name}
      </TableCell>
      <TableCell className={C.text}>
        {CATEGORY_LABELS[item.category]}
      </TableCell>
      <TableCell className={`${C.text} text-right font-mono`}>
        {item.quantity} {item.unit}
      </TableCell>
      <TableCell className={`${C.text60} text-right font-mono hidden lg:table-cell`}>
        {item.minStockLevel} {item.unit}
      </TableCell>
      <TableCell className={C.text}>
        {item.location ?? "-"}
      </TableCell>
      <TableCell className={`${C.text} font-mono hidden lg:table-cell`}>
        {item.expiryDate ?? "-"}
      </TableCell>
      <TableCell>
        <StatusBadge colorClass={getInventoryStatusColor(item.status)}>
          {getInventoryStatusLabel(item.status)}
        </StatusBadge>
      </TableCell>
      <TableCell className="text-right">
        {canEdit ? (
          <RowActionButton
            onClick={() => handleEdit(item.id)}
            aria-label={`在庫品「${item.name}」(ID: ${item.id}) を編集`}
          />
        ) : null}
      </TableCell>
    </DataTableRow>
  ), [handleEdit, canEdit]);

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
      <div className="flex flex-col gap-4">
        {/* Alert summary */}
        {summary.isError ? (
          <div
            role="status"
            className={`flex items-center gap-4 p-3 ${C.bgWarning50} ${C.borderWarning20} border rounded-lg`}
          >
            <AlertTriangle aria-hidden="true" className={`${ICON.page} ${C.textWarningIcon}`} />
            <span className={`text-base ${C.textWarning} font-medium`}>
              低在庫・欠品件数を取得できませんでした。時間をおいて再度お試しください。
            </span>
          </div>
        ) : (summary.lowStock > 0 || summary.outOfStock > 0) ? (
          <div className={`flex items-center gap-4 p-3 ${C.bgWarning50} ${C.borderWarning20} border rounded-lg`}>
            <AlertTriangle className={`${ICON.page} ${C.textWarningIcon}`} />
            <div className="flex gap-4 text-base">
              {summary.outOfStock > 0 ? (
                <span className={`${C.danger} font-medium`}>
                  在庫切れ: {summary.outOfStock}件
                </span>
              ) : null}
              {summary.lowStock > 0 ? (
                <span className={`${C.textWarning} font-medium`}>
                  残少: {summary.lowStock}件
                </span>
              ) : null}
            </div>
          </div>
        ) : null}

        {/* Search & Filters */}
        <PropertyFilter
          properties={INVENTORY_FILTER_PROPERTIES}
          activeFilters={activeFilters}
          onFilterChange={handleFilterChange}
          searchTerm={searchTerm}
          onSearchChange={handleSearchChange}
          searchPlaceholder="品名、保管場所、仕入先..."
          count={excludedFilteredItems.length}
          sortProperties={INVENTORY_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={setActiveSorts}
        />

        {/* BUG-412: 検索・並び替えは backend 非対応のため現在ページ内のみが対象。正直な注記で開示する
            （BUG-411の残存教訓・advisor指摘: sortもfilterと同じ扱いにする）。 */}
        {hasPageScopedFilter && pagination.totalPages > 1 ? (
          <div
            role="status"
            className={`flex items-start gap-2 rounded-md px-3 py-2 ${C.bgWarning50} ${C.borderWarning20} border text-sm`}
          >
            <Info aria-hidden="true" className={`${ICON.action} ${C.textWarningIcon} shrink-0 mt-0.5`} />
            <span>
              検索・並び替えは現在表示中のページ内のみが対象です。他のページにある在庫はこの検索・並び替えでは見つかりません。
              全件を確認する場合はカテゴリ・ステータスでの絞り込みをご利用ください。
            </span>
          </div>
        ) : null}

        {/* Table */}
        <FilteringIndicator isFiltering={isFiltering}>
          <DataTable
            headerRowClassName={DESIGN_TABLE_HEADER_ROW}
            headerCellClassName={DESIGN_TABLE_HEADER_CELL}
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
