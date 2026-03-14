// React/Framework
import type { ReactNode } from "react";
import { useState, useMemo } from "react";
import { useNavigate } from "react-router";

// External
import { Plus, Bed, X, Trash2 } from "lucide-react";
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
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
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
  { header: "ステータス", className: "w-[90px]", align: "right" as const },
];

// ─────────────────────────────────────────────────
// Sub-components
// ─────────────────────────────────────────────────

function NotionStatusPill({ isActive }: { isActive: boolean }) {
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-sm text-xs ${
        isActive
          ? "bg-[#D3E5EF] text-[#183B56]"
          : "bg-[#E3E2E0] text-[#37352F]/60"
      }`}
    >
      <span
        className={`size-[7px] rounded-full ${
          isActive ? "bg-[#2383E2]" : "bg-[#37352F]/10"
        }`}
      />
      {isActive ? "有効" : "無効"}
    </span>
  );
}

function PropertyRow({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) {
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
  price: string;
  description: string;
  isActive: boolean;
  bodySize: BodySize | "";
  billingUnit: BillingUnit | "";
}

const DEFAULT_FORM: HospitalizationFormData = {
  name: "",
  price: "",
  description: "",
  isActive: true,
  bodySize: "",
  billingUnit: "",
};

// ─────────────────────────────────────────────────
// HospitalizationSettings
// ─────────────────────────────────────────────────

export function HospitalizationSettings() {
  const navigate = useNavigate();
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<HospitalizationPlan | null>(
    null,
  );
  const [searchTerm, setSearchTerm] = useState("");
  const [formData, setFormData] =
    useState<HospitalizationFormData>(DEFAULT_FORM);
  const [pendingDelete, setPendingDelete] =
    useState<HospitalizationPlan | null>(null);

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

  const handleEdit = (item: HospitalizationPlan) => {
    setSelectedItem(item);
    setFormData({
      name: item.name,
      price: item.price != null ? String(item.price) : "",
      description: item.description ?? "",
      isActive: item.isActive,
      bodySize: item.bodySize ?? "",
      billingUnit: item.billingUnit ?? "",
    });
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setFormData(DEFAULT_FORM);
    setIsEditing(true);
  };

  const handleCloseEdit = () => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData(DEFAULT_FORM);
  };

  const handleSave = () => {
    if (!formData.name.trim()) {
      toast.error("名称は必須です");
      return;
    }

    const priceValue = formData.price !== "" ? Number(formData.price) : 0;

    if (selectedItem) {
      const req: UpdateHospitalizationPlanRequest = {
        name: formData.name,
        price: priceValue,
        description: formData.description || undefined,
        is_active: formData.isActive,
        body_size: formData.bodySize !== "" ? formData.bodySize : null,
        billing_unit: formData.billingUnit !== "" ? formData.billingUnit : null,
      };
      updateMutation.mutate(
        { id: selectedItem.id, req },
        {
          onSuccess: () => {
            toast.success("更新しました");
            setIsEditing(false);
          },
          onError: () => toast.error("更新に失敗しました"),
        },
      );
    } else {
      const req: CreateHospitalizationPlanRequest = {
        name: formData.name,
        price: priceValue,
        description: formData.description || undefined,
        is_active: true,
        body_size:
          formData.bodySize !== "" ? formData.bodySize : undefined,
        billing_unit:
          formData.billingUnit !== "" ? formData.billingUnit : undefined,
      };
      createMutation.mutate(req, {
        onSuccess: () => {
          toast.success("登録しました");
          setIsEditing(false);
        },
        onError: () => toast.error("登録に失敗しました"),
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
      onError: () => toast.error("削除に失敗しました"),
    });
  };

  return (
    <>
      <div className="flex h-full">
        {/* ── Left: List ── */}
        <div className="flex-1 min-w-0">
          <PageLayout
            title="入院マスタ"
            icon={<Bed className="size-5 text-[#37352F]" />}
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
                placeholder="名称で検索..."
                count={filteredItems.length}
              />

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
                      {item.price != null
                        ? `¥${item.price.toLocaleString()}`
                        : "-"}
                    </TableCell>
                    <TableCell className="text-right py-2.5">
                      <span className="inline-flex items-center gap-1.5">
                        <span
                          className={`size-[7px] rounded-full ${
                            item.isActive
                              ? "bg-[#2383E2]"
                              : "bg-[#37352F]/20"
                          }`}
                        />
                        <span
                          className={`text-sm ${
                            item.isActive
                              ? "text-[#37352F]/65"
                              : "text-[#37352F]/35"
                          }`}
                        >
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
                新しい入院プランを追加...
              </button>
            </div>
          </PageLayout>
        </div>

        {/* ── Right: Side Peek ── */}
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
                      setFormData({ ...formData, name: e.target.value })
                    }
                    placeholder="無題"
                  />
                </div>

                <div className={`${STYLE.sectionDivider} mb-1`} />

                <div className="py-1">
                  {/* Status */}
                  <PropertyRow label="ステータス">
                    <button
                      type="button"
                      onClick={() =>
                        setFormData({
                          ...formData,
                          isActive: !formData.isActive,
                        })
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
                        setFormData({
                          ...formData,
                          bodySize: v as BodySize,
                        })
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
                        setFormData({
                          ...formData,
                          billingUnit: v as BillingUnit,
                        })
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
                      <span className={`text-sm ${C.text65} select-none`}>
                        ¥
                      </span>
                      <input
                        type="number"
                        min={0}
                        className={`w-32 bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
                        value={formData.price}
                        onChange={(e) =>
                          setFormData({ ...formData, price: e.target.value })
                        }
                        placeholder="0"
                      />
                    </div>
                  </PropertyRow>

                  {/* Description */}
                  <PropertyRow label="備考">
                    <PropInput
                      value={formData.description}
                      onChange={(v) =>
                        setFormData({ ...formData, description: v })
                      }
                      placeholder="補足情報など"
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
        title="入院プランを削除しますか？"
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}
