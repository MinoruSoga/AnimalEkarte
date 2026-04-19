// React/Framework
import { memo } from "react";
import type { ReactNode } from "react";
import { useCallback, useDeferredValue, useMemo, useState, useActionState, useEffect, useTransition } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

// External
import { Plus, Building2, X, Percent } from "lucide-react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";

// Internal
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { C, STYLE, LAYOUT, ICON } from "@/lib/design-tokens";
import { NavigationBlocker } from "@/components/shared/NavigationBlocker/NavigationBlocker";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { useGetClinics, useCreateClinic, useUpdateClinic, useDeleteClinic } from "../api/clinics";

// Types
import type {
  Clinic,
  CreateClinicRequest,
  UpdateClinicRequest,
} from "../api/clinics";
import { ResourceHospitalSettings } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

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

const PropertyRow = memo(function PropertyRow({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) {
  return (
    <div className={STYLE.propertyRow}>
      <div className={`${LAYOUT.propertyRow.labelW} shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}>
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
});

// ─────────────────────────────────────────────────
// Status Pill
// ─────────────────────────────────────────────────

const STATUS_CONFIG = {
  active: {
    dot: `${C.bgAccent}`,
    label: "有効",
    bg: `${C.bgAccentLight}`,
    text: `${C.textAccentDark}`,
  },
  inactive: {
    dot: C.bgPrimary10,
    label: "無効",
    bg: `${C.bgInactive}`,
    text: C.text60,
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
  standard_tax_rate: number;
  reduced_tax_rate: number;
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
  standard_tax_rate: 0.1,
  reduced_tax_rate: 0.08,
};

// ─────────────────────────────────────────────────
// PropertyInput class
// ─────────────────────────────────────────────────

const PROP_INPUT_CLASS = `w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`;

// ─────────────────────────────────────────────────
// ClinicMasterSettings
// ─────────────────────────────────────────────────

export function ClinicMasterSettings() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceHospitalSettings);
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<Clinic | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [formData, setFormData] = useState<ClinicFormData>(DEFAULT_FORM_DATA);
  const [pendingDelete, setPendingDelete] = useState<Clinic | null>(null);
  const deferredSearch = useDeferredValue(searchTerm);

  const { data: rawClinics } = useGetClinics();
  const createMutation = useCreateClinic();
  const updateMutation = useUpdateClinic();
  const deleteMutation = useDeleteClinic();

  interface FormState {
    success: boolean;
    timestamp: number;
    nameError?: string;
  }

  /**
   * React 19 useActionState を使用したフォームアクション
   */
  const [formState, formAction, _isPending] = useActionState(
    async (_prevState: FormState, _formData: FormData): Promise<FormState> => {
      const fd = formData;
      if (!fd.name) {
        return { success: false, timestamp: Date.now(), nameError: "院名は必須です" };
      }

      try {
        if (selectedItem?.id) {
          const req: UpdateClinicRequest = {
            name: fd.name,
            postal_code: fd.postal_code || undefined,
            address: fd.address || undefined,
            phone_number: fd.phone_number || undefined,
            fax_number: fd.fax_number || undefined,
            registration_number: fd.registration_number || undefined,
            director_name: fd.director_name || undefined,
            email: fd.email || undefined,
            website: fd.website || undefined,
            is_active: fd.is_active,
            standard_tax_rate: fd.standard_tax_rate,
            reduced_tax_rate: fd.reduced_tax_rate,
          };
          await updateMutation.mutateAsync({ id: selectedItem.id, req });
          toast.success("更新しました");
        } else {
          const req: CreateClinicRequest = {
            name: fd.name,
            postal_code: fd.postal_code || undefined,
            address: fd.address || undefined,
            phone_number: fd.phone_number || undefined,
            fax_number: fd.fax_number || undefined,
            registration_number: fd.registration_number || undefined,
            director_name: fd.director_name || undefined,
            email: fd.email || undefined,
            website: fd.website || undefined,
          };
          await createMutation.mutateAsync(req);
          toast.success("登録しました");
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    { success: false, timestamp: 0 }
  );

  // React 19 Action の成功を検知してパネルを閉じる
  useEffect(() => {
    if (formState.success) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- ActionState の成功を検知してUIをリセットするパターン
      setIsEditing(false);
      setSelectedItem(null);
      setFormData(DEFAULT_FORM_DATA);
    }
  }, [formState.success, formState.timestamp]);

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

  const handleEdit = useCallback((item: Clinic) => {
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
      standard_tax_rate: item.standardTaxRate ?? 0.1,
      reduced_tax_rate: item.reducedTaxRate ?? 0.08,
    });
    setIsEditing(true);
  }, []);

  const handleCreate = useCallback(() => {
    setSelectedItem(null);
    setFormData(DEFAULT_FORM_DATA);
    setIsEditing(true);
  }, []);

  const handleCloseEdit = useCallback(() => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData(DEFAULT_FORM_DATA);
  }, []);

  const pendingDeleteId = pendingDelete?.id ?? null;
  const [isDeletePending, startDeleteTransition] = useTransition();

  const { mutate: deleteClinicMasterFn } = deleteMutation;
  const handleDeleteConfirm = useCallback(() => {
    if (pendingDeleteId === null) return;
    startDeleteTransition(() => {
      deleteClinicMasterFn(pendingDeleteId, {
        onSuccess: () => {
          setPendingDelete(null);
          setIsEditing(false);
          toast.success("削除しました");
        },
        onError: (error) => {
          handleApiError(error, "削除");
        },
      });
    });
  }, [pendingDeleteId, deleteClinicMasterFn]);

  return (
    <>
      {/* BUG-053: 編集パネルが開いている間はナビゲーションをブロック */}
      <NavigationBlocker when={isEditing} />
      <div className="flex h-full">
        {/* Left: List */}
        <div className="flex-1 min-w-0">
          <PageLayout
            title="医院マスタ"
            resource={ResourceHospitalSettings}
            icon={<Building2 className={`${ICON.page} ${C.text}`} />}
            onBack={() => navigate(paths.settings.getHref())}
            headerAction={
              canCreate ? (
                <PrimaryButton onClick={handleCreate}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
            maxWidth="max-w-full"
          >
            <div className="flex flex-col gap-4">
              <NotionFilter
          properties={[]}
          activeFilters={[]}
          onFilterChange={() => {}}
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                searchPlaceholder="院名、電話番号、メールで検索..."
                count={filteredItems.length}
              />

              <DataTable
                columns={COLUMNS}
                data={filteredItems}
                emptyMessage="医院が登録されていません"
                renderRow={(item) => (
                  <DataTableRow key={item.id} onClick={canEdit ? () => handleEdit(item) : undefined}>
                    <TableCell className={`font-medium text-sm ${C.text} py-2.5`}>
                      {item.name}
                    </TableCell>
                    <TableCell className={`font-mono text-sm ${C.text80} py-2.5`}>
                      {item.phoneNumber || "-"}
                    </TableCell>
                    <TableCell className={`text-sm ${C.text80} py-2.5`}>
                      {item.email || "-"}
                    </TableCell>
                    <TableCell className="text-right py-2.5">
                      <span className="inline-flex items-center gap-1.5">
                        <span className={`size-[7px] rounded-full ${item.isActive ? C.bgAccent : C.bgPrimary10}`} />
                        <span className={`text-sm ${item.isActive ? C.text65 : C.text35}`}>
                          {item.isActive ? "有効" : "無効"}
                        </span>
                      </span>
                    </TableCell>
                  </DataTableRow>
                )}
              />
              {canCreate ? (
                <button
                  type="button"
                  onClick={handleCreate}
                  className={`flex items-center gap-1.5 w-full px-3 py-2 text-sm ${C.text40} ${C.hoverText60} ${C.hoverBgLight} transition-colors rounded`}
                >
                  <Plus className={`${ICON.xs}`} />
                  新しい医院を追加...
                </button>
              ) : null}
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
              <span className={`text-xs ${C.text35} pl-1 select-none`}>
                {selectedItem ? "編集" : "新規作成"}
              </span>
              <div className="flex items-center gap-1">
                {selectedItem && canDelete ? (
                  <DeleteIconButton onClick={() => setPendingDelete(selectedItem)} />
                ) : null}
                <button
                  type="button"
                  onClick={handleCloseEdit}
                  className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
                >
                  <X className={ICON.action} />
                </button>
              </div>
            </div>

            {/* Body */}
            <form action={formAction} className="flex-1 flex flex-col min-h-0">
            <fieldset disabled={!canEdit} className="border-0 p-0 m-0 min-w-0">
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
                    className={`w-full bg-transparent ${C.text} ${C.textPlaceholderFaint} outline-none border-none p-0`}
                    style={{
                      fontSize: LAYOUT.pageTitle.fontSize,
                      fontWeight: LAYOUT.pageTitle.fontWeight,
                      lineHeight: LAYOUT.pageTitle.lineHeight,
                    }}
                    value={formData.name}
                    onChange={(e) =>
                      setFormData((prev) => ({ ...prev, name: e.target.value }))
                    }
                    placeholder="無題"
                  />
                  <FormFieldError message={formState.nameError} />
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
                        setFormData((prev) => ({
                          ...prev,
                          is_active: !prev.is_active,
                        }))
                      }
                      className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
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
                        setFormData((prev) => ({
                          ...prev,
                          postal_code: e.target.value,
                        }))
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
                        setFormData((prev) => ({ ...prev, address: e.target.value }))
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
                        setFormData((prev) => ({
                          ...prev,
                          phone_number: e.target.value,
                        }))
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
                        setFormData((prev) => ({
                          ...prev,
                          fax_number: e.target.value,
                        }))
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
                        setFormData((prev) => ({
                          ...prev,
                          registration_number: e.target.value,
                        }))
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
                        setFormData((prev) => ({
                          ...prev,
                          director_name: e.target.value,
                        }))
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
                        setFormData((prev) => ({ ...prev, email: e.target.value }))
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
                        setFormData((prev) => ({ ...prev, website: e.target.value }))
                      }
                      placeholder="例: https://example.com"
                    />
                  </PropertyRow>

                  {/* 税率セクション */}
                  <div className={`${STYLE.sectionDivider} my-2`} />
                  <div className={`flex items-center gap-1.5 py-1.5 text-xs ${C.text45} select-none`}>
                    <Percent className={ICON.xs} />
                    消費税率
                  </div>

                  {/* 通常課税 */}
                  <PropertyRow label="通常課税">
                    <div className="flex items-center gap-1.5">
                      <input
                        type="number"
                        min={0}
                        max={100}
                        step={1}
                        className={`${PROP_INPUT_CLASS} w-20`}
                        value={Math.round(formData.standard_tax_rate * 100)}
                        onChange={(e) =>
                          setFormData((prev) => ({
                            ...prev,
                            standard_tax_rate: Number(e.target.value) / 100,
                          }))
                        }
                      />
                      <span className={`text-sm ${C.text50}`}>%</span>
                    </div>
                  </PropertyRow>

                  {/* 軽減税率 */}
                  <PropertyRow label="軽減税率">
                    <div className="flex items-center gap-1.5">
                      <input
                        type="number"
                        min={0}
                        max={100}
                        step={1}
                        className={`${PROP_INPUT_CLASS} w-20`}
                        value={Math.round(formData.reduced_tax_rate * 100)}
                        onChange={(e) =>
                          setFormData((prev) => ({
                            ...prev,
                            reduced_tax_rate: Number(e.target.value) / 100,
                          }))
                        }
                      />
                      <span className={`text-sm ${C.text50}`}>%</span>
                    </div>
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
              {canEdit ? (
                <SubmitButton
                  className={STYLE.sidePeekSaveBtn}
                >
                  保存
                </SubmitButton>
              ) : null}
            </div>
            </fieldset>
            </form>
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
        isPending={isDeletePending}
      />
    </>
  );
}
