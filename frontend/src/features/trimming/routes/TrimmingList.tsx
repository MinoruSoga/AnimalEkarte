import { useState } from "react";
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import { TableCell } from "../../../components/ui/table";
import { Plus, Scissors, Calendar } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { getTrimmingStatusColor } from "../../../lib/status-helpers";
import { useTrimmingRecords } from "../hooks/useTrimmingRecords";

import { PageLayout } from "../../../components/shared/PageLayout";
import { SearchFilterBar } from "../../../components/shared/SearchFilterBar";
import { DataTable } from "../../../components/shared/DataTable";
import { PrimaryButton } from "../../../components/shared/PrimaryButton";
import { StatusBadge } from "../../../components/shared/StatusBadge";
import { DataTableRow } from "../../../components/shared/DataTableRow";
import { RowActionButton } from "../../../components/shared/RowActionButton";

export const TrimmingList = () => {
  const navigate = useNavigate();
  const [searchDate, setSearchDate] = useState({ from: "", to: "" });
  const [searchKeyword, setSearchKeyword] = useState("");
  const { data: filteredRecords } = useTrimmingRecords(searchKeyword, searchDate);

  const handleClear = () => {
    setSearchDate({ from: "", to: "" });
    setSearchKeyword("");
  };

  const handleEdit = (id: string) => {
    navigate(`/trimming/${id}`);
  };

  const handleNew = () => {
    navigate("/trimming/select-pet");
  };

  const columns = [
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
            <div className="flex items-center gap-2 w-full lg:w-auto">
                <div className="relative flex-1 lg:flex-none">
                    <Calendar className="absolute left-2.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
                    <Input
                        type="date"
                        value={searchDate.from}
                        onChange={(e) =>
                        setSearchDate({ ...searchDate, from: e.target.value })
                        }
                        className="pl-9 w-full lg:w-[140px] bg-white h-10 text-sm"
                    />
                </div>
                <span className="text-muted-foreground text-sm">〜</span>
                <div className="relative flex-1 lg:flex-none">
                    <Calendar className="absolute left-2.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
                    <Input
                        type="date"
                        value={searchDate.to}
                        onChange={(e) =>
                        setSearchDate({ ...searchDate, to: e.target.value })
                        }
                        className="pl-9 w-full lg:w-[140px] bg-white h-10 text-sm"
                    />
                </div>
                <Button
                    variant="outline"
                    onClick={handleClear}
                    className="text-[#37352F] h-10 text-sm px-4 ml-2"
                >
                    クリア
                </Button>
            </div>
        </SearchFilterBar>

        {/* Table */}
        <DataTable
            columns={columns}
            data={filteredRecords}
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
                <TableCell className="text-sm text-[#37352F] py-2">{record.staff}</TableCell>
                <TableCell className="py-2">
                    <StatusBadge colorClass={getTrimmingStatusColor(record.status)}>
                        {record.status}
                    </StatusBadge>
                </TableCell>
                <TableCell className="text-right py-2">
                    <RowActionButton onClick={() => handleEdit(record.id)} />
                </TableCell>
                </DataTableRow>
            )}
        />
      </div>
    </PageLayout>
  );
}
