// React/Framework
import type { ReactNode } from "react";
import { useState, useMemo, useCallback } from "react";
import { useNavigate, useSearchParams } from "react-router";

// DnD
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy, useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";

// Shared hooks
import { useSortableList } from "@/hooks/useSortableList";

// External
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import ChevronDown from "lucide-react/dist/esm/icons/chevron-down";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import Plus from "lucide-react/dist/esm/icons/plus";
import GripVertical from "lucide-react/dist/esm/icons/grip-vertical";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import X from "lucide-react/dist/esm/icons/x";
import { toast } from "sonner";

// Radix UI Tabs (primitive — same as DiagnosisSettings)
import * as TabsPrimitive from "@radix-ui/react-tabs";

// Internal shared
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout";
import { SearchFilterBar } from "@/components/shared/SearchFilterBar";
import { DataTable, DataTableRow } from "@/components/shared/DataTable";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";

// API hooks
import {
  useGetAllConsultations,
  useCreateConsultation,
  useUpdateConsultation,
  useDeleteConsultation,
  useReorderConsultations,
} from "@/features/master/api/consultations";
import {
  useGetAllExaminationTypes,
  useCreateExaminationType,
  useUpdateExaminationType,
  useDeleteExaminationType,
  useReorderExaminationTypes,
} from "@/features/master/api/exam-types-master";
import {
  useGetAllProcedures,
  useCreateProcedure,
  useUpdateProcedure,
  useDeleteProcedure,
  useReorderProcedures,
} from "@/features/master/api/procedures";
import {
  useGetAllVaccinesMaster,
  useCreateVaccineMaster,
  useUpdateVaccineMaster,
  useDeleteVaccineMaster,
  useReorderVaccinesMaster,
} from "@/features/master/api/vaccines-master";
import {
  useGetAllCheckupTypes,
  useCreateCheckupType,
  useUpdateCheckupType,
  useDeleteCheckupType,
  useReorderCheckupTypes,
} from "@/features/master/api/checkup-types";

// Types
import type { TreatmentItem } from "@/lib/transforms/treatment";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

type TreeItem = TreatmentItem & { children: TreatmentItem[] };

type TreatmentFormData = {
  name: string;
  price: number;
  description: string;
  isActive: boolean;
};

type MutateCallbacks = {
  onSuccess: () => void;
  onError: () => void;
};

interface TreatmentTabConfig {
  data: TreatmentItem[] | undefined;
  entityLabel: string;
  emptyMessage: string;
  searchPlaceholder: string;
  onCreate: (data: TreatmentFormData, callbacks: MutateCallbacks) => void;
  onUpdate: (id: string, data: TreatmentFormData, callbacks: MutateCallbacks) => void;
  onDelete: (id: string, callbacks: MutateCallbacks) => void;
  onReorder: (ids: number[]) => void;
}

type VirtualRow =
  | { type: "root"; item: TreeItem; isExpanded: boolean }
  | { type: "child"; item: TreatmentItem };

// ─────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────

const TABS = [
  { value: "consultation", label: "診察" },
  { value: "examination", label: "検査" },
  { value: "procedure", label: "処置" },
  { value: "vaccine", label: "予防接種" },
  { value: "checkup", label: "定期健診" },
] as const;

const TREATMENT_COLUMNS = [
  { header: "", className: "w-[32px]" },
  { header: "名称" },
  { header: "単価(税込)", className: "w-[120px]", align: "right" as const },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

// ─────────────────────────────────────────────────
// buildTree: flat array → tree
// ─────────────────────────────────────────────────

function buildTree(items: TreatmentItem[]): TreeItem[] {
  const map = new Map<string, TreeItem>(
    items.map((i) => [i.id, { ...i, children: [] }]),
  );
  const roots: TreeItem[] = [];
  for (const item of map.values()) {
    if (item.parentId) {
      map.get(item.parentId)?.children.push(item);
    } else {
      roots.push(item);
    }
  }
  return roots;
}

// ─────────────────────────────────────────────────
// NotionStatusPill
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
// PropertyRow (Notion-style)
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
// PropInput
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
// TreatmentItemSidePanel
// ─────────────────────────────────────────────────

function TreatmentItemSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: {
  item: TreatmentItem | null;
  onClose: () => void;
  onSave: (data: TreatmentFormData) => void;
  onDeleteRequest: () => void;
}) {
  const [formData, setFormData] = useState<TreatmentFormData>({
    name: item?.name ?? "",
    price: item?.price ?? 0,
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
  });

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
              onClick={onDeleteRequest}
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
              <Stethoscope className={LAYOUT.pageIcon.innerIcon} />
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
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="無題"
              autoFocus
            />
          </div>
          <div className={`${STYLE.sectionDivider} mb-1`} />
          <div className="py-1">
            <PropertyRow label="ステータス">
              <button
                type="button"
                onClick={() => setFormData({ ...formData, isActive: !formData.isActive })}
                className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
              >
                <NotionStatusPill isActive={formData.isActive} />
              </button>
            </PropertyRow>
            <PropertyRow label="単価(税込)">
              <input
                type="number"
                min={0}
                className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder} font-mono`}
                value={formData.price === 0 ? "" : formData.price}
                onChange={(e) =>
                  setFormData({ ...formData, price: Number(e.target.value) || 0 })
                }
                placeholder="0"
              />
            </PropertyRow>
            <PropertyRow label="備考">
              <PropInput
                value={formData.description}
                onChange={(v) => setFormData({ ...formData, description: v })}
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
// SortableTreatmentRow (root items with DnD)
// ─────────────────────────────────────────────────

function SortableTreatmentRow({
  item,
  isExpanded,
  onToggle,
  onEdit,
}: {
  item: TreeItem;
  isExpanded: boolean;
  onToggle: (id: string) => void;
  onEdit: () => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: item.id });
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };
  const hasChildren = item.children.length > 0;

  return (
    <DataTableRow ref={setNodeRef} style={style} {...attributes} onClick={onEdit}>
      <TableCell
        className="w-[32px] py-2.5 text-[#37352F]/20 cursor-grab"
        {...listeners}
      >
        <GripVertical className="size-4" />
      </TableCell>
      <TableCell className="py-2.5">
        <div className="flex items-center gap-1">
          {hasChildren ? (
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                onToggle(item.id);
              }}
              className={`size-[22px] flex items-center justify-center rounded-[3px] ${C.text40} ${C.hoverBgMedium} transition-colors shrink-0`}
            >
              {isExpanded ? (
                <ChevronDown className="size-3.5" />
              ) : (
                <ChevronRight className="size-3.5" />
              )}
            </button>
          ) : (
            <span className="size-[22px] shrink-0" />
          )}
          <span className={`text-sm font-medium ${C.text}`}>{item.name}</span>
          {hasChildren ? (
            <span className={`text-xs ${C.text25} ml-0.5`}>{item.children.length}</span>
          ) : null}
        </div>
      </TableCell>
      <TableCell className="py-2.5 text-right">
        <span className={`text-sm ${C.text70} font-mono`}>
          {item.price > 0 ? `¥${item.price.toLocaleString()}` : "-"}
        </span>
      </TableCell>
      <TableCell className="text-center py-2.5">
        <NotionStatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="text-right py-2.5">
        <RowActionButton onClick={onEdit} />
      </TableCell>
    </DataTableRow>
  );
}

// ─────────────────────────────────────────────────
// ChildTreatmentRow (child items — not sortable)
// ─────────────────────────────────────────────────

function ChildTreatmentRow({
  item,
  onEdit,
}: {
  item: TreatmentItem;
  onEdit: () => void;
}) {
  return (
    <DataTableRow onClick={onEdit}>
      <TableCell className="w-[32px] py-2.5" />
      <TableCell className="py-2.5">
        <div className="flex items-center gap-1 pl-[22px]">
          <span className="size-[22px] shrink-0" />
          <span className={`text-sm ${C.text}`}>{item.name}</span>
        </div>
      </TableCell>
      <TableCell className="py-2.5 text-right">
        <span className={`text-sm ${C.text70} font-mono`}>
          {item.price > 0 ? `¥${item.price.toLocaleString()}` : "-"}
        </span>
      </TableCell>
      <TableCell className="text-center py-2.5">
        <NotionStatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="text-right py-2.5">
        <RowActionButton onClick={onEdit} />
      </TableCell>
    </DataTableRow>
  );
}

// ─────────────────────────────────────────────────
// TreatmentTabContent (generic — shared across all 5 tabs)
// ─────────────────────────────────────────────────

function TreatmentTabContent({
  data: rawData,
  entityLabel,
  emptyMessage,
  searchPlaceholder,
  onCreate,
  onUpdate,
  onDelete,
  onReorder,
}: TreatmentTabConfig) {
  const [selectedItem, setSelectedItem] = useState<TreatmentItem | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [pendingDelete, setPendingDelete] = useState<TreatmentItem | null>(null);
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());

  const treeItems = useMemo(() => buildTree(rawData ?? []), [rawData]);

  const { orderedItems: orderedRoots, sensors, handleDragEnd } = useSortableList({
    items: treeItems,
    onReorder: (newIds) => {
      onReorder(newIds.map(Number));
    },
  });

  const filteredRoots = useMemo(() => {
    if (!searchTerm) return orderedRoots;
    const lower = searchTerm.toLowerCase();
    return orderedRoots.filter(
      (r) =>
        r.name.toLowerCase().includes(lower) ||
        r.children.some((c) => c.name.toLowerCase().includes(lower)),
    );
  }, [orderedRoots, searchTerm]);

  const toggleExpanded = useCallback((id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const flatRows = useMemo<VirtualRow[]>(
    () =>
      filteredRoots.flatMap((root) => {
        const isExpanded = expandedIds.has(root.id);
        const rows: VirtualRow[] = [{ type: "root", item: root, isExpanded }];
        if (isExpanded && root.children.length > 0) {
          for (const child of root.children) {
            rows.push({ type: "child", item: child });
          }
        }
        return rows;
      }),
    [filteredRoots, expandedIds],
  );

  const totalCount = (rawData ?? []).length;

  const handleEdit = (item: TreatmentItem) => {
    setSelectedItem(item);
    setIsEditing(true);
  };

  const handleCreate = () => {
    setSelectedItem(null);
    setIsEditing(true);
  };

  const handleClose = () => {
    setIsEditing(false);
    setSelectedItem(null);
  };

  const handleSave = (data: TreatmentFormData) => {
    if (!data.name.trim()) {
      toast.error("名称は必須です");
      return;
    }
    if (selectedItem) {
      onUpdate(selectedItem.id, data, {
        onSuccess: () => {
          toast.success("更新しました");
          handleClose();
        },
        onError: () => toast.error("更新に失敗しました"),
      });
    } else {
      onCreate(data, {
        onSuccess: () => {
          toast.success("登録しました");
          handleClose();
        },
        onError: () => toast.error("登録に失敗しました"),
      });
    }
  };

  const handleDeleteConfirm = () => {
    if (!pendingDelete) return;
    onDelete(pendingDelete.id, {
      onSuccess: () => {
        setPendingDelete(null);
        handleClose();
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  };

  return (
    <>
      <div className="flex h-full">
        {/* Table area */}
        <div className="flex flex-col gap-4 flex-1 min-w-0">
          <div className="flex items-center gap-3">
            <div className="flex-1 min-w-0">
              <SearchFilterBar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                placeholder={searchPlaceholder}
                count={totalCount}
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

          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={handleDragEnd}
          >
            <SortableContext
              items={filteredRoots.map((r) => r.id)}
              strategy={verticalListSortingStrategy}
            >
              <DataTable
                columns={TREATMENT_COLUMNS}
                data={flatRows}
                emptyMessage={emptyMessage}
                renderRow={(row) => {
                  if (row.type === "root") {
                    return (
                      <SortableTreatmentRow
                        key={row.item.id}
                        item={row.item}
                        isExpanded={row.isExpanded}
                        onToggle={toggleExpanded}
                        onEdit={() => handleEdit(row.item)}
                      />
                    );
                  }
                  return (
                    <ChildTreatmentRow
                      key={row.item.id}
                      item={row.item}
                      onEdit={() => handleEdit(row.item)}
                    />
                  );
                }}
              />
            </SortableContext>
          </DndContext>
        </div>

        {/* Side peek */}
        {isEditing ? (
          <TreatmentItemSidePanel
            key={selectedItem ? String(selectedItem.id) : "new-item"}
            item={selectedItem}
            onClose={handleClose}
            onSave={handleSave}
            onDeleteRequest={() => setPendingDelete(selectedItem)}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => setPendingDelete(null)}
        title={`${entityLabel}を削除しますか？`}
        description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleDeleteConfirm}
      />
    </>
  );
}

// ─────────────────────────────────────────────────
// TreatmentPlanMaster (main page)
// ─────────────────────────────────────────────────

export function TreatmentPlanMaster() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const activeTab = searchParams.get("tab") ?? "consultation";

  const handleTabChange = (tab: string) => {
    setSearchParams({ tab });
  };

  // ── Consultations ──────────────────────────────────
  const { data: consultationData } = useGetAllConsultations();
  const createConsultation = useCreateConsultation();
  const updateConsultation = useUpdateConsultation();
  const deleteConsultation = useDeleteConsultation();
  const reorderConsultations = useReorderConsultations();

  // ── Examination Types ──────────────────────────────
  const { data: examinationData } = useGetAllExaminationTypes();
  const createExamination = useCreateExaminationType();
  const updateExamination = useUpdateExaminationType();
  const deleteExamination = useDeleteExaminationType();
  const reorderExaminations = useReorderExaminationTypes();

  // ── Procedures ─────────────────────────────────────
  const { data: procedureData } = useGetAllProcedures();
  const createProcedure = useCreateProcedure();
  const updateProcedure = useUpdateProcedure();
  const deleteProcedure = useDeleteProcedure();
  const reorderProcedures = useReorderProcedures();

  // ── Vaccines ───────────────────────────────────────
  const { data: vaccineData } = useGetAllVaccinesMaster();
  const createVaccine = useCreateVaccineMaster();
  const updateVaccine = useUpdateVaccineMaster();
  const deleteVaccine = useDeleteVaccineMaster();
  const reorderVaccines = useReorderVaccinesMaster();

  // ── Checkup Types ──────────────────────────────────
  const { data: checkupData } = useGetAllCheckupTypes();
  const createCheckup = useCreateCheckupType();
  const updateCheckup = useUpdateCheckupType();
  const deleteCheckup = useDeleteCheckupType();
  const reorderCheckups = useReorderCheckupTypes();

  // ── Tab configs ────────────────────────────────────

  const tabConfigs: Record<string, TreatmentTabConfig> = {
    consultation: {
      data: consultationData,
      entityLabel: "診察",
      emptyMessage: "診察が登録されていません",
      searchPlaceholder: "診察名で検索...",
      onCreate: (data, cb) =>
        createConsultation.mutate(
          {
            name: data.name,
            price: data.price,
            description: data.description || undefined,
            is_active: data.isActive,
          },
          { onSuccess: () => cb.onSuccess(), onError: () => cb.onError() },
        ),
      onUpdate: (id, data, cb) =>
        updateConsultation.mutate(
          {
            id,
            req: {
              name: data.name,
              price: data.price,
              description: data.description || undefined,
              is_active: data.isActive,
            },
          },
          { onSuccess: () => cb.onSuccess(), onError: () => cb.onError() },
        ),
      onDelete: (id, cb) =>
        deleteConsultation.mutate(id, {
          onSuccess: () => cb.onSuccess(),
          onError: () => cb.onError(),
        }),
      onReorder: (ids) => reorderConsultations.mutate({ ids }),
    },

    examination: {
      data: examinationData,
      entityLabel: "検査",
      emptyMessage: "検査が登録されていません",
      searchPlaceholder: "検査名で検索...",
      onCreate: (data, cb) =>
        createExamination.mutate(
          {
            name: data.name,
            price: data.price,
            description: data.description || undefined,
            is_active: data.isActive,
          },
          { onSuccess: () => cb.onSuccess(), onError: () => cb.onError() },
        ),
      onUpdate: (id, data, cb) =>
        updateExamination.mutate(
          {
            id,
            req: {
              name: data.name,
              price: data.price,
              description: data.description || undefined,
              is_active: data.isActive,
            },
          },
          { onSuccess: () => cb.onSuccess(), onError: () => cb.onError() },
        ),
      onDelete: (id, cb) =>
        deleteExamination.mutate(id, {
          onSuccess: () => cb.onSuccess(),
          onError: () => cb.onError(),
        }),
      onReorder: (ids) => reorderExaminations.mutate({ ids }),
    },

    procedure: {
      data: procedureData,
      entityLabel: "処置",
      emptyMessage: "処置が登録されていません",
      searchPlaceholder: "処置名で検索...",
      onCreate: (data, cb) =>
        createProcedure.mutate(
          {
            name: data.name,
            price: data.price,
            description: data.description || undefined,
            is_active: data.isActive,
          },
          { onSuccess: () => cb.onSuccess(), onError: () => cb.onError() },
        ),
      onUpdate: (id, data, cb) =>
        updateProcedure.mutate(
          {
            id,
            req: {
              name: data.name,
              price: data.price,
              description: data.description || undefined,
              is_active: data.isActive,
            },
          },
          { onSuccess: () => cb.onSuccess(), onError: () => cb.onError() },
        ),
      onDelete: (id, cb) =>
        deleteProcedure.mutate(id, {
          onSuccess: () => cb.onSuccess(),
          onError: () => cb.onError(),
        }),
      onReorder: (ids) => reorderProcedures.mutate({ ids }),
    },

    vaccine: {
      data: vaccineData,
      entityLabel: "予防接種",
      emptyMessage: "予防接種が登録されていません",
      searchPlaceholder: "予防接種名で検索...",
      onCreate: (data, cb) =>
        createVaccine.mutate(
          {
            name: data.name,
            price: data.price,
            description: data.description || undefined,
            is_active: data.isActive,
          },
          { onSuccess: () => cb.onSuccess(), onError: () => cb.onError() },
        ),
      onUpdate: (id, data, cb) =>
        updateVaccine.mutate(
          {
            id,
            req: {
              name: data.name,
              price: data.price,
              description: data.description || undefined,
              is_active: data.isActive,
            },
          },
          { onSuccess: () => cb.onSuccess(), onError: () => cb.onError() },
        ),
      onDelete: (id, cb) =>
        deleteVaccine.mutate(id, {
          onSuccess: () => cb.onSuccess(),
          onError: () => cb.onError(),
        }),
      onReorder: (ids) => reorderVaccines.mutate({ ids }),
    },

    checkup: {
      data: checkupData,
      entityLabel: "定期健診",
      emptyMessage: "定期健診が登録されていません",
      searchPlaceholder: "定期健診名で検索...",
      onCreate: (data, cb) =>
        createCheckup.mutate(
          {
            name: data.name,
            price: data.price,
            description: data.description || undefined,
            is_active: data.isActive,
          },
          { onSuccess: () => cb.onSuccess(), onError: () => cb.onError() },
        ),
      onUpdate: (id, data, cb) =>
        updateCheckup.mutate(
          {
            id,
            req: {
              name: data.name,
              price: data.price,
              description: data.description || undefined,
              is_active: data.isActive,
            },
          },
          { onSuccess: () => cb.onSuccess(), onError: () => cb.onError() },
        ),
      onDelete: (id, cb) =>
        deleteCheckup.mutate(id, {
          onSuccess: () => cb.onSuccess(),
          onError: () => cb.onError(),
        }),
      onReorder: (ids) => reorderCheckups.mutate({ ids }),
    },
  };

  return (
    <PageLayout
      title="治療プランマスタ"
      icon={<Stethoscope className="size-5 text-[#37352F]" />}
      onBack={() => navigate("/settings")}
      maxWidth="max-w-full"
    >
      <TabsPrimitive.Root
        value={activeTab}
        onValueChange={handleTabChange}
        className="flex flex-col gap-4"
      >
        <TabsPrimitive.List
          className={`flex h-9 border-b ${C.borderLight} gap-0`}
        >
          {TABS.map((tab) => (
            <TabsPrimitive.Trigger
              key={tab.value}
              value={tab.value}
              className={`h-9 border-b-2 border-b-transparent px-4 text-sm ${C.text60} outline-none transition-colors cursor-pointer
                data-[state=active]:border-b-[#37352F] data-[state=active]:text-[#37352F] data-[state=active]:font-medium`}
            >
              {tab.label}
            </TabsPrimitive.Trigger>
          ))}
        </TabsPrimitive.List>

        {TABS.map((tab) => {
          const config = tabConfigs[tab.value];
          return (
            <TabsPrimitive.Content key={tab.value} value={tab.value} className="mt-4">
              <TreatmentTabContent {...config} />
            </TabsPrimitive.Content>
          );
        })}
      </TabsPrimitive.Root>
    </PageLayout>
  );
}
