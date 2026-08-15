import { useMemo, useCallback, useRef, useLayoutEffect } from "react";
import { UserRound } from "lucide-react";
import { toast } from "sonner";
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
import {
  validateOptionalEmail,
  validateOptionalPassword,
} from "@/lib/validate-credentials";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { StaffSidePanel } from "../components/StaffSidePanel";
import type { StaffFormData } from "../components/staff-side-panel-model";
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
} from "./staff-settings-model";
import { ResourceMasterStaff, ResourceMasterPermission } from "@/types/generated/models";

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
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterStaff);
  // 権限グループ割当は backend が master-staff:edit と master-permission:edit の両方を要求する
  // （`backend/internal/staff/handler.go` の PUT /staffs/:id/permission-groups）。
  const { canEdit: canEditPermission } = usePermission(ResourceMasterPermission);

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
    permissions: { canDelete },
  });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const { handleSave } = useMasterSave<Staff, StaffFormData, CreateStaffRequest, UpdateStaffRequest>({
    crud,
    createMutation,
    updateMutation,
    // R-1③ の裁定: backend は account を持たない staff を許容する（`staffs.account_id` は
    // nullable、`resource` 種別は部屋・機材等の非人物リソースで原理的にログインを持ち得ない）。
    // email/password を新規作成の必須項目にすると正当な業務データを作成できなくなるため、
    // 必須化はせず「入力された場合のみ形式・強度を検証する」に揃える。password の要件は
    // backend の auth.ValidatePassword と同契約（新規・編集の双方へ適用する）。
    validate: (d) => {
      if (!d.name.trim()) return "氏名は必須です";
      return validateOptionalEmail(d.email) ?? validateOptionalPassword(d.password);
    },
    toCreateRequest: buildStaffCreateRequest,
    toUpdateRequest: buildStaffUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  // R-1⑤: 基本staff保存は useMasterSave が権限を再検査するが、関連付けmutationは素通しだった。
  // read-only画面でEnterによりform actionが発火した場合、基本保存が拒否されても関連付けだけが
  // 独立して発行され得る。backend は両経路とも fail-closed（403）だが、UI が実行できない操作を
  // 試みる状態は残る。mutation直前に最新権限を再検査する（Active execution rules 2）。
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
          <DataTableRow key={item.id}>
            <TableCell className={`font-medium ${C.text}`}>{item.name}</TableCell>
            <TableCell className={C.text}>{item.occupationName ?? "—"}</TableCell>
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
            <TableCell className="text-right">
              {canEdit ? (
                <RowActionButton
                  onClick={() => onEdit(item)}
                  aria-label={`スタッフ「${item.name}」(ID: ${item.id}) を編集`}
                />
              ) : null}
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
