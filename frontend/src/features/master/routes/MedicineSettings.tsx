// React/Framework
import { useState, useMemo, useCallback, useDeferredValue, useTransition } from "react";
import { flushSync } from "react-dom";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";

// DnD
import type { DragEndEvent } from "@dnd-kit/core";

// External
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import Pill from "lucide-react/dist/esm/icons/pill";
import Plus from "lucide-react/dist/esm/icons/plus";

// Internal – shared
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, ICON } from "@/lib/design-tokens";
import { useReducedMotion } from "@/hooks/use-reduced-motion";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useMasterCRUD } from "../hooks/use-master-crud";
import type { UseMasterCRUDReturn } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MedicineTable } from "../components/MedicineTable";
import { MedicineSidePanel, type MedicineFormData } from "../components/MedicineSidePanel";

// Internal – feature API (direct import, no barrel)
import { useGetAllMedicines, useCreateMedicine, useUpdateMedicine, useDeleteMedicine, useReorderMedicines } from "../api/medicines";
import type { CreateMedicineRequest, UpdateMedicineRequest } from "@/types/medicine";
import { ResourceMasterMedical } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { Medicine } from "@/types";

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

  // BUG-380: 未保存破棄ガード（14 マスタ画面と統一パターン）
  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

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
    medicineCrud.handleNew();
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
    } as UseMasterCRUDReturn<Medicine>,
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
        is_non_insurance: data.isNonInsurance,
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
        is_non_insurance: data.isNonInsurance,
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

  const tableContent = (
    <MedicineTable
      sensors={sensors}
      activeId={activeId}
      groupedMedicines={groupedMedicines}
      ungroupedMedicines={ungroupedMedicines}
      collapsedGroups={collapsedGroups}
      orderedMedicinesById={orderedMedicinesById}
      canCreate={canCreate}
      canEdit={canEdit}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
      onToggleGroup={toggleGroup}
      onEdit={handleEdit}
      onCreate={handleCreate}
    />
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
