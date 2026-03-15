// React/Framework
import { useState, useMemo, useCallback, memo, useDeferredValue, useTransition } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

// External
import { Plus, Building2 } from "lucide-react";
import { toast } from "sonner";

// Internal
import { TableCell } from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { MoneyInput } from "@/components/shared/SidePeek/MoneyInput";
import { PropInput } from "@/components/shared/SidePeek/PropInput";
import { SidePeekPanel } from "@/components/shared/SidePeek/SidePeekPanel";
import { SidePeekToolbar } from "@/components/shared/SidePeek/SidePeekToolbar";
import { SidePeekBody } from "@/components/shared/SidePeek/SidePeekBody";
import { SidePeekTitleInput } from "@/components/shared/SidePeek/SidePeekTitleInput";
import { SidePeekFooter } from "@/components/shared/SidePeek/SidePeekFooter";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar/SearchFilterBar";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import {
  useGetAllCages,
  useCreateCage,
  useUpdateCage,
  useDeleteCage,
} from "@/features/master/api/cages";

// Types
import type { Cage, CageType, CageSize } from "@/types/index";
import type { CreateCageRequest, UpdateCageRequest } from "@/features/master/api/cages";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const CAGE_TYPE_LABELS: Record<CageType, string> = {
  icu: "ICU",
  dog: "犬舎",
  cat: "猫舎",
  general: "汎用",
};

const CAGE_SIZE_LABELS: Record<CageSize, string> = {
  small: "小型",
  medium: "中型",
  large: "大型",
};

const COLUMNS = [
  { header: "ケージ名" },
  { header: "エリア", className: "w-[100px]" },
  { header: "サイズ", className: "w-[90px]" },
  { header: "単価(税込)", className: "w-[120px]", align: "right" as const },
  { header: "ステータス", className: "w-[90px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

// Hoisted static JSX (rendering-hoist-jsx)
const CAGE_TYPE_SELECT_ITEMS = (
  [
    { value: "icu" as CageType, label: "ICU" },
    { value: "dog" as CageType, label: "犬舎" },
    { value: "cat" as CageType, label: "猫舎" },
    { value: "general" as CageType, label: "汎用" },
  ] satisfies { value: CageType; label: string }[]
).map((opt) => (
  <SelectItem key={opt.value} value={opt.value}>
    {opt.label}
  </SelectItem>
));

const CAGE_SIZE_SELECT_ITEMS = (
  [
    { value: "small" as CageSize, label: "小型" },
    { value: "medium" as CageSize, label: "中型" },
    { value: "large" as CageSize, label: "大型" },
  ] satisfies { value: CageSize; label: string }[]
).map((opt) => (
  <SelectItem key={opt.value} value={opt.value}>
    {opt.label}
  </SelectItem>
));

// ─────────────────────────────────────────────────
// Form state
// ─────────────────────────────────────────────────

interface CageFormData {
  name: string;
  cageType: CageType;
  cageSize: CageSize;
  price: number;
  description: string;
  isActive: boolean;
}

// ─────────────────────────────────────────────────
// CageSidePanel
// ─────────────────────────────────────────────────

interface CageSidePanelProps {
  item: Cage | null;
  onClose: () => void;
  onSave: (data: CageFormData) => void;
  onDeleteRequest: (item: Cage) => void;
}

const CageSidePanel = memo(function CageSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: CageSidePanelProps) {
  // rerender-lazy-state-init: 初回マウント時のみ item から初期化
  const [formData, setFormData] = useState<CageFormData>(() => ({
    name: item?.name ?? "",
    cageType: item?.cageType ?? "general",
    cageSize: item?.cageSize ?? "medium",
    price: item?.price ?? 0,
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
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
            <Building2 className={LAYOUT.pageIcon.innerIcon} />
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

          <PropertyRow label="エリア">
            <Select
              value={formData.cageType}
              onValueChange={(v) => setFormData((prev) => ({ ...prev, cageType: v as CageType }))}
            >
              <SelectTrigger className={STYLE.selectCompact}>
                <SelectValue placeholder="選択" />
              </SelectTrigger>
              <SelectContent>{CAGE_TYPE_SELECT_ITEMS}</SelectContent>
            </Select>
          </PropertyRow>

          <PropertyRow label="サイズ">
            <Select
              value={formData.cageSize}
              onValueChange={(v) => setFormData((prev) => ({ ...prev, cageSize: v as CageSize }))}
            >
              <SelectTrigger className={STYLE.selectCompact}>
                <SelectValue placeholder="選択" />
              </SelectTrigger>
              <SelectContent>{CAGE_SIZE_SELECT_ITEMS}</SelectContent>
            </Select>
          </PropertyRow>

          <MoneyInput
            value={formData.price}
            onChange={(v) => setFormData((prev) => ({ ...prev, price: v }))}
          />

          <PropertyRow label="備考">
            <PropInput
              value={formData.description}
              onChange={(v) => setFormData((prev) => ({ ...prev, description: v }))}
              placeholder="補足情報など"
            />
          </PropertyRow>
        </div>
      </SidePeekBody>
      <SidePeekFooter onCancel={onClose} onSave={() => onSave(formData)} />
    </SidePeekPanel>
  );
});

// ─────────────────────────────────────────────────
// CageSettings (main page)
// ─────────────────────────────────────────────────

export function CageSettings() {
  const navigate = useNavigate();

  // null=closed, "new"=create mode, Cage=edit mode
  const [editTarget, setEditTarget] = useState<Cage | "new" | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<Cage | null>(null);
  const [, startSaveTransition] = useTransition();

  const { data: rawCages } = useGetAllCages();
  const createMutation = useCreateCage();
  const updateMutation = useUpdateCage();
  const deleteMutation = useDeleteCage();

  // rerender-transitions: 検索フィルタを低優先度に遅延
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    const cages = rawCages ?? [];
    if (!deferredSearch) return cages;
    const lower = deferredSearch.toLowerCase();
    return cages.filter((c) => c.name.toLowerCase().includes(lower));
  }, [rawCages, deferredSearch]);

  const handleClose = useCallback(() => setEditTarget(null), []);

  const handleSave = useCallback(
    (data: CageFormData) => {
      if (!data.name.trim()) {
        toast.error("ケージ名は必須です");
        return;
      }

      // rerender-transitions: API書き込みを非緊急マーク
      startSaveTransition(() => {
        if (editTarget !== null && editTarget !== "new") {
          const req: UpdateCageRequest = {
            name: data.name,
            cage_type: data.cageType,
            cage_size: data.cageSize,
            price: data.price,
            description: data.description || undefined,
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
          const req: CreateCageRequest = {
            name: data.name,
            cage_type: data.cageType,
            cage_size: data.cageSize,
            price: data.price,
            description: data.description || undefined,
            is_active: data.isActive,
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
      title="ケージマスタ"
      icon={<Building2 className="size-5 text-[#37352F]" />}
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
                placeholder="ケージ名で検索..."
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
            emptyMessage="ケージが登録されていません"
            renderRow={(item) => (
              <DataTableRow key={item.id} onClick={() => setEditTarget(item)}>
                <TableCell className={`font-medium text-sm ${C.text}`}>
                  {item.name}
                </TableCell>
                <TableCell className={`text-sm ${C.text70}`}>
                  {CAGE_TYPE_LABELS[item.cageType]}
                </TableCell>
                <TableCell className={`text-sm ${C.text70}`}>
                  {CAGE_SIZE_LABELS[item.cageSize]}
                </TableCell>
                <TableCell className={`text-right font-mono text-sm ${C.text}`}>
                  {item.price != null ? `¥${item.price.toLocaleString()}` : "-"}
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
          <CageSidePanel
            key={panelItem ? String(panelItem.id) : "new-cage"}
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
        title="ケージを削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </PageLayout>
  );
}
