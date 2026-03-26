// React/Framework
import { useState, useMemo, useCallback, memo, useDeferredValue } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { paths } from "@/config/paths";
import { TaxTypeSelector } from "@/components/shared/TaxTypeSelector/TaxTypeSelector";
import { TaxRateSelector } from "@/components/shared/TaxRateSelector/TaxRateSelector";

// DnD
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";

// Shared hooks
import { useSortableList } from "@/hooks/use-sortable-list";

// External
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import ChevronDown from "lucide-react/dist/esm/icons/chevron-down";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import Plus from "lucide-react/dist/esm/icons/plus";
import { toast } from "sonner";

// Radix UI Tabs (primitive — same as DiagnosisSettings)
import * as TabsPrimitive from "@radix-ui/react-tabs";

// Internal shared
import { TableCell } from "@/components/ui/table";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { DataTableRow } from "@/components/shared/DataTable/DataTableRow";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { PropertyInput } from "@/components/shared/SidePeek/PropertyInput";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { MoneyInput } from "@/components/shared/SidePeek/MoneyInput";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, LAYOUT } from "@/lib/design-tokens";

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
import type { TaxType } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

type TreeItem = TreatmentItem & { children: TreatmentItem[] };

type TreatmentFormData = {
  name: string;
  price: number;
  description: string;
  isActive: boolean;
  taxType: TaxType;
  taxRate: number;
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
// TreatmentItemSidePanel
// ─────────────────────────────────────────────────

interface TreatmentItemSidePanelProps {
  item: TreatmentItem | null;
  onClose: () => void;
  onSave: (data: TreatmentFormData) => void;
  onDeleteRequest: () => void;
}

const TreatmentItemSidePanel = memo(function TreatmentItemSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: TreatmentItemSidePanelProps) {
  // rerender-lazy-state-init: 初回マウント時のみ item から初期化
  const [formData, setFormData] = useState<TreatmentFormData>(() => ({
    name: item?.name ?? "",
    price: item?.price ?? 0,
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
    taxType: (item?.taxType ?? "excluded") as TaxType,
    taxRate: item?.taxRate ?? 0.1,
  }));

  return (
    <MasterSidePanel
      isNew={item === null}
      title={formData.name}
      onTitleChange={(v) => setFormData((prev) => ({ ...prev, name: v }))}
      onClose={onClose}
      onSave={() => onSave(formData)}
      onDelete={item !== null ? onDeleteRequest : undefined}
      icon={<Stethoscope className={LAYOUT.pageIcon.innerIcon} />}
    >
      <StatusToggleButton
        isActive={formData.isActive}
        onToggle={() => setFormData((prev) => ({ ...prev, isActive: !prev.isActive }))}
      />
      <MoneyInput
        value={formData.price}
        onChange={(v) => setFormData((prev) => ({ ...prev, price: v }))}
      />
      <PropertyRow label="課税区分">
        <TaxTypeSelector
          value={formData.taxType}
          onChange={(v) => setFormData((prev) => ({ ...prev, taxType: v }))}
        />
      </PropertyRow>
      <PropertyRow label="税率">
        <TaxRateSelector
          value={formData.taxRate}
          onChange={(v) => setFormData((prev) => ({ ...prev, taxRate: v }))}
        />
      </PropertyRow>
      <PropertyRow label="備考">
        <PropertyInput
          value={formData.description}
          onChange={(v) => setFormData((prev) => ({ ...prev, description: v }))}
          placeholder="補足情報など"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});

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
      <TableCell className="w-[32px]" />
      <TableCell>
        <div className="flex items-center gap-1 pl-[22px]">
          <span className="size-[22px] shrink-0" />
          <span className={`text-base ${C.text}`}>{item.name}</span>
        </div>
      </TableCell>
      <TableCell className="text-right">
        <span className={`text-base ${C.text70} font-mono`}>
          {item.price > 0 ? `¥${item.price.toLocaleString()}` : "-"}
        </span>
      </TableCell>
      <TableCell className="text-center">
        <NotionStatusPill isActive={item.isActive} />
      </TableCell>
      <TableCell className="p-0 text-right">
        <RowActionButton onClick={onEdit} />
      </TableCell>
    </DataTableRow>
  );
}

// ─────────────────────────────────────────────────
// TreatmentTabContent (generic — shared across all 5 tabs)
// ─────────────────────────────────────────────────

interface TreatmentTabContentProps extends TreatmentTabConfig {
  editTarget: TreatmentItem | "new" | null;
  onEditTargetChange: (v: TreatmentItem | "new" | null) => void;
  onSave: (data: TreatmentFormData) => void;
  onDeleteRequest: () => void;
  pendingDelete: TreatmentItem | null;
  onPendingDeleteChange: (item: TreatmentItem | null) => void;
}

function TreatmentTabContent({
  data: rawData,
  entityLabel,
  emptyMessage,
  searchPlaceholder,
  onDelete,
  onReorder,
  onEditTargetChange,
  pendingDelete,
  onPendingDeleteChange,
}: TreatmentTabContentProps) {
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());

  const treeItems = useMemo(() => buildTree(rawData ?? []), [rawData]);

  const { orderedItems: orderedRoots, sensors, handleDragEnd } = useSortableList({
    items: treeItems,
    onReorder: (newIds) => {
      onReorder(newIds.map(Number));
    },
  });

  const filteredRoots = useMemo(() => {
    let items = orderedRoots;
    for (const f of activeFilters) {
      if (f.key === "status" && typeof f.value === "string") {
        const want = f.value === "active";
        items = items.filter((r) => f.condition === "is" ? r.isActive === want : r.isActive !== want);
      }
    }
    if (deferredSearch) {
      const lower = deferredSearch.toLowerCase();
      items = items.filter(
        (r) =>
          r.name.toLowerCase().includes(lower) ||
          r.children.some((c) => c.name.toLowerCase().includes(lower)),
      );
    }
    return items;
  }, [orderedRoots, activeFilters, deferredSearch]);

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

  const handleEdit = useCallback((item: TreatmentItem) => {
    onEditTargetChange(item);
  }, [onEditTargetChange]);

  const handleClose = useCallback(() => {
    onEditTargetChange(null);
  }, [onEditTargetChange]);

  const handleDeleteConfirm = useCallback(() => {
    if (!pendingDelete) return;
    onDelete(pendingDelete.id, {
      onSuccess: () => {
        onPendingDeleteChange(null);
        handleClose();
        toast.success("削除しました");
      },
      onError: () => toast.error("削除に失敗しました"),
    });
  }, [pendingDelete, onDelete, handleClose, onPendingDeleteChange]);

  return (
    <>
      <div className="flex flex-col gap-4">
        <NotionFilter
          properties={[MASTER_STATUS_FILTER]}
          activeFilters={activeFilters}
          onFilterChange={setActiveFilters}
          searchTerm={searchTerm}
          onSearchChange={setSearchTerm}
          searchPlaceholder={searchPlaceholder}
          count={totalCount}
        />

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
                    <SortableDataTableRow
                      key={row.item.id}
                      id={row.item.id}
                      onClick={() => handleEdit(row.item)}
                    >
                      <TableCell>
                        <div className="flex items-center gap-1">
                          {row.item.children.length > 0 ? (
                            <button
                              type="button"
                              onClick={(e) => {
                                e.stopPropagation();
                                toggleExpanded(row.item.id);
                              }}
                              className={`size-[22px] flex items-center justify-center rounded-[3px] ${C.text40} ${C.hoverBgMedium} transition-colors shrink-0`}
                            >
                              {row.isExpanded ? (
                                <ChevronDown className="size-3.5" />
                              ) : (
                                <ChevronRight className="size-3.5" />
                              )}
                            </button>
                          ) : (
                            <span className="size-[22px] shrink-0" />
                          )}
                          <span className={`text-base font-medium ${C.text}`}>{row.item.name}</span>
                          {row.item.children.length > 0 ? (
                            <span className={`text-base ${C.text25} ml-0.5`}>{row.item.children.length}</span>
                          ) : null}
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        <span className={`text-base ${C.text70} font-mono`}>
                          {row.item.price > 0 ? `¥${row.item.price.toLocaleString()}` : "-"}
                        </span>
                      </TableCell>
                      <TableCell className="text-center">
                        <NotionStatusPill isActive={row.item.isActive} />
                      </TableCell>
                      <TableCell className="p-0 text-right">
                        <RowActionButton onClick={() => handleEdit(row.item)} />
                      </TableCell>
                    </SortableDataTableRow>
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

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={() => onPendingDeleteChange(null)}
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
  const [editTarget, setEditTarget] = useState<TreatmentItem | "new" | null>(null);
  const [pendingDelete, setPendingDelete] = useState<TreatmentItem | null>(null);

  const handleTabChange = useCallback((tab: string) => {
    setSearchParams({ tab });
    setEditTarget(null);
    setPendingDelete(null);
  }, [setSearchParams]);

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

  const tabConfigs = useMemo<Record<string, TreatmentTabConfig>>(() => ({
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
            tax_type: data.taxType,
            tax_rate: data.taxRate,
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
              tax_type: data.taxType,
              tax_rate: data.taxRate,
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
            tax_type: data.taxType,
            tax_rate: data.taxRate,
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
              tax_type: data.taxType,
              tax_rate: data.taxRate,
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
  }), [
    consultationData, createConsultation, updateConsultation, deleteConsultation, reorderConsultations,
    examinationData, createExamination, updateExamination, deleteExamination, reorderExaminations,
    procedureData, createProcedure, updateProcedure, deleteProcedure, reorderProcedures,
    vaccineData, createVaccine, updateVaccine, deleteVaccine, reorderVaccines,
    checkupData, createCheckup, updateCheckup, deleteCheckup, reorderCheckups,
  ]);

  const selectedItem = editTarget !== null && editTarget !== "new" ? editTarget : null;
  const currentConfig = tabConfigs[activeTab];

  const handleClose = useCallback(() => setEditTarget(null), []);

  const handleSave = useCallback((data: TreatmentFormData) => {
    if (!data.name.trim()) {
      toast.error("名称は必須です");
      return;
    }
    if (!currentConfig) return;
    if (selectedItem) {
      currentConfig.onUpdate(selectedItem.id, data, {
        onSuccess: () => { toast.success("更新しました"); handleClose(); },
        onError: () => toast.error("更新に失敗しました"),
      });
    } else {
      currentConfig.onCreate(data, {
        onSuccess: () => { toast.success("登録しました"); handleClose(); },
        onError: () => toast.error("登録に失敗しました"),
      });
    }
  }, [currentConfig, selectedItem, handleClose]);

  const handleDeleteRequest = useCallback(() => {
    setPendingDelete(selectedItem);
  }, [selectedItem]);

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="治療プランマスタ"
            icon={<Stethoscope className="size-5 text-[#37352F]" />}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth="max-w-full"
            headerAction={
              <PrimaryButton onClick={() => setEditTarget("new")}>
                <Plus className="mr-1.5 size-4" />
                新規登録
              </PrimaryButton>
            }
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
                    className={`h-9 border-b-2 border-b-transparent px-4 text-base ${C.text60} outline-none transition-colors cursor-pointer
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
                    <TreatmentTabContent
                      {...config}
                      editTarget={editTarget}
                      onEditTargetChange={setEditTarget}
                      onSave={handleSave}
                      onDeleteRequest={handleDeleteRequest}
                      pendingDelete={pendingDelete}
                      onPendingDeleteChange={setPendingDelete}
                    />
                  </TabsPrimitive.Content>
                );
              })}
            </TabsPrimitive.Root>
          </PageLayout>
        </div>

        {editTarget !== null ? (
          <TreatmentItemSidePanel
            key={selectedItem ? String(selectedItem.id) : "new-item"}
            item={selectedItem}
            onClose={handleClose}
            onSave={handleSave}
            onDeleteRequest={handleDeleteRequest}
          />
        ) : null}
      </div>
    </>
  );
}
