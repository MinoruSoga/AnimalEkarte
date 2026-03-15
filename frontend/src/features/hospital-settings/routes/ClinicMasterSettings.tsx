// React/Framework
import type { ReactNode } from "react";
import { useDeferredValue, useMemo, useState } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

// External
import { Plus, Building2, X, Trash2 } from "lucide-react";
import { toast } from "sonner";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import {
  useListClinics,
  useCreateClinic,
  useUpdateClinic,
  useDeleteClinic,
} from "@/features/hospital-settings/api/clinics";

// Types
import type {
  Clinic,
  CreateClinicRequest,
  UpdateClinicRequest,
} from "@/features/hospital-settings/api/clinics";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const COLUMNS = [
  { header: "院名" },
  { header: "電話番号", className: "w-[150px]" },
  { header: "メール" },
  { header: "ステータス", className: "w-[90px]", align: "right" as const },
];

// ─────────────────────────────────────────────────
// Property Row (Notion-style)
// ─────────────────────────────────────────────────

function PropertyRow({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) {
  return (
    <div className="flex gap-2 py-2 px-2 -mx-2 rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] transition-colors min-h-[40px]">
      <div className="w-[140px] shrink-0 text-sm text-[#37352F]/65 select-none truncate flex items-center">
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// Status Pill
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
// Form state
// ─────────────────────────────────────────────────

interface ClinicFormData {
  name: string;
  postal_code: string;
  address: string;
  phone_number: string;
  fax_number: string;
  registration_number: string;
  director_name: string;
  email: string;
  website: string;
  is_active: boolean;
}

const DEFAULT_FORM_DATA: ClinicFormData = {
  name: "",
  postal_code: "",
  address: "",
  phone_number: "",
  fax_number: "",
  registration_number: "",
  director_name: "",
  email: "",
  website: "",
  is_active: true,
};

// ─────────────────────────────────────────────────
// PropertyInput class
// ─────────────────────────────────────────────────

const PROP_INPUT_CLASS = `w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] focus:bg-[rgba(55,53,47,0.04)] transition-colors placeholder:text-[rgba(55,53,47,0.3)]`;

// ─────────────────────────────────────────────────
// ClinicMasterSettings
// ─────────────────────────────────────────────────

export function ClinicMasterSettings() {
  const navigate = useNavigate();
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<Clinic | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [formData, setFormData] = useState<ClinicFormData>(DEFAULT_FORM_DATA);
  const [pendingDelete, setPendingDelete] = useState<Clinic | null>(null);
  const deferredSearch = useDeferredValue(searchTerm);

  const { data: rawClinics } = useListClinics();
  const createMutation = useCreateClinic();
  const updateMutation = useUpdateClinic();
  const deleteMutation = useDeleteClinic();

  const filteredItems = useMemo(() => {
    const clinics = rawClinics ?? [];
    if (!deferredSearch) return clinics;
    const lower = deferredSearch.toLowerCase();
    return clinics.filter(
      (c) =>
        c.name.toLowerCase().includes(lower) ||
        c.phoneNumber.toLowerCase().includes(lower) ||
        c.email.toLowerCase().includes(lower),
    );
  }, [rawClinics, deferredSearch]);

  const handleEdit = (item: Clinic) => {
    setSelectedItem(item);
    setFormData({
      name: item.name,
      postal_code: item.postalCode,
      address: item.address,
      phone_number: item.phoneNumber,
      fax_number: item.faxNumber,
      registration_number: item.registrationNumber,
      director_name: item.directorName,
      email: item.email,
      website: item.website,
      is_active: item.isActive,
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
      toast.error("院名は必須です");
      return;
    }

    if (selectedItem) {
      const req: UpdateClinicRequest = {
        name: formData.name,
        postal_code: formData.postal_code || undefined,
        address: formData.address || undefined,
        phone_number: formData.phone_number || undefined,
        fax_number: formData.fax_number || undefined,
        registration_number: formData.registration_number || undefined,
        director_name: formData.director_name || undefined,
        email: formData.email || undefined,
        website: formData.website || undefined,
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
      const req: CreateClinicRequest = {
        name: formData.name,
        postal_code: formData.postal_code || undefined,
        address: formData.address || undefined,
        phone_number: formData.phone_number || undefined,
        fax_number: formData.fax_number || undefined,
        registration_number: formData.registration_number || undefined,
        director_name: formData.director_name || undefined,
        email: formData.email || undefined,
        website: formData.website || undefined,
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
        {/* Left: List */}
        <div className="flex-1 min-w-0">
          <PageLayout
            title="医院マスタ"
            icon={<Building2 className="size-5 text-[#37352F]" />}
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
                placeholder="院名、電話番号、メールで検索..."
                count={filteredItems.length}
              />

              <DataTable
                columns={COLUMNS}
                data={filteredItems}
                emptyMessage="医院が登録されていません"
                renderRow={(item) => (
                  <DataTableRow key={item.id} onClick={() => handleEdit(item)}>
                    <TableCell className="font-medium text-sm text-[#37352F] py-2.5">
                      {item.name}
                    </TableCell>
                    <TableCell className="font-mono text-sm text-[#37352F]/80 py-2.5">
                      {item.phoneNumber || "-"}
                    </TableCell>
                    <TableCell className="text-sm text-[#37352F]/80 py-2.5">
                      {item.email || "-"}
                    </TableCell>
                    <TableCell className="text-right py-2.5">
                      <span className="inline-flex items-center gap-1.5">
                        <span className={`size-[7px] rounded-full ${item.isActive ? "bg-[#2383E2]" : "bg-[#37352F]/20"}`} />
                        <span className={`text-sm ${item.isActive ? "text-[#37352F]/65" : "text-[#37352F]/35"}`}>
                          {item.isActive ? "有効" : "無効"}
                        </span>
                      </span>
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
                新しい医院を追加...
              </button>
            </div>
          </PageLayout>
        </div>

        {/* Right: Side Peek Panel */}
        {isEditing ? (
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
                    <Building2 className={LAYOUT.pageIcon.innerIcon} />
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
                    onChange={(e) =>
                      setFormData({ ...formData, name: e.target.value })
                    }
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

                  {/* 郵便番号 */}
                  <PropertyRow label="郵便番号">
                    <input
                      type="text"
                      className={PROP_INPUT_CLASS}
                      value={formData.postal_code}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          postal_code: e.target.value,
                        })
                      }
                      placeholder="例: 150-0001"
                    />
                  </PropertyRow>

                  {/* 住所 */}
                  <PropertyRow label="住所">
                    <input
                      type="text"
                      className={PROP_INPUT_CLASS}
                      value={formData.address}
                      onChange={(e) =>
                        setFormData({ ...formData, address: e.target.value })
                      }
                      placeholder="例: 東京都渋谷区..."
                    />
                  </PropertyRow>

                  {/* 電話番号 */}
                  <PropertyRow label="電話番号">
                    <input
                      type="text"
                      className={PROP_INPUT_CLASS}
                      value={formData.phone_number}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          phone_number: e.target.value,
                        })
                      }
                      placeholder="例: 03-1234-5678"
                    />
                  </PropertyRow>

                  {/* FAX番号 */}
                  <PropertyRow label="FAX番号">
                    <input
                      type="text"
                      className={PROP_INPUT_CLASS}
                      value={formData.fax_number}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          fax_number: e.target.value,
                        })
                      }
                      placeholder="例: 03-1234-5679"
                    />
                  </PropertyRow>

                  {/* 登録番号 */}
                  <PropertyRow label="登録番号">
                    <input
                      type="text"
                      className={PROP_INPUT_CLASS}
                      value={formData.registration_number}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          registration_number: e.target.value,
                        })
                      }
                      placeholder="例: 東京都獣医師会 第12345号"
                    />
                  </PropertyRow>

                  {/* 院長名 */}
                  <PropertyRow label="院長名">
                    <input
                      type="text"
                      className={PROP_INPUT_CLASS}
                      value={formData.director_name}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          director_name: e.target.value,
                        })
                      }
                      placeholder="例: 山田 太郎"
                    />
                  </PropertyRow>

                  {/* メール */}
                  <PropertyRow label="メール">
                    <input
                      type="email"
                      className={PROP_INPUT_CLASS}
                      value={formData.email}
                      onChange={(e) =>
                        setFormData({ ...formData, email: e.target.value })
                      }
                      placeholder="例: info@clinic.com"
                    />
                  </PropertyRow>

                  {/* Webサイト */}
                  <PropertyRow label="Webサイト">
                    <input
                      type="text"
                      className={PROP_INPUT_CLASS}
                      value={formData.website}
                      onChange={(e) =>
                        setFormData({ ...formData, website: e.target.value })
                      }
                      placeholder="例: https://example.com"
                    />
                  </PropertyRow>
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
        ) : null}
      </div>

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        title="医院を削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}
