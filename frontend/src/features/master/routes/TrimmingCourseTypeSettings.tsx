import { useCallback } from "react";
import { Scissors } from "lucide-react";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C, ICON } from "@/lib/design-tokens";
import { ResourceMasterTrimming } from "@/types/generated/models";
import { MASTER_STATUS_FILTER, MASTER_TABLE_COL } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { TrimmingCourseTypeSidePanel } from "../components/TrimmingCourseTypeSidePanel";
import type { TrimmingCourseTypeFormData } from "../lib/trimming-course-type-side-panel-model";
import {
  useGetTrimmingCourseTypes,
  useCreateTrimmingCourseType,
  useUpdateTrimmingCourseType,
  useDeleteTrimmingCourseType,
} from "../api/trimming-course-type";
import type {
  CreateTrimmingCourseTypeRequest,
  TrimmingCourseType,
  UpdateTrimmingCourseTypeRequest,
} from "../api/trimming-course-type";
import {
  buildTrimmingCourseTypeCreateRequest,
  buildTrimmingCourseTypeUpdateRequest,
} from "./trimming-course-type-settings-model";

// ─── Constants ───
const COLUMNS = [
  { header: "名称", className: "flex-1" },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

// ─── Page ───
export function TrimmingCourseTypeSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterTrimming);
  const { data } = useGetTrimmingCourseTypes();
  const createMutation = useCreateTrimmingCourseType();
  const updateMutation = useUpdateTrimmingCourseType();
  const deleteMutation = useDeleteTrimmingCourseType();

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<TrimmingCourseType>({
    data,
    deleteMutation,
    entityLabel: "コース種別",
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
    TrimmingCourseType,
    TrimmingCourseTypeFormData,
    CreateTrimmingCourseTypeRequest,
    UpdateTrimmingCourseTypeRequest
  >({
    crud,
    createMutation,
    updateMutation,
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: buildTrimmingCourseTypeCreateRequest,
    toUpdateRequest: buildTrimmingCourseTypeUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  return (
    <>
      <MasterCRUDPage
        title="コース種別マスタ"
        icon={<Scissors className={`${ICON.page} ${C.text}`} />}
        resource={ResourceMasterTrimming}
        entityLabel="コース種別"
        searchPlaceholder="種別名で検索..."
        emptyMessage="コース種別が登録されていません"
        crud={crud}
        handleSave={handleSave}
        columns={COLUMNS}
        filterProperties={[MASTER_STATUS_FILTER]}
        renderRow={(item, onEdit, canEdit) => (
          <DataTableRow key={item.id}>
            <TableCell className={`font-medium ${C.text}`}>
              <DataTableRowButton
                aria-label={`詳細: コース種別 ${item.name} (ID ${item.id})`}
                onClick={() => onEdit(item)}
              >
                {item.name}
              </DataTableRowButton>
            </TableCell>
            <TableCell className="text-center">
              <StatusPill isActive={item.isActive} />
            </TableCell>
            <TableCell className="text-right">
              {canEdit ? (
                <RowActionButton
                  onClick={() => onEdit(item)}
                  aria-label={`コース種別「${item.name}」(ID: ${item.id}) を編集`}
                />
              ) : null}
            </TableCell>
          </DataTableRow>
        )}
        renderSidePanel={(props) => (
          <TrimmingCourseTypeSidePanel
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
