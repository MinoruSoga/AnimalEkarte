// React/Framework
import { useCallback, useDeferredValue, useMemo, useState } from "react";
import { useNavigate, useLoaderData } from "react-router";

// External
import { Plus, CreditCard, CircleDot, FileText } from "lucide-react";

// Internal
import { TableCell } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { getAccountingStatusColor } from "@/utils/status-helpers";
import { paths } from "@/config/paths";

// Types
import type { Accounting as AccountingType, AccountingStatus, PaymentMethod } from "../types";
import type { AccountingsLoaderData } from "../loaders";
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
  ActiveSort,
} from "@/components/shared/NotionFilter/types";

type SortKey = "scheduledDate" | "ownerName" | "petName" | "totalAmount" | "status";

// ── 静的定数（rendering-hoist-jsx）──────────────────────────
const PAYMENT_METHOD_LABELS: Record<PaymentMethod, string> = {
  cash: "現金",
  credit_card: "クレジットカード",
  electronic_money: "電子マネー",
};

const ACCOUNTING_STATUS_LABELS: Record<AccountingStatus, string> = {
  waiting: "会計待ち",
  pending: "会計待ち",
  completed: "会計済",
  cancelled: "キャンセル",
};

const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    options: [
      { value: "waiting", label: "会計待ち" },
      { value: "completed", label: "会計済" },
      { value: "cancelled", label: "キャンセル" },
    ],
  },
];

// rendering-hoist-jsx: 静的ソートプロパティ定義
const ACCOUNTING_SORT_PROPERTIES: SortProperty[] = [
  { key: "scheduledDate", label: "日時" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "totalAmount", label: "請求金額" },
  { key: "status", label: "ステータス" },
];

function formatCurrency(amount: number) {
  return new Intl.NumberFormat("ja-JP", {
    style: "currency",
    currency: "JPY",
  }).format(amount);
}

function calculateTotal(accounting: AccountingType) {
  if (accounting.payment) return accounting.payment.totalAmount;

  return accounting.items.reduce((sum: number, item) => {
    const price = item.unitPrice * item.quantity;
    const tax = Math.floor(price * item.taxRate);
    return sum + price + tax;
  }, 0);
}

export function Accounting() {
  const navigate = useNavigate();
  const { accountings } = useLoaderData<AccountingsLoaderData>();

  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearch;

  // js-cache-function-results: フィルタ結果を useMemo でキャッシュ
  const filteredRecords = useMemo(() => {
    let result = accountings;

    // ActiveFilter からフィルタ適用（condition 対応）
    const statusFilter = activeFilters.find((f) => f.key === "status");
    if (statusFilter && typeof statusFilter.value === "string") {
      result = result.filter((r) => {
        switch (statusFilter.condition) {
          case "is":
            return r.status === statusFilter.value;
          case "is_not":
            return r.status !== statusFilter.value;
          case "is_empty":
            return !r.status;
          case "is_not_empty":
            return !!r.status;
          default:
            return r.status === statusFilter.value;
        }
      });
    }

    // テキスト検索
    if (deferredSearch) {
      const lowerTerm = deferredSearch.toLowerCase();
      result = result.filter(
        (r) =>
          r.ownerName.toLowerCase().includes(lowerTerm) ||
          r.petName.toLowerCase().includes(lowerTerm),
      );
    }

    return result;
  }, [accountings, activeFilters, deferredSearch]);

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
    if (activeSorts.length === 0) return [...filteredRecords];
    const sorted = [...filteredRecords];
    sorted.sort((a, b) => {
      for (const sort of activeSorts) {
        const key = sort.key as SortKey;
        let aVal: string;
        let bVal: string;
        if (key === "totalAmount") {
          aVal = String(calculateTotal(a));
          bVal = String(calculateTotal(b));
          const numCmp = Number(aVal) - Number(bVal);
          if (numCmp !== 0) return sort.direction === "asc" ? numCmp : -numCmp;
          continue;
        }
        aVal = String(a[key] ?? "");
        bVal = String(b[key] ?? "");
        const cmp = aVal.localeCompare(bVal, "ja");
        if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
      }
      return 0;
    });
    return sorted;
  }, [filteredRecords, activeSorts]);

  const handleCreate = useCallback(() => {
    navigate(paths.accounting.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(paths.accounting.detail.getHref(id));
  }, [navigate]);

  const columns = useMemo(() => [
    {
      header: (
        <SortableHeader
          label="日時"
          direction={directionFor("scheduledDate")}
          onToggle={() => toggleSort("scheduledDate")}
        />
      ),
      className: "w-[140px]",
    },
    {
      header: (
        <SortableHeader
          label="飼主名"
          direction={directionFor("ownerName")}
          onToggle={() => toggleSort("ownerName")}
        />
      ),
    },
    {
      header: (
        <SortableHeader
          label="ペット名"
          direction={directionFor("petName")}
          onToggle={() => toggleSort("petName")}
        />
      ),
    },
    {
      header: (
        <SortableHeader
          label="請求金額"
          direction={directionFor("totalAmount")}
          onToggle={() => toggleSort("totalAmount")}
        />
      ),
      align: "right" as const,
    },
    { header: "支払方法", align: "center" as const },
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
    { header: "カルテ", className: "w-[80px]", align: "center" as const },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ], [directionFor, toggleSort]);

  // rerender-memo: renderRow を useCallback で安定化
  const renderRow = useCallback(
    (r: AccountingType) => {
      const statusLabel = ACCOUNTING_STATUS_LABELS[r.status] ?? r.status;
      return (
        <DataTableRow
          key={r.id}
          onClick={() => handleEdit(r.id)}
        >
          <TableCell className="font-mono text-base text-[#37352F] py-2">{r.scheduledDate}</TableCell>
          <TableCell className="text-base text-[#37352F] py-2">{r.ownerName}</TableCell>
          <TableCell className="text-base text-[#37352F] py-2">{r.petName}</TableCell>
          <TableCell className="text-right font-mono font-medium text-base text-[#37352F] py-2">
            {formatCurrency(calculateTotal(r))}
          </TableCell>
          <TableCell className="text-center text-base text-[#37352F] py-2">
            {r.payment ? PAYMENT_METHOD_LABELS[r.payment.method] : "-"}
          </TableCell>
          <TableCell className="py-2">
            <StatusBadge colorClass={getAccountingStatusColor(statusLabel)}>
              {statusLabel}
            </StatusBadge>
          </TableCell>
          <TableCell className="text-center py-2">
            {r.medicalRecordId ? (
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 text-blue-500 hover:text-blue-700 hover:bg-blue-50"
                onClick={(e) => {
                  e.stopPropagation();
                  navigate(paths.medicalRecords.detail.getHref(r.medicalRecordId!));
                }}
                aria-label="カルテを開く"
              >
                <FileText className="h-4 w-4" />
              </Button>
            ) : null}
          </TableCell>
          <TableCell className="text-right py-2">
            <RowActionButton onClick={() => handleEdit(r.id)} />
          </TableCell>
        </DataTableRow>
      );
    },
    [handleEdit, navigate],
  );

  return (
    <PageLayout
      title="会計管理"
      icon={<CreditCard className="size-4 text-[#37352F]" />}
      headerAction={
        <PrimaryButton onClick={handleCreate}>
          <Plus className="mr-1.5 size-4" />
          新規会計登録
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
          searchPlaceholder="飼主名、ペット名..."
          count={filteredRecords.length}
          sortProperties={ACCOUNTING_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={handleSortChange}
        />

        <div style={{ opacity: isFiltering ? 0.7 : 1, transition: "opacity 150ms" }}>
          <DataTable
            columns={columns}
            data={sortedData}
            emptyMessage="会計データが見つかりません"
            renderRow={renderRow}
          />
        </div>
      </div>
    </PageLayout>
  );
}
