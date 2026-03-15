// React/Framework
import { type ReactNode, useState, useCallback, useDeferredValue } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, FileText, Edit, Trash2, Receipt } from "lucide-react";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { StatusBadge } from "@/components/shared/StatusBadge";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { SortableHeader } from "@/components/shared/SortableHeader/SortableHeader";
import { Pagination } from "@/components/shared/Pagination";
import { C, STYLE } from "@/lib/design-tokens";
import { getMedicalRecordStatusColor } from "@/utils/status-helpers";
import { usePagination } from "@/hooks/usePagination";
import { useTableSort } from "@/hooks/useTableSort";
import type { MedicalRecord } from "@/types";

// Relative
import { useMedicalRecords } from "../hooks/useMedicalRecords";
import { useDeleteMedicalRecord } from "../api/delete-medical-record";

type SortKey = "date" | "ownerName" | "petName" | "species" | "doctor" | "status";

type SortDirection = "ascending" | "descending" | "none";

type TableColumn = {
  header: ReactNode;
  className?: string;
  align?: "left" | "center" | "right";
  sortDirection?: SortDirection;
};

export function MedicalRecords() {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const deferredSearch = useDeferredValue(searchTerm);
  const { data: filteredRecords, isLoading, isError } = useMedicalRecords(deferredSearch);
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; label: string } | null>(null);
  const { mutate: deleteRecord } = useDeleteMedicalRecord();

  const accessor = useCallback(
    (item: MedicalRecord, key: SortKey): string => String(item[key] ?? ""),
    [],
  );

  const { sortedData, directionFor, toggleSort } = useTableSort<MedicalRecord, SortKey>(
    filteredRecords,
    { accessor },
  );

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

  const COLUMNS: TableColumn[] = [
    {
      header: <SortableHeader label="診療日" direction={directionFor("date")} onToggle={() => toggleSort("date")} />,
      className: "w-[120px]",
      sortDirection: directionFor("date"),
    },
    {
      header: <SortableHeader label="飼主名" direction={directionFor("ownerName")} onToggle={() => toggleSort("ownerName")} />,
      sortDirection: directionFor("ownerName"),
    },
    {
      header: <SortableHeader label="ペット名" direction={directionFor("petName")} onToggle={() => toggleSort("petName")} />,
      sortDirection: directionFor("petName"),
    },
    {
      header: <SortableHeader label="種" direction={directionFor("species")} onToggle={() => toggleSort("species")} />,
      className: "w-[80px]",
      sortDirection: directionFor("species"),
    },
    { header: "主訴" },
    { header: "関連", className: "w-[100px]" },
    {
      header: <SortableHeader label="担当医" direction={directionFor("doctor")} onToggle={() => toggleSort("doctor")} />,
      className: "w-[100px]",
      sortDirection: directionFor("doctor"),
    },
    {
      header: <SortableHeader label="ステータス" direction={directionFor("status")} onToggle={() => toggleSort("status")} />,
      className: "w-[100px]",
      sortDirection: directionFor("status"),
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
        <SearchFilterBar
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          placeholder="飼主名、ペット名、カルテNo、主訴で検索..."
          count={filteredRecords.length}
        />

        {/* Table */}
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
              <TableCell className={STYLE.tableCell}>{r.doctor}</TableCell>
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
