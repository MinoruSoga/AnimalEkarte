import {
  useState,
  useMemo,
  useCallback,
  useLayoutEffect,
  useRef,
  useTransition,
} from "react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { useReducedMotion } from "@/hooks/use-reduced-motion";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import type { MedicineFormData } from "../components/MedicineSidePanel";
import { useMedicineTableState } from "../hooks/use-medicine-table-state";
import {
  buildMedicineCreateRequest,
  buildMedicineUpdateRequest,
  getCategoryMedicines,
  isCategoryMedicine,
  validateMedicineForm,
} from "./medicine-settings-model";
import { useGetAllMedicines, useCreateMedicine, useUpdateMedicine, useDeleteMedicine, useReorderMedicines } from "../api/medicines";
import { upsertMedicineDoseParam } from "../api/medicine-dose-params";
import type { CreateMedicineRequest, UpdateMedicineRequest } from "@/types/medicine";
import { ResourceMasterMedical } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";
import type { Medicine } from "@/types";

export function useMedicineSettings() {
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterMedical);
  const permissionsRef = useRef({ canDelete: canDelete === true });
  useLayoutEffect(() => {
    permissionsRef.current = { canDelete: canDelete === true };
  }, [canDelete]);
  const reduced = useReducedMotion();
  const panelDuration = reduced ? 0 : 0.2;

  const [defaultParentId, setDefaultParentId] = useState<string | undefined>(undefined);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  const { data: medicines = [] } = useGetAllMedicines();
  const createMutation = useCreateMedicine();
  const updateMutation = useUpdateMedicine();
  const deleteMutation = useDeleteMedicine();
  const reorderMutation = useReorderMedicines();

  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const dirtyGuard = useMemo(
    () => ({ runWithDiscardCheck: dirty.runWithDiscardCheck }),
    [dirty],
  );

  const medicineCrud = useMasterCRUD<Medicine>({
    data: medicines,
    deleteMutation,
    entityLabel: "薬品",
    dirtyGuard,
    permissions: { canDelete },
  });

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

  const categoryMedicines = useMemo(
    () => getCategoryMedicines(medicines),
    [medicines],
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

  const [, startSaveTransition] = useTransition();
  const startSaveTransitionWrapper = useCallback((cb: () => void) => {
    startSaveTransition(cb);
  }, []);

  const setEditTargetAfterSave = useCallback((target: Medicine | "new" | null) => {
    medicineCrud.setEditTarget(target);
    if (target === null) setDefaultParentId(undefined);
  }, [medicineCrud]);

  const medicineSave = useMasterSave<Medicine, MedicineFormData, CreateMedicineRequest, UpdateMedicineRequest>({
    crud: {
      editTarget,
      setEditTarget: setEditTargetAfterSave,
      startSaveTransition: startSaveTransitionWrapper,
    },
    createMutation,
    updateMutation,
    permissions: { canCreate, canEdit },
    validate: (data) => validateMedicineForm(data, isCategory),
    toCreateRequest: (data) => buildMedicineCreateRequest(data, isCategory),
    toUpdateRequest: (data) => buildMedicineUpdateRequest({ data, isCategory, selectedMedicine }),
    onSuccess: async (saved, formData) => {
      const drafts = formData.doseParamDrafts ?? [];
      for (const draft of drafts) {
        await upsertMedicineDoseParam(saved.id, draft.species, draft.input);
      }
    },
  });

  const handleSave = useCallback((formData: MedicineFormData) => {
    return medicineSave.handleSave(formData);
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

  return {
    canCreate,
    canDelete,
    canEdit,
    panelDuration,
    defaultParentId,
    deleteConfirmOpen,
    setDeleteConfirmOpen,
    dirty,
    handleDirtyChange,
    selectedMedicine,
    isEditing,
    isCategory,
    tableState,
    categoryMedicines,
    handleCloseEdit,
    handleEdit,
    handleCreate,
    handleSave,
    handleDeleteRequest,
    executeDelete,
  };
}
