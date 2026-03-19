// React/Framework
import { type ReactNode, useState, useCallback, useMemo, useDeferredValue } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, FileText, Edit, Trash2, Receipt, AlertTriangle } from "lucide-react";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { StatusBadge } from "@/components/shared/StatusBadge/StatusBadge";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { Pagination } from "@/components/shared/Pagination";
import { C, STYLE } from "@/lib/design-tokens";
import { getMedicalRecordStatusColor } from "@/utils/status-helpers";
import { usePagination } from "@/hooks/use-pagination";
import { useStaffValidation } from "@/hooks/use-staff-validation";

// Relative
import { useFilterMedicalRecords } from "../hooks/use-medical-records";
import { useDeleteMedicalRecord } from "../api/delete-medical-record";

// Types
import type {
  SortProperty,
  ActiveSort,
} from "@/components/shared/NotionFilter/types";

type SortKey = "date" | "ownerName" | "petName" | "species" | "doctor" | "status";

// rendering-hoist-jsx: 静的ソートプロパティ定義
const MEDICAL_RECORD_SORT_PROPERTIES: SortProperty[] = [
  { key: "date", label: "診療日" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "species", label: "種" },
  { key: "doctor", label: "担当医" },
  { key: "status", label: "ステータス" },
];

export function MedicalRecords() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);
  const { data: filteredRecords, isLoading, isError } = useFilterMedicalRecords(deferredSearch);
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; label: string } | null>(null);
  const { mutate: deleteRecord } = useDeleteMedicalRecord();
  const { isValidStaff } = useStaffValidation();
  const [activeSorts, setActiveSorts] = useState<ActiveSort[]>([]);

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
        const aVal = String(a[key] ?? "");
        const bVal = String(b[key] ?? "");
        const cmp = aVal.localeCompare(bVal, "ja");
        if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
      }
      return 0;
    });
    return sorted;
  }, [filteredRecords, activeSorts]);

  const {
    paginatedData,
    currentPage,
    totalPages,
    totalCount,
    startIndex,
    endIndex,
    goToPage,
    nextPage,
    prevPage,
  } = usePagination(sortedData, { pageSize: 20, resetKey: searchTerm });

  const isFiltering = searchTerm !== deferredSearch;

  const handleNavigateToForm = useCallback((recordId?: string) => {
    navigate(
      recordId ? `/medical-records/${recordId}` : "/medical-records/select-pet",
      { state: { from: "/medical-records" } },
    );
  }, [navigate]);

  if (isLoading) return (
    <div className="flex justify-center items-center p-8">
      <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[#37352F]" />
    </div>
  );
  if (isError) return <div className="p-4 text-red-600">データの取得に失敗しました</div>;

  const COLUMNS: { header: ReactNode; className?: string; align?: "left" | "center" | "right" }[] = [
    {
      header: <SortableHeader label="診療日" direction={directionFor("date")} onToggle={() => toggleSort("date")} />,
      className: "w-[120px]",
    },
    {
      header: <SortableHeader label="飼主名" direction={directionFor("ownerName")} onToggle={() => toggleSort("ownerName")} />,
    },
    {
      header: <SortableHeader label="ペット名" direction={directionFor("petName")} onToggle={() => toggleSort("petName")} />,
    },
    {
      header: <SortableHeader label="種" direction={directionFor("species")} onToggle={() => toggleSort("species")} />,
      className: "w-[80px]",
    },
    { header: "主訴" },
    { header: "関連", className: "w-[100px]" },
    {
      header: <SortableHeader label="担当医" direction={directionFor("doctor")} onToggle={() => toggleSort("doctor")} />,
      className: "w-[100px]",
    },
    {
      header: <SortableHeader label="ステータス" direction={directionFor("status")} onToggle={() => toggleSort("status")} />,
      className: "w-[100px]",
    },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ];

  return (
    <PageLayout
      title="カルテ管理"
      icon={<FileText className={`size-5 ${C.text}`} />}
      headerAction={
        <PrimaryButton onClick={() => handleNavigateToForm()}>
          <Plus className="size-4" />
          新規カルテ作成
        </PrimaryButton>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4 flex-1 min-h-0">
        {/* Search */}
        <NotionFilter
          properties={[]}
          activeFilters={[]}
          onFilterChange={() => {}}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder="飼主名、ペット名、カルテNo、主訴で検索..."
          count={filteredRecords.length}
          sortProperties={MEDICAL_RECORD_SORT_PROPERTIES}
          activeSorts={activeSorts}
          onSortChange={handleSortChange}
        />

        {/* Table */}
        <div className={isFiltering ? "opacity-60 transition-opacity duration-150" : "transition-opacity duration-150"}>
          <DataTable
            columns={COLUMNS}
            data={paginatedData}
            emptyMessage="カルテデータが見つかりません"
            renderRow={(r) => (
              <DataTableRow
                key={r.id}
                onClick={() => handleNavigateToForm(r.id)}
              >
                <TableCell className={STYLE.tableCellMono}>{r.date}</TableCell>
                <TableCell className={STYLE.tableCell}>{r.ownerName}</TableCell>
                <TableCell className={STYLE.tableCell}>{r.petName}</TableCell>
                <TableCell className={STYLE.tableCell}>{r.species}</TableCell>
                <TableCell className={`text-sm ${C.text} max-w-[200px] truncate py-2`} title={r.chiefComplaint}>
                  {r.chiefComplaint}
                </TableCell>
                <TableCell className="py-2">
                  {r.accountingId ? (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        navigate(`/accounting/${r.accountingId}`);
                      }}
                      className={`inline-flex items-center gap-1 text-xs font-medium px-2 py-1 rounded-[3px] border ${C.textSuccess} bg-[#10B981]/10 border-[#10B981]/30 hover:bg-[#10B981]/20 transition-colors`}
                    >
                      <Receipt className="size-3" />
                      会計
                    </button>
                  ) : (
                    <span className={`text-sm ${C.text40}`}>—</span>
                  )}
                </TableCell>
                <TableCell className={STYLE.tableCell}>
                  <div className="flex items-center gap-1">
                    <span className={!isValidStaff(r.doctor) ? "text-red-500 font-medium" : ""}>
                      {r.doctor}
                    </span>
                    {!isValidStaff(r.doctor) ? (
                      <span title="担当医が無効（退職等）に設定されています"><AlertTriangle className="size-3.5 text-red-500" /></span>
                    ) : null}
                  </div>
                </TableCell>
                <TableCell className="py-2">
                  <StatusBadge colorClass={getMedicalRecordStatusColor(r.status)}>
                    {r.status}
                  </StatusBadge>
                </TableCell>
                <TableCell className="text-right py-2">
                  <RowActionDropdown
                    actions={[
                      {
                        label: "編集",
                        icon: Edit,
                        onClick: () => handleNavigateToForm(r.id),
                      },
                      {
                        label: "削除",
                        icon: Trash2,
                        onClick: () =>
                          setDeleteTarget({
                            id: r.id,
                            label: `${r.recordNo} ${r.petName}`,
                          }),
                        variant: "destructive",
                      },
                    ]}
                  />
                </TableCell>
              </DataTableRow>
            )}
          />
        </div>

        {totalPages > 1 ? (
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            totalCount={totalCount}
            startIndex={startIndex}
            endIndex={endIndex}
            onPageChange={goToPage}
            onPrev={prevPage}
            onNext={nextPage}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={deleteTarget !== null}
        onClose={() => setDeleteTarget(null)}
        onConfirm={() => {
          if (deleteTarget) {
            deleteRecord(deleteTarget.id);
          }
          setDeleteTarget(null);
        }}
        title="カルテを削除しますか？"
        description={`「${deleteTarget?.label ?? ""}」を削除します。関連する治療・検査データも削除されます。この操作は元に戻せません。`}
        confirmLabel="削除"
        variant="destructive"
      />
    </PageLayout>
  );
}
