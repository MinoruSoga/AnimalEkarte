// React/Framework
import { useState, useEffect } from "react";
import type { ReactNode } from "react";
import { useNavigate } from "react-router";

// External
import { toast } from "sonner";
import Plus from "lucide-react/dist/esm/icons/plus";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import X from "lucide-react/dist/esm/icons/x";

// Internal
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable } from "@/components/shared/DataTable";
import { DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { useMasterItems } from "@/hooks/use-master-items";
import {
  CATEGORY_CONFIG,
  CATEGORY_ALIAS_MAP,
} from "@/features/master/constants/category-config";
import type { MasterSettingsCategory } from "@/features/master/constants/category-config";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";

// Types
import type { MasterItem } from "@/types";

// ─────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────

function resolveCategory(raw: string): MasterSettingsCategory | undefined {
  if (raw in CATEGORY_ALIAS_MAP) return CATEGORY_ALIAS_MAP[raw];
  if (raw in CATEGORY_CONFIG) return raw as MasterSettingsCategory;
  return undefined;
}

function toCamelCaseCategory(cat: MasterSettingsCategory): string {
  const reverseAlias: Partial<Record<MasterSettingsCategory, string>> = {
    trimming_course: "trimmingCourse",
    trimming_option: "trimmingOption",
    diagnosis_category: "diagnosisCategory",
    diagnosis_name: "diagnosisName",
  };
  return reverseAlias[cat] ?? cat;
}

// ─────────────────────────────────────────────────
// Sub-components
// ─────────────────────────────────────────────────

function PropertyRow({
  label,
  required,
  children,
}: {
  label: string;
  required?: boolean;
  children: ReactNode;
}) {
  return (
    <div
      className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px] items-center`}
    >
      <div className={`w-[140px] shrink-0 text-sm ${C.text65} select-none`}>
        {label}
        {required && <span className={`${C.textRequired} ml-0.5`}>*</span>}
      </div>
      <div className="flex-1 min-w-0">{children}</div>
    </div>
  );
}

function PropInput({
  value,
  onChange,
  placeholder,
  type = "text",
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  type?: string;
}) {
  return (
    <input
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className={STYLE.propertyInput}
    />
  );
}

function NotionStatusPill({ active }: { active: boolean }) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-sm text-xs ${
        active ? "bg-[#D3E5EF] text-[#183B56]" : "bg-[#E3E2E0] text-[#37352F]/60"
      }`}
    >
      <span
        className={`size-[7px] rounded-full ${
          active ? "bg-[#2383E2]" : "bg-[#37352F]/10"
        }`}
      />
      {active ? "有効" : "無効"}
    </span>
  );
}

function StatusDot({ active }: { active: boolean }) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <span
        className={`size-[7px] rounded-full ${active ? "bg-[#2383E2]" : "bg-[#37352F]/20"}`}
      />
      <span
        className={`text-sm ${active ? "text-[#37352F]/65" : "text-[#37352F]/35"}`}
      >
        {active ? "有効" : "無効"}
      </span>
    </span>
  );
}

// ─────────────────────────────────────────────────
// Props
// ─────────────────────────────────────────────────

interface SettingsPageProps {
  category?: string;
  /** When true, skip PageLayout wrapper and render list+peek content only (for tabbed pages) */
  embedded?: boolean;
}

// ─────────────────────────────────────────────────
// Main Component
// ─────────────────────────────────────────────────

export function Settings({ category: propCategory, embedded = false }: SettingsPageProps) {
  const navigate = useNavigate();
  const rawCategory = propCategory ?? "examination";
  const resolvedCategory = resolveCategory(rawCategory);
  const config = resolvedCategory ? CATEGORY_CONFIG[resolvedCategory] : undefined;

  const hookCategory = resolvedCategory ? toCamelCaseCategory(resolvedCategory) : rawCategory;

  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<MasterItem | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [formData, setFormData] = useState<Partial<MasterItem>>({});
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  const { data: filteredItems, add, update, remove } = useMasterItems(hookCategory, searchTerm);

  const pageTitle = config?.label ?? "マスタ";
  const showPrice = config?.showPrice ?? false;
  const showCategory = config?.showCategory ?? false;
  const labels = config?.labels ?? { code: "コード", name: "名称", category: "カテゴリ" };

  // Reset editing state when category changes
  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    setIsEditing(false);
    setSelectedItem(null);
    setSearchTerm("");
    setFormData({});
  }, [rawCategory]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const handleEdit = (item: MasterItem) => {
    setSelectedItem(item);
    setFormData({ ...item });
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setFormData({ status: "active", price: 0 });
    setIsEditing(true);
  };

  const handleCloseEdit = () => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData({});
  };

  const handleSave = () => {
    if (!formData.name) {
      toast.error("名称は必須です");
      return;
    }
    if (selectedItem) {
      update(selectedItem.id, formData, {
        onSuccess: () => {
          toast.success("更新しました");
          handleCloseEdit();
        },
      });
    } else {
      add(formData as Omit<MasterItem, "id">, {
        onSuccess: () => {
          toast.success("登録しました");
          handleCloseEdit();
        },
      });
    }
  };

  const handleDelete = () => {
    if (!selectedItem) return;
    setDeleteConfirmOpen(true);
  };

  const executeDelete = () => {
    if (!selectedItem) return;
    remove(selectedItem.id, {
      onSuccess: () => {
        handleCloseEdit();
        toast.success("削除しました");
        setDeleteConfirmOpen(false);
      },
    });
  };

  // ── Table columns ──────────────────────────────────
  const columns = [
    { header: labels.name },
    ...(showCategory ? [{ header: labels.category, className: "w-[120px]" }] : []),
    ...(showPrice
      ? [{ header: "単価(税込)", className: "w-[110px]", align: "right" as const }]
      : []),
    { header: "ステータス", className: "w-[100px]", align: "right" as const },
  ];

  // ── List content ───────────────────────────────────
  const listContent = (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-3">
        <div className="flex-1 min-w-0">
          <SearchFilterBar
            searchTerm={searchTerm}
            onSearchChange={setSearchTerm}
            placeholder={`${labels.name}で検索...`}
            count={filteredItems.length}
          />
        </div>
        {/* New button shown here only in embedded mode; standalone uses headerAction */}
        {embedded && (
          <PrimaryButton onClick={handleCreate}>
            <Plus className="mr-1.5 size-4" />
            新規登録
          </PrimaryButton>
        )}
      </div>

      <DataTable
        columns={columns}
        data={filteredItems}
        emptyMessage="データが見つかりません"
        renderRow={(item) => (
          <DataTableRow key={item.id} onClick={() => handleEdit(item)}>
            <TableCell className={`font-medium text-sm ${C.text} py-2`}>
              {item.name}
            </TableCell>
            {showCategory && (
              <TableCell className={`text-sm ${C.text} py-2`}>
                {item.category ?? "-"}
              </TableCell>
            )}
            {showPrice && (
              <TableCell className={`text-right font-mono text-sm ${C.text} py-2`}>
                {item.price ? `¥${item.price.toLocaleString()}` : "-"}
              </TableCell>
            )}
            <TableCell className="text-right py-2">
              <StatusDot active={item.status !== "inactive"} />
            </TableCell>
          </DataTableRow>
        )}
      />

      <button
        type="button"
        onClick={handleCreate}
        className={`flex items-center gap-1.5 w-full px-3 py-2 text-sm text-[#37352F]/40 hover:text-[#37352F]/65 hover:bg-[rgba(55,53,47,0.04)] transition-colors rounded`}
      >
        <Plus className="size-3.5" />
        新しい{labels.name}を追加...
      </button>
    </div>
  );

  // ── Side peek panel ────────────────────────────────
  const sidePeekPanel = isEditing ? (
    <div
      className={`flex flex-col self-stretch bg-white border-l ${C.borderLight} shadow-[-1px_0_5px_rgba(0,0,0,0.02)] ${LAYOUT.sidePeek.width} shrink-0`}
    >
      {/* Toolbar */}
      <div className={STYLE.sidePeekToolbar}>
        <span className="text-xs text-[#37352F]/35 pl-1 select-none">
          {selectedItem ? "編集" : "新規作成"}
        </span>
        <div className="flex items-center gap-1">
          {selectedItem && (
            <button
              type="button"
              onClick={handleDelete}
              className={`${STYLE.sidePeekToolbarBtn} cursor-pointer text-[#EB5757] hover:bg-[#EB5757]/10`}
            >
              <Trash2 className="size-4" />
            </button>
          )}
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
              {config && <config.IconComponent className={LAYOUT.pageIcon.innerIcon} />}
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
              value={formData.name ?? ""}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder={config?.namePlaceholder ?? "無題"}
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
                    status: formData.status === "inactive" ? "active" : "inactive",
                  })
                }
                className="inline-flex items-center rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] transition-colors py-0.5 px-0.5 cursor-pointer"
              >
                <NotionStatusPill active={formData.status !== "inactive"} />
              </button>
            </PropertyRow>

            {/* Price */}
            {showPrice && (
              <PropertyRow label="単価(税込)">
                <div className="flex items-center gap-1">
                  <span className={`text-sm ${C.text40}`}>¥</span>
                  <input
                    type="number"
                    value={formData.price ?? 0}
                    onChange={(e) =>
                      setFormData({ ...formData, price: Number(e.target.value) })
                    }
                    placeholder="0"
                    className={`${STYLE.propertyInput} w-28`}
                  />
                </div>
              </PropertyRow>
            )}

            {/* Category */}
            {showCategory && (
              <PropertyRow label={labels.category}>
                <PropInput
                  value={formData.category ?? ""}
                  onChange={(v) => setFormData({ ...formData, category: v })}
                  placeholder="分類"
                />
              </PropertyRow>
            )}

            {/* Description / remarks */}
            <PropertyRow label="備考">
              <PropInput
                value={formData.description ?? ""}
                onChange={(v) => setFormData({ ...formData, description: v })}
                placeholder="空"
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
  ) : null;

  // ── Confirm dialog ─────────────────────────────────
  const confirmDialog = (
    <ConfirmDialog
      open={deleteConfirmOpen}
      onClose={() => setDeleteConfirmOpen(false)}
      onConfirm={executeDelete}
      title="削除しますか？"
      description={`「${selectedItem?.name ?? ""}」を削除します。この操作は取り消せません。`}
      confirmLabel="削除する"
      cancelLabel="キャンセル"
      variant="destructive"
    />
  );

  // ── Embedded mode (for tabbed pages) ──────────────
  if (embedded) {
    return (
      <>
        <div className="flex">
          <div className="flex-1 min-w-0">{listContent}</div>
          {sidePeekPanel}
        </div>
        {confirmDialog}
      </>
    );
  }

  // ── Standalone list view ──────────────────────────
  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title={pageTitle}
            icon={
              config ? (
                <config.IconComponent className="size-5 text-[#37352F]" />
              ) : undefined
            }
            onBack={() => navigate("/settings")}
            headerAction={
              <PrimaryButton onClick={handleCreate}>
                <Plus className="mr-1.5 size-4" />
                新規登録
              </PrimaryButton>
            }
            maxWidth="max-w-full"
          >
            {listContent}
          </PageLayout>
        </div>
        {sidePeekPanel}
      </div>
      {confirmDialog}
    </>
  );
}
