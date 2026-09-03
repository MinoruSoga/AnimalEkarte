import { useCallback } from "react";
import { Briefcase } from "lucide-react";
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
import {
  useGetAllOccupations,
  useCreateOccupation,
  useUpdateOccupation,
  useDeleteOccupation,
} from "../api/occupations";
import type {
  Occupation,
  CreateOccupationRequest,
  UpdateOccupationRequest,
} from "../api/occupations";
import { OccupationSidePanel } from "../components/OccupationSidePanel";
import type { OccupationFormData } from "../lib/occupation-side-panel-model";
import {
  buildOccupationCreateRequest,
  buildOccupationUpdateRequest,
} from "./occupation-settings-model";
import { ResourceMasterStaff } from "@/types/generated/models";

// ─── Constants ───
const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "説明", className: "flex-1" },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

// ─── Page ───
export function OccupationSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterStaff);
  const { data } = useGetAllOccupations();
  const createMutation = useCreateOccupation();
  const updateMutation = useUpdateOccupation();
  const deleteMutation = useDeleteOccupation();

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Occupation>({
    data,
    deleteMutation,
    entityLabel: "職種",
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
    Occupation,
    OccupationFormData,
    CreateOccupationRequest,
    UpdateOccupationRequest
  >({
    crud,
    createMutation,
    updateMutation,
    permissions: { canCreate, canEdit },
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: buildOccupationCreateRequest,
    toUpdateRequest: buildOccupationUpdateRequest,
  });

  return (
    <>
      <MasterCRUDPage
        title="職種マスタ"
        icon={<Briefcase className={`${ICON.page} ${C.text}`} />}
        resource={ResourceMasterStaff}
        entityLabel="職種"
        searchPlaceholder="職種名で検索..."
        emptyMessage="職種が登録されていません"
        crud={crud}
        handleSave={handleSave}
        columns={COLUMNS}
        filterProperties={[MASTER_STATUS_FILTER]}
        renderRow={(item, onEdit, canEdit) => (
          <DataTableRow key={item.id}>
            <TableCell className={`font-medium ${C.text}`}>
              <DataTableRowButton
                aria-label={`詳細: 職種 ${item.name} (ID ${item.id})`}
                onClick={() => onEdit(item)}
              >
                {item.name}
              </DataTableRowButton>
            </TableCell>
            <TableCell className={C.text}>{item.description || "-"}</TableCell>
            <TableCell className="text-center">
              <StatusPill isActive={item.isActive} />
            </TableCell>
            <TableCell className="text-right">
              {canEdit ? (
                <RowActionButton
                  onClick={() => onEdit(item)}
                  aria-label={`職種「${item.name}」(ID: ${item.id}) を編集`}
                />
              ) : null}
            </TableCell>
          </DataTableRow>
        )}
        renderSidePanel={(props) => (
          <OccupationSidePanel
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
