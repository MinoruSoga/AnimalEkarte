// React/Framework
import { useState, useMemo, useCallback } from "react";
import { flushSync } from "react-dom";
import { useNavigate } from "react-router";

// DnD
import {
  DndContext,
  DragOverlay,
  closestCenter,
  type DragEndEvent,
  type DragStartEvent,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import {
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
  useSortable,
  arrayMove,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";

// External
import { AnimatePresence, motion } from "motion/react";
import { toast } from "sonner";
import Pill from "lucide-react/dist/esm/icons/pill";
import Plus from "lucide-react/dist/esm/icons/plus";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import X from "lucide-react/dist/esm/icons/x";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import GripVertical from "lucide-react/dist/esm/icons/grip-vertical";
import MoreHorizontal from "lucide-react/dist/esm/icons/more-horizontal";
import Maximize2 from "lucide-react/dist/esm/icons/maximize-2";

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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import { useReducedMotion } from "@/hooks/useReducedMotion";

// Internal – feature API (direct import, no barrel)
import {
  useGetAllMedicines,
  useCreateMedicine,
  useUpdateMedicine,
  useDeleteMedicine,
  useReorderMedicines,
} from "../api/medicines";
import type { CreateMedicineRequest, UpdateMedicineRequest } from "@/types/medicine";

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
const SELECT_TRIGGER_FULL = `h-[30px] text-sm bg-transparent ${C.text} border-0 ${C.hoverBgLight} px-1.5 shadow-none rounded-[3px] w-full`;

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
}

const INITIAL_FORM: MedicineFormData = {
  name: "",
  parentId: "",
  dosageForm: "tablet",
  medicineUnit: "per_tablet",
  price: 0,
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
// SortableMedicineRow
// ─────────────────────────────────────────────────

function SortableMedicineRow({
  medicine,
  onEdit,
  grouped,
}: {
  medicine: Medicine;
  onEdit: (medicine: Medicine) => void;
  grouped: boolean;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: medicine.id });
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    // ドラッグ中は完全透明でプレースホルダーとして高さを維持
    opacity: isDragging ? 0 : 1,
  };
  return (
    <TableRow
      ref={setNodeRef}
      style={style}
      {...attributes}
      onClick={() => onEdit(medicine)}
      className={`${STYLE.tableRow} group/row`}
    >
      <TableCell
        className="w-8 px-0 py-0 pl-1 text-[#37352F]/20 group-hover/row:text-[#37352F]/50 transition-colors cursor-grab"
        {...listeners}
      >
        <GripVertical className="size-4" />
      </TableCell>
      <TableCell className={`${STYLE.tableCell} font-medium ${grouped ? "pl-12!" : "pl-2"}`}>
        {medicine.name}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} w-[130px] text-right pr-4 font-mono`}>
        {medicine.price > 0 ? `¥${medicine.price.toLocaleString()}` : "-"}
      </TableCell>
      <TableCell className="w-[110px] py-2 text-center">
        <StatusDot active={medicine.isActive} />
      </TableCell>
    </TableRow>
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
      className={`flex items-center h-12 bg-white border ${C.borderLight} rounded-[4px] shadow-[0_4px_16px_rgba(0,0,0,0.12)] cursor-grabbing`}
      style={{ width: "100%" }}
    >
      <div className="w-8 shrink-0 flex items-center justify-center text-[#37352F]/50">
        <GripVertical className="size-4" />
      </div>
      <div className={`flex-1 min-w-0 text-sm font-medium ${C.text} ${grouped ? "pl-10" : "pl-0"}`}>
        {medicine.name}
      </div>
      <div className="w-[130px] shrink-0 text-right pr-4 font-mono text-sm ${C.text}">
        {medicine.price > 0 ? `¥${medicine.price.toLocaleString()}` : "-"}
      </div>
      <div className="w-[110px] shrink-0 flex justify-center">
        <StatusDot active={medicine.isActive} />
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// Main component
// ─────────────────────────────────────────────────

export function MedicineSettings() {
  const navigate = useNavigate();
  const reduced = useReducedMotion();
  const panelDuration = reduced ? 0 : 0.2;

  // ── UI state ──
  const [searchTerm, setSearchTerm] = useState("");
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(new Set());
  const [isEditing, setIsEditing] = useState(false);
  const [selectedMedicine, setSelectedMedicine] = useState<Medicine | null>(null);
  const [formData, setFormData] = useState<MedicineFormData>(INITIAL_FORM);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [overrideOrder, setOverrideOrder] = useState<string[]>([]);
  // 値は parentId 文字列、undefined = 親なし
  const [overrideCategories, setOverrideCategories] = useState<Map<string, string | undefined>>(new Map());
  const [activeId, setActiveId] = useState<string | null>(null);

  // ── API ──
  const { data: medicines = [] } = useGetAllMedicines();
  const createMutation = useCreateMedicine();
  const updateMutation = useUpdateMedicine();
  const deleteMutation = useDeleteMedicine();
  const reorderMutation = useReorderMedicines();

  // ── DnD sensors ──
  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  // ── Derived: overrideOrder + overrideCategories 適用済みリスト ──
  const orderedMedicines = useMemo(() => {
    let result: Medicine[];
    if (overrideOrder.length === 0) {
      result = medicines;
    } else {
      const idx = new Map<string, number>(overrideOrder.map((id, i) => [id, i]));
      result = [...medicines].sort((a, b) => (idx.get(a.id) ?? 0) - (idx.get(b.id) ?? 0));
    }
    if (overrideCategories.size > 0) {
      result = result.map((m) =>
        overrideCategories.has(m.id)
          ? { ...m, parentId: overrideCategories.get(m.id) }
          : m,
      );
    }
    return result;
  }, [medicines, overrideOrder, overrideCategories]);

  // ── Derived: medicines ID → Medicine マップ (js-cache-function-results) ──
  const medicinesById = useMemo(
    () => new Map(medicines.map((m) => [m.id, m])),
    [medicines],
  );

  // ── Derived: カテゴリ medicine（parentId なし、price === 0）(js-cache-function-results) ──
  const categoryMedicines = useMemo(
    () => medicines.filter((m) => !m.parentId && m.price === 0),
    [medicines],
  );

  // ── Derived: filtered + grouped + ungrouped (js-cache-function-results) ──
  const { groupedMedicines, ungroupedMedicines, totalCount } = useMemo(() => {
    const lower = searchTerm.toLowerCase();
    const filtered = orderedMedicines.filter(
      (m) => !searchTerm || m.name.toLowerCase().includes(lower),
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
  }, [orderedMedicines, searchTerm, medicinesById]);

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

  const handleDragStart = useCallback((event: DragStartEvent) => {
    setActiveId(String(event.active.id));
  }, []);

  const handleDragCancel = useCallback(() => {
    setActiveId(null);
  }, []);

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      setActiveId(null);
      const { active, over } = event;
      if (!over || active.id === over.id) return;

      const activeItemId = String(active.id);
      const overItemId = String(over.id);

      const activeMedicine = orderedMedicines.find((m) => m.id === activeItemId);
      const overMedicine = orderedMedicines.find((m) => m.id === overItemId);
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
            onError: () => {
              toast.error("カテゴリの変更に失敗しました");
              clearOptimistic();
            },
          },
        );
      } else {
        // 同カテゴリ: 並び替え
        const currentIds = orderedMedicines.map((m) => m.id);
        const newOrder = arrayMove(
          currentIds,
          currentIds.indexOf(activeItemId),
          currentIds.indexOf(overItemId),
        );
        flushSync(() => {
          setOverrideOrder(newOrder);
        });
        reorderMutation.mutate(
          { ids: newOrder.map(Number) },
          { onSuccess: () => setOverrideOrder([]) },
        );
      }
    },
    [orderedMedicines, reorderMutation, updateMutation],
  );

  const handleCloseEdit = useCallback(() => {
    setIsEditing(false);
    setSelectedMedicine(null);
    setFormData(INITIAL_FORM);
  }, []);

  const handleEdit = useCallback((medicine: Medicine) => {
    setSelectedMedicine(medicine);
    setFormData({
      name: medicine.name,
      parentId: medicine.parentId ?? "",
      dosageForm: medicine.dosageForm ?? "",
      medicineUnit: medicine.medicineUnit ?? "",
      price: medicine.price,
      description: medicine.description,
      isActive: medicine.isActive,
    });
    setIsEditing(true);
  }, []);

  const handleCreate = useCallback((parentId?: string) => {
    setSelectedMedicine(null);
    setFormData({
      ...INITIAL_FORM,
      parentId: parentId !== "uncategorized" ? (parentId ?? "") : "",
    });
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
        dosage_form: formData.dosageForm || undefined,
        medicine_unit: formData.medicineUnit || undefined,
        price: formData.price,
        description: formData.description,
        is_active: formData.isActive,
      };
      // parent_id の処理
      if (formData.parentId) {
        req.parent_id = Number(formData.parentId);
      } else if (selectedMedicine.parentId) {
        // 元々グループに属していたが今は外す
        req.clear_parent_id = true;
      }
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
        dosage_form: formData.dosageForm || undefined,
        medicine_unit: formData.medicineUnit || undefined,
        price: formData.price,
        description: formData.description,
        is_active: formData.isActive,
        ...(formData.parentId ? { parent_id: Number(formData.parentId) } : {}),
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
              <TableHead className={`${STYLE.tableHeaderCell} w-[130px] text-right pr-4`}>
                単価(税込)
              </TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} w-[110px] text-center`}>
                ステータス
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {groupedMedicines.size === 0 && ungroupedMedicines.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className={STYLE.tableEmpty}>
                  データが見つかりません
                </TableCell>
              </TableRow>
            ) : null}

            {Array.from(groupedMedicines.entries()).map(([parentId, { header, items }]) => {
              const isCollapsed = collapsedGroups.has(parentId);

              return (
                <>
                  {/* Group header row */}
                  <TableRow
                    key={`h-${parentId}`}
                    className={`border-b ${C.borderLight} bg-[#F7F6F3]/30 h-[49px] group/header hover:bg-[#F7F6F3]/60`}
                  >
                    {/* Grip handle — left */}
                    <TableCell className="w-8 px-0 py-0">
                      <button
                        type="button"
                        tabIndex={-1}
                        className="w-8 h-8 flex items-center justify-center rounded-[3px] text-[#37352F]/20 hover:bg-[rgba(55,53,47,0.08)] hover:text-[#37352F]/50 transition-colors cursor-grab"
                      >
                        <GripVertical className="size-4" />
                      </button>
                    </TableCell>

                    {/* Chevron + label + count + plus button */}
                    <TableCell className="py-0 pl-0 pr-2">
                      <div className="flex items-center">
                        <button
                          type="button"
                          onClick={() => toggleGroup(parentId)}
                          className="flex flex-1 items-center gap-1.5 py-1.5 px-1 hover:bg-[rgba(55,53,47,0.04)] rounded-[3px] transition-colors"
                        >
                          <ChevronRight
                            className={`size-3.5 text-[#37352F]/50 transition-transform duration-150 ${
                              isCollapsed ? "" : "rotate-90"
                            }`}
                          />
                          <span className="text-xs font-medium text-[#37352F]/65">
                            {header.name}
                          </span>
                          <span className="text-xs text-[#37352F]/40 ml-0.5">{items.length}</span>
                        </button>
                        <button
                          type="button"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleCreate(parentId);
                          }}
                          className="w-8 h-8 flex items-center justify-center rounded-[3px] text-[#37352F]/40 hover:bg-[rgba(55,53,47,0.08)] hover:text-[#37352F] transition-colors opacity-0 group-hover/header:opacity-100"
                        >
                          <Plus className="size-3.5" />
                        </button>
                      </div>
                    </TableCell>

                    {/* Price column — empty */}
                    <TableCell className="w-[130px] py-0" />

                    {/* Status column — 有効 */}
                    <TableCell className="w-[110px] py-0 text-center">
                      <StatusDot active={true} />
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
                        />
                      ))}
                    </SortableContext>
                  )}
                </>
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
                />
              ))}
            </SortableContext>
          </TableBody>
        </Table>
          <DragOverlay dropAnimation={null}>
            {activeId ? (() => {
              const m = orderedMedicines.find((x) => x.id === activeId);
              if (!m) return null;
              const isGrouped = Boolean(m.parentId);
              return <MedicineRowOverlay medicine={m} grouped={isGrouped} />;
            })() : null}
          </DragOverlay>
        </DndContext>
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
  const sidePeekPanel = (
    <AnimatePresence>
      {isEditing ? (
        <motion.div
          key="side-peek"
          initial={{ width: 0, opacity: 0 }}
          animate={{ width: 680, opacity: 1 }}
          exit={{ width: 0, opacity: 0 }}
          transition={{ duration: panelDuration, ease: [0.25, 0.1, 0.25, 1] }}
          className="shrink-0 min-h-0 overflow-hidden"
        >
        <div
          className={`flex flex-col h-full w-[680px] bg-white border-l ${C.borderLight} shadow-[-1px_0_5px_rgba(0,0,0,0.02)]`}
        >
          {/* Toolbar */}
          <div className="flex items-center justify-between h-[48px] px-3 shrink-0">
            <span className={`text-xs ${C.text35} pl-1 select-none`}>
              {selectedMedicine ? "編集" : "新規作成"}
            </span>
            <div className="flex items-center gap-1">
              {/* Expand button */}
              <button
                type="button"
                onClick={() => {}}
                className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
                aria-label="全画面で開く"
              >
                <Maximize2 className="size-4" />
              </button>

              {/* 3-dots dropdown */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button
                    type="button"
                    className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
                    aria-label="その他の操作"
                  >
                    <MoreHorizontal className="size-4" />
                  </button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {selectedMedicine ? (
                    <DropdownMenuItem
                      onClick={handleDeleteRequest}
                      className="text-red-600 focus:text-red-600"
                    >
                      <Trash2 className="size-4 mr-2" />
                      削除
                    </DropdownMenuItem>
                  ) : null}
                </DropdownMenuContent>
              </DropdownMenu>

              {/* Close button */}
              <button
                type="button"
                onClick={handleCloseEdit}
                className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
                aria-label="閉じる"
              >
                <X className="size-4" />
              </button>
            </div>
          </div>

          {/* Body */}
          <div className="flex-1 overflow-y-auto">
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
                {/* Parent category select */}
                <PropertyRow label="親カテゴリ">
                  <select
                    value={formData.parentId}
                    onChange={(e) => updateForm({ parentId: e.target.value })}
                    className={SELECT_TRIGGER_FULL}
                  >
                    <option value="">なし（未分類）</option>
                    {categoryMedicines.map((cat) => (
                      <option key={cat.id} value={cat.id}>{cat.name}</option>
                    ))}
                  </select>
                </PropertyRow>

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
                <div className="flex items-center gap-1.5 py-2 mb-1">
                  <Pill className={`size-3.5 ${C.text40}`} />
                  <span
                    className={`text-xs font-medium ${C.text50} uppercase tracking-wide select-none`}
                  >
                    薬剤詳細
                  </span>
                </div>

                {/* Dosage form — full-width */}
                <PropertyRow label="剤形" required>
                  <Select
                    value={formData.dosageForm}
                    onValueChange={(v) => updateForm({ dosageForm: v })}
                  >
                    <SelectTrigger className={SELECT_TRIGGER_FULL}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>{DOSAGE_FORM_SELECT_ITEMS}</SelectContent>
                  </Select>
                </PropertyRow>

                {/* Unit — full-width */}
                <PropertyRow label="単位">
                  <Select
                    value={formData.medicineUnit}
                    onValueChange={(v) => updateForm({ medicineUnit: v })}
                  >
                    <SelectTrigger className={SELECT_TRIGGER_FULL}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>{MEDICINE_UNIT_SELECT_ITEMS}</SelectContent>
                  </Select>
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
            <button type="button" onClick={handleSave} className={STYLE.sidePeekSaveBtn}>
              保存
            </button>
          </div>
        </div>
        </motion.div>
      ) : null}
    </AnimatePresence>
  );

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
