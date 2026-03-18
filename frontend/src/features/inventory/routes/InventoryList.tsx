// React/Framework
import { useState, useDeferredValue, useCallback, useMemo } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, Package, FileSpreadsheet, AlertTriangle, CircleDot, FolderOpen } from "lucide-react";

// Types
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
  ActiveSort,
} from "@/components/shared/NotionFilter/types";

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
import {
  getInventoryStatusColor,
  getInventoryStatusLabel,
} from "@/utils/status-helpers";

// Relative
import { useInventory } from "../hooks/useInventory";

// Types
import type { InventoryItem } from "@/types";

type CategoryFilter = InventoryItem["category"] | "all";
type StatusFilter = InventoryItem["status"] | "all";
type SortKey = "name" | "category" | "quantity" | "status";

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
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);
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

  // ── Sort logic driven by activeSorts ──
  const handleSortChange = useCallback((sorts: ActiveSort[]) => {
    setActiveSorts(sorts);
  }, []);

  const toggleSort = useCallback((key: SortKey) => {
    setActiveSorts((prev) => {
      const existing = prev.find((s) => s.key === key);
      if (!existing) {
        return [{ key, direction: "asc" as const }];
      }
      if (existing.direction === "asc") {
        return prev.map((s) => s.key === key ? { ...s, direction: "desc" as const } : s);
      }
      return prev.filter((s) => s.key !== key);
    });
  }, []);

  const directionFor = useCallback(
    (key: SortKey): "ascending" | "descending" | "none" => {
      const sort = activeSorts.find((s) => s.key === key);
      if (!sort) return "none";
      return sort.direction === "asc" ? "ascending" : "descending";
    },
    [activeSorts],
  );

  const sortedData = useMemo(() => {
    if (activeSorts.length === 0) return [...filteredItems];
    const sorted = [...filteredItems];
    sorted.sort((a, b) => {
      for (const sort of activeSorts) {
        const key = sort.key as SortKey;
        if (key === "quantity") {
          const numCmp = a.quantity - b.quantity;
          if (numCmp !== 0) return sort.direction === "asc" ? numCmp : -numCmp;
          continue;
        }
        const aVal = String(a[key] ?? "");
        const bVal = String(b[key] ?? "");
        const cmp = aVal.localeCompare(bVal, "ja");
        if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
      }
      return 0;
    });
    return sorted;
  }, [filteredItems, activeSorts]);

  const handleCreate = useCallback(() => {
    navigate(paths.inventory.new.getHref());
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(`/inventory/${id}`);
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
    { header: "最低在庫", className: "w-[100px]", align: "right" as const },
    { header: "保管場所", className: "w-[120px]" },
    { header: "有効期限", className: "w-[120px]" },
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

  return (
    <PageLayout
      title="在庫管理"
      icon={<Package className="size-5 text-[#37352F]" />}
      headerAction={
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            className="h-10 text-sm gap-2 bg-white"
            onClick={() => {}}
          >
            <FileSpreadsheet className="size-4" />
            データ取込
          </Button>
          <PrimaryButton onClick={handleCreate}>
            <Plus className="mr-1.5 size-4" />
            新規登録
          </PrimaryButton>
        </div>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        {/* Alert summary */}
        {(summary.lowStock > 0 || summary.outOfStock > 0) ? (
          <div className="flex items-center gap-4 p-3 bg-amber-50 border border-amber-200 rounded-lg">
            <AlertTriangle className="size-5 text-amber-600" />
            <div className="flex gap-4 text-sm">
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
          onSortChange={handleSortChange}
        />

        {/* Table */}
        <div className={isFiltering ? "opacity-60 transition-opacity duration-150" : "transition-opacity duration-150"}>
          <DataTable
            columns={columns}
            data={sortedData}
            emptyMessage="在庫データが見つかりません"
            renderRow={(item) => (
              <DataTableRow key={item.id} onClick={() => handleEdit(item.id)}>
                <TableCell className="text-sm font-medium text-[#37352F] py-2">
                  {item.name}
                </TableCell>
                <TableCell className="text-sm text-[#37352F] py-2">
                  {CATEGORY_LABELS[item.category]}
                </TableCell>
                <TableCell className="text-sm text-[#37352F] py-2 text-right font-mono">
                  {item.quantity} {item.unit}
                </TableCell>
                <TableCell className="text-sm text-[#37352F]/60 py-2 text-right font-mono">
                  {item.minStockLevel} {item.unit}
                </TableCell>
                <TableCell className="text-sm text-[#37352F] py-2">
                  {item.location ?? "-"}
                </TableCell>
                <TableCell className="text-sm text-[#37352F] py-2 font-mono">
                  {item.expiryDate ?? "-"}
                </TableCell>
                <TableCell className="py-2">
                  <StatusBadge colorClass={getInventoryStatusColor(item.status)}>
                    {getInventoryStatusLabel(item.status)}
                  </StatusBadge>
                </TableCell>
                <TableCell className="text-right py-2">
                  <RowActionButton onClick={() => handleEdit(item.id)} />
                </TableCell>
              </DataTableRow>
            )}
          />
        </div>
      </div>
    </PageLayout>
  );
}
