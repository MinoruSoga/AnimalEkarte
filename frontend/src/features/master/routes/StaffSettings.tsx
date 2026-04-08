import { memo, useState, useMemo, useCallback } from "react";
import { UserRound, Shield, Building2 } from "lucide-react";
import { TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { C, STYLE, LAYOUT, ICON, PALETTE } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS, MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import type { FilterProperty } from "@/components/shared/NotionFilter/types";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import {
  useGetStaffs,
  useCreateStaff,
  useUpdateStaff,
  useDeleteStaff,
  useGetStaffPermissionGroups,
  useSetStaffPermissionGroups,
  useGetStaffClinics,
  useSetStaffClinics,
  useGetClinicsList,
  useGetAllStaffPermissionGroupMap,
} from "@/features/master/api/staffs";
import type { Staff, CreateStaffRequest, UpdateStaffRequest } from "@/features/master/api/staffs";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/NotionFilter/types";
import { useGetPermissionGroups } from "@/features/master/api/permission-groups";
import type { PermissionGroup } from "@/features/master/api/permission-groups";
import type { ClinicSummary } from "@/features/master/api/staffs";
import { useGetAllOccupations } from "@/features/master/api/occupations";
import type { Occupation } from "@/features/master/api/occupations";
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
// Types
// ─────────────────────────────────────────────────

interface StaffFormData {
  name: string;
  jobTitleId: string | null;
  licenseNumber: string;
  isActive: boolean;
  email: string;
  password: string;
}

// ─────────────────────────────────────────────────
// StaffSidePanel
// ─────────────────────────────────────────────────

interface StaffSidePanelProps {
  item: Staff | null;
  onClose: () => void;
  onSave: (d: StaffFormData) => void;
  onDeleteRequest?: (i: Staff) => void;
  readOnly?: boolean;
  /** All occupations (職種) available in this clinic */
  allOccupations: Occupation[];
  /** All permission groups available in this clinic */
  allGroups: PermissionGroup[];
  /** Called when groups should be saved for this staff */
  onSaveGroups: (staffId: string, groupIds: string[]) => void;
  /** All clinics available for assignment */
  allClinics: ClinicSummary[];
  /** Called when clinics should be saved for this staff */
  onSaveClinics: (staffId: string, clinicIds: string[]) => void;
}

const StaffSidePanel = memo(function StaffSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  allOccupations,
  allGroups,
  onSaveGroups,
  allClinics,
  onSaveClinics,
}: StaffSidePanelProps) {
  const isNew = item === null;
  const staffId = item?.id ?? null;

  const [f, setF] = useState<StaffFormData>(() => ({
    name: item?.name ?? "",
    jobTitleId: item?.occupationId ?? null,
    licenseNumber: item?.licenseNumber ?? "",
    isActive: item?.isActive ?? true,
    email: item?.email ?? "",
    password: "",
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  // ── Occupation select items (memoized) ────────────
  const occupationSelectItems = useMemo(
    () =>
      allOccupations
        .filter((oc) => oc.isActive)
        .map((oc) => (
          <SelectItem key={oc.id} value={oc.id}>{oc.name}</SelectItem>
        )),
    [allOccupations],
  );

  // ── Permission groups state ──────────────────────
  const { data: serverGroupIds } = useGetStaffPermissionGroups(staffId);
  const [userEditedGroupIds, setUserEditedGroupIds] = useState<string[] | null>(null);

  const groupIds = useMemo(
    () => userEditedGroupIds ?? serverGroupIds ?? [],
    [userEditedGroupIds, serverGroupIds],
  );

  // ── Clinic assignments state ───────────────────
  const { data: serverClinicIds } = useGetStaffClinics(staffId);
  const [userEditedClinicIds, setUserEditedClinicIds] = useState<string[] | null>(null);

  const clinicIds = useMemo(
    () => userEditedClinicIds ?? serverClinicIds ?? [],
    [userEditedClinicIds, serverClinicIds],
  );

  // ── Handlers ─────────────────────────────────────
  const handleClinicToggle = useCallback(
    (clinicId: string, checked: boolean) => {
      setUserEditedClinicIds((prev) => {
        const current = prev ?? serverClinicIds ?? [];
        return checked ? [...current, clinicId] : current.filter((id) => id !== clinicId);
      });
      setIsDirty(true);
    },
    [serverClinicIds],
  );

  const handleGroupToggle = useCallback(
    (groupId: string, checked: boolean) => {
      setUserEditedGroupIds((prev) => {
        const current = prev ?? serverGroupIds ?? [];
        return checked ? [...current, groupId] : current.filter((id) => id !== groupId);
      });
      setIsDirty(true);
    },
    [serverGroupIds],
  );

  const handleSave = useCallback(() => {
    if (!f.name.trim()) {
      setNameError("氏名を入力してください");
      return;
    }
    setNameError("");
    onSave(f);
    if (!isNew && staffId) {
      onSaveGroups(staffId, groupIds);
      onSaveClinics(staffId, clinicIds);
    }
    setIsDirty(false);
  }, [f, isNew, staffId, groupIds, clinicIds, onSave, onSaveGroups, onSaveClinics]);

  const handleTitleChange = useCallback((v: string) => {
    setF((p) => ({ ...p, name: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleToggleActive = useCallback(() => {
    setF((p) => ({ ...p, isActive: !p.isActive }));
    setIsDirty(true);
  }, []);

  const handleOccupationChange = useCallback((v: string) => {
    setF((p) => ({ ...p, jobTitleId: v }));
    setIsDirty(true);
  }, []);

  const handleLicenseNumberChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setF((p) => ({ ...p, licenseNumber: e.target.value }));
    setIsDirty(true);
  }, []);

  const handleEmailChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setF((p) => ({ ...p, email: e.target.value }));
    setIsDirty(true);
  }, []);

  const handlePasswordChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setF((p) => ({ ...p, password: e.target.value }));
    setIsDirty(true);
  }, []);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel
      isNew={isNew}
      title={f.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={handleSave}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<UserRound className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton
        isActive={f.isActive}
        onToggle={handleToggleActive}
      />

      <PropertyRow label="職種">
        <Select
          value={f.jobTitleId ?? undefined}
          onValueChange={handleOccupationChange}
        >
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>{occupationSelectItems}</SelectContent>
        </Select>
      </PropertyRow>

      <PropertyRow label="資格番号">
        <input
          type="text"
          className={MASTER_INPUT_CLASS}
          value={f.licenseNumber}
          onChange={handleLicenseNumberChange}
          placeholder="空"
        />
      </PropertyRow>

      {isNew ? (
        <>
          <PropertyRow label="メールアドレス">
            <input
              type="email"
              className={MASTER_INPUT_CLASS}
              value={f.email}
              onChange={handleEmailChange}
              placeholder="例: staff@clinic.com"
            />
          </PropertyRow>
          <PropertyRow label="パスワード">
            <input
              type="password"
              className={MASTER_INPUT_CLASS}
              value={f.password}
              onChange={handlePasswordChange}
              placeholder="8文字以上"
            />
          </PropertyRow>
        </>
      ) : (
        <>
          <PropertyRow label="メールアドレス">
            <span className={`text-sm ${C.text65}`}>{item?.email || "未設定"}</span>
          </PropertyRow>
          <PropertyRow label="パスワード">
            <input
              type="password"
              className={MASTER_INPUT_CLASS}
              value={f.password}
              onChange={handlePasswordChange}
              placeholder="変更する場合のみ入力"
            />
          </PropertyRow>
        </>
      )}

      {/* ── Clinic assignments ─────────────────────── */}
      <div className="mt-4 border-t pt-4">
        <div className="flex items-center gap-1.5 mb-2">
          <Building2 className={`${ICON.xs} text-muted-foreground`} />
          <p className="text-xs font-medium text-muted-foreground">所属医院</p>
        </div>

        {isNew ? (
          <p className="text-xs text-muted-foreground pl-0.5">
            スタッフ登録後に所属医院を設定できます
          </p>
        ) : allClinics.length === 0 ? (
          <p className="text-xs text-muted-foreground pl-0.5">
            医院が登録されていません
          </p>
        ) : (
          <div className="space-y-0.5">
            {allClinics.map((clinic) => (
              <label
                key={clinic.id}
                className="flex items-center gap-2.5 py-1.5 px-0.5 rounded cursor-pointer hover:bg-muted/40 transition-colors"
              >
                <Checkbox
                  checked={clinicIds.includes(clinic.id)}
                  onCheckedChange={(checked) =>
                    handleClinicToggle(clinic.id, checked === true)
                  }
                />
                <span className="text-sm">{clinic.name}</span>
              </label>
            ))}
          </div>
        )}
      </div>

      {/* ── Permission groups ─────────────────────── */}
      <div className="mt-4 border-t pt-4">
        <div className="flex items-center gap-1.5 mb-2">
          <Shield className={`${ICON.xs} text-muted-foreground`} />
          <p className="text-xs font-medium text-muted-foreground">権限グループ</p>
        </div>

        {isNew ? (
          <p className="text-xs text-muted-foreground pl-0.5">
            スタッフ登録後に権限グループを設定できます
          </p>
        ) : allGroups.length === 0 ? (
          <p className="text-xs text-muted-foreground pl-0.5">
            権限グループが登録されていません
          </p>
        ) : (
          <div className="space-y-0.5">
            {allGroups.map((group) => (
              <label
                key={group.id}
                className="flex items-center gap-2.5 py-1.5 px-0.5 rounded cursor-pointer hover:bg-muted/40 transition-colors"
              >
                <Checkbox
                  checked={groupIds.includes(group.id)}
                  onCheckedChange={(checked) =>
                    handleGroupToggle(group.id, checked === true)
                  }
                />
                <div
                  className="w-2.5 h-2.5 rounded-full flex-shrink-0"
                  style={{ backgroundColor: group.color ?? PALETTE.defaultGray }}
                />
                <span className="text-sm">{group.name}</span>
              </label>
            ))}
          </div>
        )}
      </div>
    </MasterSidePanel>
  );
});

// ─────────────────────────────────────────────────
// StaffSettings (page)
// ─────────────────────────────────────────────────

export function StaffSettings() {
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

  // Clinics — shown as checkboxes in the panel
  const { data: allClinicsData } = useGetClinicsList("all");
  const allClinics = useMemo(() => allClinicsData ?? [], [allClinicsData]);

  // Group assignment mutation
  const setGroupsMutation = useSetStaffPermissionGroups();
  const setClinicsMutation = useSetStaffClinics();

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
  });

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
    }),
    toUpdateRequest: (d) => ({
      name: d.name,
      license_number: d.licenseNumber || undefined,
      is_active: d.isActive,
      occupation_id: d.jobTitleId ?? undefined,
      password: d.password || undefined,
    }),
  });

  const handleSaveGroups = useCallback(
    (staffId: string, groupIds: string[]) => {
      setGroupsMutation.mutate({ staffId, groupIds });
    },
    [setGroupsMutation],
  );

  const handleSaveClinics = useCallback(
    (staffId: string, clinicIds: string[]) => {
      setClinicsMutation.mutate({ staffId, clinicIds });
    },
    [setClinicsMutation],
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
                        className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-[3px] text-xs"
                        style={{
                          backgroundColor: g.color ? `${g.color}18` : `${PALETTE.primary}0f`,
                          color: g.color ?? PALETTE.primary,
                        }}
                      >
                        <span
                          className="size-1.5 rounded-full shrink-0"
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
          allOccupations={allOccupations}
          allGroups={allGroups}
          onSaveGroups={handleSaveGroups}
          allClinics={allClinics}
          onSaveClinics={handleSaveClinics}
        />
      )}
    />
  );
}
