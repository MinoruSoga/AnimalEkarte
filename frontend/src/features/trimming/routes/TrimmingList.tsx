// React/Framework
import { useState } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, Scissors, AlertTriangle, Edit, Trash2 } from "lucide-react";
import { toast } from "sonner";

// Internal
import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { StatusBadge } from "@/components/shared/StatusBadge";
import { RowActionDropdown } from "@/components/shared/RowActionDropdown";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionDatePicker } from "@/components/shared/NotionDatePicker";
import { Pagination } from "@/components/shared/Pagination";
import { getTrimmingStatusColor } from "@/utils/status-helpers";
import { usePagination } from "@/hooks/usePagination";
import { useStaffValidation } from "@/hooks/useStaffValidation";
import type { TrimmingRecord } from "@/types";

// Relative
import { useTrimmingRecords } from "../hooks/useTrimmingRecords";

// ─── 静的データはモジュールレベルに配置 (rendering-hoist-jsx) ───────────────
const COLUMNS = [
  { header: "診療日", className: "w-[120px]" },
  { header: "飼主名" },
  { header: "ペット名" },
  { header: "種", className: "w-[80px]" },
  { header: "体重", className: "w-[80px]" },
  { header: "スタイル希望" },
  { header: "担当", className: "w-[100px]" },
  { header: "ステータス", className: "w-[100px]" },
  { header: "操作", className: "w-[100px]", align: "right" as const },
];

// ✅ React 19: function宣言を使用 (CLAUDE.md準拠)
export function TrimmingList() {
  const navigate = useNavigate();
  const [searchDate, setSearchDate] = useState({ from: "", to: "" });
  const [searchKeyword, setSearchKeyword] = useState("");
  const { data: filteredRecords, isLoading, error, deleteRecord } = useTrimmingRecords(searchKeyword, searchDate);
  const { isValidStaff } = useStaffValidation();

  // Pagination
  const {
    currentPage,
    paginatedData,
    totalPages,
    startIndex,
    endIndex,
    goToPage,
    prevPage,
    nextPage,
  } = usePagination(filteredRecords, { pageSize: 10 });

  // Delete confirmation
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; label: string } | null>(null);

  if (isLoading) return (
    <div className="flex justify-center items-center p-8">
      <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[#37352F]" />
    </div>
  );
  if (error) return <div className="p-4 text-red-600">データの取得に失敗しました</div>;

  const handleClear = () => {
    setSearchDate({ from: "", to: "" });
    setSearchKeyword("");
  };

  const handleEdit = (id: string) => {
    navigate(`/trimming/${id}`, { state: { from: "/trimming" } });
  };

  const handleDeleteClick = (record: TrimmingRecord) => {
    setDeleteTarget({
      id: record.id,
      label: `${record.ownerName} - ${record.petName}`,
    });
  };

  const handleDeleteConfirm = () => {
    if (deleteTarget) {
      deleteRecord(deleteTarget.id);
      toast.success("削除しました", { description: deleteTarget.label });
      setDeleteTarget(null);
    }
  };

  const handleNew = () => {
    navigate("/trimming/select-pet");
  };

  return (
    <PageLayout
      title="トリミング管理"
      icon={<Scissors className="size-5 text-[#37352F]" />}
      headerAction={
        <PrimaryButton onClick={handleNew}>
          <Plus className="mr-1.5 size-4" />
          新規登録
        </PrimaryButton>
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        {/* Filters */}
        <SearchFilterBar
          searchTerm={searchKeyword}
          onSearchChange={setSearchKeyword}
          placeholder="飼主名、ペット名..."
          count={filteredRecords.length}
        >
          <div className="flex items-center gap-2 w-full lg:w-auto flex-wrap">
            <div className="flex-1 lg:flex-none">
              <NotionDatePicker
                value={searchDate.from}
                onChange={(val) => setSearchDate({ ...searchDate, from: val })}
                placeholder="開始日"
              />
            </div>
            <span className="text-muted-foreground text-sm shrink-0">〜</span>
            <div className="flex-1 lg:flex-none">
              <NotionDatePicker
                value={searchDate.to}
                onChange={(val) => setSearchDate({ ...searchDate, to: val })}
                placeholder="終了日"
              />
            </div>
            <Button
              variant="outline"
              onClick={handleClear}
              className="text-[#37352F] h-10 text-sm px-4 ml-2 border-[rgba(55,53,47,0.16)] hover:bg-[#F7F6F3]"
            >
              クリア
            </Button>
          </div>
        </SearchFilterBar>

        {/* Table */}
        <DataTable
          columns={COLUMNS}
          data={paginatedData}
          renderRow={(record) => (
            <DataTableRow
              key={record.id}
              onClick={() => handleEdit(record.id)}
            >
              <TableCell className="font-mono text-sm text-[#37352F] py-2">
                {record.date}
              </TableCell>
              <TableCell className="text-sm text-[#37352F] py-2">{record.ownerName}</TableCell>
              <TableCell className="py-2">
                <div className="flex flex-col">
                  <span className="text-sm text-[#37352F]">{record.petName}</span>
                  <span className="text-sm text-[#37352F]/60">{record.petNumber}</span>
                </div>
              </TableCell>
              <TableCell className="text-sm text-[#37352F] py-2">{record.species}</TableCell>
              <TableCell className="text-sm text-[#37352F] py-2">{record.weight}</TableCell>
              <TableCell className="text-sm text-[#37352F] truncate max-w-[200px] py-2">
                {record.styleRequest}
              </TableCell>
              <TableCell className="text-sm text-[#37352F] py-2">
                <div className="flex items-center gap-1.5">
                  {!isValidStaff(record.staff) && (
                    <AlertTriangle className="size-4 text-amber-500" />
                  )}
                  {record.staff}
                </div>
              </TableCell>
              <TableCell className="py-2">
                <StatusBadge colorClass={getTrimmingStatusColor(record.status)}>
                  {record.status}
                </StatusBadge>
              </TableCell>
              <TableCell className="text-right py-2">
                <RowActionDropdown
                  actions={[
                    {
                      label: "編集",
                      icon: Edit,
                      onClick: () => handleEdit(record.id),
                    },
                    {
                      label: "削除",
                      icon: Trash2,
                      variant: "destructive",
                      onClick: () => handleDeleteClick(record),
                    },
                  ]}
                />
              </TableCell>
            </DataTableRow>
          )}
        />

        {/* Pagination */}
        <Pagination
          currentPage={currentPage}
          totalPages={totalPages}
          totalCount={filteredRecords.length}
          startIndex={startIndex}
          endIndex={endIndex}
          onPageChange={goToPage}
          onPrev={prevPage}
          onNext={nextPage}
        />
      </div>

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        title="削除確認"
        description={`${deleteTarget?.label} を削除してもよろしいですか？`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </PageLayout>
  );
}
