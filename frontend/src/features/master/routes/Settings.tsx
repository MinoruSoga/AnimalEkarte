// React/Framework
import { useCallback, useDeferredValue, useMemo, useState, useTransition } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

// External
import { toast } from "sonner";
import Plus from "lucide-react/dist/esm/icons/plus";

// Internal
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { TableCell } from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar/SearchFilterBar";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { useMasterItems } from "@/hooks/use-master-items";
import {
  CATEGORY_CONFIG,
  CATEGORY_ALIAS_MAP,
} from "@/features/master/constants/category-config";
import type { MasterSettingsCategory } from "@/features/master/constants/category-config";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { PropInput } from "@/components/shared/SidePeek/PropInput";
import { SidePeekPanel } from "@/components/shared/SidePeek/SidePeekPanel";
import { SidePeekToolbar } from "@/components/shared/SidePeek/SidePeekToolbar";
import { SidePeekBody } from "@/components/shared/SidePeek/SidePeekBody";
import { SidePeekTitleInput } from "@/components/shared/SidePeek/SidePeekTitleInput";
import { SidePeekFooter } from "@/components/shared/SidePeek/SidePeekFooter";

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
  const [, startSaveTransition] = useTransition();
  const deferredSearch = useDeferredValue(searchTerm);

  const { data: filteredItems, add, update, remove } = useMasterItems(hookCategory, deferredSearch);

  const pageTitle = config?.label ?? "マスタ";
  const showPrice = config?.showPrice ?? false;
  const showCode = config?.showCode ?? false;
  const showCategory = config?.showCategory ?? false;
  const labels = useMemo(
    () => config?.labels ?? { code: "コード", name: "名称", category: "カテゴリ" },
    [config],
  );

  const handleCloseEdit = useCallback(() => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData({});
  }, []);

  const handleEdit = useCallback((item: MasterItem) => {
    setSelectedItem(item);
    setFormData({ ...item });
    setIsEditing(true);
  }, []);

  const handleCreate = useCallback(() => {
    setSelectedItem(null);
    setFormData({ status: "active", price: 0 });
    setIsEditing(true);
  }, []);

  const handleSave = useCallback(() => {
    if (!formData.name) {
      toast.error("名称は必須です");
      return;
    }
    // rerender-transitions: API書き込みを非緊急マーク
    startSaveTransition(() => {
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
    });
  }, [formData, selectedItem, update, add, handleCloseEdit, startSaveTransition]);

  const handleDelete = useCallback(() => {
    if (!selectedItem) return;
    setDeleteConfirmOpen(true);
  }, [selectedItem]);

  const executeDelete = useCallback(() => {
    if (!selectedItem) return;
    remove(selectedItem.id, {
      onSuccess: () => {
        handleCloseEdit();
        toast.success("削除しました");
        setDeleteConfirmOpen(false);
      },
    });
  }, [selectedItem, remove, handleCloseEdit]);

  // ── Table columns ──────────────────────────────────
  const columns = useMemo(
    () => [
      { header: labels.name },
      ...(showCode ? [{ header: labels.code, className: "w-[120px]" }] : []),
      ...(showCategory ? [{ header: labels.category, className: "w-[120px]" }] : []),
      ...(showPrice
        ? [{ header: "単価(税込)", className: "w-[110px]", align: "right" as const }]
        : []),
      { header: "ステータス", className: "w-[100px]", align: "center" as const },
      { header: "操作", className: "w-[80px]", align: "right" as const },
    ],
    [labels, showCode, showCategory, showPrice],
  );

  // ── renderRow ──────────────────────────────────────
  const renderRow = useCallback(
    (item: MasterItem) => (
      <DataTableRow key={item.id} onClick={() => handleEdit(item)}>
        <TableCell className={`font-medium text-sm ${C.text}`}>
          {item.name}
        </TableCell>
        {showCode ? (
          <TableCell className={`text-sm ${C.text}`}>
            {item.code ?? "-"}
          </TableCell>
        ) : null}
        {showCategory ? (
          <TableCell className={`text-sm ${C.text}`}>
            {item.category ?? "-"}
          </TableCell>
        ) : null}
        {showPrice ? (
          <TableCell className={`text-right font-mono text-sm ${C.text}`}>
            {item.price ? `¥${item.price.toLocaleString()}` : "-"}
          </TableCell>
        ) : null}
        <TableCell className="text-center">
          <NotionStatusPill isActive={item.status !== "inactive"} />
        </TableCell>
        <TableCell className="p-0 text-right">
          <RowActionButton onClick={() => handleEdit(item)} />
        </TableCell>
      </DataTableRow>
    ),
    [handleEdit, showCode, showCategory, showPrice],
  );

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
        {embedded ? (
          <PrimaryButton onClick={handleCreate}>
            <Plus className="mr-1.5 size-4" />
            新規登録
          </PrimaryButton>
        ) : null}
      </div>

      <DataTable
        columns={columns}
        data={filteredItems}
        emptyMessage="データが見つかりません"
        renderRow={renderRow}
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
    <SidePeekPanel>
      <SidePeekToolbar
        isNew={selectedItem === null}
        onClose={handleCloseEdit}
        onDelete={selectedItem !== null ? handleDelete : undefined}
      />
      <SidePeekBody>
        <div className="pt-4 pb-2">
          <div className={STYLE.pageIcon}>
            {config ? <config.IconComponent className={LAYOUT.pageIcon.innerIcon} /> : null}
          </div>
        </div>
        <SidePeekTitleInput
          value={formData.name ?? ""}
          onChange={(v) => setFormData((prev) => ({ ...prev, name: v }))}
          placeholder={config?.namePlaceholder ?? "無題"}
        />
        <div className={`${STYLE.sectionDivider} mb-1`} />
        <div className="py-1">
          {/* Status */}
          <PropertyRow label="ステータス">
            <button
              type="button"
              onClick={() =>
                setFormData((prev) => ({
                  ...prev,
                  status: prev.status === "inactive" ? "active" : "inactive",
                }))
              }
              className="inline-flex items-center rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] transition-colors py-0.5 px-0.5 cursor-pointer"
            >
              <NotionStatusPill isActive={formData.status !== "inactive"} />
            </button>
          </PropertyRow>

          {/* Price */}
          {showPrice ? (
            <PropertyRow label="単価(税込)">
              <div className="flex items-center gap-1">
                <span className={`text-sm ${C.text40}`}>¥</span>
                <input
                  type="number"
                  value={formData.price ?? 0}
                  onChange={(e) =>
                    setFormData((prev) => ({ ...prev, price: Number(e.target.value) }))
                  }
                  placeholder="0"
                  className={`${STYLE.propertyInput} w-28`}
                />
              </div>
            </PropertyRow>
          ) : null}

          {/* Category */}
          {showCategory ? (
            <PropertyRow label={labels.category}>
              <PropInput
                value={formData.category ?? ""}
                onChange={(v) => setFormData((prev) => ({ ...prev, category: v }))}
                placeholder="分類"
              />
            </PropertyRow>
          ) : null}

          {/* Description / remarks */}
          <PropertyRow label="備考">
            <PropInput
              value={formData.description ?? ""}
              onChange={(v) => setFormData((prev) => ({ ...prev, description: v }))}
              placeholder="空"
            />
          </PropertyRow>

          {/* Consultation-specific: time_condition */}
          {resolvedCategory === "consultation" ? (
            <PropertyRow label="適用区分">
              <Select
                value={formData.timeCondition ?? "anytime"}
                onValueChange={(v) => setFormData((prev) => ({ ...prev, timeCondition: v }))}
              >
                <SelectTrigger className={STYLE.selectCompact}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="anytime">常時</SelectItem>
                  <SelectItem value="first_visit">初診</SelectItem>
                  <SelectItem value="revisit">再診</SelectItem>
                  <SelectItem value="after_hours">時間外</SelectItem>
                  <SelectItem value="emergency">緊急</SelectItem>
                </SelectContent>
              </Select>
            </PropertyRow>
          ) : null}

          {/* Consultation-specific: duration */}
          {resolvedCategory === "consultation" ? (
            <PropertyRow label="標準診察時間 (分)">
              <input
                type="number"
                min={0}
                value={formData.duration ?? ""}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    duration: e.target.value === "" ? null : Number(e.target.value),
                  }))
                }
                placeholder="例: 15"
                className={`${STYLE.propertyInput} w-28`}
              />
            </PropertyRow>
          ) : null}
        </div>
      </SidePeekBody>
      <SidePeekFooter onCancel={handleCloseEdit} onSave={handleSave} />
    </SidePeekPanel>
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
            onBack={() => navigate(paths.settings.getHref())}
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
