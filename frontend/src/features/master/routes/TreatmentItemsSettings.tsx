// React/Framework
import { useState, useMemo, useCallback } from "react";
import { flushSync } from "react-dom";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";

// DnD
import {
  DndContext,
  DragOverlay,
  closestCenter,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";

// External
import { AnimatePresence, motion } from "motion/react";
import { toast } from "sonner";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
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
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { SidePeekPanel } from "@/components/shared/SidePeek/SidePeekPanel";
import { SidePeekBody } from "@/components/shared/SidePeek/SidePeekBody";
import { SidePeekTitleInput } from "@/components/shared/SidePeek/SidePeekTitleInput";
import { SidePeekFooter } from "@/components/shared/SidePeek/SidePeekFooter";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { useSortableList } from "@/hooks/useSortableList";

// Internal – feature API (direct import, no barrel)
import {
  useGetAllConsultations,
  useCreateConsultation,
  useUpdateConsultation,
  useDeleteConsultation,
  useReorderConsultations,
} from "../api/consultations";
import {
  useGetAllExaminationTypes,
  useCreateExaminationType,
  useUpdateExaminationType,
  useDeleteExaminationType,
  useReorderExaminationTypes,
} from "../api/exam-types-master";
import {
  useGetAllProcedures,
  useCreateProcedure,
  useUpdateProcedure,
  useDeleteProcedure,
  useReorderProcedures,
} from "../api/procedures";
import {
  useGetAllVaccinesMaster,
  useCreateVaccineMaster,
  useUpdateVaccineMaster,
  useDeleteVaccineMaster,
  useReorderVaccinesMaster,
} from "../api/vaccines-master";
import {
  useGetAllCheckupTypes,
  useCreateCheckupType,
  useUpdateCheckupType,
  useDeleteCheckupType,
  useReorderCheckupTypes,
} from "../api/checkup-types";

// Types
import type { TreatmentItem } from "@/lib/transforms/treatment";
import type {
  CreateConsultationRequest,
  UpdateConsultationRequest,
  CreateExaminationTypeRequest,
  UpdateExaminationTypeRequest,
  CreateProcedureRequest,
  UpdateProcedureRequest,
  CreateVaccineRequest,
  UpdateVaccineRequest,
  CreateCheckupTypeRequest,
  UpdateCheckupTypeRequest,
} from "@/types/treatment";

// ─────────────────────────────────────────────────
// Module-level constants
// ─────────────────────────────────────────────────

const TABS = [
  { value: "consultation", label: "診察" },
  { value: "examination", label: "検査" },
  { value: "procedure", label: "処置" },
  { value: "vaccine", label: "予防接種" },
  { value: "checkup", label: "定期健診" },
] as const;

type TabValue = (typeof TABS)[number]["value"];

// Full-width select trigger style
const SELECT_TRIGGER_FULL = `h-[30px] text-sm bg-transparent ${C.text} border-0 ${C.hoverBgLight} px-1.5 shadow-none rounded-[3px] w-full`;

// ─────────────────────────────────────────────────
// Form data type (共通)
// ─────────────────────────────────────────────────

interface TreatmentFormData {
  name: string;
  parentId: string;
  price: number;
  isActive: boolean;
  description: string;
}

const INITIAL_FORM: TreatmentFormData = {
  name: "",
  parentId: "",
  price: 0,
  isActive: true,
  description: "",
};

// ─────────────────────────────────────────────────
// TreatmentTabContent props
// ─────────────────────────────────────────────────

interface TreatmentTabContentProps {
  items: TreatmentItem[];
  entityLabel: string;
  entityIcon: React.ReactNode;
  onCreate: (req: {
    name: string;
    parent_id?: number;
    price?: number;
    is_active?: boolean;
    description?: string;
  }) => void;
  onUpdate: (
    id: string,
    req: {
      name?: string;
      parent_id?: number;
      clear_parent_id?: boolean;
      price?: number;
      is_active?: boolean;
      description?: string;
    },
  ) => void;
  onDelete: (id: string) => void;
  onReorder: (newIds: string[]) => void;
  onCrossGroupUpdate: (
    activeId: string,
    req: { parent_id?: number; clear_parent_id?: boolean },
  ) => void;
  isCreatePending: boolean;
  isUpdatePending: boolean;
  isDeletePending: boolean;
}

// ─────────────────────────────────────────────────
// Sub-components
// ─────────────────────────────────────────────────

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
// TreatmentItemRow
// ─────────────────────────────────────────────────

function TreatmentItemRow({
  item,
  onEdit,
  grouped,
}: {
  item: TreatmentItem;
  onEdit: (item: TreatmentItem) => void;
  grouped: boolean;
}) {
  return (
    <SortableDataTableRow
      id={item.id}
      onClick={() => onEdit(item)}
      className="group/row"
      gripClassName="w-8 px-0 py-0 pl-1 text-[#37352F]/20 group-hover/row:text-[#37352F]/50 transition-colors cursor-grab"
      isDraggingOpacity={0}
    >
      <TableCell className={`${STYLE.tableCell} font-medium ${grouped ? "pl-12!" : "pl-2"}`}>
        {item.name}
      </TableCell>
      <TableCell className={`${STYLE.tableCell} w-[130px] text-right pr-4 font-mono`}>
        {item.price > 0 ? `¥${item.price.toLocaleString()}` : "-"}
      </TableCell>
      <TableCell className="w-[110px] py-2.5 text-center">
        <StatusDot active={item.isActive} />
      </TableCell>
    </SortableDataTableRow>
  );
}

// ─────────────────────────────────────────────────
// TreatmentRowOverlay — DragOverlay 用
// ─────────────────────────────────────────────────

function TreatmentRowOverlay({
  item,
  grouped,
}: {
  item: TreatmentItem;
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
      <div
        className={`flex-1 min-w-0 text-sm font-medium ${C.text} ${grouped ? "pl-10" : "pl-0"}`}
      >
        {item.name}
      </div>
      <div className="w-[130px] shrink-0 text-right pr-4 font-mono text-sm">
        {item.price > 0 ? `¥${item.price.toLocaleString()}` : "-"}
      </div>
      <div className="w-[110px] shrink-0 flex justify-center">
        <StatusDot active={item.isActive} />
      </div>
    </div>
  );
}

// ─────────────────────────────────────────────────
// TreatmentTabContent — 各タブの共通UI
// ─────────────────────────────────────────────────

function TreatmentTabContent({
  items,
  entityLabel,
  entityIcon,
  onCreate,
  onUpdate,
  onDelete,
  onReorder,
  onCrossGroupUpdate,
  isCreatePending: _isCreatePending,
  isUpdatePending: _isUpdatePending,
  isDeletePending: _isDeletePending,
}: TreatmentTabContentProps) {
  const reduced = useReducedMotion();
  const panelDuration = reduced ? 0 : 0.2;

  // ── UI state ──
  const [searchTerm, setSearchTerm] = useState("");
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(new Set());
  const [isEditing, setIsEditing] = useState(false);
  const [selectedItem, setSelectedItem] = useState<TreatmentItem | null>(null);
  const [formData, setFormData] = useState<TreatmentFormData>(INITIAL_FORM);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [overrideCategories, setOverrideCategories] = useState<
    Map<string, string | undefined>
  >(new Map());

  // ── DnD ──
  const {
    orderedItems: sortedItems,
    sensors,
    activeId,
    handleDragStart,
    handleDragCancel,
    handleDragEnd: handleFlatSortDragEnd,
    resetOrder,
  } = useSortableList({
    items,
    onReorder: (newIds) => {
      onReorder(newIds);
      resetOrder();
    },
  });

  // ── Derived: overrideCategories 適用済みリスト ──
  const orderedItems = useMemo(() => {
    if (overrideCategories.size === 0) return sortedItems;
    return sortedItems.map((m) =>
      overrideCategories.has(m.id) ? { ...m, parentId: overrideCategories.get(m.id) } : m,
    );
  }, [sortedItems, overrideCategories]);

  // ── Derived: ID → item マップ ──
  const itemsById = useMemo(() => new Map(items.map((m) => [m.id, m])), [items]);

  // ── Derived: カテゴリアイテム（parentId なし、price === 0）──
  const categoryItems = useMemo(
    () => items.filter((m) => !m.parentId && m.price === 0),
    [items],
  );

  // ── Derived: filtered + grouped + ungrouped ──
  const { groupedItems, ungroupedItems, totalCount } = useMemo(() => {
    const lower = searchTerm.toLowerCase();
    const filtered = orderedItems.filter(
      (m) => !searchTerm || m.name.toLowerCase().includes(lower),
    );

    const groups = new Map<string, { header: TreatmentItem; items: TreatmentItem[] }>();
    const ungrouped: TreatmentItem[] = [];

    for (const m of filtered) {
      if (m.parentId) {
        const existing = groups.get(m.parentId);
        if (existing) {
          existing.items.push(m);
        } else {
          const parent = itemsById.get(m.parentId);
          if (parent) {
            groups.set(m.parentId, { header: parent, items: [m] });
          } else {
            ungrouped.push(m);
          }
        }
      } else if (m.price === 0) {
        if (!groups.has(m.id)) {
          groups.set(m.id, { header: m, items: [] });
        }
      } else {
        ungrouped.push(m);
      }
    }

    return { groupedItems: groups, ungroupedItems: ungrouped, totalCount: filtered.length };
  }, [orderedItems, searchTerm, itemsById]);

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

      const activeItem = orderedItems.find((m) => m.id === activeItemId);
      const overItem = orderedItems.find((m) => m.id === overItemId);
      if (!activeItem || !overItem) return;

      const activeCat = activeItem.parentId ?? null;
      const overCat = overItem.parentId ?? null;

      if (activeCat !== overCat) {
        // クロスグループ: parent_id を変更
        flushSync(() => {
          setOverrideCategories((prev) => {
            const next = new Map(prev);
            next.set(activeItemId, overCat ?? undefined);
            return next;
          });
        });

        const req: { parent_id?: number; clear_parent_id?: boolean } = overCat
          ? { parent_id: Number(overCat) }
          : { clear_parent_id: true };

        const clearOptimistic = () => {
          setOverrideCategories((prev) => {
            const next = new Map(prev);
            next.delete(activeItemId);
            return next;
          });
        };

        onCrossGroupUpdate(activeItemId, req);
        clearOptimistic();
      } else {
        handleFlatSortDragEnd(event);
      }
    },
    [orderedItems, onCrossGroupUpdate, handleFlatSortDragEnd],
  );

  const handleCloseEdit = useCallback(() => {
    setIsEditing(false);
    setSelectedItem(null);
    setFormData(INITIAL_FORM);
  }, []);

  const handleEdit = useCallback((item: TreatmentItem) => {
    setSelectedItem(item);
    setFormData({
      name: item.name,
      parentId: item.parentId ?? "",
      price: item.price,
      description: item.description,
      isActive: item.isActive,
    });
    setIsEditing(true);
  }, []);

  const handleCreate = useCallback((parentId?: string) => {
    setSelectedItem(null);
    setFormData({
      ...INITIAL_FORM,
      parentId: parentId !== "uncategorized" ? (parentId ?? "") : "",
    });
    setIsEditing(true);
  }, []);

  const updateForm = useCallback((updates: Partial<TreatmentFormData>) => {
    setFormData((prev) => ({ ...prev, ...updates }));
  }, []);

  const handleSave = useCallback(() => {
    if (!formData.name.trim()) {
      toast.error(`${entityLabel}名は必須です`);
      return;
    }

    if (selectedItem) {
      const req: {
        name?: string;
        parent_id?: number;
        clear_parent_id?: boolean;
        price?: number;
        is_active?: boolean;
        description?: string;
      } = {
        name: formData.name,
        price: formData.price,
        description: formData.description,
        is_active: formData.isActive,
      };

      if (formData.parentId) {
        req.parent_id = Number(formData.parentId);
      } else if (selectedItem.parentId) {
        req.clear_parent_id = true;
      }

      onUpdate(selectedItem.id, req);
      toast.success("更新しました");
      handleCloseEdit();
    } else {
      const req: {
        name: string;
        parent_id?: number;
        price?: number;
        is_active?: boolean;
        description?: string;
      } = {
        name: formData.name,
        price: formData.price,
        description: formData.description,
        is_active: formData.isActive,
        ...(formData.parentId ? { parent_id: Number(formData.parentId) } : {}),
      };

      onCreate(req);
      toast.success("登録しました");
      handleCloseEdit();
    }
  }, [formData, selectedItem, onUpdate, onCreate, handleCloseEdit, entityLabel]);

  const handleDeleteRequest = useCallback(() => {
    if (!selectedItem) return;
    setDeleteConfirmOpen(true);
  }, [selectedItem]);

  const executeDelete = useCallback(() => {
    if (!selectedItem) return;
    onDelete(selectedItem.id);
    toast.success("削除しました");
    setDeleteConfirmOpen(false);
    handleCloseEdit();
  }, [selectedItem, onDelete, handleCloseEdit]);

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
                <TableHead className={`${STYLE.tableHeaderCell} pl-3`}>
                  {entityLabel}名
                </TableHead>
                <TableHead className={`${STYLE.tableHeaderCell} w-[130px] text-right pr-4`}>
                  単価(税込)
                </TableHead>
                <TableHead className={`${STYLE.tableHeaderCell} w-[110px] text-center`}>
                  ステータス
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {groupedItems.size === 0 && ungroupedItems.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className={STYLE.tableEmpty}>
                    データが見つかりません
                  </TableCell>
                </TableRow>
              ) : null}

              {Array.from(groupedItems.entries()).map(([parentId, { header, items: groupRows }]) => {
                const isCollapsed = collapsedGroups.has(parentId);
                return (
                  <>
                    {/* Group header row */}
                    <TableRow
                      key={`h-${parentId}`}
                      className={`border-b ${C.borderLight} bg-[#F7F6F3]/30 h-12 group/header hover:bg-[#F7F6F3]/60`}
                    >
                      <TableCell className="w-8 px-0 py-0">
                        <button
                          type="button"
                          tabIndex={-1}
                          className="w-8 h-8 flex items-center justify-center rounded-[3px] text-[#37352F]/20 hover:bg-[rgba(55,53,47,0.08)] hover:text-[#37352F]/50 transition-colors cursor-grab"
                        >
                          <GripVertical className="size-4" />
                        </button>
                      </TableCell>

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
                            <span className="text-xs text-[#37352F]/40 ml-0.5">
                              {groupRows.length}
                            </span>
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

                      <TableCell className="w-[130px] py-0" />
                      <TableCell className="w-[110px] py-0 text-center">
                        <StatusDot active={true} />
                      </TableCell>
                    </TableRow>

                    {/* Group item rows */}
                    {isCollapsed ? null : (
                      <SortableContext
                        items={groupRows.map((m) => m.id)}
                        strategy={verticalListSortingStrategy}
                      >
                        {groupRows.map((item) => (
                          <TreatmentItemRow
                            key={item.id}
                            item={item}
                            onEdit={handleEdit}
                            grouped
                          />
                        ))}
                      </SortableContext>
                    )}
                  </>
                );
              })}

              {/* Ungrouped items */}
              <SortableContext
                items={ungroupedItems.map((m) => m.id)}
                strategy={verticalListSortingStrategy}
              >
                {ungroupedItems.map((item) => (
                  <TreatmentItemRow
                    key={item.id}
                    item={item}
                    onEdit={handleEdit}
                    grouped={false}
                  />
                ))}
              </SortableContext>
            </TableBody>
          </Table>

          <DragOverlay dropAnimation={null}>
            {activeId ? (() => {
              const m = orderedItems.find((x) => x.id === activeId);
              if (!m) return null;
              return <TreatmentRowOverlay item={m} grouped={Boolean(m.parentId)} />;
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
        新しい{entityLabel}を追加...
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
          <SidePeekPanel>
            {/* Toolbar */}
            <div className="flex items-center justify-between h-[48px] px-3 shrink-0">
              <span className={`text-xs ${C.text35} pl-1 select-none`}>
                {selectedItem ? "編集" : "新規作成"}
              </span>
              <div className="flex items-center gap-1">
                <button
                  type="button"
                  onClick={() => {}}
                  className={`${STYLE.sidePeekToolbarBtn} cursor-pointer`}
                  aria-label="全画面で開く"
                >
                  <Maximize2 className="size-4" />
                </button>

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
                    {selectedItem ? (
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

            <SidePeekBody>
              {/* Page icon */}
              <div className="pt-4 pb-2">
                <div className={STYLE.pageIcon}>
                  <span className={LAYOUT.pageIcon.innerIcon}>{entityIcon}</span>
                </div>
              </div>

              <SidePeekTitleInput
                value={formData.name}
                onChange={(v) => updateForm({ name: v })}
                placeholder={`${entityLabel}名`}
                autoFocus={false}
              />

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
                    {categoryItems.map((cat) => (
                      <option key={cat.id} value={cat.id}>
                        {cat.name}
                      </option>
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
                    <NotionStatusPill isActive={formData.isActive} />
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
            </SidePeekBody>

            <SidePeekFooter
              onCancel={handleCloseEdit}
              onSave={handleSave}
            />
          </SidePeekPanel>
        </motion.div>
      ) : null}
    </AnimatePresence>
  );

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <div className="flex flex-col gap-4">
            <div className="flex items-center gap-3">
              <div className="flex-1 min-w-0">
                <SearchFilterBar
                  searchTerm={searchTerm}
                  onSearchChange={setSearchTerm}
                  placeholder={`${entityLabel}名で検索...`}
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
        </div>
        {sidePeekPanel}
      </div>

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
    </>
  );
}

// ─────────────────────────────────────────────────
// Tab sub-components — 各タブが独自の hooks を呼ぶ
// ─────────────────────────────────────────────────

function ConsultationTab() {
  const { data: items = [] } = useGetAllConsultations();
  const createMutation = useCreateConsultation();
  const updateMutation = useUpdateConsultation();
  const deleteMutation = useDeleteConsultation();
  const reorderMutation = useReorderConsultations();

  const handleCreate = useCallback(
    (req: CreateConsultationRequest) => {
      createMutation.mutate(req, {
        onError: () => toast.error("登録に失敗しました"),
      });
    },
    [createMutation],
  );

  const handleUpdate = useCallback(
    (id: string, req: UpdateConsultationRequest) => {
      updateMutation.mutate(
        { id, req },
        { onError: () => toast.error("更新に失敗しました") },
      );
    },
    [updateMutation],
  );

  const handleDelete = useCallback(
    (id: string) => {
      deleteMutation.mutate(id, {
        onError: () => toast.error("削除に失敗しました"),
      });
    },
    [deleteMutation],
  );

  const handleReorder = useCallback(
    (newIds: string[]) => {
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        { onError: () => toast.error("並び替えに失敗しました") },
      );
    },
    [reorderMutation],
  );

  const handleCrossGroupUpdate = useCallback(
    (activeId: string, req: { parent_id?: number; clear_parent_id?: boolean }) => {
      updateMutation.mutate(
        { id: activeId, req },
        { onError: () => toast.error("カテゴリの変更に失敗しました") },
      );
    },
    [updateMutation],
  );

  return (
    <TreatmentTabContent
      items={items}
      entityLabel="診察"
      entityIcon={<Stethoscope className="size-5 text-[#37352F]" />}
      onCreate={handleCreate}
      onUpdate={handleUpdate}
      onDelete={handleDelete}
      onReorder={handleReorder}
      onCrossGroupUpdate={handleCrossGroupUpdate}
      isCreatePending={createMutation.isPending}
      isUpdatePending={updateMutation.isPending}
      isDeletePending={deleteMutation.isPending}
    />
  );
}

function ExaminationTab() {
  const { data: items = [] } = useGetAllExaminationTypes();
  const createMutation = useCreateExaminationType();
  const updateMutation = useUpdateExaminationType();
  const deleteMutation = useDeleteExaminationType();
  const reorderMutation = useReorderExaminationTypes();

  const handleCreate = useCallback(
    (req: CreateExaminationTypeRequest) => {
      createMutation.mutate(req, {
        onError: () => toast.error("登録に失敗しました"),
      });
    },
    [createMutation],
  );

  const handleUpdate = useCallback(
    (id: string, req: UpdateExaminationTypeRequest) => {
      updateMutation.mutate(
        { id, req },
        { onError: () => toast.error("更新に失敗しました") },
      );
    },
    [updateMutation],
  );

  const handleDelete = useCallback(
    (id: string) => {
      deleteMutation.mutate(id, {
        onError: () => toast.error("削除に失敗しました"),
      });
    },
    [deleteMutation],
  );

  const handleReorder = useCallback(
    (newIds: string[]) => {
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        { onError: () => toast.error("並び替えに失敗しました") },
      );
    },
    [reorderMutation],
  );

  const handleCrossGroupUpdate = useCallback(
    (activeId: string, req: { parent_id?: number; clear_parent_id?: boolean }) => {
      updateMutation.mutate(
        { id: activeId, req },
        { onError: () => toast.error("カテゴリの変更に失敗しました") },
      );
    },
    [updateMutation],
  );

  return (
    <TreatmentTabContent
      items={items}
      entityLabel="検査"
      entityIcon={<Stethoscope className="size-5 text-[#37352F]" />}
      onCreate={handleCreate}
      onUpdate={handleUpdate}
      onDelete={handleDelete}
      onReorder={handleReorder}
      onCrossGroupUpdate={handleCrossGroupUpdate}
      isCreatePending={createMutation.isPending}
      isUpdatePending={updateMutation.isPending}
      isDeletePending={deleteMutation.isPending}
    />
  );
}

function ProcedureTab() {
  const { data: items = [] } = useGetAllProcedures();
  const createMutation = useCreateProcedure();
  const updateMutation = useUpdateProcedure();
  const deleteMutation = useDeleteProcedure();
  const reorderMutation = useReorderProcedures();

  const handleCreate = useCallback(
    (req: CreateProcedureRequest) => {
      createMutation.mutate(req, {
        onError: () => toast.error("登録に失敗しました"),
      });
    },
    [createMutation],
  );

  const handleUpdate = useCallback(
    (id: string, req: UpdateProcedureRequest) => {
      updateMutation.mutate(
        { id, req },
        { onError: () => toast.error("更新に失敗しました") },
      );
    },
    [updateMutation],
  );

  const handleDelete = useCallback(
    (id: string) => {
      deleteMutation.mutate(id, {
        onError: () => toast.error("削除に失敗しました"),
      });
    },
    [deleteMutation],
  );

  const handleReorder = useCallback(
    (newIds: string[]) => {
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        { onError: () => toast.error("並び替えに失敗しました") },
      );
    },
    [reorderMutation],
  );

  const handleCrossGroupUpdate = useCallback(
    (activeId: string, req: { parent_id?: number; clear_parent_id?: boolean }) => {
      updateMutation.mutate(
        { id: activeId, req },
        { onError: () => toast.error("カテゴリの変更に失敗しました") },
      );
    },
    [updateMutation],
  );

  return (
    <TreatmentTabContent
      items={items}
      entityLabel="処置"
      entityIcon={<Stethoscope className="size-5 text-[#37352F]" />}
      onCreate={handleCreate}
      onUpdate={handleUpdate}
      onDelete={handleDelete}
      onReorder={handleReorder}
      onCrossGroupUpdate={handleCrossGroupUpdate}
      isCreatePending={createMutation.isPending}
      isUpdatePending={updateMutation.isPending}
      isDeletePending={deleteMutation.isPending}
    />
  );
}

function VaccineTab() {
  const { data: items = [] } = useGetAllVaccinesMaster();
  const createMutation = useCreateVaccineMaster();
  const updateMutation = useUpdateVaccineMaster();
  const deleteMutation = useDeleteVaccineMaster();
  const reorderMutation = useReorderVaccinesMaster();

  const handleCreate = useCallback(
    (req: CreateVaccineRequest) => {
      createMutation.mutate(req, {
        onError: () => toast.error("登録に失敗しました"),
      });
    },
    [createMutation],
  );

  const handleUpdate = useCallback(
    (id: string, req: UpdateVaccineRequest) => {
      updateMutation.mutate(
        { id, req },
        { onError: () => toast.error("更新に失敗しました") },
      );
    },
    [updateMutation],
  );

  const handleDelete = useCallback(
    (id: string) => {
      deleteMutation.mutate(id, {
        onError: () => toast.error("削除に失敗しました"),
      });
    },
    [deleteMutation],
  );

  const handleReorder = useCallback(
    (newIds: string[]) => {
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        { onError: () => toast.error("並び替えに失敗しました") },
      );
    },
    [reorderMutation],
  );

  const handleCrossGroupUpdate = useCallback(
    (activeId: string, req: { parent_id?: number; clear_parent_id?: boolean }) => {
      updateMutation.mutate(
        { id: activeId, req },
        { onError: () => toast.error("カテゴリの変更に失敗しました") },
      );
    },
    [updateMutation],
  );

  return (
    <TreatmentTabContent
      items={items}
      entityLabel="予防接種"
      entityIcon={<Stethoscope className="size-5 text-[#37352F]" />}
      onCreate={handleCreate}
      onUpdate={handleUpdate}
      onDelete={handleDelete}
      onReorder={handleReorder}
      onCrossGroupUpdate={handleCrossGroupUpdate}
      isCreatePending={createMutation.isPending}
      isUpdatePending={updateMutation.isPending}
      isDeletePending={deleteMutation.isPending}
    />
  );
}

function CheckupTab() {
  const { data: items = [] } = useGetAllCheckupTypes();
  const createMutation = useCreateCheckupType();
  const updateMutation = useUpdateCheckupType();
  const deleteMutation = useDeleteCheckupType();
  const reorderMutation = useReorderCheckupTypes();

  const handleCreate = useCallback(
    (req: CreateCheckupTypeRequest) => {
      createMutation.mutate(req, {
        onError: () => toast.error("登録に失敗しました"),
      });
    },
    [createMutation],
  );

  const handleUpdate = useCallback(
    (id: string, req: UpdateCheckupTypeRequest) => {
      updateMutation.mutate(
        { id, req },
        { onError: () => toast.error("更新に失敗しました") },
      );
    },
    [updateMutation],
  );

  const handleDelete = useCallback(
    (id: string) => {
      deleteMutation.mutate(id, {
        onError: () => toast.error("削除に失敗しました"),
      });
    },
    [deleteMutation],
  );

  const handleReorder = useCallback(
    (newIds: string[]) => {
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        { onError: () => toast.error("並び替えに失敗しました") },
      );
    },
    [reorderMutation],
  );

  const handleCrossGroupUpdate = useCallback(
    (activeId: string, req: { parent_id?: number; clear_parent_id?: boolean }) => {
      updateMutation.mutate(
        { id: activeId, req },
        { onError: () => toast.error("カテゴリの変更に失敗しました") },
      );
    },
    [updateMutation],
  );

  return (
    <TreatmentTabContent
      items={items}
      entityLabel="定期健診"
      entityIcon={<Stethoscope className="size-5 text-[#37352F]" />}
      onCreate={handleCreate}
      onUpdate={handleUpdate}
      onDelete={handleDelete}
      onReorder={handleReorder}
      onCrossGroupUpdate={handleCrossGroupUpdate}
      isCreatePending={createMutation.isPending}
      isUpdatePending={updateMutation.isPending}
      isDeletePending={deleteMutation.isPending}
    />
  );
}

// ─────────────────────────────────────────────────
// Main component
// ─────────────────────────────────────────────────

export function TreatmentItemsSettings() {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState<TabValue>("consultation");

  return (
    <PageLayout
      title="診療項目マスタ"
      icon={<Stethoscope className="size-5 text-[#37352F]" />}
      onBack={() => navigate(paths.settings.getHref())}
      maxWidth="max-w-full"
    >
      <div className="flex flex-col gap-4">
        <Tabs
          value={activeTab}
          onValueChange={(v) => setActiveTab(v as TabValue)}
        >
          <TabsList
            className={`h-9 bg-transparent border-b ${C.borderLight} rounded-none w-full justify-start gap-0 p-0`}
          >
            {TABS.map((tab) => (
              <TabsTrigger
                key={tab.value}
                value={tab.value}
                className={`h-9 rounded-none border-b-2 border-transparent px-4 text-sm ${C.text60}
                  data-[state=active]:border-[#37352F] data-[state=active]:${C.text}
                  data-[state=active]:shadow-none data-[state=active]:bg-transparent`}
              >
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>

          <TabsContent value="consultation" className="mt-4">
            <ConsultationTab />
          </TabsContent>
          <TabsContent value="examination" className="mt-4">
            <ExaminationTab />
          </TabsContent>
          <TabsContent value="procedure" className="mt-4">
            <ProcedureTab />
          </TabsContent>
          <TabsContent value="vaccine" className="mt-4">
            <VaccineTab />
          </TabsContent>
          <TabsContent value="checkup" className="mt-4">
            <CheckupTab />
          </TabsContent>
        </Tabs>
      </div>
    </PageLayout>
  );
}
