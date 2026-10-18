import { useCallback, useRef, useLayoutEffect } from "react";

import { toast } from "sonner";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { ErrorFallback, LoadingFallback } from "@/components/shared/DataStates";
import { MASTER_TABLE_COL } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { usePermission } from "@/hooks/use-permission";
import { validateOptionalEmail, validateOptionalPassword } from "@/lib/validate-credentials";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { StaffSidePanel } from "../components/StaffSidePanel";
import type { StaffFormData } from "../lib/staff-side-panel-model";
import { StaffSettingsRow } from "../components/StaffSettingsRow";
import {
  useCreateStaff,
  useUpdateStaff,
  useDeleteStaff,
  useGetStaffs,
  useUpdateStaffClinics,
  useUpdateStaffCapableReservationTypes,
  useUpdateStaffPermissionGroups,
} from "../api/staffs";
import type { Staff, CreateStaffRequest, UpdateStaffRequest } from "../api/staffs";
import {
  buildStaffCreateRequest,
  buildStaffUpdateRequest,
  filterStaffByMasterFilters,
  searchStaff,
  STAFF_DEFAULT_ACTIVE_FILTERS,
} from "./staff-settings-model";
import { useStaffSettingsLookups } from "../hooks/use-staff-settings-lookups";
import { ResourceMasterStaff, ResourceMasterPermission } from "@/types/generated/models";

const COLUMNS = [
  { header: "氏名", className: "flex-1 min-w-[140px]" },
  { header: "職種", className: `${MASTER_TABLE_COL.w130} whitespace-nowrap` },
  { header: "権限グループ", className: `${MASTER_TABLE_COL.w200} whitespace-nowrap` },
  {
    header: "ステータス",
    className: `${MASTER_TABLE_COL.w100} min-w-[100px] whitespace-nowrap`,
    align: "center" as const,
  },
  {
    header: "操作",
    className: `${MASTER_TABLE_COL.w80} whitespace-nowrap`,
    align: "right" as const,
  },
];

export function StaffSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterStaff);
  const { canEdit: canEditPermission } = usePermission(ResourceMasterPermission);

  const lookups = useStaffSettingsLookups();
  // 一覧クエリの状態を DataStates に接続する（useGetStaffs は lookups 内と同一 queryKey で
  // キャッシュ共有されるため追加 fetch は発生しない）。
  const { isPending, isError } = useGetStaffs();
  const createMutation = useCreateStaff();
  const updateMutation = useUpdateStaff();
  const deleteMutation = useDeleteStaff();
  const setGroupsMutation = useUpdateStaffPermissionGroups();
  const { mutateAsync: setGroupsAsync } = setGroupsMutation;
  const setClinicsMutation = useUpdateStaffClinics();
  const { mutateAsync: setClinicsAsync } = setClinicsMutation;
  const setCapableMutation = useUpdateStaffCapableReservationTypes();
  const { mutateAsync: setCapableAsync } = setCapableMutation;

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Staff>({
    data: lookups.data,
    deleteMutation,
    entityLabel: "スタッフ",
    searchFilter: searchStaff,
    activeFilterApply: filterStaffByMasterFilters,
    initialFilters: STAFF_DEFAULT_ACTIVE_FILTERS,
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

  const { handleSave, handleSaveWithResult } = useMasterSave<
    Staff,
    StaffFormData,
    CreateStaffRequest,
    UpdateStaffRequest
  >({
    crud,
    createMutation,
    updateMutation,
    validate: (d) => {
      if (!d.name.trim()) return "氏名は必須です";
      return validateOptionalEmail(d.email) ?? validateOptionalPassword(d.password);
    },
    toCreateRequest: buildStaffCreateRequest,
    toUpdateRequest: buildStaffUpdateRequest,
    permissions: { canCreate, canEdit },
    closeOnSuccess: false,
    showSuccessToast: false,
  });

  const associationPermissionsRef = useRef({ canEdit, canEditPermission });
  useLayoutEffect(() => {
    associationPermissionsRef.current = { canEdit, canEditPermission };
  }, [canEdit, canEditPermission]);

  const handleSaveGroups = useCallback(
    async (staffId: string, groupIds: string[]) => {
      const current = associationPermissionsRef.current;
      if (!current.canEdit || !current.canEditPermission) {
        toast.error("権限グループを変更する権限がありません");
        return false;
      }
      try {
        await setGroupsAsync({ staffId, groupIds });
        return true;
      } catch {
        return false;
      }
    },
    [setGroupsAsync],
  );

  const handleSaveClinics = useCallback(
    async (staffId: string, clinicIds: string[]) => {
      if (!associationPermissionsRef.current.canEdit) {
        toast.error("所属医院を変更する権限がありません");
        return false;
      }
      try {
        await setClinicsAsync({ staffId, clinicIds });
        return true;
      } catch {
        return false;
      }
    },
    [setClinicsAsync],
  );

  const handleSaveCapableReservationTypes = useCallback(
    async (staffId: string, reservationTypeIds: string[]) => {
      if (!associationPermissionsRef.current.canEdit) {
        toast.error("対応区分を変更する権限がありません");
        return false;
      }
      try {
        await setCapableAsync({ staffId, reservationTypeIds });
        return true;
      } catch {
        return false;
      }
    },
    [setCapableAsync],
  );

  return (
    <>
      <MasterCRUDPage
        title="スタッフマスタ"
        resource={ResourceMasterStaff}
        entityLabel="スタッフ"
        searchPlaceholder="氏名、職種で検索..."
        emptyMessage="スタッフが登録されていません"
        crud={crud}
        handleSave={handleSave}
        columns={COLUMNS}
        filterProperties={lookups.staffFilterProperties}
        renderRow={(item, onEdit, rowCanEdit) => (
          <StaffSettingsRow
            item={item}
            groups={lookups.groupsByStaffId.get(item.id) ?? []}
            onEdit={onEdit}
            canEdit={rowCanEdit}
          />
        )}
        renderSidePanel={({ item, onClose, onSave, onDeleteRequest, readOnly }) => (
          <StaffSidePanel
            key={item?.id ?? "new"}
            item={item}
            onClose={onClose}
            onSave={onSave}
            onSaveWithResult={handleSaveWithResult}
            canEditPersisted={canEdit}
            onDeleteRequest={onDeleteRequest}
            readOnly={readOnly}
            onDirtyChange={handleDirtyChange}
            allOccupations={lookups.allOccupations}
            allGroups={lookups.allGroups}
            onSaveGroups={handleSaveGroups}
            allClinics={lookups.allClinics}
            onSaveClinics={handleSaveClinics}
            allReservationTypes={lookups.allReservationTypes}
            onSaveCapableReservationTypes={handleSaveCapableReservationTypes}
          />
        )}
      >
        {isError ? (
          <ErrorFallback message="スタッフの取得に失敗しました" />
        ) : isPending ? (
          <LoadingFallback />
        ) : null}
      </MasterCRUDPage>
      {dirty.discardDialog}
    </>
  );
}
