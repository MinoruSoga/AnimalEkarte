import { Building2, Plus } from "lucide-react";
import type { ReactNode } from "react";

import { TableCell } from "@/components/ui/table";
import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
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
  /** 医院一覧の上部に差し込む任意セクション（例: 法人情報/インボイス）。 */
  topSection?: ReactNode;
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
  topSection,
}: ClinicMasterListProps) {
  return (
    <PageLayout
      title="医院マスタ"
      resource={ResourceHospitalSettings}
      icon={<Building2 className={`${ICON.page} ${C.text}`} />}
      onBack={onBack}
      headerAction={
        canCreate ? (
          <PrimaryButton colorVariant="primary" onClick={onCreate}>
            <Plus className={`mr-1.5 ${ICON.action}`} />
            新規登録
          </PrimaryButton>
        ) : null
      }
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      <div className="flex flex-col gap-4">
        {topSection}
        <PropertyFilter
          properties={[]}
          activeFilters={[]}
          onFilterChange={() => {}}
          searchTerm={searchTerm}
          onSearchChange={onSearchChange}
          searchPlaceholder="院名、電話番号、メールで検索..."
          count={items.length}
        />

        <DataTable
          headerRowClassName={DESIGN_TABLE_HEADER_ROW}
          headerCellClassName={DESIGN_TABLE_HEADER_CELL}
          columns={COLUMNS}
          data={items}
          emptyMessage="医院が登録されていません"
          renderRow={(item) => (
            <DataTableRow key={item.id}>
              <TableCell className={`font-medium text-sm ${C.text}`}>
                {item.name}
              </TableCell>
              <TableCell className={`font-mono text-sm ${C.text80}`}>
                {item.phoneNumber || "-"}
              </TableCell>
              <TableCell className={`text-sm ${C.text80}`}>
                {item.email || "-"}
              </TableCell>
              <TableCell className="text-right">
                <span className="inline-flex items-center gap-1.5">
                  <span className={`size-[7px] rounded-full ${item.isActive ? C.bgBrandDot : C.bgPrimary10}`} />
                  <span className={`text-sm ${item.isActive ? C.text65 : C.text35}`}>
                    {item.isActive ? "有効" : "無効"}
                  </span>
                </span>
              </TableCell>
              <TableCell className="text-right">
                {canEdit ? (
                  <RowActionButton
                    onClick={() => onEdit(item)}
                    aria-label={`医院「${item.name}」(ID: ${item.id}) を編集`}
                  />
                ) : null}
              </TableCell>
            </DataTableRow>
          )}
        />
        {canCreate ? (
          <button
            type="button"
            onClick={onCreate}
            className={`flex min-h-11 items-center gap-1.5 w-full px-3 py-2 text-sm ${C.text40} ${C.hoverText60} ${C.hoverBgLight} transition-colors rounded`}
          >
            <Plus className={ICON.xs} />
            新しい医院を追加...
          </button>
        ) : null}
      </div>
    </PageLayout>
  );
}
