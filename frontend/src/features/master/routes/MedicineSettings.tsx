import { useNavigate } from "react-router";
import { paths } from "@/config/paths";
import Pill from "lucide-react/dist/esm/icons/pill";
import Plus from "lucide-react/dist/esm/icons/plus";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { MedicineDeleteDialog } from "../components/MedicineDeleteDialog";
import { MedicineSidePanel } from "../components/MedicineSidePanel";
import { MedicineTableSection } from "../components/MedicineTableSection";
import { ResourceMasterMedical } from "@/types/generated/models";
import { useMedicineSettings } from "../hooks/use-medicine-settings";

export function MedicineSettings() {
  const navigate = useNavigate();
  const s = useMedicineSettings();

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
              s.canCreate ? (
                <PrimaryButton onClick={() => s.handleCreate()}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            <MedicineTableSection
              searchTerm={s.tableState.searchTerm}
              onSearchChange={s.tableState.setSearchTerm}
              activeFilters={s.tableState.activeFilters}
              onFilterChange={s.tableState.setActiveFilters}
              totalCount={s.tableState.totalCount}
              tableProps={{
                sensors: s.tableState.sensors,
                activeId: s.tableState.activeId,
                groupedMedicines: s.tableState.groupedMedicines,
                ungroupedMedicines: s.tableState.ungroupedMedicines,
                collapsedGroups: s.tableState.collapsedGroups,
                orderedMedicinesById: s.tableState.orderedMedicinesById,
                canCreate: s.canCreate,
                canEdit: s.canEdit,
                onDragStart: s.tableState.handleDragStart,
                onDragEnd: s.tableState.handleDragEnd,
                onDragCancel: s.tableState.handleDragCancel,
                onToggleGroup: s.tableState.toggleGroup,
                onEdit: s.handleEdit,
                onCreate: s.handleCreate,
              }}
            />
          </PageLayout>
        </div>
        <MedicineSidePanel
          isEditing={s.isEditing}
          selectedMedicine={s.selectedMedicine}
          isCategory={s.isCategory}
          defaultParentId={s.defaultParentId}
          categoryMedicines={s.categoryMedicines}
          panelDuration={s.panelDuration}
          onCloseEdit={s.handleCloseEdit}
          onSave={s.handleSave}
          onDeleteRequest={s.handleDeleteRequest}
          readOnly={!s.canEdit}
          canDelete={s.canDelete}
          onDirtyChange={s.handleDirtyChange}
        />
      </div>

      <MedicineDeleteDialog
        open={s.deleteConfirmOpen}
        onClose={() => s.setDeleteConfirmOpen(false)}
        onConfirm={s.executeDelete}
        medicine={s.selectedMedicine}
      />
      {s.dirty.discardDialog}
    </>
  );
}
