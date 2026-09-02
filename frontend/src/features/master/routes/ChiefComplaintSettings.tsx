import { useCallback } from "react";
import { MessageSquareText } from "lucide-react";
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
import { ChiefComplaintSidePanel } from "../components/ChiefComplaintSidePanel";
import type { ChiefComplaintFormData } from "../components/chief-complaint-side-panel-model";
import { useGetChiefComplaintTypes, useCreateChiefComplaintType, useUpdateChiefComplaintType, useDeleteChiefComplaintType } from "../api/chief-complaint-types";
import type {
  ChiefComplaintType,
  CreateChiefComplaintTypeRequest,
  UpdateChiefComplaintTypeRequest,
} from "../api/chief-complaint-types";
import {
  buildChiefComplaintCreateRequest,
  buildChiefComplaintUpdateRequest,
} from "./chief-complaint-settings-model";
import { ResourceMasterMedical } from "@/types/generated/models";

const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "説明", className: "flex-1" },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

export function ChiefComplaintSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);
  const { data } = useGetChiefComplaintTypes();
  const createMutation = useCreateChiefComplaintType();
  const updateMutation = useUpdateChiefComplaintType();
  const deleteMutation = useDeleteChiefComplaintType();
  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<ChiefComplaintType>({
    data,
    deleteMutation,
    entityLabel: "主訴マスタ",
    dirtyGuard: dirty,
    permissions: { canDelete },
  });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);
  const { handleSave } = useMasterSave<ChiefComplaintType, ChiefComplaintFormData, CreateChiefComplaintTypeRequest, UpdateChiefComplaintTypeRequest>({
    crud, createMutation, updateMutation,
    permissions: { canCreate, canEdit },
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: buildChiefComplaintCreateRequest,
    toUpdateRequest: buildChiefComplaintUpdateRequest,
  });

  return (
    <>
    <MasterCRUDPage title="主訴マスタ" icon={<MessageSquareText className={`${ICON.page} ${C.text}`} />} resource={ResourceMasterMedical}
      entityLabel="主訴マスタ" searchPlaceholder="名称で検索..." emptyMessage="主訴マスタが登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={(item, onEdit, canEdit) => (
        <DataTableRow key={item.id}>
          <TableCell className={`font-medium ${C.text}`}>
            <DataTableRowButton
              aria-label={`詳細: 主訴 ${item.name} (ID ${item.id})`}
              onClick={() => onEdit(item)}
            >
              {item.name}
            </DataTableRowButton>
          </TableCell>
          <TableCell className={C.text}>{item.description}</TableCell>
          <TableCell className="text-center"><StatusPill isActive={item.isActive} /></TableCell>
          <TableCell className="text-right">
            {canEdit ? (
              <RowActionButton
                onClick={() => onEdit(item)}
                aria-label={`主訴「${item.name}」(ID: ${item.id}) を編集`}
              />
            ) : null}
          </TableCell>
        </DataTableRow>
      )}
      renderSidePanel={(props) => <ChiefComplaintSidePanel key={props.item?.id ?? "new"} {...props} onDirtyChange={handleDirtyChange} />}
    />
    {dirty.discardDialog}
    </>
  );
}
