import { useCallback } from "react";
import { Bed } from "lucide-react";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C, ICON } from "@/lib/design-tokens";
import { formatCurrencyOrDash } from "@/lib/format/number";
import { MASTER_STATUS_FILTER, MASTER_TABLE_COL } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import {
  useGetAllHospitalizationPlans,
  useCreateHospitalizationPlan,
  useUpdateHospitalizationPlan,
  useDeleteHospitalizationPlan,
  BODY_SIZE_LABELS,
  BILLING_UNIT_LABELS,
} from "../api/hospitalization-plans";
import type {
  HospitalizationPlan,
  CreateHospitalizationPlanRequest,
  UpdateHospitalizationPlanRequest,
} from "../api/hospitalization-plans";
import { HospitalizationSidePanel } from "../components/HospitalizationSidePanel";
import type { HospitalizationFormData } from "../lib/hospitalization-side-panel-model";
import {
  buildHospitalizationCreateRequest,
  buildHospitalizationUpdateRequest,
} from "./hospitalization-settings-model";
import { ResourceMasterHospitalization } from "@/types/generated/models";

const COLUMNS = [
  { header: "名称" },
  { header: "対象体格", className: MASTER_TABLE_COL.w100 },
  { header: "料金単位", className: MASTER_TABLE_COL.w120 },
  { header: "単価(税込)", className: MASTER_TABLE_COL.w120, align: "right" as const },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

export function HospitalizationSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterHospitalization);
  const { data } = useGetAllHospitalizationPlans();
  const createMutation = useCreateHospitalizationPlan();
  const updateMutation = useUpdateHospitalizationPlan();
  const deleteMutation = useDeleteHospitalizationPlan();
  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<HospitalizationPlan>({
    data,
    deleteMutation,
    entityLabel: "入院プラン",
    dirtyGuard: dirty,
    permissions: { canDelete },
  });
  const handleDirtyChange = useCallback(
    (d: boolean) => {
      if (d) dirty.markDirty();
      else dirty.markClean();
    },
    [dirty],
  );
  const { handleSave } = useMasterSave<
    HospitalizationPlan,
    HospitalizationFormData,
    CreateHospitalizationPlanRequest,
    UpdateHospitalizationPlanRequest
  >({
    crud,
    createMutation,
    updateMutation,
    permissions: { canCreate, canEdit },
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: buildHospitalizationCreateRequest,
    toUpdateRequest: buildHospitalizationUpdateRequest,
  });

  return (
    <>
      <MasterCRUDPage
        title="入院マスタ"
        icon={<Bed className={`${ICON.page} ${C.text}`} />}
        resource={ResourceMasterHospitalization}
        entityLabel="入院プラン"
        searchPlaceholder="名称で検索..."
        emptyMessage="入院プランが登録されていません"
        crud={crud}
        handleSave={handleSave}
        columns={COLUMNS}
        filterProperties={[MASTER_STATUS_FILTER]}
        renderRow={(item, onEdit, canEdit) => (
          <DataTableRow key={item.id}>
            <TableCell className={`font-medium ${C.text}`}>
              <DataTableRowButton
                aria-label={`詳細: 入院プラン ${item.name} (ID ${item.id})`}
                onClick={() => onEdit(item)}
              >
                {item.name}
              </DataTableRowButton>
            </TableCell>
            <TableCell className={C.text70}>
              {item.bodySize ? (BODY_SIZE_LABELS[item.bodySize] ?? item.bodySize) : "-"}
            </TableCell>
            <TableCell className={C.text70}>
              {item.billingUnit ? (BILLING_UNIT_LABELS[item.billingUnit] ?? item.billingUnit) : "-"}
            </TableCell>
            <TableCell className={`text-right font-mono ${C.text}`}>
              {formatCurrencyOrDash(item.price)}
            </TableCell>
            <TableCell className="text-center">
              <StatusPill isActive={item.isActive} />
            </TableCell>
            <TableCell className="text-right">
              {canEdit ? (
                <RowActionButton
                  onClick={() => onEdit(item)}
                  aria-label={`入院プラン「${item.name}」(ID: ${item.id}) を編集`}
                />
              ) : null}
            </TableCell>
          </DataTableRow>
        )}
        renderSidePanel={(props) => (
          <HospitalizationSidePanel
            key={props.item?.id ?? "new"}
            {...props}
            onDirtyChange={handleDirtyChange}
          />
        )}
      />
      {dirty.discardDialog}
    </>
  );
}
