// React/Framework
import { useState, useMemo } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

// External
import { Plus, UserRound } from "lucide-react";
import { toast } from "sonner";

// Internal
import { TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import {
  PropertyRow,
  SidePeekPanel,
  SidePeekToolbar,
  SidePeekBody,
  SidePeekTitleInput,
  SidePeekFooter,
} from "@/components/shared/SidePeek";
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
  { header: "ステータス", className: "w-[90px]", align: "right" as const },
];

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
  const createMutation = useCreateStaff();
  const updateMutation = useUpdateStaff();
  const deleteMutation = useDeleteStaff();

  // rerender-dependencies: rawStaffs ?? [] は毎回新参照を生成するため useMemo 内に移動
  const filteredItems = useMemo(() => {
    const staffs = rawStaffs ?? [];
    if (!searchTerm) return staffs;
    const lower = searchTerm.toLowerCase();
    return staffs.filter(
      (s) =>
        s.name.toLowerCase().includes(lower) ||
        STAFF_ROLE_LABELS[s.staffRole].toLowerCase().includes(lower),
    );
  }, [rawStaffs, searchTerm]);

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
            onBack={() => navigate(paths.settings.getHref())}
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
                    <TableCell className="font-mono text-sm text-[#37352F]/80">
                      {item.code}
                    </TableCell>
                    <TableCell className="font-medium text-sm text-[#37352F]">
                      {item.name}
                    </TableCell>
                    <TableCell className="text-sm text-[#37352F]">
                      {STAFF_ROLE_LABELS[item.staffRole]}
                    </TableCell>
                    <TableCell className="text-right">
                      <NotionStatusPill isActive={item.isActive} />
                    </TableCell>
                  </DataTableRow>
                )}
              />
              <button
                type="button"
                onClick={handleCreate}
                className="flex items-center gap-1.5 w-full px-3 py-2 text-sm text-[#37352F]/40 hover:text-[#37352F]/65 hover:bg-[rgba(55,53,47,0.04)] transition-colors rounded"
              >
                <Plus className="size-3.5" />
                新しいスタッフを追加...
              </button>
            </div>
          </PageLayout>
        </div>

        {/* ── Right: Side Peek Panel ── */}
        {isEditing ? (
          <SidePeekPanel>
            <SidePeekToolbar
              isNew={selectedItem === null}
              onClose={handleCloseEdit}
              onDelete={selectedItem !== null ? () => setPendingDelete(selectedItem) : undefined}
            />
            <SidePeekBody>
              <div className="pt-4 pb-2">
                <div className={STYLE.pageIcon}>
                  <UserRound className={LAYOUT.pageIcon.innerIcon} />
                </div>
              </div>
              <SidePeekTitleInput
                value={formData.name}
                onChange={(v) => setFormData({ ...formData, name: v })}
              />
              <div className={`${STYLE.sectionDivider} mb-1`} />
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
                    <NotionStatusPill isActive={formData.is_active} />
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
                {selectedItem === null ? (
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
                ) : null}
              </div>
            </SidePeekBody>
            <SidePeekFooter onCancel={handleCloseEdit} onSave={handleSave} />
          </SidePeekPanel>
        ) : null}
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
