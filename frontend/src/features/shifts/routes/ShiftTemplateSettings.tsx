import { useCallback, useMemo, useState } from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";

import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { useSortableList } from "@/hooks/use-sortable-list";
import { paths } from "@/config/paths";
import { ResourceShifts } from "@/types/generated/models";
import { useCreateShiftTemplate } from "../api/create-shift-template";
import { useDeleteShiftTemplate } from "../api/delete-shift-template";
import { useGetShiftTemplates } from "../api/get-shift-templates";
import { useReorderShiftTemplates } from "../api/reorder-shift-templates";
import { useUpdateShiftTemplate } from "../api/update-shift-template";
import { ShiftTemplateSettingsWorkspace } from "../components/shift-template-settings-workspace";
import type { TemplateFormData } from "../components/shift-template-form-model";
import { filterShiftTemplates } from "../components/shift-template-table-model";
import {
  toShiftTemplateCreateInput,
  toShiftTemplateUpdateInput,
} from "../components/shift-template-write-model";
import type { ShiftTemplate } from "../types";

export function ShiftTemplateSettings() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceShifts);
  const { data: templates = [] } = useGetShiftTemplates();
  const createMutation = useCreateShiftTemplate();
  const updateMutation = useUpdateShiftTemplate();
  const deleteMutation = useDeleteShiftTemplate();
  const reorderMutation = useReorderShiftTemplates();

  const [selectedItem, setSelectedItem] = useState<ShiftTemplate | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [pendingDelete, setPendingDelete] = useState<ShiftTemplate | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);

  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback(
    (isDirty: boolean) => {
      if (isDirty) dirty.markDirty();
      else dirty.markClean();
    },
    [dirty],
  );

  const { orderedItems, sensors, handleDragEnd } = useSortableList({
    items: templates,
    onReorder: (newIds) => {
      if (!canEdit) return;
      reorderMutation.mutate(newIds.map(Number));
    },
  });

  const filteredItems = useMemo(
    () => filterShiftTemplates(orderedItems, searchTerm, activeFilters),
    [orderedItems, searchTerm, activeFilters],
  );

  const handleCreate = useCallback(() => {
    if (!canCreate) return;
    dirty.runWithDiscardCheck(() => {
      setSelectedItem(null);
      setIsEditing(true);
    });
  }, [canCreate, dirty.runWithDiscardCheck]);

  const handleEdit = useCallback((item: ShiftTemplate) => {
    dirty.runWithDiscardCheck(() => {
      setSelectedItem(item);
      setIsEditing(true);
    });
  }, [dirty.runWithDiscardCheck]);

  const handleClose = useCallback(() => {
    dirty.runWithDiscardCheck(() => {
      setIsEditing(false);
      setSelectedItem(null);
    });
  }, [dirty.runWithDiscardCheck]);

  const handleSave = useCallback((formData: TemplateFormData) => {
    const canSave = selectedItem !== null ? canEdit : canCreate;
    if (!canSave) {
      toast.error("シフトテンプレートを保存する権限がありません");
      return;
    }

    if (selectedItem !== null) {
      updateMutation.mutate(
        {
          id: selectedItem.id,
          input: toShiftTemplateUpdateInput(formData),
        },
        {
          onSuccess: () => {
            toast.success("テンプレートを更新しました");
            dirty.markClean();
            handleClose();
          },
        },
      );
    } else {
      createMutation.mutate(
        toShiftTemplateCreateInput(formData),
        {
          onSuccess: () => {
            toast.success("テンプレートを作成しました");
            dirty.markClean();
            handleClose();
          },
        },
      );
    }
  }, [selectedItem, canEdit, canCreate, createMutation, updateMutation, handleClose, dirty]);

  const handleDeleteConfirm = useCallback(() => {
    if (!canDelete || !pendingDelete) return;
    deleteMutation.mutate(pendingDelete.id, {
      onSuccess: () => {
        toast.success("テンプレートを削除しました");
        setPendingDelete(null);
        if (selectedItem?.id === pendingDelete.id) {
          dirty.markClean();
          handleClose();
        }
      },
    });
  }, [canDelete, pendingDelete, deleteMutation, selectedItem, handleClose, dirty]);

  return (
    <ShiftTemplateSettingsWorkspace
      canCreate={canCreate}
      canEdit={canEdit}
      searchTerm={searchTerm}
      onSearchChange={setSearchTerm}
      activeFilters={activeFilters}
      onFilterChange={setActiveFilters}
      filteredItems={filteredItems}
      sensors={sensors}
      onDragEnd={handleDragEnd}
      onCreate={handleCreate}
      onEdit={handleEdit}
      onBack={() => navigate(paths.settings.getHref())}
      isEditing={isEditing}
      selectedItem={selectedItem}
      onClose={handleClose}
      onSave={handleSave}
      onDeleteRequest={canDelete ? () => {
        if (selectedItem) setPendingDelete(selectedItem);
      } : undefined}
      isSaving={createMutation.isPending || updateMutation.isPending}
      isPanelReadOnly={selectedItem !== null ? !canEdit : !canCreate}
      onDirtyChange={handleDirtyChange}
      pendingDelete={pendingDelete}
      onDeleteCancel={() => setPendingDelete(null)}
      onDeleteConfirm={handleDeleteConfirm}
      discardDialog={dirty.discardDialog}
    />
  );
}
