import Calendar from "lucide-react/dist/esm/icons/calendar";
import Plus from "lucide-react/dist/esm/icons/plus";

import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { ResourceShifts } from "@/types/generated/models";
import { ShiftTemplateDeleteDialog, ShiftTemplateSidePanel } from "./ShiftTemplateSettingsParts";
import { ShiftTemplateSettingsList } from "./ShiftTemplateSettingsList";
import type { TemplateFormData } from "../lib/shift-template-form-model";
import type { ShiftTemplate } from "../types";
import { useSortableList } from "@/hooks/use-sortable-list";

type SortableSensors = ReturnType<typeof useSortableList<ShiftTemplate>>["sensors"];
type SortableDragEnd = ReturnType<typeof useSortableList<ShiftTemplate>>["handleDragEnd"];

interface ShiftTemplateSettingsWorkspaceProps {
  canCreate: boolean;
  canEdit: boolean;
  searchTerm: string;
  onSearchChange: (value: string) => void;
  activeFilters: ActiveFilter[];
  onFilterChange: (filters: ActiveFilter[]) => void;
  filteredItems: ShiftTemplate[];
  sensors: SortableSensors;
  onDragEnd: SortableDragEnd;
  onCreate: () => void;
  onEdit: (item: ShiftTemplate) => void;
  onBack: () => void;
  isEditing: boolean;
  selectedItem: ShiftTemplate | null;
  onClose: () => void;
  onSave: (formData: TemplateFormData) => void;
  onDeleteRequest: (() => void) | undefined;
  isSaving: boolean;
  isPanelReadOnly: boolean;
  onDirtyChange: (isDirty: boolean) => void;
  pendingDelete: ShiftTemplate | null;
  onDeleteCancel: () => void;
  onDeleteConfirm: () => void;
  discardDialog: React.ReactNode;
}

export function ShiftTemplateSettingsWorkspace({
  canCreate,
  canEdit,
  searchTerm,
  onSearchChange,
  activeFilters,
  onFilterChange,
  filteredItems,
  sensors,
  onDragEnd,
  onCreate,
  onEdit,
  onBack,
  isEditing,
  selectedItem,
  onClose,
  onSave,
  onDeleteRequest,
  isSaving,
  isPanelReadOnly,
  onDirtyChange,
  pendingDelete,
  onDeleteCancel,
  onDeleteConfirm,
  discardDialog,
}: ShiftTemplateSettingsWorkspaceProps) {
  return (
    <>
      <div className="flex h-full overflow-hidden">
        <div className="flex-1 min-w-0 overflow-auto">
          <PageLayout
            title="シフトテンプレートマスタ"
            icon={<Calendar className={`${ICON.page} ${C.text}`} />}
            resource={ResourceShifts}
            onBack={onBack}
            maxWidth={LAYOUT.pageContentMaxWidth.full}
            headerAction={
              canCreate ? (
                <PrimaryButton onClick={onCreate}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            <ShiftTemplateSettingsList
              searchTerm={searchTerm}
              onSearchChange={onSearchChange}
              activeFilters={activeFilters}
              onFilterChange={onFilterChange}
              filteredItems={filteredItems}
              canEdit={canEdit}
              sensors={sensors}
              onDragEnd={onDragEnd}
              onEdit={onEdit}
            />
          </PageLayout>
        </div>

        {isEditing ? (
          <ShiftTemplateSidePanel
            key={selectedItem ? selectedItem.id : "new"}
            item={selectedItem}
            onClose={onClose}
            onSave={onSave}
            onDeleteRequest={onDeleteRequest}
            isSaving={isSaving}
            readOnly={isPanelReadOnly}
            onDirtyChange={onDirtyChange}
          />
        ) : null}

        <ShiftTemplateDeleteDialog
          pendingDelete={pendingDelete}
          onClose={onDeleteCancel}
          onConfirm={onDeleteConfirm}
        />
      </div>
      {discardDialog}
    </>
  );
}
