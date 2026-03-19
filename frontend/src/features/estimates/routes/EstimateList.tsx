import { useState, useMemo, useDeferredValue, useCallback } from "react";
import { useNavigate } from "react-router";
import { Plus, FileText, Trash2, ExternalLink, CircleDot, Calendar } from "lucide-react";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { EstimateStatusBadge } from "../components/EstimateStatusBadge/EstimateStatusBadge";
import { useGetEstimates } from "../api/get-estimates";
import { useDeleteEstimate } from "../api/delete-estimate";
import type { Estimate } from "../types";
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
  ActiveSort,
} from "@/components/shared/NotionFilter/types";

// rendering-hoist-jsx: 静的定義はモジュール定数に巻き上げ
const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    options: [
      { value: "draft", label: "下書き" },
      { value: "sent", label: "送付済み" },
      { value: "approved", label: "承認済み" },
      { value: "rejected", label: "却下" },
    ],
  },
  {
    key: "validUntil",
    label: "有効期限",
    type: "date-range",
    icon: Calendar,
  },
];

const SORT_PROPERTIES: SortProperty[] = [
  { key: "estimateNo", label: "見積番号" },
  { key: "title", label: "タイトル" },
  { key: "ownerName", label: "飼主名" },
  { key: "validUntil", label: "有効期限" },
  { key: "totalAmount", label: "合計金額" },
];

const formatCurrency = (amount: number) =>
  new Intl.NumberFormat("ja-JP", { style: "currency", currency: "JPY" }).format(amount);

const COLUMNS = [
  { header: "見積番号", className: "w-[140px]" },
  { header: "タイトル" },
  { header: "飼主名", className: "w-[130px]" },
  { header: "有効期限", className: "w-[110px]" },
  { header: "合計金額", align: "right" as const },
  { header: "ステータス", className: "w-[110px]" },
  { header: "操作", className: "w-[60px]", align: "right" as const },
];

export function EstimateList() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);
  const [deleteTargetId, setDeleteTargetId] = useState<string | null>(null);

  const deferredSearch = useDeferredValue(searchTerm);

  const { data: result, isLoading, isError } = useGetEstimates();
  const { mutate: deleteEstimate } = useDeleteEstimate();

  const estimates = result?.data ?? [];

  // フィルタ + 検索 + ソートを適用
  const filtered = useMemo(() => {
    let items = [...estimates];

    // ActiveFilter 適用
    for (const filter of activeFilters) {
      if (filter.key === "status" && typeof filter.value === "string") {
        items = items.filter((e) => {
          switch (filter.condition) {
            case "is":
              return e.status === filter.value;
            case "is_not":
              return e.status !== filter.value;
            case "is_empty":
              return !e.status;
            case "is_not_empty":
              return !!e.status;
            default:
              return e.status === filter.value;
          }
        });
      }
      if (filter.key === "validUntil" && typeof filter.value === "object" && !Array.isArray(filter.value)) {
        const dateVal = filter.value as { from?: string; to?: string };
        items = items.filter((e) => {
          if (!e.validUntil) return filter.condition === "is_empty";
          const d = e.validUntil.slice(0, 10);
          switch (filter.condition) {
            case "is":
              return dateVal.from ? d === dateVal.from : true;
            case "is_before":
              return dateVal.from ? d < dateVal.from : true;
            case "is_after":
              return dateVal.from ? d > dateVal.from : true;
            case "is_between":
              return (dateVal.from ? d >= dateVal.from : true) && (dateVal.to ? d <= dateVal.to : true);
            case "is_empty":
              return false;
            case "is_not_empty":
              return true;
            default:
              return true;
          }
        });
      }
    }

    // テキスト検索
    if (deferredSearch) {
      const lower = deferredSearch.toLowerCase();
      items = items.filter(
        (e) =>
          e.title.toLowerCase().includes(lower) ||
          (e.ownerName ?? "").toLowerCase().includes(lower) ||
          e.estimateNo.toLowerCase().includes(lower),
      );
    }

    // ソート適用
    if (activeSorts.length > 0) {
      items.sort((a, b) => {
        for (const sort of activeSorts) {
          let cmp = 0;
          if (sort.key === "totalAmount") {
            cmp = a.totalAmount - b.totalAmount;
          } else {
            const aVal = String((a as unknown as Record<string, unknown>)[sort.key] ?? "");
            const bVal = String((b as unknown as Record<string, unknown>)[sort.key] ?? "");
            cmp = aVal.localeCompare(bVal, "ja");
          }
          if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
        }
        return 0;
      });
    }

    return items;
  }, [estimates, activeFilters, deferredSearch, activeSorts]);

  const handleDeleteConfirm = useCallback(() => {
    if (deleteTargetId == null) return;
    deleteEstimate(deleteTargetId);
    setDeleteTargetId(null);
  }, [deleteTargetId, deleteEstimate]);

  const handleSortChange = useCallback((sorts: ActiveSort[]) => {
    setActiveSorts(sorts);
  }, []);

  const renderRow = (estimate: Estimate) => (
    <DataTableRow key={estimate.id} onClick={() => navigate(`/estimates/${estimate.id}`)}>
      <TableCell className="font-mono text-base text-[#37352F]/60 py-2">{estimate.estimateNo}</TableCell>
      <TableCell className="text-base text-[#37352F] py-2 font-medium">{estimate.title}</TableCell>
      <TableCell className="text-base text-[#37352F] py-2">{estimate.ownerName ?? "-"}</TableCell>
      <TableCell className="text-base text-[#37352F]/60 py-2">
        {estimate.validUntil ? estimate.validUntil.slice(0, 10) : "-"}
      </TableCell>
      <TableCell className="text-right font-mono font-medium text-base text-[#37352F] py-2">
        {formatCurrency(estimate.totalAmount)}
      </TableCell>
      <TableCell className="py-2">
        <EstimateStatusBadge status={estimate.status} />
      </TableCell>
      <TableCell className="text-right py-2">
        <RowActionDropdown
          actions={[
            {
              label: "詳細",
              icon: ExternalLink,
              onClick: () => navigate(`/estimates/${estimate.id}`),
            },
            {
              label: "編集",
              icon: FileText,
              onClick: () => navigate(`/estimates/${estimate.id}/edit`),
            },
            {
              label: "削除",
              icon: Trash2,
              variant: "destructive",
              onClick: () => setDeleteTargetId(estimate.id),
            },
          ]}
        />
      </TableCell>
    </DataTableRow>
  );

  if (isLoading) {
    return (
      <div className="flex justify-center items-center p-8">
        <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[#37352F]" />
      </div>
    );
  }
  if (isError) {
    return <div className="p-4 text-red-600">データの取得に失敗しました</div>;
  }

  return (
    <PageLayout
      title="見積書管理"
      icon={<FileText className="size-4 text-[#37352F]" />}
      headerAction={
        <PrimaryButton onClick={() => navigate("/estimates/new")}>
          <Plus className="mr-1.5 size-4" />
          新規見積書作成
        </PrimaryButton>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        <NotionFilter
          properties={FILTER_PROPERTIES}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder="見積番号、タイトル、飼主名..."
          count={filtered.length}
          sortProperties={SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={handleSortChange}
        />

        <DataTable
          columns={COLUMNS}
          data={filtered}
          emptyMessage="見積書が見つかりません"
          renderRow={renderRow}
        />
      </div>

      <ConfirmDialog
        open={deleteTargetId != null}
        onClose={() => setDeleteTargetId(null)}
        onConfirm={handleDeleteConfirm}
        title="見積書を削除しますか?"
        description="この操作は取り消せません。"
        confirmLabel="削除"
        variant="destructive"
      />
    </PageLayout>
  );
}
