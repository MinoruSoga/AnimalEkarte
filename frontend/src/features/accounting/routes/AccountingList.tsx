// React/Framework
import { ICON, C } from "@/lib/design-tokens";
import { useCallback, useDeferredValue, useEffect, useMemo, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";

// External
import { Plus, CreditCard, CircleDot, FileText, Calendar, RotateCcw } from "lucide-react";

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
import { usePagination } from "@/hooks/use-pagination";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { formatCurrency } from "@/utils/format/number";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";

// Relative
import { useGetAccountings } from "../api/get-accountings";
import type { AccountingFilters } from "../api/get-accountings";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { Accounting as AccountingType, AccountingStatus, PaymentMethod } from "../types";
import { CONDITIONS_NO_EMPTY, CONDITIONS_WITH_EMPTY } from "@/components/shared/NotionFilter/types";
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
} from "@/components/shared/NotionFilter/types";
import { ResourceAccounting } from "@/types/generated/models";

// ── 静的定数（rendering-hoist-jsx）──────────────────────────
const PAYMENT_METHOD_LABELS: Record<PaymentMethod, string> = {
  cash: "現金",
  credit_card: "クレジットカード",
  electronic_money: "電子マネー",
};

const ACCOUNTING_STATUS_LABELS: Record<AccountingStatus, string> = {
  waiting: "会計待ち",
  pending: "会計待ち", // ENUM に存在するが billing service では使用しない（表示用のフォールバックとして保持）
  completed: "会計済",
  cancelled: "キャンセル",
};

const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    // billing service が実際にセットする値は waiting/completed/cancelled の3値のみ。
    // pending は ENUM 定義にあるが service 層で使われていないため選択肢から除外。
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "waiting", label: "会計待ち" },
      { value: "completed", label: "会計済" },
      { value: "cancelled", label: "キャンセル" },
    ],
  },
  {
    key: "paymentMethod",
    label: "支払方法",
    type: "select",
    icon: CreditCard,
    // payments レコードは会計完了時に作成されるため、会計待ち中は payment が存在しない（空）。
    // 「支払方法 = 空」= 未会計の絞り込みとして有効。
    conditions: CONDITIONS_WITH_EMPTY,
    options: [
      { value: "cash", label: "現金" },
      { value: "credit_card", label: "クレジットカード" },
      { value: "electronic_money", label: "電子マネー" },
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


function calculateTotal(accounting: AccountingType) {
  if (accounting.payment) return accounting.payment.totalAmount;

  return accounting.items.reduce((sum: number, item) => {
    const price = item.unitPrice * item.quantity;
    const tax = Math.floor(price * item.taxRate);
    return sum + price + tax;
  }, 0);
}

export function AccountingList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit } = usePermission("accounting");

  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearch;

  // activeFilters から日付フィルタのみを抽出してAPIに渡す
  // ステータスフィルタは is_not/is_empty/is_not_empty 条件があるためクライアントサイドのまま維持
  const apiFilters = useMemo<AccountingFilters>(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
    };
  }, [activeFilters]);

  const { data: accountings = [], isLoading, isError } = useGetAccountings(apiFilters);

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

    // 支払方法フィルタ（クライアントサイド）
    const paymentMethodFilter = activeFilters.find((f) => f.key === "paymentMethod");
    if (paymentMethodFilter && typeof paymentMethodFilter.value === "string") {
      result = result.filter((r) => {
        const method = r.payment?.method ?? "";
        switch (paymentMethodFilter.condition) {
          case "is":           return method === paymentMethodFilter.value;
          case "is_not":       return method !== paymentMethodFilter.value;
          case "is_empty":     return !method;
          case "is_not_empty": return !!method;
          default:             return method === paymentMethodFilter.value;
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

  const getSortValue = useCallback((item: AccountingType, key: string) => {
    if (key === "totalAmount") return calculateTotal(item);
    return String(item[key as keyof AccountingType] ?? "");
  }, []);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords, { getSortValue });

  const pagination = usePagination(sortedData, {
    pageSize: 20,
    resetKey: deferredSearch,
  });

  // FE-144: URLクエリパラメータからページ番号を読み取る
  const urlPage = Number(searchParams.get("page") ?? 1);

  // FE-144: URLのページ番号とローカル状態を同期（URLが変わったときのみ）
  // rerender-dependencies: pagination（オブジェクト）を destructure し primitive を deps に使用
  const { totalPages, currentPage, goToPage } = pagination;
  useEffect(() => {
    const clampedPage = Math.max(1, Math.min(urlPage, totalPages));
    if (clampedPage !== currentPage) {
      goToPage(clampedPage);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlPage, totalPages]);

  // FE-144: ページ変更時にURLクエリパラメータを更新
  const handlePageChange = useCallback((page: number) => {
    goToPage(page);
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (page === 1) {
        next.delete("page");
      } else {
        next.set("page", String(page));
      }
      return next;
    }, { replace: true });
  }, [goToPage, setSearchParams]);

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
      className: "w-[120px]",
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
    { header: "カルテ", className: "w-[80px] hidden lg:table-cell", align: "center" as const },
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
          <TableCell className={`font-mono text-base ${C.text} py-2`}>{r.scheduledDate}</TableCell>
          <TableCell className={`text-base ${C.text} py-2`}>{r.ownerName}</TableCell>
          <TableCell className={`text-base ${C.text} py-2`}>{r.petName}</TableCell>
          <TableCell className={`text-right font-mono font-medium text-base ${C.text} py-2`}>
            {formatCurrency(calculateTotal(r))}
          </TableCell>
          <TableCell className={`text-center text-base ${C.text} py-2`}>
            {r.payment ? PAYMENT_METHOD_LABELS[r.payment.method] : "-"}
          </TableCell>
          <TableCell className="py-2">
            <div className="flex flex-wrap gap-1 items-center">
              <StatusBadge colorClass={getAccountingStatusColor(statusLabel)}>
                {statusLabel}
              </StatusBadge>
              {r.totalRefundedAmount > 0 ? (
                <span className={`inline-flex items-center gap-0.5 text-[10px] font-medium px-1.5 py-0.5 rounded ${C.bgDiscountLight} ${C.textDiscount} ${C.borderOrangeBadge}`}>
                  <RotateCcw className={ICON.action} />
                  返金あり
                </span>
              ) : null}
            </div>
          </TableCell>
          <TableCell className="text-center py-2 hidden lg:table-cell">
            {r.medicalRecordId ? (
              <Button
                variant="ghost"
                size="icon"
                className={`h-8 w-8 ${C.accent} ${C.hoverTextAccent} ${C.hoverBgAccent5}`}
                onClick={(e) => {
                  e.stopPropagation();
                  navigate(paths.medicalRecords.detail.getHref(r.medicalRecordId!));
                }}
                aria-label="カルテを開く"
              >
                <FileText className={ICON.action} />
              </Button>
            ) : null}
          </TableCell>
          <TableCell className="text-right py-2">
            {canEdit ? <RowActionButton onClick={() => handleEdit(r.id)} /> : null}
          </TableCell>
        </DataTableRow>
      );
    },
    [handleEdit, navigate, canEdit],
  );

  // 全フック呼び出し後に早期リターン（Rules of Hooks 準拠）
  if (isLoading) return <LoadingFallback />;
  if (isError) return <ErrorFallback />;

  return (
    <PageLayout
      title="会計管理"
      resource={ResourceAccounting}
      icon={<CreditCard className={`${ICON.page} ${C.text}`} />}
      headerAction={
        canCreate ? (
          <PrimaryButton onClick={handleCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規会計登録
          </PrimaryButton>
        ) : null
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
          onSortChange={setActiveSorts}
        />

        <FilteringIndicator isFiltering={isFiltering}>
          <DataTable
            columns={columns}
            data={pagination.paginatedData}
            emptyMessage="会計データが見つかりません"
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
