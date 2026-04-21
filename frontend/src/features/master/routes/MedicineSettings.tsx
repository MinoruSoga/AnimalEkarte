// React/Framework
import { useState, useMemo, useCallback, useDeferredValue, useTransition, useEffect, memo, Fragment } from "react";
import { flushSync } from "react-dom";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";

// DnD
import { DndContext, DragOverlay, closestCenter, type DragEndEvent } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

// External
import { AnimatePresence, motion } from "motion/react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import Pill from "lucide-react/dist/esm/icons/pill";
import Plus from "lucide-react/dist/esm/icons/plus";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import GripVertical from "lucide-react/dist/esm/icons/grip-vertical";
import MoreHorizontal from "lucide-react/dist/esm/icons/more-horizontal";
import Maximize2 from "lucide-react/dist/esm/icons/maximize-2";

// Internal – shared
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { PropertyRow, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, STYLE, LAYOUT, ICON } from "@/lib/design-tokens";
import { useReducedMotion } from "@/hooks/use-reduced-motion";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";

// Internal – feature API (direct import, no barrel)
import { useGetAllMedicines, useCreateMedicine, useUpdateMedicine, useDeleteMedicine, useReorderMedicines } from "../api/medicines";
import type { CreateMedicineRequest, UpdateMedicineRequest } from "@/types/medicine";
import { TaxTypeSelector } from "@/components/shared/TaxTypeSelector/TaxTypeSelector";
import { TaxRateSelector } from "@/components/shared/TaxRateSelector/TaxRateSelector";
import type { TaxType } from "@/types/generated/models";
import { ResourceMasterMedical } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

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
    <SelectItem value="per_tablet">1錠あたり</SelectItem>
    <SelectItem value="per_ml">1mlあたり</SelectItem>
    <SelectItem value="per_dose">1回あたり</SelectItem>
    <SelectItem value="per_gram">1gあたり</SelectItem>
  </>
);

// Full-width Select trigger — matches Figma (h-[30px], no border, rounded-[3px])
const SELECT_TRIGGER_FULL = `h-[30px] text-base bg-transparent ${C.text} border-0 ${C.hoverBgLight} px-1.5 shadow-none rounded-[3px] w-full`;

// 剤形ラベルマップ
const DOSAGE_FORM_LABELS: Record<string, string> = {
  tablet: "錠剤",
  liquid: "液剤",
  injection: "注射剤",
  topical: "外用剤",
  powder: "散剤",
};

// ─────────────────────────────────────────────────
// Form data type
// ─────────────────────────────────────────────────

interface MedicineFormData {
  name: string;
  parentId: string; // medicine の ID 文字列、空文字 = 親なし
  dosageForm: string;
  medicineUnit: string;
  price: number;
  description: string;
  isActive: boolean;
  taxType: TaxType;
  taxRate: number;
}

const INITIAL_FORM: MedicineFormData = {
  name: "",
  parentId: "",
  dosageForm: "tablet",
  medicineUnit: "per_tablet",
  price: 0,
  description: "",
  isActive: true,
  taxType: "excluded",
  taxRate: 0.1,
};

// 編集対象 medicine から初期 formData を構築（マウント時 1 回）
function medicineToFormData(m: Medicine | null, defaultParentId?: string): MedicineFormData {
  if (!m) {
    return {
      ...INITIAL_FORM,
      parentId: defaultParentId && defaultParentId !== "uncategorized" ? defaultParentId : "",
    };
  }
  return {
    name: m.name,
    parentId: m.parentId ?? "",
    dosageForm: m.dosageForm ?? "",
    medicineUnit: m.medicineUnit ?? "",
    price: m.price,
    description: m.description,
    isActive: m.isActive,
    taxType: m.taxType ?? "excluded",
    taxRate: m.taxRate ?? 0.1,
  };
}

// ─────────────────────────────────────────────────
// SortableMedicineRow
// ─────────────────────────────────────────────────

function SortableMedicineRow({
  medicine,
  onEdit,
  grouped,
  canEdit,
}: {
  medicine: Medicine;
  onEdit: (medicine: Medicine) => void;
  grouped: boolean;
  canEdit: boolean;
}) {
  return (
    <SortableDataTableRow id={medicine.id} onClick={canEdit ? () => onEdit(medicine) : undefined}>
      <TableCell className={`${STYLE.tableCell} font-medium ${grouped ? "pl-12!" : "pl-2"}`}>
        {medicine.name}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} w-[100px] text-center text-base`}>
        {medicine.dosageForm ? (DOSAGE_FORM_LABELS[medicine.dosageForm] ?? medicine.dosageForm) : "-"}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} w-[130px] text-right pr-4 font-mono`}>
        {medicine.price > 0 ? `¥${medicine.price.toLocaleString()}` : "-"}
      </TableCell>
      <TableCell className="w-[110px] py-2.5 text-center">
        <NotionStatusPill isActive={medicine.isActive} />
      </TableCell>
      <TableCell className="w-[80px] py-2.5 text-center" onClick={(e) => e.stopPropagation()}>
        {canEdit ? (
          <div className="flex items-center justify-center gap-1">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  type="button"
                  className={`${STYLE.iconBtn28} ${C.text40} ${C.hoverBgMedium} ${C.hoverText}`}
                >
                  <MoreHorizontal className={ICON.action} />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => onEdit(medicine)}>編集</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            <button
              type="button"
              onClick={() => onEdit(medicine)}
              className={`${STYLE.iconBtn28} ${C.text40} ${C.hoverBgMedium} ${C.hoverText}`}
            >
              <Maximize2 className={`${ICON.xs}`} />
            </button>
          </div>
        ) : null}
      </TableCell>
    </SortableDataTableRow>
  );
}

// ─────────────────────────────────────────────────
// MedicineRowOverlay — DragOverlay 用（テーブル外ポータルに描画）
// ─────────────────────────────────────────────────

function MedicineRowOverlay({
  medicine,
  grouped,
}: {
  medicine: Medicine;
  grouped: boolean;
}) {
  return (
    <div
      className={`flex items-center h-12 ${C.bgWhite} border ${C.borderLight} rounded-[4px] shadow-[0_4px_16px_rgba(0,0,0,0.12)] cursor-grabbing`}
      style={{ width: "100%" }}
    >
      <div className={`w-8 shrink-0 flex items-center justify-center ${C.text50}`}>
        <GripVertical className={ICON.action} />
      </div>
      <div className={`flex-1 min-w-0 text-base font-medium ${C.text} ${grouped ? "pl-10" : "pl-0"}`}>
        {medicine.name}
      </div>
      <div className={`w-[100px] shrink-0 text-center text-base ${C.text65}`}>
        {medicine.dosageForm ? (DOSAGE_FORM_LABELS[medicine.dosageForm] ?? medicine.dosageForm) : "-"}
      </div>
      <div className={`w-[130px] shrink-0 text-right pr-4 font-mono text-base ${C.text}`}>
        {medicine.price > 0 ? `¥${medicine.price.toLocaleString()}` : "-"}
      </div>
      <div className="w-[110px] shrink-0 flex justify-center">
        <NotionStatusPill isActive={medicine.isActive} />
      </div>
      <div className="w-[80px] shrink-0" />
    </div>
  );
}

// ─────────────────────────────────────────────────
// MedicineSidePanel
// ─────────────────────────────────────────────────

// MedicineSidePanelBody は formData を内部で所有し、14 マスタ画面と同じ
// child-marks 戦略 (useEffect → onDirtyChange) を採用する。
// 編集対象切替時は親の <key> で再マウント → useState lazy init で formData がリセットされる。
interface MedicineSidePanelBodyProps {
  selectedMedicine: Medicine | null;
  isCategory: boolean;
  defaultParentId?: string;
  categoryMedicines: Medicine[];
  onCloseEdit: () => void;
  onSave: (data: MedicineFormData) => void;
  onDeleteRequest: () => void;
  readOnly?: boolean;
  canDelete?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

const MedicineSidePanelBody = memo(function MedicineSidePanelBody({
  selectedMedicine,
  isCategory,
  defaultParentId,
  categoryMedicines,
  onCloseEdit,
  onSave,
  onDeleteRequest,
  readOnly,
  canDelete,
  onDirtyChange,
}: MedicineSidePanelBodyProps) {
  const [formData, setFormData] = useState<MedicineFormData>(() =>
    medicineToFormData(selectedMedicine, defaultParentId),
  );
  const [nameError, setNameError] = useState("");
  const [isDirty, setIsDirty] = useState(false);

  // BUG-380
  useEffect(() => { onDirtyChange?.(isDirty); }, [isDirty, onDirtyChange]);
  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(formData);
    setIsDirty(false);
  }, [formData, onSave]);

  const handleTitleChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: v }));
    if (v.trim()) setNameError("");
  }, [setFormDataDirty]);

  return (
    <MasterSidePanel
      isNew={!selectedMedicine}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={onCloseEdit}
      action={handleAction}
      onDelete={selectedMedicine && canDelete ? onDeleteRequest : undefined}
      icon={<Pill className={LAYOUT.pageIcon.innerIcon} />}
      titlePlaceholder="薬品名"
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      {/* Properties */}
      <PropertyRow label="親カテゴリ">
        {isCategory ? (
          <span className={`text-base ${C.text}`}>なし（ルート）</span>
        ) : (
          <Select
            value={formData.parentId || "__none__"}
            onValueChange={(v) => setFormDataDirty((prev) => ({ ...prev, parentId: v === "__none__" ? "" : v }))}
          >
            <SelectTrigger className={SELECT_TRIGGER_FULL}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="__none__">なし（未分類）</SelectItem>
              {categoryMedicines.map((cat) => (
                <SelectItem key={cat.id} value={cat.id}>{cat.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </PropertyRow>

      <PropertyRow label="単価(税込)">
        {isCategory ? (
          <span className={`text-base ${C.text35} select-none`}>子項目に金額を設定</span>
        ) : (
          <div className="flex items-center gap-1">
            <span className={`text-base ${C.text40}`}>¥</span>
            <input
              type="number"
              min={0}
              value={formData.price}
              onChange={(e) => setFormDataDirty((prev) => ({ ...prev, price: Number(e.target.value) }))}
              placeholder="0"
              className={`${STYLE.propertyInput} w-28`}
            />
          </div>
        )}
      </PropertyRow>

      <PropertyRow label="課税区分">
        <TaxTypeSelector
          value={formData.taxType}
          onChange={(v) => setFormDataDirty((prev) => ({ ...prev, taxType: v }))}
          disabled={isCategory}
        />
      </PropertyRow>

      <PropertyRow label="税率">
        <TaxRateSelector
          value={formData.taxRate}
          onChange={(v) => setFormDataDirty((prev) => ({ ...prev, taxRate: v }))}
          disabled={isCategory}
        />
      </PropertyRow>

      <PropertyRow label="ステータス">
        <button
          type="button"
          onClick={() => setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }))}
          className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <NotionStatusPill isActive={formData.isActive} />
        </button>
      </PropertyRow>

      <PropertyRow label="備考">
        <PropertyInput
          value={formData.description}
          onChange={(v) => setFormDataDirty((prev) => ({ ...prev, description: v }))}
          placeholder="空"
        />
      </PropertyRow>

      <div className={`${STYLE.sectionDivider} mt-3 mb-1`} />
      <div className="py-1">
        <div className="flex items-center gap-1.5 py-2 mb-1">
          <Pill className={`${ICON.xs} ${C.text40}`} />
          <span
            className={`text-base font-medium ${C.text50} uppercase tracking-wide select-none`}
          >
            薬剤詳細
          </span>
        </div>

        <PropertyRow label="剤形">
          <Select
            value={formData.dosageForm}
            onValueChange={(v) => setFormDataDirty((prev) => ({ ...prev, dosageForm: v }))}
          >
            <SelectTrigger className={SELECT_TRIGGER_FULL}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{DOSAGE_FORM_SELECT_ITEMS}</SelectContent>
          </Select>
        </PropertyRow>

        <PropertyRow label="単位">
          <Select
            value={formData.medicineUnit}
            onValueChange={(v) => setFormDataDirty((prev) => ({ ...prev, medicineUnit: v }))}
          >
            <SelectTrigger className={SELECT_TRIGGER_FULL}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{MEDICINE_UNIT_SELECT_ITEMS}</SelectContent>
          </Select>
        </PropertyRow>
      </div>
    </MasterSidePanel>
  );
});

// 外側ラッパは AnimatePresence + motion.div の責務のみ。Body の formData は
// key={selectedMedicine?.id ?? "new"} で再マウント → 新規 lazy init される。
interface MedicineSidePanelProps {
  isEditing: boolean;
  selectedMedicine: Medicine | null;
  isCategory: boolean;
  defaultParentId?: string;
  categoryMedicines: Medicine[];
  panelDuration: number;
  onCloseEdit: () => void;
  onSave: (data: MedicineFormData) => void;
  onDeleteRequest: () => void;
  readOnly?: boolean;
  canDelete?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

const MedicineSidePanel = memo(function MedicineSidePanel({
  isEditing,
  selectedMedicine,
  isCategory,
  defaultParentId,
  categoryMedicines,
  panelDuration,
  onCloseEdit,
  onSave,
  onDeleteRequest,
  readOnly,
  canDelete,
  onDirtyChange,
}: MedicineSidePanelProps) {
  return (
    <AnimatePresence>
      {isEditing ? (
        <motion.div
          key="side-peek"
          initial={{ width: 0, opacity: 0 }}
          animate={{ width: LAYOUT.sidePeek.widthPx, opacity: 1 }}
          exit={{ width: 0, opacity: 0 }}
          transition={{ duration: panelDuration, ease: [0.25, 0.1, 0.25, 1] }}
          className="shrink-0 min-h-0 overflow-hidden"
        >
          <MedicineSidePanelBody
            key={selectedMedicine?.id ?? "new"}
            selectedMedicine={selectedMedicine}
            isCategory={isCategory}
            defaultParentId={defaultParentId}
            categoryMedicines={categoryMedicines}
            onCloseEdit={onCloseEdit}
            onSave={onSave}
            onDeleteRequest={onDeleteRequest}
            readOnly={readOnly}
            canDelete={canDelete}
            onDirtyChange={onDirtyChange}
          />
        </motion.div>
      ) : null}
    </AnimatePresence>
  );
});

// ─────────────────────────────────────────────────
// Main component
// ─────────────────────────────────────────────────

// カテゴリかどうかの判定ヘルパー
function isCategoryMedicine(m: Medicine | null): boolean {
  return m !== null && !m.parentId && m.price === 0;
}

export function MedicineSettings() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);
  const reduced = useReducedMotion();
  const panelDuration = reduced ? 0 : 0.2;

  // ── UI state (non-CRUD: kept external per Option A) ──
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(new Set());
  // BUG-380 統一: formData は SidePanel が所有。新規作成時の親カテゴリ起点だけ親で保持。
  const [defaultParentId, setDefaultParentId] = useState<string | undefined>(undefined);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  // 値は parentId 文字列、undefined = 親なし
  const [overrideCategories, setOverrideCategories] = useState<Map<string, string | undefined>>(new Map());

  // ── API ──
  const { data: medicines = [] } = useGetAllMedicines();
  const createMutation = useCreateMedicine();
  const updateMutation = useUpdateMedicine();
  const deleteMutation = useDeleteMedicine();
  const reorderMutation = useReorderMedicines();

  // ── FR1: useMasterCRUD (editTarget state only; deletion modal kept external) ──
  const medicineCrud = useMasterCRUD<Medicine>({
    data: medicines,
    deleteMutation,
    entityLabel: "薬品",
    dirtyGuard: dirty,
  });

  // ── Derived: editTarget → selectedMedicine / isEditing / isCategory ──
  const { editTarget } = medicineCrud;
  const selectedMedicine = editTarget !== null && editTarget !== "new" ? editTarget : null;
  const isEditing = editTarget !== null;
  const isCategory = isCategoryMedicine(selectedMedicine);

  // DnD (flat sort 担当) ──
  const {
    orderedItems: sortedMedicines,
    sensors,
    activeId,
    handleDragStart,
    handleDragCancel,
    handleDragEnd: handleFlatSortDragEnd,
    resetOrder,
  } = useSortableList({
    items: medicines,
    onReorder: (newIds) => {
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        { onSuccess: resetOrder },
      );
    },
  });

  // ── Derived: overrideCategories 適用済みリスト ──
  const orderedMedicines = useMemo(() => {
    if (overrideCategories.size === 0) return sortedMedicines;
    return sortedMedicines.map((m) =>
      overrideCategories.has(m.id)
        ? { ...m, parentId: overrideCategories.get(m.id) }
        : m,
    );
  }, [sortedMedicines, overrideCategories]);

  // ── Derived: medicines ID → Medicine マップ (js-cache-function-results) ──
  const medicinesById = useMemo(
    () => new Map(medicines.map((m) => [m.id, m])),
    [medicines],
  );

  // ── Derived: orderedMedicines ID → Medicine マップ（DnD handleDragEnd 用 O(1) 検索） ──
  const orderedMedicinesById = useMemo(
    () => new Map(orderedMedicines.map((m) => [m.id, m])),
    [orderedMedicines],
  );

  // ── Derived: カテゴリ medicine（parentId なし、price === 0）(js-cache-function-results) ──
  const categoryMedicines = useMemo(
    () => medicines.filter((m) => !m.parentId && m.price === 0),
    [medicines],
  );

  // ── Derived: filtered + grouped + ungrouped (js-cache-function-results) ──
  const deferredSearch = useDeferredValue(searchTerm);

  const { groupedMedicines, ungroupedMedicines, totalCount } = useMemo(() => {
    let items = orderedMedicines;
    for (const f of activeFilters) {
      if (f.key === "status" && typeof f.value === "string") {
        const want = f.value === "active";
        items = items.filter((m) => f.condition === "is" ? m.isActive === want : m.isActive !== want);
      }
    }
    const lower = deferredSearch.toLowerCase();
    const filtered = items.filter(
      (m) => !deferredSearch || m.name.toLowerCase().includes(lower),
    );

    const groups = new Map<string, { header: Medicine; items: Medicine[] }>();
    const ungrouped: Medicine[] = [];

    for (const m of filtered) {
      if (m.parentId) {
        // 子 medicine → 親グループに追加
        const existing = groups.get(m.parentId);
        if (existing) {
          existing.items.push(m);
        } else {
          // 親が filtered にない場合でも medicinesById から取得
          const parent = medicinesById.get(m.parentId);
          if (parent) {
            groups.set(m.parentId, { header: parent, items: [m] });
          } else {
            ungrouped.push(m); // 孤立アイテム
          }
        }
      } else if (m.price === 0) {
        // カテゴリ medicine（price=0, parentId なし）→ グループヘッダー
        if (!groups.has(m.id)) {
          groups.set(m.id, { header: m, items: [] });
        }
      } else {
        // 通常の ungrouped medicine（price > 0, parentId なし）
        ungrouped.push(m);
      }
    }

    return { groupedMedicines: groups, ungroupedMedicines: ungrouped, totalCount: filtered.length };
  }, [orderedMedicines, activeFilters, deferredSearch, medicinesById]);

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

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event;
      if (!over || active.id === over.id) return;

      const activeItemId = String(active.id);
      const overItemId = String(over.id);

      // js-index-maps: orderedMedicinesById Map で O(1) 検索（orderedMedicines.find は O(n)）
      const activeMedicine = orderedMedicinesById.get(activeItemId);
      const overMedicine = orderedMedicinesById.get(overItemId);
      if (!activeMedicine || !overMedicine) return;

      const activeCat = activeMedicine.parentId ?? null;
      const overCat = overMedicine.parentId ?? null;

      if (activeCat !== overCat) {
        // クロスグループ: parent_id を変更
        flushSync(() => {
          setOverrideCategories((prev) => {
            const next = new Map(prev);
            next.set(activeItemId, overCat ?? undefined);
            return next;
          });
        });
        const req: UpdateMedicineRequest = overCat
          ? { parent_id: Number(overCat) }
          : { clear_parent_id: true };
        const clearOptimistic = () => {
          setOverrideCategories((prev) => {
            const next = new Map(prev);
            next.delete(activeItemId);
            return next;
          });
        };
        updateMutation.mutate(
          { id: activeItemId, req },
          {
            onSuccess: clearOptimistic,
            onError: (error: unknown) => {
              handleApiError(error, "カテゴリの変更");
              clearOptimistic();
            },
          },
        );
      } else {
        // 同カテゴリ: 並び替え — useSortableList に委譲
        handleFlatSortDragEnd(event);
      }
    },
    [orderedMedicinesById, updateMutation, handleFlatSortDragEnd],
  );

  // BUG-380: 未保存破棄ガード（14 マスタ画面と統一パターン）
  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const handleCloseEdit = useCallback(() => {
    medicineCrud.handleClose();
    setDefaultParentId(undefined);
  }, [medicineCrud]);

  const handleEdit = useCallback((medicine: Medicine) => {
    setDefaultParentId(undefined);
    medicineCrud.handleEdit(medicine);
  }, [medicineCrud]);

  const handleCreate = useCallback((parentId?: string) => {
    setDefaultParentId(parentId);
    medicineCrud.handleCreate();
  }, [medicineCrud]);

  // ── startSaveTransition wrapper for useMasterSave ──
  const [, startSaveTransition] = useTransition();
  const startSaveTransitionWrapper = useCallback((cb: () => void) => {
    startSaveTransition(cb);
  }, []);

  // ── FR2: useMasterSave (with complex price/parent logic) ──
  const medicineSave = useMasterSave<Medicine, MedicineFormData, CreateMedicineRequest, UpdateMedicineRequest>({
    crud: {
      editTarget,
      handleClose: handleCloseEdit,
      startSaveTransition: startSaveTransitionWrapper,
    },
    createMutation,
    updateMutation,
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: (data) => {
      // カテゴリ編集時は price を 0 に固定
      const effectivePrice = isCategory ? 0 : data.price;
      return {
        name: data.name,
        dosage_form: data.dosageForm || undefined,
        medicine_unit: data.medicineUnit || undefined,
        price: effectivePrice,
        description: data.description,
        is_active: data.isActive,
        tax_type: data.taxType,
        tax_rate: data.taxRate,
        ...(data.parentId ? { parent_id: Number(data.parentId) } : {}),
      };
    },
    toUpdateRequest: (data) => {
      // カテゴリ編集時は price を 0 に固定
      const effectivePrice = isCategory ? 0 : data.price;
      const req: UpdateMedicineRequest = {
        name: data.name,
        dosage_form: data.dosageForm || undefined,
        medicine_unit: data.medicineUnit || undefined,
        price: effectivePrice,
        description: data.description,
        is_active: data.isActive,
        tax_type: data.taxType,
        tax_rate: data.taxRate,
      };
      // parent_id の処理
      if (data.parentId) {
        req.parent_id = Number(data.parentId);
      } else if (selectedMedicine?.parentId) {
        // 元々グループに属していたが今は外す
        req.clear_parent_id = true;
      }
      return req;
    },
  });

  // handleSave delegates to hook
  const handleSave = useCallback((formData: MedicineFormData) => {
    medicineSave.handleSave(formData);
  }, [medicineSave]);

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
      onError: (error) => handleApiError(error, "薬剤の削除"),
    });
  }, [selectedMedicine, deleteMutation, handleCloseEdit]);

  // ── Table ──
  const tableContent = (
    <div className={STYLE.tableContainer}>
      <div className="flex-1 overflow-auto relative">
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
          onDragCancel={handleDragCancel}
        >
        <Table>
          <TableHeader className="sticky top-0 z-10">
            <TableRow className={STYLE.tableHeaderRow}>
              <TableHead className="w-8 px-0" />
              <TableHead className={`${STYLE.tableHeaderCell} pl-3`}>薬品名</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} w-[100px] text-center`}>剤形</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} w-[130px] text-right pr-4`}>
                単価(税込)
              </TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} w-[110px] text-center`}>
                ステータス
              </TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} w-[80px] text-center`}>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {groupedMedicines.size === 0 && ungroupedMedicines.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className={STYLE.tableEmpty}>
                  データが見つかりません
                </TableCell>
              </TableRow>
            ) : null}

            {Array.from(groupedMedicines.entries()).map(([parentId, { header, items }]) => {
              const isCollapsed = collapsedGroups.has(parentId);

              return (
                <Fragment key={parentId}>
                  {/* Group header row — クリックでカテゴリ編集パネルを開く */}
                  <TableRow
                    onClick={() => handleEdit(header)}
                    className={`${STYLE.tableRow} border-b ${C.borderLight} ${C.bgPage30} group/header ${C.hoverBgPage60}`}
                  >
                    {/* Grip handle — left (クリック伝播を止める) */}
                    <TableCell className="w-8 px-0 py-0">
                      <button
                        type="button"
                        tabIndex={-1}
                        onClick={(e) => e.stopPropagation()}
                        className={`${STYLE.iconBtn32} ${C.text20} ${C.hoverBgMedium} ${C.hoverText60} cursor-grab`}
                      >
                        <GripVertical className={ICON.action} />
                      </button>
                    </TableCell>

                    {/* Chevron + label + count + plus button */}
                    <TableCell className="py-0 pl-0 pr-2">
                      <div className="flex items-center">
                        <button
                          type="button"
                          onClick={(e) => {
                            e.stopPropagation();
                            toggleGroup(parentId);
                          }}
                          className={`flex items-center gap-1.5 py-1.5 px-1 ${C.hoverBgLight} rounded-[3px] transition-colors`}
                        >
                          <ChevronRight
                            className={`${ICON.xs} ${C.text50} transition-transform duration-150 ${
                              isCollapsed ? "" : "rotate-90"
                            }`}
                          />
                          <span className={`text-base font-medium ${C.text65}`}>
                            {header.name}
                          </span>
                          <span className={`text-base ${C.text40} ml-0.5`}>{items.length}</span>
                        </button>
                        <div className="flex-1" />
                        {canCreate ? (
                          <button
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleCreate(parentId);
                            }}
                            className={`${STYLE.iconBtn32} ${C.text40} ${C.hoverBgMedium} ${C.hoverText} opacity-0 group-hover/header:opacity-100`}
                          >
                            <Plus className={`${ICON.xs}`} />
                          </button>
                        ) : null}
                      </div>
                    </TableCell>

                    {/* Dosage form column — empty for category */}
                    <TableCell className="w-[100px] py-0" />

                    {/* Price column — empty */}
                    <TableCell className="w-[130px] py-0" />

                    {/* Status column — 有効 */}
                    <TableCell className="w-[110px] py-0 text-center">
                      <NotionStatusPill isActive={true} />
                    </TableCell>

                    {/* Actions column — category */}
                    <TableCell className="w-[80px] py-0 text-center" onClick={(e) => e.stopPropagation()}>
                      {canEdit ? (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <button
                              type="button"
                              className={`${STYLE.iconBtn28} ${C.text40} ${C.hoverBgMedium} ${C.hoverText} opacity-0 group-hover/header:opacity-100`}
                            >
                              <MoreHorizontal className={ICON.action} />
                            </button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => handleEdit(header)}>編集</DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      ) : null}
                    </TableCell>
                  </TableRow>

                  {/* Medicine rows — sortable within this group */}
                  {isCollapsed ? null : (
                    <SortableContext
                      items={items.map((m) => m.id)}
                      strategy={verticalListSortingStrategy}
                    >
                      {items.map((medicine) => (
                        <SortableMedicineRow
                          key={medicine.id}
                          medicine={medicine}
                          onEdit={handleEdit}
                          grouped
                          canEdit={canEdit}
                        />
                      ))}
                    </SortableContext>
                  )}
                </Fragment>
              );
            })}

            {/* Ungrouped medicines — flat rows, sortable */}
            <SortableContext
              items={ungroupedMedicines.map((m) => m.id)}
              strategy={verticalListSortingStrategy}
            >
              {ungroupedMedicines.map((medicine) => (
                <SortableMedicineRow
                  key={medicine.id}
                  medicine={medicine}
                  onEdit={handleEdit}
                  grouped={false}
                  canEdit={canEdit}
                />
              ))}
            </SortableContext>
          </TableBody>
        </Table>
          <DragOverlay dropAnimation={null}>
            {activeId ? (() => {
              const m = orderedMedicinesById.get(String(activeId));
              if (!m) return null;
              const isGrouped = Boolean(m.parentId);
              return <MedicineRowOverlay medicine={m} grouped={isGrouped} />;
            })() : null}
          </DragOverlay>
        </DndContext>
      </div>

      {/* Inline add row — BUG-157: create 権限で制御 */}
      {canCreate ? (
        <button
          type="button"
          onClick={() => handleCreate()}
          className={STYLE.inlineAddBtn}
        >
          <Plus className={`${ICON.xs}`} />
          新しい薬剤を追加...
        </button>
      ) : null}
    </div>
  );

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="薬剤マスタ"
            icon={<Pill className={`${ICON.page} ${C.text}`} />}
            resource={ResourceMasterMedical}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth="max-w-full"
            headerAction={
              canCreate ? (
                <PrimaryButton onClick={() => handleCreate()}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            <div className="flex flex-col gap-4">
              <NotionFilter
                properties={[MASTER_STATUS_FILTER]}
                activeFilters={activeFilters}
                onFilterChange={setActiveFilters}
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                searchPlaceholder="薬品名で検索..."
                count={totalCount}
              />
              {tableContent}
            </div>
          </PageLayout>
        </div>
        <MedicineSidePanel
          isEditing={isEditing}
          selectedMedicine={selectedMedicine}
          isCategory={isCategory}
          defaultParentId={defaultParentId}
          categoryMedicines={categoryMedicines}
          panelDuration={panelDuration}
          onCloseEdit={handleCloseEdit}
          onSave={handleSave}
          onDeleteRequest={handleDeleteRequest}
          readOnly={!canEdit}
          canDelete={canDelete}
          onDirtyChange={handleDirtyChange}
        />
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
