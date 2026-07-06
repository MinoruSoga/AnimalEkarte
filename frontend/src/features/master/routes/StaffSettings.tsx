import { useMemo, useCallback } from "react";
import { UserRound } from "lucide-react";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C, LAYOUT, ICON, PALETTE } from "@/lib/design-tokens";
import { MASTER_TABLE_COL } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { usePermission } from "@/hooks/use-permission";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { StaffSidePanel } from "../components/StaffSidePanel";
import type { StaffFormData } from "../components/StaffSidePanelModel";
import {
  useGetStaffs,
  useCreateStaff,
  useUpdateStaff,
  useDeleteStaff,
  useUpdateStaffClinics,
  useGetClinicsList,
  useGetAllStaffPermissionGroupMap,
  useUpdateStaffCapableReservationTypes,
  useUpdateStaffPermissionGroups,
} from "../api/staffs";
import type { Staff, CreateStaffRequest, UpdateStaffRequest } from "../api/staffs";
import { useGetPermissionGroups } from "../api/permission-groups";
import { useGetAllOccupations } from "../api/occupations";
import { useGetReservationTypes } from "../api/reservation-types";
import {
  buildGroupsByStaffId,
  buildStaffFilterProperties,
  buildStaffCreateRequest,
  buildStaffIds,
  buildStaffUpdateRequest,
  filterStaffByMasterFilters,
  searchStaff,
} from "./StaffSettingsModel";
import { ResourceMasterStaff } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const COLUMNS = [
  { header: "氏名", className: "flex-1" },
  { header: "職種", className: MASTER_TABLE_COL.w130 },
  { header: "権限グループ", className: MASTER_TABLE_COL.w200 },
  { header: "ステータス", className: MASTER_TABLE_COL.w100, align: "center" as const },
  { header: "操作", className: MASTER_TABLE_COL.w80, align: "right" as const },
];

// ─────────────────────────────────────────────────
// StaffSettings (page)
// ─────────────────────────────────────────────────

export function StaffSettings() {
  usePermission(ResourceMasterStaff);

  const { data } = useGetStaffs();
  const createMutation = useCreateStaff();
  const updateMutation = useUpdateStaff();
  const deleteMutation = useDeleteStaff();

  // Occupations (職種) — shown as select in the panel + filter
  const { data: allOccupationsData } = useGetAllOccupations();
  const allOccupations = useMemo(() => allOccupationsData ?? [], [allOccupationsData]);

  // Permission groups — shown as checkboxes in the panel
  const { data: allGroupsData } = useGetPermissionGroups();
  const allGroups = useMemo(() => allGroupsData ?? [], [allGroupsData]);

  // Service Types — shown as capability checkboxes in the panel
  const { data: allReservationTypesData } = useGetReservationTypes();
  const allReservationTypes = useMemo(() => allReservationTypesData ?? [], [allReservationTypesData]);

  // Clinics — shown as checkboxes in the panel
  const { data: allClinicsData } = useGetClinicsList("all");
  const allClinics = useMemo(() => allClinicsData ?? [], [allClinicsData]);

  // Group / Clinic / Capability mutation
  const setGroupsMutation = useUpdateStaffPermissionGroups();
  const { mutate: setGroupsFn } = setGroupsMutation;
  const setClinicsMutation = useUpdateStaffClinics();
  const { mutate: setClinicsFn } = setClinicsMutation;
  const setCapableMutation = useUpdateStaffCapableReservationTypes();
  const { mutate: setCapableFn } = setCapableMutation;

  // スタッフ全員の権限グループIDマップ（テーブル表示用）
  const staffIds = useMemo(() => buildStaffIds(data), [data]);
  const { data: staffGroupMap } = useGetAllStaffPermissionGroupMap(staffIds);

  // staffId → PermissionGroup[] のルックアップ
  const groupsByStaffId = useMemo(
    () => buildGroupsByStaffId({ staffGroupMap, groups: allGroups }),
    [staffGroupMap, allGroups],
  );

  // フィルタの職種選択肢を occupations マスタから動的生成
  const staffFilterProperties = useMemo(
    () => buildStaffFilterProperties(allOccupations),
    [allOccupations],
  );

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Staff>({
    data,
    deleteMutation,
    entityLabel: "スタッフ",
    searchFilter: searchStaff,
    activeFilterApply: filterStaffByMasterFilters,
    dirtyGuard: dirty,
  });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const { handleSave } = useMasterSave<Staff, StaffFormData, CreateStaffRequest, UpdateStaffRequest>({
    crud,
    createMutation,
    updateMutation,
    validate: (d) => {
      if (!d.name.trim()) return "氏名は必須です";
      if (crud.editTarget === null || crud.editTarget === "new") {
        if (!d.email) return "メールアドレスは必須です";
        if (!d.password || d.password.length < 8) return "パスワードは8文字以上で入力してください";
      }
      return null;
    },
    toCreateRequest: buildStaffCreateRequest,
    toUpdateRequest: buildStaffUpdateRequest,
  });

  const handleSaveGroups = useCallback(
    (staffId: string, groupIds: string[]) => {
      setGroupsFn({ staffId, groupIds });
    },
    [setGroupsFn],
  );

  const handleSaveClinics = useCallback(
    (staffId: string, clinicIds: string[]) => {
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
      filterProperties={staffFilterProperties}
      renderRow={(item, onEdit, canEdit) => {
        const groups = groupsByStaffId.get(item.id) ?? [];
        const visibleGroups = groups.slice(0, 2);
        const extraCount = groups.length - visibleGroups.length;
        return (
          <DataTableRow key={item.id} onClick={canEdit ? () => onEdit(item) : undefined}>
            <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
            <TableCell className={`text-base ${C.text}`}>{item.occupationName ?? "—"}</TableCell>
            <TableCell>
              <div className="flex flex-wrap items-center gap-1">
                {visibleGroups.length === 0 ? (
                  <span className={`text-sm ${C.text40}`}>—</span>
                ) : (
                  <>
                    {visibleGroups.map((g) => (
                      <span
                        key={g.id}
                        className={`inline-flex items-center gap-1 ${LAYOUT.inputCompact} text-xs`}
                        style={{
                          backgroundColor: g.color ? `${g.color}18` : `${PALETTE.primary}0f`,
                          color: g.color ?? PALETTE.primary,
                        }}
                      >
                        <span
                          className={`${ICON.dotSm} rounded-full shrink-0`}
                          style={{ backgroundColor: g.color ?? PALETTE.defaultGray }}
                        />
                        {g.name}
                      </span>
                    ))}
                    {extraCount > 0 ? (
                      <span className={`text-xs ${C.text40}`}>+{extraCount}</span>
                    ) : null}
                  </>
                )}
              </div>
            </TableCell>
            <TableCell className="text-center">
              <StatusPill isActive={item.isActive} />
            </TableCell>
            <TableCell className="p-0 text-right">
              {canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}
            </TableCell>
          </DataTableRow>
        );
      }}
      renderSidePanel={({ item, onClose, onSave, onDeleteRequest, readOnly }) => (
        <StaffSidePanel
          key={item?.id ?? "new"}
          item={item}
          onClose={onClose}
          onSave={onSave}
          onDeleteRequest={onDeleteRequest}
          readOnly={readOnly}
          onDirtyChange={handleDirtyChange}
          allOccupations={allOccupations}
          allGroups={allGroups}
          onSaveGroups={handleSaveGroups}
          allClinics={allClinics}
          onSaveClinics={handleSaveClinics}
          allReservationTypes={allReservationTypes}
          onSaveCapableReservationTypes={handleSaveCapableReservationTypes}
        />
      )}
    />
  );
}
