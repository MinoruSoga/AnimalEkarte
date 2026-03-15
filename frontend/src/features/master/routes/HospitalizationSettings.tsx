// React/Framework
import { useState, useMemo, useCallback, memo, useDeferredValue, useTransition } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

// External
import { Plus, Bed } from "lucide-react";
import { toast } from "sonner";

// Shared
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar/SearchFilterBar";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { PropInput } from "@/components/shared/SidePeek/PropInput";
import { SidePeekPanel } from "@/components/shared/SidePeek/SidePeekPanel";
import { SidePeekToolbar } from "@/components/shared/SidePeek/SidePeekToolbar";
import { SidePeekBody } from "@/components/shared/SidePeek/SidePeekBody";
import { SidePeekTitleInput } from "@/components/shared/SidePeek/SidePeekTitleInput";
import { SidePeekFooter } from "@/components/shared/SidePeek/SidePeekFooter";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import {
  useGetAllHospitalizationPlans,
  useCreateHospitalizationPlan,
  useUpdateHospitalizationPlan,
  useDeleteHospitalizationPlan,
  BODY_SIZE_OPTIONS,
  BODY_SIZE_LABELS,
  BILLING_UNIT_OPTIONS,
  BILLING_UNIT_LABELS,
} from "@/features/master/api/hospitalization-plans";

// Types
import type {
  HospitalizationPlan,
  CreateHospitalizationPlanRequest,
  UpdateHospitalizationPlanRequest,
} from "@/features/master/api/hospitalization-plans";
import type { BodySize, BillingUnit } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const COLUMNS = [
  { header: "名称" },
  { header: "対象体格", className: "w-[100px]" },
  { header: "料金単位", className: "w-[120px]" },
  { header: "単価(税込)", className: "w-[120px]", align: "right" as const },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

// Hoisted static JSX (rendering-hoist-jsx)
const BODY_SIZE_SELECT_ITEMS = BODY_SIZE_OPTIONS.map((opt) => (
  <SelectItem key={opt.value} value={opt.value}>
    {opt.label}
  </SelectItem>
));

const BILLING_UNIT_SELECT_ITEMS = BILLING_UNIT_OPTIONS.map((opt) => (
  <SelectItem key={opt.value} value={opt.value}>
    {opt.label}
  </SelectItem>
));

// ─────────────────────────────────────────────────
// Form state
// ─────────────────────────────────────────────────

interface HospitalizationFormData {
  name: string;
  price: number;
  description: string;
  isActive: boolean;
  bodySize: BodySize | "";
  billingUnit: BillingUnit | "";
}

// ─────────────────────────────────────────────────
// HospitalizationSidePanel
// ─────────────────────────────────────────────────

interface HospitalizationSidePanelProps {
  item: HospitalizationPlan | null;
  onClose: () => void;
  onSave: (data: HospitalizationFormData) => void;
  onDeleteRequest: (item: HospitalizationPlan) => void;
}

const HospitalizationSidePanel = memo(function HospitalizationSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: HospitalizationSidePanelProps) {
  const [formData, setFormData] = useState<HospitalizationFormData>(() => ({
    name: item?.name ?? "",
    price: item?.price ?? 0,
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
    bodySize: item?.bodySize ?? "",
    billingUnit: item?.billingUnit ?? "",
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
            <Bed className={LAYOUT.pageIcon.innerIcon} />
          </div>
        </div>
        <SidePeekTitleInput
          value={formData.name}
          onChange={(v) => setFormData((prev) => ({ ...prev, name: v }))}
        />
        <div className={`${STYLE.sectionDivider} mb-1`} />
        <div className="py-1">
          {/* Status */}
          <PropertyRow label="ステータス">
            <button
              type="button"
              onClick={() =>
                setFormData((prev) => ({ ...prev, isActive: !prev.isActive }))
              }
              className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
            >
              <NotionStatusPill isActive={formData.isActive} />
            </button>
          </PropertyRow>

          {/* Body Size */}
          <PropertyRow label="対象体格">
            <Select
              value={formData.bodySize}
              onValueChange={(v) =>
                setFormData((prev) => ({ ...prev, bodySize: v as BodySize }))
              }
            >
              <SelectTrigger className={STYLE.selectCompact}>
                <SelectValue placeholder="選択" />
              </SelectTrigger>
              <SelectContent>{BODY_SIZE_SELECT_ITEMS}</SelectContent>
            </Select>
          </PropertyRow>

          {/* Billing Unit */}
          <PropertyRow label="料金単位">
            <Select
              value={formData.billingUnit}
              onValueChange={(v) =>
                setFormData((prev) => ({ ...prev, billingUnit: v as BillingUnit }))
              }
            >
              <SelectTrigger className={STYLE.selectCompact}>
                <SelectValue placeholder="選択" />
              </SelectTrigger>
              <SelectContent>{BILLING_UNIT_SELECT_ITEMS}</SelectContent>
            </Select>
          </PropertyRow>

          {/* Price */}
          <PropertyRow label="単価(税込)">
            <div className="flex items-center gap-1">
              <span className={`text-sm ${C.text65} select-none`}>¥</span>
              <input
                type="number"
                min={0}
                className={`w-32 bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
                value={formData.price === 0 ? "" : String(formData.price)}
                onChange={(e) => {
                  if (e.target.value === "") {
                    setFormData((prev) => ({ ...prev, price: 0 }));
                    return;
                  }
                  const parsed = Number(e.target.value);
                  if (!Number.isNaN(parsed) && parsed >= 0) {
                    setFormData((prev) => ({ ...prev, price: parsed }));
                  }
                }}
                placeholder="0"
              />
            </div>
          </PropertyRow>

          {/* Description */}
          <PropertyRow label="備考">
            <PropInput
              value={formData.description}
              onChange={(v) =>
                setFormData((prev) => ({ ...prev, description: v }))
              }
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
// HospitalizationSettings (main page)
// ─────────────────────────────────────────────────

export function HospitalizationSettings() {
  const navigate = useNavigate();

  // null=closed, "new"=create mode, HospitalizationPlan=edit mode
  const [editTarget, setEditTarget] = useState<HospitalizationPlan | "new" | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<HospitalizationPlan | null>(null);
  const [, startSaveTransition] = useTransition();

  const { data: rawPlans } = useGetAllHospitalizationPlans();
  const createMutation = useCreateHospitalizationPlan();
  const updateMutation = useUpdateHospitalizationPlan();
  const deleteMutation = useDeleteHospitalizationPlan();

  // rerender-transitions: 検索フィルタを低優先度に遅延
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredItems = useMemo(() => {
    const plans = rawPlans ?? [];
    if (!deferredSearch) return plans;
    const lower = deferredSearch.toLowerCase();
    return plans.filter((p) => p.name.toLowerCase().includes(lower));
  }, [rawPlans, deferredSearch]);

  const handleClose = useCallback(() => setEditTarget(null), []);

  const handleSave = useCallback(
    (data: HospitalizationFormData) => {
      if (!data.name.trim()) {
        toast.error("名称は必須です");
        return;
      }

      // rerender-transitions: API書き込みを非緊急マーク
      startSaveTransition(() => {
        if (editTarget !== null && editTarget !== "new") {
          const req: UpdateHospitalizationPlanRequest = {
            name: data.name,
            price: data.price,
            description: data.description || undefined,
            is_active: data.isActive,
            body_size: data.bodySize !== "" ? data.bodySize : null,
            billing_unit: data.billingUnit !== "" ? data.billingUnit : null,
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
          const req: CreateHospitalizationPlanRequest = {
            name: data.name,
            price: data.price || undefined,
            description: data.description || undefined,
            is_active: data.isActive,
            body_size: data.bodySize !== "" ? data.bodySize : undefined,
            billing_unit: data.billingUnit !== "" ? data.billingUnit : undefined,
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
      title="入院マスタ"
      icon={<Bed className="size-5 text-[#37352F]" />}
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
                placeholder="名称で検索..."
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
            emptyMessage="入院プランが登録されていません"
            renderRow={(item) => (
              <DataTableRow key={item.id} onClick={() => setEditTarget(item)}>
                <TableCell className={`font-medium text-sm ${C.text}`}>
                  {item.name}
                </TableCell>
                <TableCell className={`text-sm ${C.text70}`}>
                  {item.bodySize
                    ? (BODY_SIZE_LABELS[item.bodySize] ?? item.bodySize)
                    : "-"}
                </TableCell>
                <TableCell className={`text-sm ${C.text70}`}>
                  {item.billingUnit
                    ? (BILLING_UNIT_LABELS[item.billingUnit] ?? item.billingUnit)
                    : "-"}
                </TableCell>
                <TableCell className={`text-right font-mono text-sm ${C.text}`}>
                  {item.price > 0 ? `¥${item.price.toLocaleString()}` : "-"}
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
          <HospitalizationSidePanel
            key={panelItem ? String(panelItem.id) : "new-hospitalization"}
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
        title="入院プランを削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </PageLayout>
  );
}
