// React/Framework
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { useState, useDeferredValue, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useTransition } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";
import { uniqueSortedOptions } from "@/lib/unique-sorted-options";
import { isPastJSTDate } from "@/lib/jst-date";

// External
import { Plus, Syringe, Calendar, User, Pencil, Trash2, AlertTriangle } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowLink } from "@/components/shared/DataTable/DataTableRowLink";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown/RowActionDropdown";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { usePagination } from "@/hooks/use-pagination";
import { useGetPet } from "@/hooks/use-pet";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";

// Relative
import { useFilterVaccinations } from "../hooks/use-vaccinations";
import { useDeleteVaccination } from "../api/delete-vaccination";
import { usePermission } from "@/hooks/use-permission";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";

// Types
import type { VaccinationRecord } from "@/types";
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
} from "@/components/shared/PropertyFilter/types";
import { CONDITIONS_WITH_EMPTY } from "@/components/shared/PropertyFilter/types";
import { ResourceVaccinations } from "@/types/generated/models";

// rendering-hoist-jsx: 静的フィルタプロパティ（担当医は動的オプションのためコンポーネント内で構築）
const STATIC_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
];

// rendering-hoist-jsx: 静的ソートプロパティ定義
const VACCINATION_SORT_PROPERTIES: SortProperty[] = [
  { key: "date", label: "実施日" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "vaccineName", label: "予防接種名" },
  { key: "nextDate", label: "次回予定" },
];

interface VaccinationRowActionsProps {
  record: VaccinationRecord;
  canEdit: boolean;
  canDelete: boolean;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
}

function VaccinationRowActions({
  record,
  canEdit,
  canDelete,
  onEdit,
  onDelete,
}: VaccinationRowActionsProps) {
  const { data: pet } = useGetPet(record.petId ?? "");
  const actions = [
    ...(canEdit && pet?.status === "生存" ? [{
      label: "編集",
      icon: Pencil,
      onClick: () => onEdit(record.id),
    }] : []),
    ...(canDelete ? [{
      label: "削除",
      icon: Trash2,
      onClick: () => onDelete(record.id),
      variant: "destructive" as const,
    }] : []),
  ];

  return actions.length > 0 ? (
    <RowActionDropdown
      actions={actions}
      ariaLabel={`予防接種操作: ${record.petName} ${record.date} ID ${record.id}`}
    />
  ) : null;
}

export function VaccinationList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit, canDelete } = usePermission("vaccinations");
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null);
  const { mutate: deleteVaccinationFn } = useDeleteVaccination();
  const [isDeletePending, startDeleteTransition] = useTransition();
  const canDeleteRef = useRef(canDelete);
  useLayoutEffect(() => {
    canDeleteRef.current = canDelete;
  }, [canDelete]);
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearchTerm = useDeferredValue(searchTerm);
  const isFiltering = searchTerm !== deferredSearchTerm;

  // activeFilters から日付フィルタを抽出
  const filters = useMemo(() => {
    const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
      | { from?: string; to?: string }
      | undefined;
    return {
      startDate: dateFilter?.from,
      endDate: dateFilter?.to,
    };
  }, [activeFilters]);

  const { data: filteredRecords, allVaccinations, isLoading, error } = useFilterVaccinations(deferredSearchTerm, filters, activeFilters);
  const pendingDeletePetId = allVaccinations.find(
    (record) => record.id === pendingDeleteId,
  )?.petId ?? "";
  const { data: pendingDeletePet } = useGetPet(pendingDeletePetId);

  // js-cache-function-results: ロード済みデータから担当医の選択肢を動的生成
  const filterProperties = useMemo<FilterProperty[]>(() => {
    const doctorOptions = uniqueSortedOptions(allVaccinations, (r) => r.doctor);
    return [
      ...STATIC_FILTER_PROPERTIES,
      // vaccinations.doctor_id nullable（未割当あり）
      { key: "doctor", label: "担当医", type: "select" as const, icon: User, conditions: CONDITIONS_WITH_EMPTY, options: doctorOptions },
    ];
  }, [allVaccinations]);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords);

  const pagination = usePagination(sortedData, {
    resetKey: deferredSearchTerm,
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
    navigate(paths.vaccinations.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(paths.vaccinations.detail.getHref(id));
  }, [navigate]);

  const handleDeleteConfirm = useCallback(() => {
    if (
      canDeleteRef.current !== true ||
      !pendingDeleteId ||
      pendingDeletePet?.status !== "生存"
    ) {
      return;
    }
    startDeleteTransition(() => {
      deleteVaccinationFn(pendingDeleteId, {
        onSuccess: () => {
          toast.success("予防接種記録を削除しました");
          setPendingDeleteId(null);
        },
        onError: (error) => {
          handleApiError(error, "削除");
        },
      });
    });
  }, [pendingDeleteId, pendingDeletePet?.status, deleteVaccinationFn]);

  const columns = useMemo(() => [
    {
      header: (
        <SortableHeader
          label="実施日"
          direction={directionFor("date")}
          onToggle={() => toggleSort("date")}
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
          label="予防接種名"
          direction={directionFor("vaccineName")}
          onToggle={() => toggleSort("vaccineName")}
        />
      ),
    },
    {
      header: (
        <SortableHeader
          label="次回予定"
          direction={directionFor("nextDate")}
          onToggle={() => toggleSort("nextDate")}
        />
      ),
      className: "w-[140px]",
    },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ], [directionFor, toggleSort]);

  // rerender-memo: renderRow を useCallback でメモ化（DataTable への参照を安定化）
  const renderRow = useCallback((r: VaccinationRecord) => {
    const overdue = isPastJSTDate(r.nextDate);
    return (
      <DataTableRow key={r.id}>
        <TableCell className={`font-mono ${C.text}`}>{r.date}</TableCell>
        <TableCell className={C.text}>{r.ownerName}</TableCell>
        <TableCell className={C.text}>
          <DataTableRowLink
            to={paths.vaccinations.detail.getHref(r.id)}
            aria-label={`予防接種詳細: ${r.petName} ${r.date} ID ${r.id}`}
          >
            {r.petName}
          </DataTableRowLink>
        </TableCell>
        <TableCell className={`font-medium ${C.text}`}>{r.vaccineName}</TableCell>
        <TableCell className={`font-mono ${overdue ? C.danger : C.text}`}>
          {overdue ? (
            <span className="inline-flex items-center gap-1.5">
              <AlertTriangle className={`${ICON.xs} shrink-0`} />
              <span>
                {r.nextDate}
                <span className="ml-1.5 text-xs font-medium">（期限超過）</span>
              </span>
            </span>
          ) : r.nextDate}
        </TableCell>
        <TableCell className="text-right">
          {/* BUG-089: 行操作ドロップダウン（編集・削除） */}
          {canEdit || canDelete ? (
            <VaccinationRowActions
              record={r}
              canEdit={canEdit}
              canDelete={canDelete}
              onEdit={handleEdit}
              onDelete={setPendingDeleteId}
            />
          ) : null}
        </TableCell>
      </DataTableRow>
    );
  }, [handleEdit, canEdit, canDelete]);

  if (isLoading) return <LoadingFallback />;
  if (error) return <ErrorFallback />;

  return (
    <>
    <PageLayout
      title="予防接種管理"
      resource={ResourceVaccinations}
      icon={<Syringe className={`${ICON.page} ${C.text}`} />}
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
        <PropertyFilter
          properties={filterProperties}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder="飼主名、ペット名、予防接種名..."
          count={filteredRecords.length}
          sortProperties={VACCINATION_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={setActiveSorts}
        />

        <FilteringIndicator isFiltering={isFiltering}>
          <DataTable
            headerRowClassName={DESIGN_TABLE_HEADER_ROW}
            headerCellClassName={DESIGN_TABLE_HEADER_CELL}
            columns={columns}
            data={pagination.paginatedData}
            emptyMessage="データが見つかりません"
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

    {/* BUG-089: 削除確認ダイアログ */}
    <ConfirmDialog
      open={pendingDeleteId !== null}
      onClose={() => setPendingDeleteId(null)}
      title="予防接種記録を削除しますか？"
      description="この操作は取り消せません。"
      confirmLabel="削除"
      variant="destructive"
      onConfirm={handleDeleteConfirm}
      isPending={isDeletePending}
    />
    </>
  );
}
