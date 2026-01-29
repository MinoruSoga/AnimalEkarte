// React/Framework
import { useState } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, Package, FileSpreadsheet, AlertTriangle } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusBadge } from "@/components/shared/StatusBadge";
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

const CATEGORY_LABELS: Record<InventoryItem["category"], string> = {
  medicine: "医薬品",
  consumable: "消耗品",
  food: "フード",
  other: "その他",
};

export function InventoryList() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [category, setCategory] = useState<CategoryFilter>("all");
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");

  const { data: filteredItems, summary } = useInventory({
    searchTerm,
    category,
    statusFilter,
  });

  const handleCreate = () => {
    navigate("/inventory/new");
  };

  const handleEdit = (id: string) => {
    navigate(`/inventory/${id}`);
  };

  const columns = [
    { header: "品名", className: "min-w-[200px]" },
    { header: "カテゴリ", className: "w-[100px]" },
    { header: "在庫数", className: "w-[100px]", align: "right" as const },
    { header: "最低在庫", className: "w-[100px]", align: "right" as const },
    { header: "保管場所", className: "w-[120px]" },
    { header: "有効期限", className: "w-[120px]" },
    { header: "ステータス", className: "w-[100px]" },
    { header: "操作", className: "w-[80px]", align: "right" as const },
  ];

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
        {(summary.lowStock > 0 || summary.outOfStock > 0) && (
          <div className="flex items-center gap-4 p-3 bg-amber-50 border border-amber-200 rounded-lg">
            <AlertTriangle className="size-5 text-amber-600" />
            <div className="flex gap-4 text-sm">
              {summary.outOfStock > 0 && (
                <span className="text-red-600 font-medium">
                  在庫切れ: {summary.outOfStock}件
                </span>
              )}
              {summary.lowStock > 0 && (
                <span className="text-amber-600 font-medium">
                  残少: {summary.lowStock}件
                </span>
              )}
            </div>
          </div>
        )}

        {/* Search & Filters */}
        <div className="flex items-center gap-3">
          <div className="flex-1">
            <SearchFilterBar
              searchTerm={searchTerm}
              onSearchChange={setSearchTerm}
              placeholder="品名、保管場所、仕入先..."
              count={filteredItems.length}
            />
          </div>
          <Select
            value={category}
            onValueChange={(v) => setCategory(v as CategoryFilter)}
          >
            <SelectTrigger className="w-[140px] h-10 bg-white">
              <SelectValue placeholder="カテゴリ" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全カテゴリ</SelectItem>
              <SelectItem value="medicine">医薬品</SelectItem>
              <SelectItem value="consumable">消耗品</SelectItem>
              <SelectItem value="food">フード</SelectItem>
              <SelectItem value="other">その他</SelectItem>
            </SelectContent>
          </Select>
          <Select
            value={statusFilter}
            onValueChange={(v) => setStatusFilter(v as StatusFilter)}
          >
            <SelectTrigger className="w-[140px] h-10 bg-white">
              <SelectValue placeholder="ステータス" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全ステータス</SelectItem>
              <SelectItem value="sufficient">十分</SelectItem>
              <SelectItem value="low">残少</SelectItem>
              <SelectItem value="out_of_stock">在庫切れ</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Table */}
        <DataTable
          columns={columns}
          data={filteredItems}
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
    </PageLayout>
  );
}
