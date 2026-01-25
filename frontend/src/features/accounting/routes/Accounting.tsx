// React/Framework
import { useState } from "react";
import { useNavigate } from "react-router-dom";

// External
import { Plus, CreditCard } from "lucide-react";

// Internal
import { TableCell } from "../../../components/ui/table";
import { PageLayout } from "../../../components/shared/PageLayout";
import { SearchFilterBar } from "../../../components/shared/SearchFilterBar";
import { DataTable } from "../../../components/shared/DataTable";
import { PrimaryButton } from "../../../components/shared/PrimaryButton";
import { StatusBadge } from "../../../components/shared/StatusBadge";
import { DataTableRow } from "../../../components/shared/DataTableRow";
import { RowActionButton } from "../../../components/shared/RowActionButton";
import { getAccountingStatusColor } from "../../../lib/status-helpers";

// Relative
import { useAccountingRecords } from "../hooks/useAccountingRecords";

// Types
import type { Accounting as AccountingType } from "../types";

export const Accounting = () => {
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState("");
  const { data: filteredRecords } = useAccountingRecords(searchTerm);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("ja-JP", {
      style: "currency",
      currency: "JPY",
    }).format(amount);
  };

  const calculateTotal = (accounting: AccountingType) => {
    if (accounting.payment) return accounting.payment.totalAmount;
    
    return accounting.items.reduce((sum: number, item) => {
      const price = item.unitPrice * item.quantity;
      const tax = Math.floor(price * item.taxRate);
      return sum + price + tax;
    }, 0);
  };

  const handleCreate = () => {
    navigate("/accounting/select-pet");
  };

  const handleEdit = (id: string) => {
    navigate(`/accounting/${id}`);
  };

  const columns = [
    { header: "日時", className: "w-[140px]" },
    { header: "飼主名" },
    { header: "ペット名" },
    { header: "請求金額", align: "right" as const },
    { header: "支払方法", align: "center" as const },
    { header: "ステータス", className: "w-[100px]" },
    { header: "操作", className: "w-[100px]", align: "right" as const },
  ];

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
        {/* Filter & Search */}
        <SearchFilterBar
            searchTerm={searchTerm}
            onSearchChange={setSearchTerm}
            placeholder="飼主名、ペット名..."
            count={filteredRecords.length}
        />

        {/* Table */}
        <DataTable
            columns={columns}
            data={filteredRecords}
            emptyMessage="会計データが見つかりません"
            renderRow={(r) => (
                <DataTableRow 
                    key={r.id} 
                    onClick={() => handleEdit(r.id)}
                >
                    <TableCell className="font-mono text-sm text-[#37352F] py-2">{r.scheduledDate}</TableCell>
                    <TableCell className="text-sm text-[#37352F] py-2">{r.ownerName}</TableCell>
                    <TableCell className="text-sm text-[#37352F] py-2">{r.petName}</TableCell>
                    <TableCell className="text-right font-mono font-medium text-sm text-[#37352F] py-2">
                    {formatCurrency(calculateTotal(r))}
                    </TableCell>
                    <TableCell className="text-center text-sm text-[#37352F] py-2">{r.payment?.method || "-"}</TableCell>
                    <TableCell className="py-2">
                    <StatusBadge colorClass={getAccountingStatusColor(r.status)}>
                        {r.status}
                    </StatusBadge>
                    </TableCell>
                    <TableCell className="text-right py-2">
                        <RowActionButton onClick={() => handleEdit(r.id)} />
                    </TableCell>
                </DataTableRow>
            )}
        />
      </div>
    </PageLayout>
  );
}
