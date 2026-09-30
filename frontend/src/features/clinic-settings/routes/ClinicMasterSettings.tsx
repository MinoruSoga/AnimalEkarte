import { NavigationBlocker } from "@/components/shared/NavigationBlocker/NavigationBlocker";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
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

  // EMR-227: 読み込み/エラーを空リストとして扱わず、専用フォールバックを表示する。
  if (s.isPending) return <LoadingFallback />;
  if (s.isError) return <ErrorFallback message="医院一覧の取得に失敗しました" />;

  return (
    <>
      <NavigationBlocker when={s.isEditing} />
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <ClinicMasterList
            canCreate={s.canCreate}
            canEdit={s.canEdit}
            topSection={<CompanyInvoiceSection canEdit={s.canEdit} />}
            items={s.filteredItems}
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
