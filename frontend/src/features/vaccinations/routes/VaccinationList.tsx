// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { useState, useDeferredValue, useCallback, useEffect, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";

// Hooks
import { useSortableData } from "@/hooks/use-sortable-data";

// External
import { Plus, Syringe, FileSpreadsheet, Calendar, User, Pencil, Trash2 } from "lucide-react";

// Internal
import { paths } from "@/config/paths";
import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown/RowActionDropdown";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { usePagination } from "@/hooks/use-pagination";
import { Pagination } from "@/components/shared/Pagination/Pagination";
import { FilteringIndicator } from "@/components/shared/FilteringIndicator/FilteringIndicator";

// Relative
import { useFilterVaccinations } from "../hooks/use-vaccinations";
import { useDeleteVaccination } from "../api/delete-vaccination";
import { usePermission } from "@/features/auth";

// Types
import type { VaccinationRecord } from "@/types";
import type {
  FilterProperty,
  ActiveFilter,
  SortProperty,
} from "@/components/shared/NotionFilter/types";
import { CONDITIONS_WITH_EMPTY } from "@/components/shared/NotionFilter/types";

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

export function VaccinationList() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { canCreate, canEdit, canDelete } = usePermission("vaccinations");
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null);
  const deleteMutation = useDeleteVaccination();
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

  const { data: filteredRecords, allVaccinations } = useFilterVaccinations(deferredSearchTerm, filters, activeFilters);

  // js-cache-function-results: ロード済みデータから担当医の選択肢を動的生成
  const filterProperties = useMemo<FilterProperty[]>(() => {
    const doctorOptions = Array.from(new Set(allVaccinations.map((r) => r.doctor).filter(Boolean)))
      .sort()
      .map((d) => ({ value: d, label: d }));
    return [
      ...STATIC_FILTER_PROPERTIES,
      // vaccinations.doctor_id nullable（未割当あり）
      { key: "doctor", label: "担当医", type: "select" as const, icon: User, conditions: CONDITIONS_WITH_EMPTY, options: doctorOptions },
    ];
  }, [allVaccinations]);

  const { activeSorts, setActiveSorts, toggleSort, directionFor, sortedData } =
    useSortableData(filteredRecords);

  const pagination = usePagination(sortedData, {
    pageSize: 20,
    resetKey: deferredSearchTerm,
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
    navigate(paths.vaccinations.selectPet.getHref());
  }, [navigate]);

  const handleEdit = useCallback((id: string) => {
    navigate(`/vaccinations/${id}`);
  }, [navigate]);

  const handleDeleteConfirm = useCallback(() => {
    if (!pendingDeleteId) return;
    deleteMutation.mutate(pendingDeleteId, {
      onSuccess: () => setPendingDeleteId(null),
    });
  }, [pendingDeleteId, deleteMutation]);

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
    const actions = [
      ...(canEdit ? [{ label: "編集", icon: Pencil, onClick: () => handleEdit(r.id) }] : []),
      ...(canDelete ? [{ label: "削除", icon: Trash2, onClick: () => setPendingDeleteId(r.id), variant: "destructive" as const }] : []),
    ];
    return (
      <DataTableRow key={r.id} onClick={() => handleEdit(r.id)}>
        <TableCell className={`font-mono text-base ${C.text} py-2`}>{r.date}</TableCell>
        <TableCell className={`text-base ${C.text} py-2`}>{r.ownerName}</TableCell>
        <TableCell className={`text-base ${C.text} py-2`}>{r.petName}</TableCell>
        <TableCell className={`text-base font-medium ${C.text} py-2`}>{r.vaccineName}</TableCell>
        <TableCell className={`font-mono text-base ${C.text} py-2`}>{r.nextDate}</TableCell>
        <TableCell className="text-right py-2">
          {/* BUG-089: 行操作ドロップダウン（編集・削除） */}
          {actions.length > 0 ? <RowActionDropdown actions={actions} /> : null}
        </TableCell>
      </DataTableRow>
    );
  }, [handleEdit, canEdit, canDelete]);

  return (
    <>
    <PageLayout
      title="予防接種管理"
      icon={<Syringe className={`${ICON.page} ${C.text}`} />}
      headerAction={
        <div className="flex items-center gap-2">
          <Button variant="outline" className="h-10 text-base gap-2 bg-white" onClick={() => {}}>
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
        <NotionFilter
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
    />
    </>
  );
}
