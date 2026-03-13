// React/Framework
import type { ReactNode } from "react";
import { useState, useMemo } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, UserRound, X, Trash2 } from "lucide-react";
import { toast } from "sonner";

// Internal
import { TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { StatusBadge } from "@/components/shared/StatusBadge";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { getMasterStatusColor } from "@/utils/status-helpers";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import {
  useListStaffs,
  useCreateStaff,
  useUpdateStaff,
  useDeleteStaff,
  STAFF_ROLE_LABELS,
} from "@/features/master/api/staffs";

// Types
import type { Staff, StaffRoleValue, CreateStaffRequest, UpdateStaffRequest } from "@/features/master/api/staffs";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const STAFF_ROLE_OPTIONS: { value: StaffRoleValue; label: string }[] = [
  { value: "veterinarian", label: "獣医師" },
  { value: "nurse", label: "看護師" },
  { value: "trimmer", label: "トリマー" },
  { value: "reception", label: "受付" },
  { value: "manager", label: "運営管理者" },
];

const COLUMNS = [
  { header: "社員番号", className: "w-[130px]" },
  { header: "氏名" },
  { header: "職種", className: "w-[130px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

// ─────────────────────────────────────────────────
// Notion Status Pill
// ─────────────────────────────────────────────────

const STATUS_CONFIG = {
  active: {
    dot: "bg-[#2383E2]",
    label: "有効",
    bg: "bg-[#D3E5EF]",
    text: "text-[#183B56]",
  },
  inactive: {
    dot: "bg-[#37352F]/10",
    label: "無効",
    bg: "bg-[#E3E2E0]",
    text: "text-[#37352F]/60",
  },
} as const;

function NotionStatusPill({ status }: { status: "active" | "inactive" }) {
  const cfg = STATUS_CONFIG[status];
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-sm text-xs ${cfg.bg} ${cfg.text}`}
    >
      <span className={`size-[7px] rounded-full ${cfg.dot}`} />
      {cfg.label}
    </span>
  );
}

// ─────────────────────────────────────────────────
// Property Row (Notion-style)
// ─────────────────────────────────────────────────

function PropertyRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div
      className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] transition-colors min-h-[40px]`}
    >
      <div className="w-[140px] shrink-0 text-sm text-[#37352F]/65 select-none truncate flex items-center">
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// Form state types
// ─────────────────────────────────────────────────

interface StaffFormData {
  name: string;
  staff_role: StaffRoleValue;
  code: string;
  license_number: string;
  is_active: boolean;
  email: string;
  password: string;
}

const DEFAULT_FORM_DATA: StaffFormData = {
  name: "",
  staff_role: "veterinarian",
  code: "",
  license_number: "",
  is_active: true,
  email: "",
  password: "",
};

// ─────────────────────────────────────────────────
// StaffSettings
// ─────────────────────────────────────────────────

export function StaffSettings() {
  const navigate = useNavigate();
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<Staff | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [formData, setFormData] = useState<StaffFormData>(DEFAULT_FORM_DATA);
  const [pendingDelete, setPendingDelete] = useState<Staff | null>(null);

  const { data: rawStaffs } = useListStaffs();
  const staffs = rawStaffs ?? [];
  const createMutation = useCreateStaff();
  const updateMutation = useUpdateStaff();
  const deleteMutation = useDeleteStaff();

  const filteredItems = useMemo(() => {
    if (!searchTerm) return staffs;
    const lower = searchTerm.toLowerCase();
    return staffs.filter(
      (s) =>
        s.name.toLowerCase().includes(lower) ||
        STAFF_ROLE_LABELS[s.staffRole].toLowerCase().includes(lower),
    );
  }, [staffs, searchTerm]);

  const handleEdit = (item: Staff) => {
    setSelectedItem(item);
    setFormData({
      name: item.name,
      staff_role: item.staffRole,
      code: item.code,
      license_number: item.licenseNumber,
      is_active: item.isActive,
      email: "",
      password: "",
    });
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setFormData(DEFAULT_FORM_DATA);
    setIsEditing(true);
  };

  const handleCloseEdit = () => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData(DEFAULT_FORM_DATA);
  };

  const handleSave = () => {
    if (!formData.name) {
      toast.error("氏名は必須です");
      return;
    }

    if (selectedItem) {
      const req: UpdateStaffRequest = {
        name: formData.name,
        staff_role: formData.staff_role,
        code: formData.code || undefined,
        license_number: formData.license_number || undefined,
        is_active: formData.is_active,
      };
      updateMutation.mutate(
        { id: selectedItem.id, req },
        {
          onSuccess: () => {
            toast.success("更新しました");
            setIsEditing(false);
          },
          onError: () => {
            toast.error("更新に失敗しました");
          },
        },
      );
    } else {
      if (!formData.email) {
        toast.error("メールアドレスは必須です");
        return;
      }
      if (!formData.password) {
        toast.error("パスワードは必須です");
        return;
      }
      const req: CreateStaffRequest = {
        name: formData.name,
        staff_role: formData.staff_role,
        email: formData.email,
        password: formData.password,
        code: formData.code || undefined,
        license_number: formData.license_number || undefined,
      };
      createMutation.mutate(req, {
        onSuccess: () => {
          toast.success("登録しました");
          setIsEditing(false);
        },
        onError: () => {
          toast.error("登録に失敗しました");
        },
      });
    }
  };

  const handleDeleteConfirm = () => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        setIsEditing(false);
        toast.success("削除しました");
      },
      onError: () => {
        toast.error("削除に失敗しました");
      },
    });
  };

  return (
    <>
      <div className="flex h-full">
        {/* ── Left: List ── */}
        <div className="flex-1 min-w-0">
          <PageLayout
            title="スタッフマスタ"
            icon={<UserRound className="size-5 text-[#37352F]" />}
            onBack={() => navigate("/settings")}
            headerAction={
              <PrimaryButton onClick={handleCreate}>
                <Plus className="mr-1.5 size-4" />
                新規登録
              </PrimaryButton>
            }
            maxWidth="max-w-full"
          >
            <div className="flex flex-col gap-4">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="社員番号、氏名、職種で検索..."
                count={filteredItems.length}
              />

              <DataTable
                columns={COLUMNS}
                data={filteredItems}
                emptyMessage="スタッフが登録されていません"
                renderRow={(item) => (
                  <DataTableRow key={item.id} onClick={() => handleEdit(item)}>
                    <TableCell className="font-mono text-sm text-[#37352F]/80 py-2.5">
                      {item.code}
                    </TableCell>
                    <TableCell className="font-medium text-sm text-[#37352F] py-2.5">
                      {item.name}
                    </TableCell>
                    <TableCell className="text-sm text-[#37352F] py-2.5">
                      {STAFF_ROLE_LABELS[item.staffRole]}
                    </TableCell>
                    <TableCell className="text-center py-2.5">
                      <StatusBadge colorClass={getMasterStatusColor(item.isActive ? "active" : "inactive")}>
                        {item.isActive ? "有効" : "無効"}
                      </StatusBadge>
                    </TableCell>
                    <TableCell className="text-right py-2.5">
                      <RowActionButton onClick={() => handleEdit(item)} />
                    </TableCell>
                  </DataTableRow>
                )}
              />
            </div>
          </PageLayout>
        </div>

        {/* ── Right: Side Peek Panel ── */}
        {isEditing && (
          <div
            className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0 flex flex-col`}
          >
            {/* Toolbar */}
            <div className={STYLE.sidePeekToolbar}>
              <span className="text-xs text-[#37352F]/35 pl-1 select-none">
                {selectedItem ? "編集" : "新規作成"}
              </span>
              <div className="flex items-center gap-1">
                {selectedItem ? (
                  <button
                    type="button"
                    onClick={() => setPendingDelete(selectedItem)}
                    className={`${STYLE.sidePeekToolbarBtn} cursor-pointer text-[#EB5757] hover:bg-[#EB5757]/10`}
                  >
                    <Trash2 className="size-4" />
                  </button>
                ) : null}
                <button
                  type="button"
                  onClick={handleCloseEdit}
                  className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
                >
                  <X className="size-4" />
                </button>
              </div>
            </div>

            {/* Body */}
            <div className={STYLE.sidePeekBody}>
              <div className="px-16 pb-8">
                {/* Page icon */}
                <div className="pt-4 pb-2">
                  <div className={STYLE.pageIcon}>
                    <UserRound className={LAYOUT.pageIcon.innerIcon} />
                  </div>
                </div>

                {/* Title input (name) */}
                <div className="pb-1 mb-4">
                  <input
                    type="text"
                    className={`w-full bg-transparent ${C.text} placeholder:text-[rgba(55,53,47,0.15)] outline-none border-none p-0`}
                    style={{
                      fontSize: LAYOUT.pageTitle.fontSize,
                      fontWeight: LAYOUT.pageTitle.fontWeight,
                      lineHeight: LAYOUT.pageTitle.lineHeight,
                    }}
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="無題"
                  />
                </div>

                {/* Separator */}
                <div className={`${STYLE.sectionDivider} mb-1`} />

                {/* Properties */}
                <div className="py-1">
                  {/* Status */}
                  <PropertyRow label="ステータス">
                    <button
                      type="button"
                      onClick={() =>
                        setFormData({
                          ...formData,
                          is_active: !formData.is_active,
                        })
                      }
                      className="inline-flex items-center rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] transition-colors py-0.5 px-0.5 cursor-pointer"
                    >
                      <NotionStatusPill
                        status={formData.is_active ? "active" : "inactive"}
                      />
                    </button>
                  </PropertyRow>

                  {/* 社員番号 */}
                  <PropertyRow label="社員番号">
                    <input
                      type="text"
                      className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] focus:bg-[rgba(55,53,47,0.04)] transition-colors placeholder:text-[rgba(55,53,47,0.3)]`}
                      value={formData.code}
                      onChange={(e) => setFormData({ ...formData, code: e.target.value })}
                      placeholder="例: ST-001"
                    />
                  </PropertyRow>

                  {/* 職種 */}
                  <PropertyRow label="職種">
                    <Select
                      value={formData.staff_role}
                      onValueChange={(val) =>
                        setFormData({ ...formData, staff_role: val as StaffRoleValue })
                      }
                    >
                      <SelectTrigger className={STYLE.selectCompact}>
                        <SelectValue placeholder="選択" />
                      </SelectTrigger>
                      <SelectContent>
                        {STAFF_ROLE_OPTIONS.map((r) => (
                          <SelectItem key={r.value} value={r.value}>
                            {r.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </PropertyRow>

                  {/* 資格番号 */}
                  <PropertyRow label="資格番号">
                    <input
                      type="text"
                      className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] focus:bg-[rgba(55,53,47,0.04)] transition-colors placeholder:text-[rgba(55,53,47,0.3)]`}
                      value={formData.license_number}
                      onChange={(e) =>
                        setFormData({ ...formData, license_number: e.target.value })
                      }
                      placeholder="空"
                    />
                  </PropertyRow>

                  {/* 新規作成時のみ: email / password */}
                  {!selectedItem && (
                    <>
                      <PropertyRow label="メールアドレス">
                        <input
                          type="email"
                          className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] focus:bg-[rgba(55,53,47,0.04)] transition-colors placeholder:text-[rgba(55,53,47,0.3)]`}
                          value={formData.email}
                          onChange={(e) =>
                            setFormData({ ...formData, email: e.target.value })
                          }
                          placeholder="例: staff@clinic.com"
                        />
                      </PropertyRow>

                      <PropertyRow label="パスワード">
                        <input
                          type="password"
                          className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] focus:bg-[rgba(55,53,47,0.04)] transition-colors placeholder:text-[rgba(55,53,47,0.3)]`}
                          value={formData.password}
                          onChange={(e) =>
                            setFormData({ ...formData, password: e.target.value })
                          }
                          placeholder="8文字以上"
                        />
                      </PropertyRow>
                    </>
                  )}
                </div>
              </div>
            </div>

            {/* Footer */}
            <div className={STYLE.sidePeekFooter}>
              <button
                type="button"
                onClick={handleCloseEdit}
                className={STYLE.sidePeekCancelBtn}
              >
                キャンセル
              </button>
              <button
                type="button"
                onClick={handleSave}
                className={STYLE.sidePeekSaveBtn}
              >
                保存
              </button>
            </div>
          </div>
        )}
      </div>

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        title="スタッフを削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}
