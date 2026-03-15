// React/Framework
import { useState, useMemo, useCallback, memo, useDeferredValue, useTransition } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

// External
import { Plus, UserRound } from "lucide-react";
import { toast } from "sonner";

// Internal
import { TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar/SearchFilterBar";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { SidePeekPanel } from "@/components/shared/SidePeek/SidePeekPanel";
import { SidePeekToolbar } from "@/components/shared/SidePeek/SidePeekToolbar";
import { SidePeekBody } from "@/components/shared/SidePeek/SidePeekBody";
import { SidePeekTitleInput } from "@/components/shared/SidePeek/SidePeekTitleInput";
import { SidePeekFooter } from "@/components/shared/SidePeek/SidePeekFooter";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import {
  useGetStaffs,
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

const COLUMNS = [
  { header: "氏名", className: "flex-1" },
  { header: "職種", className: "w-[130px]" },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

// Hoisted CSS class string (rendering-hoist-jsx)
const INPUT_CLASS = `w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`;

// Hoisted static JSX (rendering-hoist-jsx)
const STAFF_ROLE_SELECT_ITEMS = (
  [
    { value: "veterinarian" as StaffRoleValue, label: "獣医師" },
    { value: "nurse" as StaffRoleValue, label: "看護師" },
    { value: "trimmer" as StaffRoleValue, label: "トリマー" },
    { value: "reception" as StaffRoleValue, label: "受付" },
    { value: "manager" as StaffRoleValue, label: "運営管理者" },
  ] satisfies { value: StaffRoleValue; label: string }[]
).map((r) => (
  <SelectItem key={r.value} value={r.value}>
    {r.label}
  </SelectItem>
));

// ─────────────────────────────────────────────────
// Form state types
// ─────────────────────────────────────────────────

interface StaffFormData {
  name: string;
  staffRole: StaffRoleValue;
  code: string;
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
  onSave: (data: StaffFormData) => void;
  onDeleteRequest: (item: Staff) => void;
}

const StaffSidePanel = memo(function StaffSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: StaffSidePanelProps) {
  // rerender-lazy-state-init: 初回マウント時のみ item から初期化
  const [formData, setFormData] = useState<StaffFormData>(() => ({
    name: item?.name ?? "",
    staffRole: item?.staffRole ?? "veterinarian",
    code: item?.code ?? "",
    licenseNumber: item?.licenseNumber ?? "",
    isActive: item?.isActive ?? true,
    email: "",
    password: "",
  }));

  return (
    <SidePeekPanel>
      <SidePeekToolbar
        isNew={item === null}
        onClose={onClose}
        onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      />
      <SidePeekBody>
        <div className="pt-4 pb-2">
          <div className={STYLE.pageIcon}>
            <UserRound className={LAYOUT.pageIcon.innerIcon} />
          </div>
        </div>
        <SidePeekTitleInput
          value={formData.name}
          onChange={(v) => setFormData((prev) => ({ ...prev, name: v }))}
        />
        <div className={`${STYLE.sectionDivider} mb-1`} />
        <div className="py-1">
          <StatusToggleButton
            isActive={formData.isActive}
            onToggle={() => setFormData((prev) => ({ ...prev, isActive: !prev.isActive }))}
          />

          <PropertyRow label="社員番号">
            <input
              type="text"
              className={INPUT_CLASS}
              value={formData.code}
              onChange={(e) => setFormData((prev) => ({ ...prev, code: e.target.value }))}
              placeholder="例: ST-001"
            />
          </PropertyRow>

          <PropertyRow label="職種">
            <Select
              value={formData.staffRole}
              onValueChange={(v) => setFormData((prev) => ({ ...prev, staffRole: v as StaffRoleValue }))}
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
              className={INPUT_CLASS}
              value={formData.licenseNumber}
              onChange={(e) => setFormData((prev) => ({ ...prev, licenseNumber: e.target.value }))}
              placeholder="空"
            />
          </PropertyRow>

          {/* 新規作成時のみ: email / password */}
          {item === null ? (
            <>
              <PropertyRow label="メールアドレス">
                <input
                  type="email"
                  className={INPUT_CLASS}
                  value={formData.email}
                  onChange={(e) => setFormData((prev) => ({ ...prev, email: e.target.value }))}
                  placeholder="例: staff@clinic.com"
                />
              </PropertyRow>

              <PropertyRow label="パスワード">
                <input
                  type="password"
                  className={INPUT_CLASS}
                  value={formData.password}
                  onChange={(e) => setFormData((prev) => ({ ...prev, password: e.target.value }))}
                  placeholder="8文字以上"
                />
              </PropertyRow>
            </>
          ) : null}
        </div>
      </SidePeekBody>
      <SidePeekFooter onCancel={onClose} onSave={() => onSave(formData)} />
    </SidePeekPanel>
  );
});

// ─────────────────────────────────────────────────
// StaffSettings (main page)
// ─────────────────────────────────────────────────

export function StaffSettings() {
  const navigate = useNavigate();

  // null=closed, "new"=create mode, Staff=edit mode
  const [editTarget, setEditTarget] = useState<Staff | "new" | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<Staff | null>(null);
  const [, startSaveTransition] = useTransition();

  const { data: rawStaffs } = useGetStaffs();
  const createMutation = useCreateStaff();
  const updateMutation = useUpdateStaff();
  const deleteMutation = useDeleteStaff();

  // rerender-transitions: 検索フィルタを低優先度に遅延
  const deferredSearch = useDeferredValue(searchTerm);

  // rerender-dependencies: rawStaffs ?? [] は毎回新参照を生成するため useMemo 内に移動
  const filteredItems = useMemo(() => {
    const staffs = rawStaffs ?? [];
    if (!deferredSearch) return staffs;
    const lower = deferredSearch.toLowerCase();
    return staffs.filter(
      (s) =>
        s.name.toLowerCase().includes(lower) ||
        STAFF_ROLE_LABELS[s.staffRole].toLowerCase().includes(lower),
    );
  }, [rawStaffs, deferredSearch]);

  const handleClose = useCallback(() => setEditTarget(null), []);

  const handleSave = useCallback(
    (data: StaffFormData) => {
      // バリデーションは startSaveTransition の外で実施
      if (!data.name.trim()) {
        toast.error("氏名は必須です");
        return;
      }
      if (editTarget === null || editTarget === "new") {
        if (!data.email) {
          toast.error("メールアドレスは必須です");
          return;
        }
        if (!data.password || data.password.length < 8) {
          toast.error("パスワードは8文字以上で入力してください");
          return;
        }
      }

      // rerender-transitions: API書き込みを非緊急マーク
      startSaveTransition(() => {
        if (editTarget !== null && editTarget !== "new") {
          const req: UpdateStaffRequest = {
            name: data.name,
            staff_role: data.staffRole,
            code: data.code || undefined,
            license_number: data.licenseNumber || undefined,
            is_active: data.isActive,
          };
          updateMutation.mutate(
            { id: editTarget.id, req },
            {
              onSuccess: () => {
                toast.success("更新しました");
                handleClose();
              },
              onError: () => toast.error("更新に失敗しました"),
            },
          );
        } else {
          const req: CreateStaffRequest = {
            name: data.name,
            staff_role: data.staffRole,
            email: data.email,
            password: data.password,
            code: data.code || undefined,
            license_number: data.licenseNumber || undefined,
          };
          createMutation.mutate(req, {
            onSuccess: () => {
              toast.success("登録しました");
              handleClose();
            },
            onError: () => toast.error("登録に失敗しました"),
          });
        }
      });
    },
    [editTarget, updateMutation, createMutation, handleClose],
  );

  const handleDeleteConfirm = useCallback(() => {
    if (!pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        handleClose();
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  }, [pendingDelete, deleteMutation, handleClose]);

  const panelItem = editTarget !== null && editTarget !== "new" ? editTarget : null;

  return (
    <PageLayout
      title="スタッフマスタ"
      icon={<UserRound className="size-5 text-[#37352F]" />}
      onBack={() => navigate(paths.settings.getHref())}
      maxWidth="max-w-full"
    >
      <div className="flex h-full">
        {/* Table area */}
        <div className="flex flex-col gap-4 flex-1 min-w-0">
          <div className="flex items-center gap-3">
            <div className="flex-1 min-w-0">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder="氏名、職種で検索..."
                count={filteredItems.length}
              />
            </div>
            <button
              type="button"
              onClick={() => setEditTarget("new")}
              className="inline-flex items-center gap-1 text-sm font-medium text-[#2383E2] hover:text-[#1B6EC2] cursor-pointer transition-colors"
            >
              <Plus className="size-4" />
              新規登録
            </button>
          </div>

          <DataTable
            columns={COLUMNS}
            data={filteredItems}
            emptyMessage="スタッフが登録されていません"
            renderRow={(item) => (
              <DataTableRow key={item.id} onClick={() => setEditTarget(item)}>
                <TableCell className={`font-medium text-sm ${C.text}`}>
                  {item.name}
                </TableCell>
                <TableCell className={`text-sm ${C.text}`}>
                  {STAFF_ROLE_LABELS[item.staffRole]}
                </TableCell>
                <TableCell className="text-center">
                  <NotionStatusPill isActive={item.isActive} />
                </TableCell>
                <TableCell className="p-0 text-right">
                  <RowActionButton onClick={() => setEditTarget(item)} />
                </TableCell>
              </DataTableRow>
            )}
          />
        </div>

        {/* Side peek */}
        {editTarget !== null ? (
          <StaffSidePanel
            key={panelItem ? String(panelItem.id) : "new-staff"}
            item={panelItem}
            onClose={handleClose}
            onSave={handleSave}
            onDeleteRequest={setPendingDelete}
          />
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
    </PageLayout>
  );
}
