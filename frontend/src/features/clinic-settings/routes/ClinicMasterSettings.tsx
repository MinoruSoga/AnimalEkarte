import { NavigationBlocker } from "@/components/shared/NavigationBlocker/NavigationBlocker";
import { paths } from "@/config/paths";
import {
  ClinicDeleteDialog,
  ClinicMasterList,
  ClinicMasterSidePanel,
} from "../components/ClinicMasterSettingsPanels";
import { CompanyInvoiceSection } from "../components/CompanyInvoiceSection";
import { useClinicMasterSettings } from "../hooks/use-clinic-master-settings";

export function ClinicMasterSettings() {
  const s = useClinicMasterSettings();

  return (
    <>
      <NavigationBlocker when={s.isEditing} />
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <ClinicMasterList
            canCreate={s.canCreate}
            canEdit={s.canEdit}
            topSection={<CompanyInvoiceSection canEdit={s.canEdit} />}
            items={s.isPending || s.isError ? [] : s.filteredItems}
            searchTerm={s.searchTerm}
            onSearchChange={s.setSearchTerm}
            activeFilters={s.activeFilters}
            onFilterChange={s.setActiveFilters}
            emptyMessage={s.emptyMessage}
            onBack={() => s.navigate(paths.settings.getHref())}
            onCreate={s.handleCreate}
            onEdit={s.handleEdit}
          />
        </div>

        {s.isEditing ? (
          <ClinicMasterSidePanel
            selectedItem={s.selectedItem}
            formData={s.formData}
            setFormData={s.setFormData}
            formAction={s.formAction}
            nameError={s.formState.fieldErrors?.name}
            canEdit={s.canEdit}
            canDelete={s.canDelete}
            onClose={s.handleCloseEdit}
            onDeleteClick={s.setPendingDelete}
          />
        ) : null}
      </div>

      <ClinicDeleteDialog
        pendingDelete={s.pendingDelete}
        isPending={s.isDeletePending}
        onClose={() => s.setPendingDelete(null)}
        onConfirm={s.handleDeleteConfirm}
      />
    </>
  );
}
