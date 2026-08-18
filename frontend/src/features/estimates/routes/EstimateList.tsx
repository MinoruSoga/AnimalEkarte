import { ICON, C, LAYOUT } from "@/lib/design-tokens";
import { normalizeKana } from "@/lib/normalize-kana";
import { LoadingFallback } from "@/components/shared/DataStates";
import { useState, useMemo, useDeferredValue, useCallback, useTransition } from "react";
import { useNavigate } from "react-router";
import { usePermission } from "@/hooks/use-permission";
import { useModalState } from "@/hooks/use-modal-state";
import { usePagination } from "@/hooks/use-pagination";
import { formatCurrency } from "@/lib/format/number";
import { formatDate } from "@/lib/format/date";
import { Plus, FileText, Trash2, ExternalLink, CircleDot, Calendar } from "lucide-react";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { EstimateStatusBadge } from "../components/EstimateStatusBadge/EstimateStatusBadge";
import { paths } from "@/config/paths";
import { useGetEstimates } from "../api/get-estimates";
import { useDeleteEstimate } from "../api/delete-estimate";
import type { Estimate } from "../types";
import { isEstimateLockedStatus } from "../lib/is-estimate-locked-status";
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
  ActiveSort,
} from "@/components/shared/PropertyFilter/types";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/PropertyFilter/types";
import { ResourceEstimates } from "@/types/generated/models";

// rendering-hoist-jsx: 静的定義はモジュール定数に巻き上げ
const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "status",
    label: "ステータス",
    type: "select",
    icon: CircleDot,
    // estimates.status DEFAULT 'draft' — 空値は存在しない
    conditions: CONDITIONS_NO_EMPTY,
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


const COLUMNS = [
  { header: "見積番号", className: "w-[140px]" },
  { header: "タイトル" },
  { header: "飼主名", className: "w-[130px]" },
  { header: "有効期限", className: "w-[110px]" },
  { header: "合計金額", align: "right" as const },
  { header: "ステータス", className: "w-[100px]" },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

export function EstimateList() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission("estimates");
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);
  const deleteModal = useModalState<string>();

  const deferredSearch = useDeferredValue(searchTerm);

  const { data: result, isLoading, isError } = useGetEstimates();
  const { mutate: deleteEstimate } = useDeleteEstimate();
  const [isDeletePending, startDeleteTransition] = useTransition();

  // フィルタ + 検索 + ソートを適用
  const filtered = useMemo(() => {
    let items = [...(result?.data ?? [])];

    // ActiveFilter 適用
    for (const filter of activeFilters) {
      if (filter.key === "status" && typeof filter.value === "string") {
        items = items.filter((estimate) => {
          switch (filter.condition) {
            case "is":
              return estimate.status === filter.value;
            case "is_not":
              return estimate.status !== filter.value;
            case "is_empty":
              return !estimate.status;
            case "is_not_empty":
              return !!estimate.status;
            default:
              return estimate.status === filter.value;
          }
        });
      }
      if (filter.key === "validUntil" && typeof filter.value === "object" && !Array.isArray(filter.value)) {
        const dateVal = filter.value as { from?: string; to?: string };
        items = items.filter((estimate) => {
          if (!estimate.validUntil) return filter.condition === "is_empty";
          const d = estimate.validUntil.slice(0, 10);
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

    // テキスト検索（カタカナ・ひらがな非区別）
    if (deferredSearch) {
      const normalizedTerm = normalizeKana(deferredSearch).toLowerCase();
      items = items.filter(
        (estimate) =>
          normalizeKana(estimate.title).toLowerCase().includes(normalizedTerm) ||
          normalizeKana(estimate.ownerName ?? "").toLowerCase().includes(normalizedTerm) ||
          estimate.estimateNo.toLowerCase().includes(normalizedTerm),
      );
    }

    // ソート適用
    if (activeSorts.length > 0) {
      items.sort((a, b) => {
        for (const sort of activeSorts) {
          if (sort.key === "totalAmount") {
            const cmp = a.totalAmount - b.totalAmount;
            if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
          } else {
            const aVal = String(a[sort.key as keyof Estimate] ?? "");
            const bVal = String(b[sort.key as keyof Estimate] ?? "");
            const cmp = aVal.localeCompare(bVal, "ja");
            if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
          }
        }
        return 0;
      });
    }

    return items;
  }, [result?.data, activeFilters, deferredSearch, activeSorts]);

  // rerender-transitions: ページネーション状態管理
  const pagination = usePagination(filtered, {
    resetKey: deferredSearch,
  });

  // rerender-dependencies: deleteModal オブジェクトでなく primitive/安定参照を deps に
  const deleteItemId = deleteModal.item;
  const closeDeleteModal = deleteModal.close;
  const handleDeleteConfirm = useCallback(() => {
    if (deleteItemId == null) return;
    startDeleteTransition(() => {
      deleteEstimate(deleteItemId, {
        onSuccess: () => closeDeleteModal(),
      });
    });
  }, [deleteItemId, closeDeleteModal, deleteEstimate]);

  const handleSortChange = useCallback((sorts: ActiveSort[]) => {
    setActiveSorts(sorts);
  }, []);

  const openDeleteModal = deleteModal.open;
  // rerender-memo: renderRow を useCallback でメモ化（DataTable への参照を安定化）
  const renderRow = useCallback((estimate: Estimate) => {
    const isLocked = isEstimateLockedStatus(estimate.status);
    const showEdit = canEdit && !isLocked;
    const showDelete = canDelete && !isLocked;

    return (
      <DataTableRow key={estimate.id}>
        <TableCell className={`font-mono ${C.text60}`}>
          <DataTableRowLink
            to={paths.estimates.detail.getHref(estimate.id)}
            aria-label={`見積書「${estimate.estimateNo} / ${estimate.title}」(ID: ${estimate.id}) の詳細を開く`}
          >
            {estimate.estimateNo}
          </DataTableRowLink>
        </TableCell>
        <TableCell className={`${C.text} font-medium`}>{estimate.title}</TableCell>
        <TableCell className={C.text}>
          {estimate.ownerName ?? "-"}
          {estimate.petName ? (
            <span className={`block text-xs ${C.text50}`}>{estimate.petName}</span>
          ) : null}
        </TableCell>
        <TableCell className={C.text60}>
          {formatDate(estimate.validUntil)}
        </TableCell>
        <TableCell className={`text-right font-mono font-medium ${C.text}`}>
          {formatCurrency(estimate.totalAmount)}
        </TableCell>
        <TableCell>
          <EstimateStatusBadge status={estimate.status} />
        </TableCell>
        <TableCell className="text-right">
          {(canEdit || canDelete) ? (
            <RowActionDropdown
              ariaLabel={`見積書「${estimate.estimateNo} / ${estimate.title}」(ID: ${estimate.id}) の操作`}
              actions={[
                {
                  label: "詳細",
                  icon: ExternalLink,
                  onClick: () => navigate(paths.estimates.detail.getHref(estimate.id)),
                },
                ...(showEdit ? [{
                  label: "編集",
                  icon: FileText,
                  onClick: () => navigate(paths.estimates.edit.getHref(estimate.id)),
                }] : []),
                ...(showDelete ? [{
                  label: "削除",
                  icon: Trash2,
                  variant: "destructive" as const,
                  onClick: () => openDeleteModal(estimate.id),
                }] : []),
              ]}
            />
          ) : null}
        </TableCell>
      </DataTableRow>
    );
  }, [navigate, canEdit, canDelete, openDeleteModal]);

  if (isLoading) {
    return <LoadingFallback />;
  }
  if (isError) {
    return <div className={`p-4 ${C.danger}`}>データの取得に失敗しました</div>;
  }

  return (
    <PageLayout
      title="見積書管理"
      resource={ResourceEstimates}
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      headerAction={
        canCreate ? (
          <PrimaryButton colorVariant="primary" onClick={() => navigate(paths.estimates.new.getHref())}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規見積書登録
          </PrimaryButton>
        ) : null
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <div className="flex flex-col gap-4">
        <PropertyFilter
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
          headerRowClassName={DESIGN_TABLE_HEADER_ROW}
          headerCellClassName={DESIGN_TABLE_HEADER_CELL}
          columns={COLUMNS}
          data={pagination.paginatedData}
          emptyMessage="見積書が見つかりません"
          renderRow={renderRow}
        />

        {filtered.length > 0 ? (
          <Pagination
            currentPage={pagination.currentPage}
            totalPages={pagination.totalPages}
            totalCount={pagination.totalCount}
            startIndex={pagination.startIndex}
            endIndex={pagination.endIndex}
            onPageChange={pagination.goToPage}
            onPrev={pagination.prevPage}
            onNext={pagination.nextPage}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={deleteModal.isOpen}
        onClose={deleteModal.close}
        onConfirm={handleDeleteConfirm}
        title="見積書を削除しますか?"
        description="この操作は取り消せません。"
        confirmLabel="削除"
        variant="destructive"
        isPending={isDeletePending}
      />
    </PageLayout>
  );
}
