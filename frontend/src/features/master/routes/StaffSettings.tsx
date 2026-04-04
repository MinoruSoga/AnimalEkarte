import { memo, useState, useCallback } from "react";
import { UserRound } from "lucide-react";
import { TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { C, STYLE, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS, MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import type { FilterProperty } from "@/components/shared/NotionFilter/types";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import { useGetStaffs, useCreateStaff, useUpdateStaff, useDeleteStaff, STAFF_ROLE_LABELS } from "@/features/master/api/staffs";
import type { Staff, StaffRoleValue, CreateStaffRequest, UpdateStaffRequest } from "@/features/master/api/staffs";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/NotionFilter/types";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const COLUMNS = [
  { header: "氏名", className: "flex-1" },
  { header: "職種", className: "w-[130px]" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

const STAFF_ROLE_SELECT_ITEMS = (
  [
    { value: "veterinarian" as StaffRoleValue, label: "獣医師" },
    { value: "nurse" as StaffRoleValue, label: "看護師" },
    { value: "trimmer" as StaffRoleValue, label: "トリマー" },
    { value: "reception" as StaffRoleValue, label: "受付" },
    { value: "manager" as StaffRoleValue, label: "運営管理者" },
  ] satisfies { value: StaffRoleValue; label: string }[]
).map((r) => (
  <SelectItem key={r.value} value={r.value}>{r.label}</SelectItem>
));

const STAFF_FILTER_PROPERTIES: FilterProperty[] = [
  MASTER_STATUS_FILTER,
  {
    key: "staffRole",
    label: "職種",
    type: "select",
    icon: UserRound,
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "veterinarian", label: "獣医師" },
      { value: "nurse", label: "看護師" },
      { value: "trimmer", label: "トリマー" },
      { value: "reception", label: "受付" },
      { value: "manager", label: "運営管理者" },
    ],
  },
];

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

interface StaffFormData {
  name: string;
  staffRole: StaffRoleValue;
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
  onDeleteRequest: (i: Staff) => void;
}

const StaffSidePanel = memo(function StaffSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: StaffSidePanelProps) {
  const isNew = item === null;

  const [f, setF] = useState<StaffFormData>(() => ({
    name: item?.name ?? "",
    staffRole: item?.staffRole ?? "veterinarian",
    licenseNumber: item?.licenseNumber ?? "",
    isActive: item?.isActive ?? true,
    email: "",
    password: "",
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  // ── Handlers ─────────────────────────────────────
  const handleSave = useCallback(() => {
    if (!f.name.trim()) {
      setNameError("氏名を入力してください");
      return;
    }
    setNameError("");
    onSave(f);
    setIsDirty(false);
  }, [f, onSave]);

  const handleTitleChange = useCallback((v: string) => {
    setF((p) => ({ ...p, name: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleToggleActive = useCallback(() => {
    setF((p) => ({ ...p, isActive: !p.isActive }));
    setIsDirty(true);
  }, []);

  const handleStaffRoleChange = useCallback((v: string) => {
    setF((p) => ({ ...p, staffRole: v as StaffRoleValue }));
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
      onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      icon={<UserRound className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
    >
      <StatusToggleButton
        isActive={f.isActive}
        onToggle={handleToggleActive}
      />

      <PropertyRow label="職種">
        <Select
          value={f.staffRole}
          onValueChange={handleStaffRoleChange}
        >
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>{STAFF_ROLE_SELECT_ITEMS}</SelectContent>
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
      ) : null}
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

  const crud = useMasterCRUD<Staff>({
    data,
    deleteMutation,
    entityLabel: "スタッフ",
    searchFilter: (s, lower) =>
      s.name.toLowerCase().includes(lower) ||
      STAFF_ROLE_LABELS[s.staffRole].toLowerCase().includes(lower),
    activeFilterApply: (item, filters) => {
      for (const f of filters) {
        if (f.key === "status" && typeof f.value === "string") {
          const want = f.value === "active";
          if (f.condition === "is" && item.isActive !== want) return false;
          if (f.condition === "is_not" && item.isActive === want) return false;
        }
        if (f.key === "staffRole" && typeof f.value === "string") {
          if (f.condition === "is" && item.staffRole !== f.value) return false;
          if (f.condition === "is_not" && item.staffRole === f.value) return false;
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
      staff_role: d.staffRole,
      email: d.email,
      password: d.password,
      license_number: d.licenseNumber || undefined,
    }),
    toUpdateRequest: (d) => ({
      name: d.name,
      staff_role: d.staffRole,
      license_number: d.licenseNumber || undefined,
      is_active: d.isActive,
    }),
  });


  return (
    <MasterCRUDPage
      title="スタッフマスタ"
      icon={<UserRound className={`${ICON.page} ${C.text}`} />}
      entityLabel="スタッフ"
      searchPlaceholder="氏名、職種で検索..."
      emptyMessage="スタッフが登録されていません"
      crud={crud}
      handleSave={handleSave}
      columns={COLUMNS}
      filterProperties={STAFF_FILTER_PROPERTIES}
      renderRow={(item, onEdit) => (
        <DataTableRow key={item.id} onClick={() => onEdit(item)}>
          <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
          <TableCell className={`text-base ${C.text}`}>{STAFF_ROLE_LABELS[item.staffRole]}</TableCell>
          <TableCell className="text-center">
            <NotionStatusPill isActive={item.isActive} />
          </TableCell>
          <TableCell className="p-0 text-right">
            <RowActionButton onClick={() => onEdit(item)} />
          </TableCell>
        </DataTableRow>
      )}
      renderSidePanel={({ item, onClose, onSave, onDeleteRequest }) => (
        <StaffSidePanel
          key={item?.id ?? "new"}
          item={item}
          onClose={onClose}
          onSave={onSave}
          onDeleteRequest={onDeleteRequest}
        />
      )}
    />
  );
}
