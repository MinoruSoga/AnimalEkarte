import { Building2, Plus } from "lucide-react";

import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { C, ICON } from "@/lib/design-tokens";
import { ResourceHospitalSettings } from "@/types/generated/models";
import type { Clinic } from "../api/clinics";

const COLUMNS = [
  { header: "院名" },
  { header: "電話番号", className: "w-[150px]" },
  { header: "メール" },
  { header: "ステータス", className: "w-[100px]", align: "right" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface ClinicMasterListProps {
  canCreate: boolean;
  canEdit: boolean;
  items: Clinic[];
  searchTerm: string;
  onSearchChange: (term: string) => void;
  onBack: () => void;
  onCreate: () => void;
  onEdit: (item: Clinic) => void;
}

export function ClinicMasterList({
  canCreate,
  canEdit,
  items,
  searchTerm,
  onSearchChange,
  onBack,
  onCreate,
  onEdit,
}: ClinicMasterListProps) {
  return (
    <PageLayout
      title="医院マスタ"
      resource={ResourceHospitalSettings}
      icon={<Building2 className={`${ICON.page} ${C.text}`} />}
      onBack={onBack}
      headerAction={
        canCreate ? (
          <PrimaryButton onClick={onCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規登録
          </PrimaryButton>
        ) : null
      }
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        <NotionFilter
          properties={[]}
          activeFilters={[]}
          onFilterChange={() => {}}
          searchTerm={searchTerm}
          onSearchChange={onSearchChange}
          searchPlaceholder="院名、電話番号、メールで検索..."
          count={items.length}
        />

        <DataTable
          columns={COLUMNS}
          data={items}
          emptyMessage="医院が登録されていません"
          renderRow={(item) => (
            <DataTableRow key={item.id} onClick={canEdit ? () => onEdit(item) : undefined}>
              <TableCell className={`font-medium text-sm ${C.text} py-2.5`}>
                {item.name}
              </TableCell>
              <TableCell className={`font-mono text-sm ${C.text80} py-2.5`}>
                {item.phoneNumber || "-"}
              </TableCell>
              <TableCell className={`text-sm ${C.text80} py-2.5`}>
                {item.email || "-"}
              </TableCell>
              <TableCell className="text-right py-2.5">
                <span className="inline-flex items-center gap-1.5">
                  <span className={`size-[7px] rounded-full ${item.isActive ? C.bgAccent : C.bgPrimary10}`} />
                  <span className={`text-sm ${item.isActive ? C.text65 : C.text35}`}>
                    {item.isActive ? "有効" : "無効"}
                  </span>
                </span>
              </TableCell>
              <TableCell className="p-0 text-right">
                {canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}
              </TableCell>
            </DataTableRow>
          )}
        />
        {canCreate ? (
          <button
            type="button"
            onClick={onCreate}
            className={`flex items-center gap-1.5 w-full px-3 py-2 text-sm ${C.text40} ${C.hoverText60} ${C.hoverBgLight} transition-colors rounded`}
          >
            <Plus className={ICON.xs} />
            新しい医院を追加...
          </button>
        ) : null}
      </div>
    </PageLayout>
  );
}
