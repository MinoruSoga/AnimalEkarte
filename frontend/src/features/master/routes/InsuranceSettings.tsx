import { useCallback } from "react";
import { Shield } from "lucide-react";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER, MASTER_TABLE_COL } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { InsuranceSidePanel } from "../components/InsuranceSidePanel";
import type { InsuranceFormData } from "../components/insurance-side-panel-model";
import { useGetAllInsurances, useCreateInsurance, useUpdateInsurance, useDeleteInsurance } from "../api/insurances";
import type { Insurance, CreateInsuranceRequest, UpdateInsuranceRequest } from "../api/insurances";
import {
  buildInsuranceCreateRequest,
  buildInsuranceUpdateRequest,
  validateInsuranceForm,
} from "./insurance-settings-model";
import { ResourceMasterInsurance } from "@/types/generated/models";

const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "補償率", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "連絡先", className: MASTER_TABLE_COL.w140 },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

export function InsuranceSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterInsurance);
  const { data } = useGetAllInsurances();
  const createMutation = useCreateInsurance();
  const updateMutation = useUpdateInsurance();
  const deleteMutation = useDeleteInsurance();
  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Insurance>({
    data,
    deleteMutation,
    entityLabel: "保険",
    dirtyGuard: dirty,
    permissions: { canDelete },
  });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);
  const { handleSave } = useMasterSave<Insurance, InsuranceFormData, CreateInsuranceRequest, UpdateInsuranceRequest>({
    crud, createMutation, updateMutation,
    permissions: { canCreate, canEdit },
    validate: validateInsuranceForm,
    toCreateRequest: buildInsuranceCreateRequest,
    toUpdateRequest: buildInsuranceUpdateRequest,
  });

  return (
    <MasterCRUDPage title="保険マスタ" icon={<Shield className={`${ICON.page} ${C.text}`} />} resource={ResourceMasterInsurance}
      entityLabel="保険" searchPlaceholder="保険名で検索..." emptyMessage="保険が登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={(item, onEdit, canEdit) => (
        <DataTableRow key={item.id}>
          <TableCell className={`font-medium ${C.text}`}>
            <DataTableRowButton
              aria-label={`詳細: 保険 ${item.name} (ID ${item.id})`}
              onClick={() => onEdit(item)}
            >
              {item.name}
            </DataTableRowButton>
          </TableCell>
          <TableCell className={`text-center ${C.text}`}>{item.coverageRate > 0 ? `${item.coverageRate}%` : "-"}</TableCell>
          <TableCell className={C.text70}>{item.contactPhone || "-"}</TableCell>
          <TableCell className="text-center"><StatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="text-right">
            {canEdit ? (
              <RowActionButton
                onClick={() => onEdit(item)}
                aria-label={`保険「${item.name}」(ID: ${item.id}) を編集`}
              />
            ) : null}
          </TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => <InsuranceSidePanel key={props.item?.id ?? "new"} {...props} onDirtyChange={handleDirtyChange} />}
    />
  );
}
