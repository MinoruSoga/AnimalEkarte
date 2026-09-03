import { useNavigate } from "react-router";
import { Activity, Plus } from "lucide-react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { ResourceMasterReservationType } from "@/types/generated/models";
import { ReservationTypeDeleteDialogs } from "../components/ReservationTypeDeleteDialogs";
import { ReservationTypeSettingsContent } from "../components/ReservationTypeSettingsContent";
import { ReservationTypeSettingsSidePanels } from "../components/ReservationTypeSettingsSidePanels";
import { useReservationTypeSettings } from "../hooks/use-reservation-type-settings";

export function ReservationTypeSettings() {
  const navigate = useNavigate();
  const s = useReservationTypeSettings();

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="予約区分マスタ"
            icon={<Activity className={`${ICON.page} ${C.text}`} />}
            resource={ResourceMasterReservationType}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth={LAYOUT.pageContentMaxWidth.full}
            headerAction={
              s.canCreate ? (
                <PrimaryButton onClick={() => s.handleCategoryAddInGroup(undefined)}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            <ReservationTypeSettingsContent
              groups={s.groupsRaw}
              categories={s.categoryCrud.filteredItems}
              activeFilters={s.categoryCrud.activeFilters}
              onFilterChange={s.categoryCrud.setActiveFilters}
              searchTerm={s.categoryCrud.searchTerm}
              onSearchChange={s.categoryCrud.setSearchTerm}
              count={s.categoryCrud.filteredItems.length}
              canCreate={s.canCreate}
              canEdit={s.canEdit}
              onCategoryEdit={s.handleCategoryEdit}
              onGroupEdit={s.handleGroupEdit}
              onCategoryAddInGroup={s.handleCategoryAddInGroup}
              onGroupAdd={s.handleGroupAdd}
            />
          </PageLayout>
        </div>

        <ReservationTypeSettingsSidePanels
          groupEditTarget={s.groupCrud.editTarget}
          groupPanelItem={s.groupCrud.panelItem}
          categoryEditTarget={s.categoryCrud.editTarget}
          categoryPanelItem={s.categoryCrud.panelItem}
          activeGroups={s.activeGroups}
          categoryDefaultGroupId={s.categoryDefaultGroupId}
          canDelete={s.canDelete}
          canEdit={s.canEdit}
          onGroupClose={s.groupCrud.handleClose}
          onGroupSave={s.groupSave.handleSave}
          onGroupDeleteRequest={s.handleGroupDeleteRequest}
          onCategoryClose={s.categoryCrud.handleClose}
          onCategorySave={s.categorySave.handleSave}
          onCategoryDeleteRequest={s.handleCategoryDeleteRequest}
          onDirtyChange={s.handleDirtyChange}
        />
      </div>

      <ReservationTypeDeleteDialogs
        pendingGroupDelete={s.groupCrud.pendingDelete}
        pendingCategoryDelete={s.categoryCrud.pendingDelete}
        onGroupDeleteCancel={s.groupCrud.handleDeleteCancel}
        onGroupDeleteConfirm={s.groupCrud.handleDeleteConfirm}
        onCategoryDeleteCancel={s.categoryCrud.handleDeleteCancel}
        onCategoryDeleteConfirm={s.categoryCrud.handleDeleteConfirm}
      />
      {s.dirty.discardDialog}
    </>
  );
}
