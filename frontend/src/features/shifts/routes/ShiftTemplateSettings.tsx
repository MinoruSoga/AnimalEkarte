import { useCallback, useMemo, useState } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import Calendar from "lucide-react/dist/esm/icons/calendar";
import Plus from "lucide-react/dist/esm/icons/plus";
import { useNavigate } from "react-router";
import { toast } from "sonner";

import { DataTable, DESIGN_TABLE_HEADER_ROW, DESIGN_TABLE_HEADER_CELL } from "@/components/shared/DataTable/DataTable";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PropertyFilter } from "@/components/shared/PropertyFilter/PropertyFilter";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { usePermission } from "@/hooks/use-permission";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { useSortableList } from "@/hooks/use-sortable-list";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { ResourceShifts, ShiftTypeOff, ShiftTypePaidLeave } from "@/types/generated/models";
import { useCreateShiftTemplate } from "../api/create-shift-template";
import { useDeleteShiftTemplate } from "../api/delete-shift-template";
import { useGetShiftTemplates } from "../api/get-shift-templates";
import { useReorderShiftTemplates } from "../api/reorder-shift-templates";
import { useUpdateShiftTemplate } from "../api/update-shift-template";
import {
  ShiftTemplateDeleteDialog,
  ShiftTemplateRow,
  ShiftTemplateSidePanel,
} from "../components/ShiftTemplateSettingsParts";
import type { TemplateFormData } from "../components/shift-template-form-model";
import { SHIFT_STATUS_FILTER, SHIFT_TEMPLATE_COLUMNS, filterShiftTemplates } from "../components/shift-template-table-model";
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

    const breaks = formData.breaks.filter((b) => b.break_start && b.break_end);
    const isTimeHidden =
      formData.shift_type === ShiftTypeOff || formData.shift_type === ShiftTypePaidLeave;

    if (selectedItem !== null) {
      updateMutation.mutate(
        {
          id: selectedItem.id,
          input: {
            name: formData.name,
            shift_type: formData.shift_type,
            start_time: isTimeHidden ? null : formData.start_time || undefined,
            end_time: isTimeHidden ? null : formData.end_time || undefined,
            notes: formData.notes,
            is_active: formData.is_active,
            breaks: isTimeHidden ? [] : breaks,
          },
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
        {
          name: formData.name,
          shift_type: formData.shift_type,
          start_time: isTimeHidden ? undefined : formData.start_time || undefined,
          end_time: isTimeHidden ? undefined : formData.end_time || undefined,
          notes: formData.notes,
          is_active: formData.is_active,
          breaks: isTimeHidden ? [] : breaks,
        },
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

  const isSaving = createMutation.isPending || updateMutation.isPending;
  const isPanelReadOnly = selectedItem !== null ? !canEdit : !canCreate;

  return (
    <>
    <div className="flex h-full overflow-hidden">
      <div className="flex-1 min-w-0 overflow-auto">
        <PageLayout
          title="シフトテンプレートマスタ"
          icon={<Calendar className={`${ICON.page} ${C.text}`} />}
          resource={ResourceShifts}
          onBack={() => navigate(paths.settings.getHref())}
          maxWidth={LAYOUT.pageContentMaxWidth.full}
          headerAction={
            canCreate ? (
              <PrimaryButton onClick={handleCreate}>
                <Plus className={`mr-1.5 ${ICON.action}`} />
                新規登録
              </PrimaryButton>
            ) : null
          }
        >
          <div className="flex flex-col gap-4">
            <PropertyFilter
              properties={[SHIFT_STATUS_FILTER]}
              activeFilters={activeFilters}
              onFilterChange={setActiveFilters}
              searchTerm={searchTerm}
              onSearchChange={setSearchTerm}
              searchPlaceholder="テンプレート名で検索..."
              count={filteredItems.length}
            />
            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragEnd={handleDragEnd}
            >
              <SortableContext
                items={filteredItems.map((item) => item.id)}
                strategy={verticalListSortingStrategy}
              >
                <DataTable
                  headerRowClassName={DESIGN_TABLE_HEADER_ROW}
                  headerCellClassName={DESIGN_TABLE_HEADER_CELL}
                  columns={SHIFT_TEMPLATE_COLUMNS}
                  data={filteredItems}
                  emptyMessage="テンプレートがありません"
                  renderRow={(item) => (
                    <ShiftTemplateRow
                      key={item.id}
                      item={item}
                      canEdit={canEdit}
                      onEdit={() => handleEdit(item)}
                    />
                  )}
                />
              </SortableContext>
            </DndContext>
          </div>
        </PageLayout>
      </div>

      {isEditing ? (
        <ShiftTemplateSidePanel
          key={selectedItem ? selectedItem.id : "new"}
          item={selectedItem}
          onClose={handleClose}
          onSave={handleSave}
          onDeleteRequest={canDelete ? () => {
            if (selectedItem) setPendingDelete(selectedItem);
          } : undefined}
          isSaving={isSaving}
          readOnly={isPanelReadOnly}
          onDirtyChange={handleDirtyChange}
        />
      ) : null}

      <ShiftTemplateDeleteDialog
        pendingDelete={pendingDelete}
        onClose={() => setPendingDelete(null)}
        onConfirm={handleDeleteConfirm}
      />
    </div>
    {dirty.discardDialog}
    </>
  );
}
