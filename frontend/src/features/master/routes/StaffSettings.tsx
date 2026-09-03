import { useCallback, useRef, useLayoutEffect } from "react";
import { UserRound } from "lucide-react";
import { toast } from "sonner";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { C, ICON } from "@/lib/design-tokens";
import { MASTER_TABLE_COL } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { usePermission } from "@/hooks/use-permission";
import {
  validateOptionalEmail,
  validateOptionalPassword,
} from "@/lib/validate-credentials";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { StaffSidePanel } from "../components/StaffSidePanel";
import type { StaffFormData } from "../components/staff-side-panel-model";
import { StaffSettingsRow } from "../components/StaffSettingsRow";
import {
  useCreateStaff,
  useUpdateStaff,
  useDeleteStaff,
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
} from "./staff-settings-model";
import { useStaffSettingsLookups } from "../hooks/use-staff-settings-lookups";
import { ResourceMasterStaff, ResourceMasterPermission } from "@/types/generated/models";

const COLUMNS = [
  { header: "氏名", className: "flex-1" },
  { header: "職種", className: MASTER_TABLE_COL.w130 },
  { header: "権限グループ", className: MASTER_TABLE_COL.w200 },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

export function StaffSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterStaff);
  const { canEdit: canEditPermission } = usePermission(ResourceMasterPermission);

  const lookups = useStaffSettingsLookups();
  const createMutation = useCreateStaff();
  const updateMutation = useUpdateStaff();
  const deleteMutation = useDeleteStaff();
  const setGroupsMutation = useUpdateStaffPermissionGroups();
  const { mutate: setGroupsFn } = setGroupsMutation;
  const setClinicsMutation = useUpdateStaffClinics();
  const { mutate: setClinicsFn } = setClinicsMutation;
  const setCapableMutation = useUpdateStaffCapableReservationTypes();
  const { mutate: setCapableFn } = setCapableMutation;

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Staff>({
    data: lookups.data,
    deleteMutation,
    entityLabel: "スタッフ",
    searchFilter: searchStaff,
    activeFilterApply: filterStaffByMasterFilters,
    dirtyGuard: dirty,
    permissions: { canDelete },
  });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const { handleSave } = useMasterSave<Staff, StaffFormData, CreateStaffRequest, UpdateStaffRequest>({
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
  });

  const associationPermissionsRef = useRef({ canEdit, canEditPermission });
  useLayoutEffect(() => {
    associationPermissionsRef.current = { canEdit, canEditPermission };
  }, [canEdit, canEditPermission]);

  const handleSaveGroups = useCallback(
    (staffId: string, groupIds: string[]) => {
      const current = associationPermissionsRef.current;
      if (!current.canEdit || !current.canEditPermission) {
        toast.error("権限グループを変更する権限がありません");
        return;
      }
      setGroupsFn({ staffId, groupIds });
    },
    [setGroupsFn],
  );

  const handleSaveClinics = useCallback(
    (staffId: string, clinicIds: string[]) => {
      if (!associationPermissionsRef.current.canEdit) {
        toast.error("所属医院を変更する権限がありません");
        return;
      }
      setClinicsFn({ staffId, clinicIds });
    },
    [setClinicsFn],
  );

  const handleSaveCapableReservationTypes = useCallback(
    (staffId: string, reservationTypeIds: string[]) => {
      setCapableFn({ staffId, reservationTypeIds });
    },
    [setCapableFn],
  );

  return (
    <>
    <MasterCRUDPage
      title="スタッフマスタ"
      icon={<UserRound className={`${ICON.page} ${C.text}`} />}
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
    />
    {dirty.discardDialog}
    </>
  );
}
