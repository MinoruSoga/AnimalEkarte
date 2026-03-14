// React/Framework
import { useState, useMemo, useCallback } from "react";
import { useNavigate } from "react-router";

// External
import { toast } from "sonner";
import Pill from "lucide-react/dist/esm/icons/pill";
import Plus from "lucide-react/dist/esm/icons/plus";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import X from "lucide-react/dist/esm/icons/x";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import GripVertical from "lucide-react/dist/esm/icons/grip-vertical";

// Internal – shared
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar/SearchFilterBar";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";

// Internal – feature API (direct import, no barrel)
import {
  useGetAllMedicines,
  useCreateMedicine,
  useUpdateMedicine,
  useDeleteMedicine,
} from "../api/medicines";
import type { CreateMedicineRequest, UpdateMedicineRequest } from "../api/medicines";

// Types
import type { Medicine } from "@/types";

// ─────────────────────────────────────────────────
// Module-level constants (hoisted — never recreated)
// ─────────────────────────────────────────────────

// Static JSX hoisted outside component (rendering-hoist-jsx)
const DOSAGE_FORM_SELECT_ITEMS = (
  <>
    <SelectItem value="tablet">錠剤</SelectItem>
    <SelectItem value="liquid">液剤</SelectItem>
    <SelectItem value="injection">注射剤</SelectItem>
    <SelectItem value="topical">外用剤</SelectItem>
    <SelectItem value="powder">散剤</SelectItem>
  </>
);

const MEDICINE_UNIT_SELECT_ITEMS = (
  <>
    <SelectItem value="per_tablet">錠</SelectItem>
    <SelectItem value="per_ml">ml</SelectItem>
    <SelectItem value="per_dose">回</SelectItem>
    <SelectItem value="per_gram">g</SelectItem>
  </>
);

// ─────────────────────────────────────────────────
// Form data type
// ─────────────────────────────────────────────────

interface MedicineFormData {
  name: string;
  drugCategory: string;
  dosageForm: string;
  medicineUnit: string;
  price: number;
  defaultQuantity: string;
  description: string;
  isActive: boolean;
}

const INITIAL_FORM: MedicineFormData = {
  name: "",
  drugCategory: "",
  dosageForm: "tablet",
  medicineUnit: "per_tablet",
  price: 0,
  defaultQuantity: "",
  description: "",
  isActive: true,
};

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
  children: React.ReactNode;
}) {
  return (
    <div
      className={`flex gap-2 py-2 px-2 -mx-2 rounded-[3px] ${C.hoverBgLight} transition-colors min-h-[40px] items-center`}
    >
      <div className={`w-[140px] shrink-0 text-sm ${C.text65} select-none`}>
        {label}
        {required ? <span className={`${C.textRequired} ml-0.5`}>*</span> : null}
      </div>
      <div className="flex-1 min-w-0">{children}</div>
    </div>
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
        className={`size-[7px] rounded-full ${active ? "bg-[#2383E2]" : "bg-[#37352F]/10"}`}
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
      <span className={`text-sm ${active ? "text-[#37352F]/65" : "text-[#37352F]/35"}`}>
        {active ? "有効" : "無効"}
      </span>
    </span>
  );
}

// ─────────────────────────────────────────────────
// Main component
// ─────────────────────────────────────────────────

export function MedicineSettings() {
  const navigate = useNavigate();

  // ── UI state ──
  const [searchTerm, setSearchTerm] = useState("");
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(new Set());
  const [isEditing, setIsEditing] = useState(false);
  const [selectedMedicine, setSelectedMedicine] = useState<Medicine | null>(null);
  const [formData, setFormData] = useState<MedicineFormData>(INITIAL_FORM);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  // ── API ──
  const { data: medicines = [] } = useGetAllMedicines();
  const createMutation = useCreateMedicine();
  const updateMutation = useUpdateMedicine();
  const deleteMutation = useDeleteMedicine();

  // ── Derived: filtered + grouped (js-cache-function-results) ──
  const { groupedMedicines, totalCount } = useMemo(() => {
    const lower = searchTerm.toLowerCase();
    const filtered = medicines.filter(
      (m) => !searchTerm || m.name.toLowerCase().includes(lower),
    );

    const groups = new Map<string, Medicine[]>();
    for (const m of filtered) {
      const key = m.drugCategory ?? "uncategorized";
      const existing = groups.get(key);
      if (existing) {
        existing.push(m);
      } else {
        groups.set(key, [m]);
      }
    }

    return { groupedMedicines: groups, totalCount: filtered.length };
  }, [medicines, searchTerm]);

  // ── Handlers ──

  const toggleGroup = useCallback((key: string) => {
    setCollapsedGroups((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  }, []);

  const handleCloseEdit = useCallback(() => {
    setIsEditing(false);
    setSelectedMedicine(null);
    setFormData(INITIAL_FORM);
  }, []);

  const handleEdit = useCallback((medicine: Medicine) => {
    setSelectedMedicine(medicine);
    setFormData({
      name: medicine.name,
      drugCategory: medicine.drugCategory ?? "",
      dosageForm: medicine.dosageForm,
      medicineUnit: medicine.medicineUnit,
      price: medicine.price,
      defaultQuantity: medicine.defaultQuantity?.toString() ?? "",
      description: medicine.description,
      isActive: medicine.isActive,
    });
    setIsEditing(true);
  }, []);

  const handleCreate = useCallback((drugCategory?: string) => {
    setSelectedMedicine(null);
    setFormData({ ...INITIAL_FORM, drugCategory: drugCategory !== "uncategorized" ? (drugCategory ?? "") : "" });
    setIsEditing(true);
  }, []);

  const updateForm = useCallback((updates: Partial<MedicineFormData>) => {
    setFormData((prev) => ({ ...prev, ...updates }));
  }, []);

  const handleSave = useCallback(() => {
    if (!formData.name.trim()) {
      toast.error("薬品名は必須です");
      return;
    }

    if (selectedMedicine) {
      const req: UpdateMedicineRequest = {
        name: formData.name,
        drug_category: formData.drugCategory || undefined,
        dosage_form: formData.dosageForm,
        medicine_unit: formData.medicineUnit,
        price: formData.price,
        default_quantity: formData.defaultQuantity ? Number(formData.defaultQuantity) : undefined,
        description: formData.description,
        is_active: formData.isActive,
      };
      updateMutation.mutate(
        { id: selectedMedicine.id, req },
        {
          onSuccess: () => {
            toast.success("更新しました");
            handleCloseEdit();
          },
          onError: () => toast.error("更新に失敗しました"),
        },
      );
    } else {
      const req: CreateMedicineRequest = {
        name: formData.name,
        drug_category: formData.drugCategory || undefined,
        dosage_form: formData.dosageForm,
        medicine_unit: formData.medicineUnit,
        price: formData.price,
        default_quantity: formData.defaultQuantity ? Number(formData.defaultQuantity) : undefined,
        description: formData.description,
        is_active: formData.isActive,
      };
      createMutation.mutate(req, {
        onSuccess: () => {
          toast.success("登録しました");
          handleCloseEdit();
        },
        onError: () => toast.error("登録に失敗しました"),
      });
    }
  }, [formData, selectedMedicine, updateMutation, createMutation, handleCloseEdit]);

  const handleDeleteRequest = useCallback(() => {
    if (!selectedMedicine) return;
    setDeleteConfirmOpen(true);
  }, [selectedMedicine]);

  const executeDelete = useCallback(() => {
    if (!selectedMedicine) return;
    deleteMutation.mutate(selectedMedicine.id, {
      onSuccess: () => {
        toast.success("削除しました");
        setDeleteConfirmOpen(false);
        handleCloseEdit();
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  }, [selectedMedicine, deleteMutation, handleCloseEdit]);

  // ── Table ──
  const tableContent = (
    <div className={STYLE.tableContainer}>
      <div className="flex-1 overflow-auto relative">
        <Table>
          <TableHeader className="sticky top-0 z-10">
            <TableRow className={STYLE.tableHeaderRow}>
              <TableHead className="w-8 px-0" />
              <TableHead className={`${STYLE.tableHeaderCell} pl-3`}>薬品名</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} w-[130px] text-right pr-4`}>
                単価(税込)
              </TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} w-[110px] text-center`}>
                ステータス
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {groupedMedicines.size === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className={STYLE.tableEmpty}>
                  データが見つかりません
                </TableCell>
              </TableRow>
            ) : null}

            {Array.from(groupedMedicines.entries()).map(([key, items]) => {
              const isCollapsed = collapsedGroups.has(key);
              const label = key === "uncategorized" ? "未分類" : key;

              return (
                <>
                  {/* Group header row */}
                  <TableRow
                    key={`h-${key}`}
                    className={`border-b ${C.borderLight} bg-[#F7F6F3]/30 h-9 group/header hover:bg-[#F7F6F3]/60`}
                  >
                    <TableCell className="w-8 px-0 py-0" />
                    <TableCell colSpan={3} className="py-0 pl-0 pr-2">
                      <div className="flex items-center">
                        <button
                          type="button"
                          onClick={() => toggleGroup(key)}
                          className="flex flex-1 items-center gap-1.5 py-1.5 px-1 hover:bg-[rgba(55,53,47,0.04)] rounded-[3px] transition-colors"
                        >
                          <ChevronRight
                            className={`size-3.5 text-[#37352F]/50 transition-transform duration-150 ${
                              isCollapsed ? "" : "rotate-90"
                            }`}
                          />
                          <span className="text-xs font-medium text-[#37352F]/65 uppercase tracking-wide">
                            {label}
                          </span>
                          <span className="text-xs text-[#37352F]/40 ml-0.5">{items.length}</span>
                        </button>
                        <button
                          type="button"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleCreate(key);
                          }}
                          className="w-8 h-8 flex items-center justify-center rounded-[3px] text-[#37352F]/40 hover:bg-[rgba(55,53,47,0.08)] hover:text-[#37352F] transition-colors opacity-0 group-hover/header:opacity-100"
                        >
                          <Plus className="size-3.5" />
                        </button>
                      </div>
                    </TableCell>
                  </TableRow>

                  {/* Medicine rows */}
                  {isCollapsed
                    ? null
                    : items.map((medicine) => (
                        <TableRow
                          key={medicine.id}
                          onClick={() => handleEdit(medicine)}
                          className={`${STYLE.tableRow} group/row`}
                        >
                          <TableCell className="w-8 px-0 py-0 pl-1 text-[#37352F]/20 group-hover/row:text-[#37352F]/50 transition-colors">
                            <GripVertical className="size-4" />
                          </TableCell>
                          <TableCell className={`${STYLE.tableCell} pl-2 font-medium`}>
                            {medicine.name}
                          </TableCell>
                          <TableCell className={`${STYLE.tableCell} w-[130px] text-right pr-4 font-mono`}>
                            {medicine.price > 0 ? `¥${medicine.price.toLocaleString()}` : "-"}
                          </TableCell>
                          <TableCell className="w-[110px] py-2 text-center">
                            <StatusDot active={medicine.isActive} />
                          </TableCell>
                        </TableRow>
                      ))}
                </>
              );
            })}
          </TableBody>
        </Table>
      </div>

      {/* Inline add row */}
      <button
        type="button"
        onClick={() => handleCreate()}
        className="flex items-center gap-1.5 w-full px-3 py-2.5 text-sm text-[#37352F]/40 hover:text-[#37352F]/65 hover:bg-[#F7F6F3]/50 transition-colors rounded-b-[4px]"
      >
        <Plus className="size-3.5" />
        新しい薬剤を追加...
      </button>
    </div>
  );

  // ── Side peek panel ──
  const sidePeekPanel = isEditing ? (
    <div
      className={`flex flex-col self-stretch bg-white border-l ${C.borderLight} shadow-[-1px_0_5px_rgba(0,0,0,0.02)] ${LAYOUT.sidePeek.width} shrink-0`}
    >
      {/* Toolbar */}
      <div className={STYLE.sidePeekToolbar}>
        <span className={`text-xs ${C.text35} pl-1 select-none`}>
          {selectedMedicine ? "編集" : "新規作成"}
        </span>
        <div className="flex items-center gap-1">
          {selectedMedicine ? (
            <button
              type="button"
              onClick={handleDeleteRequest}
              className={`${STYLE.sidePeekToolbarBtn} cursor-pointer ${C.danger} hover:bg-[#EB5757]/10`}
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
              <Pill className={LAYOUT.pageIcon.innerIcon} />
            </div>
          </div>

          {/* Title input (薬品名) */}
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
              onChange={(e) => updateForm({ name: e.target.value })}
              placeholder="薬品名"
            />
          </div>

          <div className={`${STYLE.sectionDivider} mb-1`} />

          {/* Properties */}
          <div className="py-1">
            {/* Price */}
            <PropertyRow label="単価(税込)">
              <div className="flex items-center gap-1">
                <span className={`text-sm ${C.text40}`}>¥</span>
                <input
                  type="number"
                  min={0}
                  value={formData.price}
                  onChange={(e) => updateForm({ price: Number(e.target.value) })}
                  placeholder="0"
                  className={`${STYLE.propertyInput} w-28`}
                />
              </div>
            </PropertyRow>

            {/* Status */}
            <PropertyRow label="ステータス">
              <button
                type="button"
                onClick={() => updateForm({ isActive: !formData.isActive })}
                className="inline-flex items-center rounded-[3px] hover:bg-[rgba(55,53,47,0.04)] transition-colors py-0.5 px-0.5 cursor-pointer"
              >
                <NotionStatusPill active={formData.isActive} />
              </button>
            </PropertyRow>

            {/* Drug category */}
            <PropertyRow label="薬効分類">
              <input
                type="text"
                value={formData.drugCategory}
                onChange={(e) => updateForm({ drugCategory: e.target.value })}
                placeholder="例: 抗生剤、消炎剤"
                className={STYLE.propertyInput}
              />
            </PropertyRow>

            {/* Description */}
            <PropertyRow label="備考">
              <input
                type="text"
                value={formData.description}
                onChange={(e) => updateForm({ description: e.target.value })}
                placeholder="空"
                className={STYLE.propertyInput}
              />
            </PropertyRow>
          </div>

          {/* ── 薬剤詳細 section ── */}
          <div className={`${STYLE.sectionDivider} mt-3 mb-1`} />
          <div className="py-1">
            <div className={`flex items-center gap-1.5 py-2 mb-1`}>
              <Pill className={`size-3.5 ${C.text40}`} />
              <span className={`text-xs font-medium ${C.text50} uppercase tracking-wide select-none`}>
                薬剤詳細
              </span>
            </div>

            {/* Dosage form */}
            <PropertyRow label="剤形" required>
              <Select
                value={formData.dosageForm}
                onValueChange={(v) => updateForm({ dosageForm: v })}
              >
                <SelectTrigger className={STYLE.selectCompact}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>{DOSAGE_FORM_SELECT_ITEMS}</SelectContent>
              </Select>
            </PropertyRow>

            {/* Unit */}
            <PropertyRow label="単位">
              <Select
                value={formData.medicineUnit}
                onValueChange={(v) => updateForm({ medicineUnit: v })}
              >
                <SelectTrigger className={STYLE.selectCompact}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>{MEDICINE_UNIT_SELECT_ITEMS}</SelectContent>
              </Select>
            </PropertyRow>

            {/* Default quantity */}
            <PropertyRow label="デフォルト数量">
              <input
                type="number"
                min={1}
                value={formData.defaultQuantity}
                onChange={(e) => updateForm({ defaultQuantity: e.target.value })}
                placeholder="空"
                className={`${STYLE.propertyInput} w-28`}
              />
            </PropertyRow>
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className={STYLE.sidePeekFooter}>
        <button type="button" onClick={handleCloseEdit} className={STYLE.sidePeekCancelBtn}>
          キャンセル
        </button>
        <button type="button" onClick={handleSave} className={STYLE.sidePeekSaveBtn}>
          保存
        </button>
      </div>
    </div>
  ) : null;

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="薬剤マスタ"
            icon={<Pill className="size-5 text-[#37352F]" />}
            onBack={() => navigate("/settings")}
            maxWidth="max-w-full"
          >
            <div className="flex flex-col gap-4">
              <div className="flex items-center gap-3">
                <div className="flex-1 min-w-0">
                  <SearchFilterBar
                    searchTerm={searchTerm}
                    onSearchChange={setSearchTerm}
                    placeholder="薬品名で検索..."
                    count={totalCount}
                  />
                </div>
                <button
                  type="button"
                  onClick={() => handleCreate()}
                  className={`flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-[4px] ${C.accent} ${C.hoverBgAccent5} transition-colors whitespace-nowrap`}
                >
                  <Plus className="size-3.5" />
                  新規登録
                </button>
              </div>
              {tableContent}
            </div>
          </PageLayout>
        </div>
        {sidePeekPanel}
      </div>

      <ConfirmDialog
        open={deleteConfirmOpen}
        onClose={() => setDeleteConfirmOpen(false)}
        onConfirm={executeDelete}
        title="削除しますか？"
        description={`「${selectedMedicine?.name ?? ""}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除する"
        cancelLabel="キャンセル"
        variant="destructive"
      />
    </>
  );
}
