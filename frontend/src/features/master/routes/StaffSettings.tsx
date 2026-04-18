import { memo, useState, useMemo, useCallback, useEffect } from "react";
import { UserRound, Shield, Building2, MessageCircle, Ban } from "lucide-react";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { Switch } from "@/components/ui/switch";
import { TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow, StatusToggleButton, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, STYLE, LAYOUT, ICON, PALETTE } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS, MASTER_STATUS_FILTER } from "../constants/styles";
import type { FilterProperty } from "@/components/shared/NotionFilter/types";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
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
  useGetStaffExcludedReservationTypes,
  useSetStaffExcludedReservationTypes,
} from "../api/staffs";
import type { Staff, CreateStaffRequest, UpdateStaffRequest } from "../api/staffs";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/NotionFilter/types";
import { useGetPermissionGroups } from "../api/permission-groups";
import type { PermissionGroup } from "../api/permission-groups";
import type { ClinicSummary } from "../api/staffs";
import { useGetAllOccupations } from "../api/occupations";
import { useGetReservationTypes } from "../api/reservation-types";
import type { ReservationType } from "../api/reservation-types";
import type { Occupation } from "../api/occupations";
import { ResourceMasterStaff } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

// rendering-hoist-jsx: 静的 SelectItem JSX をモジュール定数に巻き上げ
const STAFF_TYPE_SELECT_ITEMS = (
  <>
    <SelectItem value="doctor">医師</SelectItem>
    <SelectItem value="nurse">看護師</SelectItem>
    <SelectItem value="resource">設備</SelectItem>
  </>
);

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
  // LINE予約用
  staffType: string;
  reservationDisplayName: string;
  reservationVisible: boolean;
  reservationComment: string;
  reservationImageUrl: string;
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
  onDirtyChange?: (dirty: boolean) => void;
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
  /** All reservation categorys for excluded courses */
  allReservationTypes: ReservationType[];
  /** Called when excluded reservation categorys should be saved */
  onSaveExcludedReservationTypes: (staffId: string, reservationTypeIds: string[]) => void;
}

const StaffSidePanel = memo(function StaffSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
  allOccupations,
  allGroups,
  onSaveGroups,
  allClinics,
  onSaveClinics,
  allReservationTypes,
  onSaveExcludedReservationTypes,
}: StaffSidePanelProps) {
  const isNew = item === null;
  const staffId = item?.id ?? null;

  const [formData, setFormData] = useState<StaffFormData>(() => ({
    name: item?.name ?? "",
    jobTitleId: item?.occupationId ?? null,
    licenseNumber: item?.licenseNumber ?? "",
    isActive: item?.isActive ?? true,
    email: item?.email ?? "",
    password: "",
    staffType: item?.staffType ?? "doctor",
    reservationDisplayName: item?.reservationDisplayName ?? "",
    reservationVisible: item?.reservationVisible ?? true,
    reservationComment: item?.reservationComment ?? "",
    reservationImageUrl: item?.reservationImageUrl ?? "",
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  // BUG-380
  useEffect(() => { onDirtyChange?.(isDirty); }, [isDirty, onDirtyChange]);
  const markDirty = useCallback(() => setIsDirty(true), []);
  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    markDirty();
  }, [markDirty]);

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

  // js-combine-iterations: filter().map() を一回の reduce に統合（active サービスタイプのみ）
  const activeReservationTypes = useMemo(
    () => allReservationTypes.filter((st) => st.isActive),
    [allReservationTypes],
  );

  // ── Permission groups state ──────────────────────
  const { data: serverGroupIds } = useGetStaffPermissionGroups(staffId);
  const [userEditedGroupIds, setUserEditedGroupIds] = useState<string[] | null>(null);

  const groupIds = useMemo(
    () => userEditedGroupIds ?? serverGroupIds ?? [],
    [userEditedGroupIds, serverGroupIds],
  );
  // js-set-map-lookups: O(n) includes() → Set.has() O(1)
  const groupIdSet = useMemo(() => new Set(groupIds), [groupIds]);

  // ── Clinic assignments state ───────────────────
  const { data: serverClinicIds } = useGetStaffClinics(staffId);
  const [userEditedClinicIds, setUserEditedClinicIds] = useState<string[] | null>(null);

  const clinicIds = useMemo(
    () => userEditedClinicIds ?? serverClinicIds ?? [],
    [userEditedClinicIds, serverClinicIds],
  );
  // js-set-map-lookups: O(n) includes() → Set.has() O(1)
  const clinicIdSet = useMemo(() => new Set(clinicIds), [clinicIds]);

  // ── Excluded reservation categorys state ─────────────────
  const { data: serverExcludedIds } = useGetStaffExcludedReservationTypes(staffId);
  const [userEditedExcludedIds, setUserEditedExcludedIds] = useState<string[] | null>(null);

  const excludedIds = useMemo(
    () => userEditedExcludedIds ?? serverExcludedIds ?? [],
    [userEditedExcludedIds, serverExcludedIds],
  );
  // js-set-map-lookups: O(n) includes() → Set.has() O(1)
  const excludedIdSet = useMemo(() => new Set(excludedIds), [excludedIds]);

  // ── Handlers ─────────────────────────────────────
  const handleExcludedToggle = useCallback(
    (reservationTypeId: string, checked: boolean) => {
      setUserEditedExcludedIds((prev) => {
        const current = prev ?? serverExcludedIds ?? [];
        return checked ? [...current, reservationTypeId] : current.filter((id) => id !== reservationTypeId);
      });
      markDirty();
    },
    [serverExcludedIds, markDirty],
  );

  const handleClinicToggle = useCallback(
    (clinicId: string, checked: boolean) => {
      setUserEditedClinicIds((prev) => {
        const current = prev ?? serverClinicIds ?? [];
        return checked ? [...current, clinicId] : current.filter((id) => id !== clinicId);
      });
      markDirty();
    },
    [serverClinicIds, markDirty],
  );

  const handleGroupToggle = useCallback(
    (groupId: string, checked: boolean) => {
      setUserEditedGroupIds((prev) => {
        const current = prev ?? serverGroupIds ?? [];
        return checked ? [...current, groupId] : current.filter((id) => id !== groupId);
      });
      markDirty();
    },
    [serverGroupIds, markDirty],
  );

  const handleSave = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("氏名を入力してください");
      return;
    }
    setNameError("");
    onSave(formData);
    if (!isNew && staffId) {
      onSaveGroups(staffId, groupIds);
      onSaveClinics(staffId, clinicIds);
      onSaveExcludedReservationTypes(staffId, excludedIds);
    }
    setIsDirty(false);
  }, [formData, isNew, staffId, groupIds, clinicIds, excludedIds, onSave, onSaveGroups, onSaveClinics, onSaveExcludedReservationTypes]);

  const handleTitleChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: v }));
    if (v.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleOccupationChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, jobTitleId: v }));
  }, [setFormDataDirty]);

  const handleLicenseNumberChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, licenseNumber: e.target.value }));
  }, [setFormDataDirty]);

  const handleEmailChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, email: e.target.value }));
  }, [setFormDataDirty]);

  const handlePasswordChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, password: e.target.value }));
  }, [setFormDataDirty]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel
      isNew={isNew}
      title={formData.name}
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
        isActive={formData.isActive}
        onToggle={handleToggleActive}
      />

      <PropertyRow label="職種">
        <Select
          value={formData.jobTitleId ?? undefined}
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
          value={formData.licenseNumber}
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
              value={formData.email}
              onChange={handleEmailChange}
              placeholder="例: staff@clinic.com"
            />
          </PropertyRow>
          <PropertyRow label="パスワード">
            <input
              type="password"
              className={MASTER_INPUT_CLASS}
              value={formData.password}
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
              value={formData.password}
              onChange={handlePasswordChange}
              placeholder="変更する場合のみ入力"
            />
          </PropertyRow>
        </>
      )}

      {/* ── LINE予約設定 ─────────────────────────── */}
      <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
        <div className="flex items-center gap-1.5 mb-3">
          <MessageCircle className={ICON.smXs} style={{ color: PALETTE.lineGreen }} />
          <p className={`text-xs font-medium ${C.text50}`}>LINE予約設定</p>
        </div>

        <PropertyRow label="LINE表示名">
          <input
            type="text"
            className={MASTER_INPUT_CLASS}
            value={formData.reservationDisplayName}
            onChange={(e) => setFormDataDirty((prev) => ({ ...prev, reservationDisplayName: e.target.value }))}
            placeholder={formData.name || "空欄なら氏名を使用"}
          />
        </PropertyRow>

        <PropertyRow label="予約ページに表示">
          <Switch
            checked={formData.reservationVisible}
            onCheckedChange={(v) => setFormDataDirty((prev) => ({ ...prev, reservationVisible: v }))}
          />
        </PropertyRow>

        <PropertyRow label="スタッフ種別">
          <Select
            value={formData.staffType}
            onValueChange={(v) => setFormDataDirty((prev) => ({ ...prev, staffType: v }))}
          >
            <SelectTrigger className={STYLE.selectCompact}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{STAFF_TYPE_SELECT_ITEMS}</SelectContent>
          </Select>
        </PropertyRow>

        <PropertyRow label="LINE説明文">
          <input
            type="text"
            className={MASTER_INPUT_CLASS}
            value={formData.reservationComment}
            onChange={(e) => setFormDataDirty((prev) => ({ ...prev, reservationComment: e.target.value }))}
            placeholder="LINE予約画面に表示する説明"
          />
        </PropertyRow>

        <PropertyRow label="画像URL">
          <input
            type="text"
            className={MASTER_INPUT_CLASS}
            value={formData.reservationImageUrl}
            onChange={(e) => setFormDataDirty((prev) => ({ ...prev, reservationImageUrl: e.target.value }))}
            placeholder="https://..."
          />
        </PropertyRow>
      </div>

      {/* ── Excluded reservation categorys ───────────────── */}
      <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
        <div className="flex items-center gap-1.5 mb-2">
          <Ban className={`${ICON.xs} ${C.text50}`} />
          <p className={`text-xs font-medium ${C.text50}`}>対応不可コース</p>
        </div>

        {isNew ? (
          <p className={`text-xs ${C.text50} pl-0.5`}>
            スタッフ登録後に設定できます
          </p>
        ) : allReservationTypes.length === 0 ? (
          <p className={`text-xs ${C.text50} pl-0.5`}>
            予約区分が登録されていません
          </p>
        ) : (
          <div className="space-y-0.5">
            {activeReservationTypes.map((st) => (
              <label
                key={st.id}
                className={`flex items-center gap-2.5 py-1.5 px-0.5 rounded cursor-pointer ${C.hoverBgLight} transition-colors`}
              >
                <Checkbox
                  checked={excludedIdSet.has(st.id)}
                  onCheckedChange={(checked) =>
                    handleExcludedToggle(st.id, checked === true)
                  }
                />
                <span
                  className={`${ICON.dotMd} rounded-full shrink-0`}
                  style={{ backgroundColor: st.color }}
                />
                <span className="text-sm">{st.name}</span>
              </label>
            ))}
          </div>
        )}
      </div>

      {/* ── Clinic assignments ─────────────────────── */}
      <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
        <div className="flex items-center gap-1.5 mb-2">
          <Building2 className={`${ICON.xs} ${C.text50}`} />
          <p className={`text-xs font-medium ${C.text50}`}>所属医院</p>
        </div>

        {isNew ? (
          <p className={`text-xs ${C.text50} pl-0.5`}>
            スタッフ登録後に所属医院を設定できます
          </p>
        ) : allClinics.length === 0 ? (
          <p className={`text-xs ${C.text50} pl-0.5`}>
            医院が登録されていません
          </p>
        ) : (
          <div className="space-y-0.5">
            {allClinics.map((clinic) => (
              <label
                key={clinic.id}
                className={`flex items-center gap-2.5 py-1.5 px-0.5 rounded cursor-pointer ${C.hoverBgLight} transition-colors`}
              >
                <Checkbox
                  checked={clinicIdSet.has(clinic.id)}
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
      <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
        <div className="flex items-center gap-1.5 mb-2">
          <Shield className={`${ICON.xs} ${C.text50}`} />
          <p className={`text-xs font-medium ${C.text50}`}>権限グループ</p>
        </div>

        {isNew ? (
          <p className={`text-xs ${C.text50} pl-0.5`}>
            スタッフ登録後に権限グループを設定できます
          </p>
        ) : allGroups.length === 0 ? (
          <p className={`text-xs ${C.text50} pl-0.5`}>
            権限グループが登録されていません
          </p>
        ) : (
          <div className="space-y-0.5">
            {allGroups.map((group) => (
              <label
                key={group.id}
                className={`flex items-center gap-2.5 py-1.5 px-0.5 rounded cursor-pointer ${C.hoverBgLight} transition-colors`}
              >
                <Checkbox
                  checked={groupIdSet.has(group.id)}
                  onCheckedChange={(checked) =>
                    handleGroupToggle(group.id, checked === true)
                  }
                />
                {/* BUG-323: Status Dot Icon Token 使用統一 */}
                <div
                  className={`${ICON.dotMd} rounded-full flex-shrink-0`}
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

  // Service Types — shown as checkboxes in excluded courses panel
  const { data: allReservationTypesData } = useGetReservationTypes();
  const allReservationTypes = useMemo(() => allReservationTypesData ?? [], [allReservationTypesData]);

  // Clinics — shown as checkboxes in the panel
  const { data: allClinicsData } = useGetClinicsList("all");
  const allClinics = useMemo(() => allClinicsData ?? [], [allClinicsData]);

  // Group / Clinic / Excluded mutation
  const setGroupsMutation = useSetStaffPermissionGroups();
  const { mutate: setGroupsFn } = setGroupsMutation;
  const setClinicsMutation = useSetStaffClinics();
  const { mutate: setClinicsFn } = setClinicsMutation;
  const setExcludedMutation = useSetStaffExcludedReservationTypes();
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
                        className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-[3px] text-xs"
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
