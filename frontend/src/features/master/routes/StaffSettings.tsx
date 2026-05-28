import { useMemo, useCallback } from "react";
import { UserRound } from "lucide-react";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { TableCell } from "@/components/ui/table";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { C, LAYOUT, ICON, PALETTE } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import type { FilterProperty } from "@/components/shared/NotionFilter/types";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { usePermission } from "@/hooks/use-permission";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { StaffSidePanel, type StaffFormData } from "../components/StaffSidePanel";
import {
  useGetStaffs,
  useCreateStaff,
  useUpdateStaff,
  useDeleteStaff,
  useUpdateStaffClinics,
  useGetClinicsList,
  useGetAllStaffPermissionGroupMap,
  useUpdateStaffExcludedReservationTypes,
  useUpdateStaffPermissionGroups,
} from "../api/staffs";
import type { Staff, CreateStaffRequest, UpdateStaffRequest } from "../api/staffs";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/NotionFilter/types";
import { useGetPermissionGroups } from "../api/permission-groups";
import { useGetAllOccupations } from "../api/occupations";
import { useGetReservationTypes } from "../api/reservation-types";
import { ResourceMasterStaff } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const COLUMNS = [
  { header: "氏名", className: "flex-1" },
  { header: "職種", className: "w-[130px]" },
  { header: "権限グループ", className: "w-[200px]" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
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

  // Service Types — shown as checkboxes in excluded courses panel
  const { data: allReservationTypesData } = useGetReservationTypes();
  const allReservationTypes = useMemo(() => allReservationTypesData ?? [], [allReservationTypesData]);

  // Clinics — shown as checkboxes in the panel
  const { data: allClinicsData } = useGetClinicsList("all");
  const allClinics = useMemo(() => allClinicsData ?? [], [allClinicsData]);

  // Group / Clinic / Excluded mutation
  const setGroupsMutation = useUpdateStaffPermissionGroups();
  const { mutate: setGroupsFn } = setGroupsMutation;
  const setClinicsMutation = useUpdateStaffClinics();
  const { mutate: setClinicsFn } = setClinicsMutation;
  const setExcludedMutation = useUpdateStaffExcludedReservationTypes();
  const { mutate: setExcludedFn } = setExcludedMutation;

  // スタッフ全員の権限グループIDマップ（テーブル表示用）
  const staffIds = useMemo(() => (data ?? []).map((s) => s.id), [data]);
  const { data: staffGroupMap } = useGetAllStaffPermissionGroupMap(staffIds);

  // staffId → PermissionGroup[] のルックアップ
  const groupsByStaffId = useMemo(() => {
    const map = new Map<string, typeof allGroups>();
    if (!staffGroupMap) return map;
    for (const [staffId, groupIds] of staffGroupMap.entries()) {
      map.set(staffId, allGroups.filter((g) => groupIds.includes(g.id)));
    }
    return map;
  }, [staffGroupMap, allGroups]);

  // フィルタの職種選択肢を occupations マスタから動的生成
  const staffFilterProperties = useMemo<FilterProperty[]>(
    () => [
      MASTER_STATUS_FILTER,
      {
        key: "occupationId",
        label: "職種",
        type: "select",
        icon: UserRound,
        conditions: CONDITIONS_NO_EMPTY,
        options: allOccupations
          .filter((oc) => oc.isActive)
          .map((oc) => ({ value: oc.id, label: oc.name })),
      },
    ],
    [allOccupations],
  );

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Staff>({
    data,
    deleteMutation,
    entityLabel: "スタッフ",
    searchFilter: (s, lower) =>
      s.name.toLowerCase().includes(lower) ||
      (s.occupationName?.toLowerCase().includes(lower) ?? false),
    activeFilterApply: (item, filters) => {
      for (const f of filters) {
        if (f.key === "status" && typeof f.value === "string") {
          const want = f.value === "active";
          if (f.condition === "is" && item.isActive !== want) return false;
          if (f.condition === "is_not" && item.isActive === want) return false;
        }
        if (f.key === "occupationId" && typeof f.value === "string") {
          if (f.condition === "is" && item.occupationId !== f.value) return false;
          if (f.condition === "is_not" && item.occupationId === f.value) return false;
        }
      }
      return true;
    },
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
    toCreateRequest: (d) => ({
      name: d.name,
      email: d.email,
      password: d.password,
      license_number: d.licenseNumber || undefined,
      occupation_id: d.jobTitleId ?? undefined,
      staff_type: d.staffType,
      reservation_display_name: d.reservationDisplayName || undefined,
      reservation_visible: d.reservationVisible,
      reservation_comment: d.reservationComment || undefined,
      reservation_image_url: d.reservationImageUrl || undefined,
    }),
    toUpdateRequest: (d) => ({
      name: d.name,
      license_number: d.licenseNumber || undefined,
      is_active: d.isActive,
      occupation_id: d.jobTitleId ?? undefined,
      password: d.password || undefined,
      staff_type: d.staffType,
      reservation_display_name: d.reservationDisplayName || undefined,
      reservation_visible: d.reservationVisible,
      reservation_comment: d.reservationComment || undefined,
      reservation_image_url: d.reservationImageUrl || undefined,
    }),
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

  const handleSaveExcludedReservationTypes = useCallback(
    (staffId: string, reservationTypeIds: string[]) => {
      setExcludedFn({ staffId, reservationTypeIds });
    },
    [setExcludedFn],
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
              <NotionStatusPill isActive={item.isActive} />
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
          onSaveExcludedReservationTypes={handleSaveExcludedReservationTypes}
        />
      )}
    />
  );
}
