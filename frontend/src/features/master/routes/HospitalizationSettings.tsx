// React/Framework
import type { ReactNode } from "react";
import { useState, useMemo, useCallback } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, Bed, X, Trash2 } from "lucide-react";
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
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
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
import type { HospitalizationPlan } from "@/features/master/api/hospitalization-plans";
import type {
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

function NotionStatusPill({ isActive }: { isActive: boolean }) {
  const cfg = STATUS_CONFIG[isActive ? "active" : "inactive"];
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
      className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px]`}
    >
      <div className="w-[140px] shrink-0 text-sm text-[#37352F]/65 select-none truncate flex items-center">
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// Inline input for property rows
// ─────────────────────────────────────────────────

function PropInput({
  value,
  onChange,
  placeholder,
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
}) {
  return (
    <input
      type="text"
      className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder ?? "空"}
    />
  );
}

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

function HospitalizationSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: {
  item: HospitalizationPlan | null;
  onClose: () => void;
  onSave: (data: HospitalizationFormData) => void;
  onDeleteRequest: (item: HospitalizationPlan) => void;
}) {
  const [formData, setFormData] = useState<HospitalizationFormData>(() => ({
    name: item?.name ?? "",
    price: item?.price ?? 0,
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
    bodySize: item?.bodySize ?? "",
    billingUnit: item?.billingUnit ?? "",
  }));

  return (
    <div className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0`}>
      {/* Toolbar */}
      <div className={STYLE.sidePeekToolbar}>
        <span className={`text-xs ${C.text35} pl-1 select-none`}>
          {item !== null ? "編集" : "新規作成"}
        </span>
        <div className="flex items-center gap-1">
          {item !== null ? (
            <button
              type="button"
              onClick={() => onDeleteRequest(item)}
              className={`${STYLE.sidePeekToolbarBtn} cursor-pointer text-[#EB5757] hover:bg-[#EB5757]/10`}
            >
              <Trash2 className="size-4" />
            </button>
          ) : null}
          <button
            type="button"
            onClick={onClose}
            className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
            aria-label="閉じる"
          >
            <X className="size-4" />
          </button>
        </div>
      </div>

      {/* Body */}
      <div className={STYLE.sidePeekBody}>
        <div className="px-16 pb-8">
          <div className="pt-4 pb-2">
            <div className={STYLE.pageIcon}>
              <Bed className={LAYOUT.pageIcon.innerIcon} />
            </div>
          </div>
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
                setFormData((prev) => ({ ...prev, name: e.target.value }))
              }
              placeholder="無題"
              autoFocus
            />
          </div>
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
                <SelectContent>
                  {BODY_SIZE_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </PropertyRow>

            {/* Billing Unit */}
            <PropertyRow label="料金単位">
              <Select
                value={formData.billingUnit}
                onValueChange={(v) =>
                  setFormData((prev) => ({
                    ...prev,
                    billingUnit: v as BillingUnit,
                  }))
                }
              >
                <SelectTrigger className={STYLE.selectCompact}>
                  <SelectValue placeholder="選択" />
                </SelectTrigger>
                <SelectContent>
                  {BILLING_UNIT_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
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
        </div>
      </div>

      {/* Footer */}
      <div className={STYLE.sidePeekFooter}>
        <button type="button" onClick={onClose} className={STYLE.sidePeekCancelBtn}>
          キャンセル
        </button>
        <button
          type="button"
          onClick={() => onSave(formData)}
          className={STYLE.sidePeekSaveBtn}
        >
          保存
        </button>
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// HospitalizationSettings (main page)
// ─────────────────────────────────────────────────

export function HospitalizationSettings() {
  const navigate = useNavigate();

  const [selectedItem, setSelectedItem] = useState<HospitalizationPlan | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<HospitalizationPlan | null>(null);

  const { data: rawPlans } = useGetAllHospitalizationPlans();
  const createMutation = useCreateHospitalizationPlan();
  const updateMutation = useUpdateHospitalizationPlan();
  const deleteMutation = useDeleteHospitalizationPlan();

  const filteredItems = useMemo(() => {
    const plans = rawPlans ?? [];
    if (!searchTerm) return plans;
    const lower = searchTerm.toLowerCase();
    return plans.filter((p) => p.name.toLowerCase().includes(lower));
  }, [rawPlans, searchTerm]);

  const handleEdit = useCallback((item: HospitalizationPlan) => {
    setSelectedItem(item);
    setIsEditing(true);
  }, []);

  const handleCreate = useCallback(() => {
    setSelectedItem(null);
    setIsEditing(true);
  }, []);

  const handleClose = useCallback(() => {
    setIsEditing(false);
    setSelectedItem(null);
  }, []);

  const handleSave = useCallback(
    (data: HospitalizationFormData) => {
      if (!data.name.trim()) {
        toast.error("名称は必須です");
        return;
      }

      if (selectedItem) {
        const req: UpdateHospitalizationPlanRequest = {
          name: data.name,
          price: data.price,
          description: data.description || undefined,
          is_active: data.isActive,
          body_size: data.bodySize !== "" ? data.bodySize : null,
          billing_unit: data.billingUnit !== "" ? data.billingUnit : null,
        };
        updateMutation.mutate(
          { id: selectedItem.id, req },
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
    },
    [selectedItem, updateMutation, createMutation, handleClose],
  );

  const handleDeleteRequest = useCallback((item: HospitalizationPlan) => {
    setPendingDelete(item);
  }, []);

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

  return (
    <PageLayout
      title="入院マスタ"
      icon={<Bed className="size-5 text-[#37352F]" />}
      onBack={() => navigate("/settings")}
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
              onClick={handleCreate}
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
              <DataTableRow key={item.id} onClick={() => handleEdit(item)}>
                <TableCell className={`font-medium text-sm ${C.text} py-2.5`}>
                  {item.name}
                </TableCell>
                <TableCell className={`text-sm ${C.text70} py-2.5`}>
                  {item.bodySize
                    ? (BODY_SIZE_LABELS[item.bodySize] ?? item.bodySize)
                    : "-"}
                </TableCell>
                <TableCell className={`text-sm ${C.text70} py-2.5`}>
                  {item.billingUnit
                    ? (BILLING_UNIT_LABELS[item.billingUnit] ?? item.billingUnit)
                    : "-"}
                </TableCell>
                <TableCell className={`text-right font-mono text-sm ${C.text} py-2.5`}>
                  {item.price > 0 ? `¥${item.price.toLocaleString()}` : "-"}
                </TableCell>
                <TableCell className="text-center py-2.5">
                  <NotionStatusPill isActive={item.isActive} />
                </TableCell>
                <TableCell className="text-right py-2.5">
                  <RowActionButton onClick={() => handleEdit(item)} />
                </TableCell>
              </DataTableRow>
            )}
          />
        </div>

        {/* Side peek */}
        {isEditing ? (
          <HospitalizationSidePanel
            key={selectedItem ? String(selectedItem.id) : "new-hospitalization"}
            item={selectedItem}
            onClose={handleClose}
            onSave={handleSave}
            onDeleteRequest={handleDeleteRequest}
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
