// React/Framework
import {
  useState,
  useMemo,
  useCallback,
  useLayoutEffect,
  useRef,
  useTransition,
} from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";

// External
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import Pill from "lucide-react/dist/esm/icons/pill";
import Plus from "lucide-react/dist/esm/icons/plus";

// Internal – shared
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { useReducedMotion } from "@/hooks/use-reduced-motion";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MedicineDeleteDialog } from "../components/MedicineDeleteDialog";
import { MedicineSidePanel, type MedicineFormData } from "../components/MedicineSidePanel";
import { MedicineTableSection } from "../components/MedicineTableSection";
import { useMedicineTableState } from "../hooks/use-medicine-table-state";
import {
  buildMedicineCreateRequest,
  buildMedicineUpdateRequest,
  getCategoryMedicines,
  isCategoryMedicine,
} from "./medicine-settings-model";

// Internal – feature API (direct import, no barrel)
import { useGetAllMedicines, useCreateMedicine, useUpdateMedicine, useDeleteMedicine, useReorderMedicines } from "../api/medicines";
import type { CreateMedicineRequest, UpdateMedicineRequest } from "@/types/medicine";
import { ResourceMasterMedical } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { Medicine } from "@/types";

export function MedicineSettings() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);
  const permissionsRef = useRef({ canDelete: canDelete === true });
  useLayoutEffect(() => {
    permissionsRef.current = { canDelete: canDelete === true };
  }, [canDelete]);
  const reduced = useReducedMotion();
  const panelDuration = reduced ? 0 : 0.2;

  // BUG-380 統一: formData は SidePanel が所有。新規作成時の親カテゴリ起点だけ親で保持。
  const [defaultParentId, setDefaultParentId] = useState<string | undefined>(undefined);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  // ── API ──
  const { data: medicines = [] } = useGetAllMedicines();
  const createMutation = useCreateMedicine();
  const updateMutation = useUpdateMedicine();
  const deleteMutation = useDeleteMedicine();
  const reorderMutation = useReorderMedicines();

  // BUG-380: 未保存破棄ガード（14 マスタ画面と統一パターン）
  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  // R-F8-S3: useSidePeekDirty は毎レンダー新規オブジェクトを返す（confirmDiscard 自体は安定）。
  // dirtyGuard を confirmDiscard だけに絞って安定化し、medicineCrud.handleEdit/handleNew の
  // 参照安定性（→ 行コンポーネント memo() の実効性）を確保する。
  const dirtyGuard = useMemo(
    () => ({ confirmDiscard: dirty.confirmDiscard }),
    [dirty.confirmDiscard],
  );

  // ── FR1: useMasterCRUD (editTarget state only; deletion modal kept external) ──
  const medicineCrud = useMasterCRUD<Medicine>({
    data: medicines,
    deleteMutation,
    entityLabel: "薬品",
    dirtyGuard,
    permissions: { canDelete },
  });

  // ── Derived: editTarget → selectedMedicine / isEditing / isCategory ──
  const { editTarget } = medicineCrud;
  const selectedMedicine = editTarget !== null && editTarget !== "new" ? editTarget : null;
  const isEditing = editTarget !== null;
  const isCategory = isCategoryMedicine(selectedMedicine);

  const tableState = useMedicineTableState({
    medicines,
    reorderMutation,
    updateMutation,
    canEdit,
  });

  // ── Derived: カテゴリ medicine（parentId なし、price === 0）(js-cache-function-results) ──
  const categoryMedicines = useMemo(
    () => getCategoryMedicines(medicines),
    [medicines],
  );

  // ── Handlers ──
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

  const setEditTargetAfterSave = useCallback((target: Medicine | "new" | null) => {
    medicineCrud.setEditTarget(target);
    if (target === null) setDefaultParentId(undefined);
  }, [medicineCrud]);

  // ── FR2: useMasterSave (with complex price/parent logic) ──
  const medicineSave = useMasterSave<Medicine, MedicineFormData, CreateMedicineRequest, UpdateMedicineRequest>({
    crud: {
      editTarget,
      setEditTarget: setEditTargetAfterSave,
      startSaveTransition: startSaveTransitionWrapper,
    },
    createMutation,
    updateMutation,
    permissions: { canCreate, canEdit },
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: (data) => buildMedicineCreateRequest(data, isCategory),
    toUpdateRequest: (data) => buildMedicineUpdateRequest({ data, isCategory, selectedMedicine }),
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
    if (permissionsRef.current.canDelete !== true) return;
    deleteMutation.mutate(selectedMedicine.id, {
      onSuccess: () => {
        toast.success("削除しました");
        setDeleteConfirmOpen(false);
        handleCloseEdit();
      },
      onError: (error) => handleApiError(error, "薬剤の削除"),
    });
  }, [selectedMedicine, deleteMutation, handleCloseEdit]);

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="薬剤マスタ"
            icon={<Pill className={`${ICON.page} ${C.text}`} />}
            resource={ResourceMasterMedical}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth={LAYOUT.pageContentMaxWidth.full}
            headerAction={
              canCreate ? (
                <PrimaryButton onClick={() => handleCreate()}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            <MedicineTableSection
              searchTerm={tableState.searchTerm}
              onSearchChange={tableState.setSearchTerm}
              activeFilters={tableState.activeFilters}
              onFilterChange={tableState.setActiveFilters}
              totalCount={tableState.totalCount}
              tableProps={{
                sensors: tableState.sensors,
                activeId: tableState.activeId,
                groupedMedicines: tableState.groupedMedicines,
                ungroupedMedicines: tableState.ungroupedMedicines,
                collapsedGroups: tableState.collapsedGroups,
                orderedMedicinesById: tableState.orderedMedicinesById,
                canCreate,
                canEdit,
                onDragStart: tableState.handleDragStart,
                onDragEnd: tableState.handleDragEnd,
                onDragCancel: tableState.handleDragCancel,
                onToggleGroup: tableState.toggleGroup,
                onEdit: handleEdit,
                onCreate: handleCreate,
              }}
            />
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

      <MedicineDeleteDialog
        open={deleteConfirmOpen}
        onClose={() => setDeleteConfirmOpen(false)}
        onConfirm={executeDelete}
        medicine={selectedMedicine}
      />
    </>
  );
}
