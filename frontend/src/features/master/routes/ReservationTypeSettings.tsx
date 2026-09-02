import { useState, useCallback, useMemo } from "react";
import { useNavigate } from "react-router";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { Activity, Plus } from "lucide-react";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { ResourceMasterReservationType } from "@/types/generated/models";
import {
  useGetReservationTypes,
  useCreateReservationType,
  useUpdateReservationType,
  useDeleteReservationType,
} from "../api/reservation-types";
import type { ReservationType } from "../api/reservation-types";
import {
  useGetReservationTypeGroups,
  useCreateReservationTypeGroup,
  useUpdateReservationTypeGroup,
  useDeleteReservationTypeGroup,
} from "../api/reservation-type-groups";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";
import type { CreateReservationTypeGroupRequest, UpdateReservationTypeGroupRequest } from "../api/reservation-type-groups";
import type { CreateReservationTypeRequest, UpdateReservationTypeRequest } from "../api/reservation-types";
import type { GroupFormData } from "../components/ReservationTypeGroupSidePanel";
import type { CategoryFormData } from "../components/ReservationTypeSidePanel";
import { ReservationTypeDeleteDialogs } from "../components/ReservationTypeDeleteDialogs";
import { ReservationTypeSettingsContent } from "../components/ReservationTypeSettingsContent";
import { ReservationTypeSettingsSidePanels } from "../components/ReservationTypeSettingsSidePanels";
import {
  buildReservationTypeCreateRequest,
  buildReservationTypeGroupCreateRequest,
  buildReservationTypeGroupUpdateRequest,
  buildReservationTypeUpdateRequest,
  getActiveReservationTypeGroupOptions,
  matchesReservationTypeSearch,
} from "./reservation-type-settings-model";

// ─────────────────────────────────────────────────────────────────
// ReservationTypeSettings
// ─────────────────────────────────────────────────────────────────

export function ReservationTypeSettings() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterReservationType);

  const { data: groupsRaw = [] } = useGetReservationTypeGroups();
  const { data: categoriesRaw = [] } = useGetReservationTypes();

  // グループ選択肢（有効のみ）
  const activeGroups = useMemo(
    () => getActiveReservationTypeGroupOptions(groupsRaw),
    [groupsRaw],
  );

  // ミューテーション
  const createGroupMutation = useCreateReservationTypeGroup();
  const updateGroupMutation = useUpdateReservationTypeGroup();
  const deleteGroupMutation = useDeleteReservationTypeGroup();
  const createCategoryMutation = useCreateReservationType();
  const updateCategoryMutation = useUpdateReservationType();
  const deleteCategoryMutation = useDeleteReservationType();

  // BUG-380: 未保存破棄ガード
  const dirty = useSidePeekDirty();
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  // ── FR1: useMasterCRUD hooks (groups & categories) ────────────
  const groupCrud = useMasterCRUD<ReservationTypeGroup>({
    data: groupsRaw,
    deleteMutation: deleteGroupMutation,
    entityLabel: "予約区分グループ",
    dirtyGuard: dirty,
    permissions: { canDelete },
  });

  const categoryCrud = useMasterCRUD<ReservationType>({
    data: categoriesRaw,
    deleteMutation: deleteCategoryMutation,
    entityLabel: "予約区分",
    dirtyGuard: dirty,
    permissions: { canDelete },
    searchFilter: matchesReservationTypeSearch,
    activeFilterApply: (item, filters) => {
      for (const f of filters) {
        if (f.key === "status" && typeof f.value === "string") {
          const want = f.value === "active";
          if (f.condition === "is" ? item.isActive !== want : item.isActive === want) return false;
        }
      }
      return true;
    },
  });

  // ── Additional state: categoryDefaultGroupId (creation context) ───
  const [categoryDefaultGroupId, setCategoryDefaultGroupId] = useState<string | undefined>(undefined);

  // ── FR2: useMasterSave hooks ───────────────────────────────────
  const groupSave = useMasterSave<ReservationTypeGroup, GroupFormData, CreateReservationTypeGroupRequest, UpdateReservationTypeGroupRequest>({
    crud: groupCrud,
    createMutation: createGroupMutation,
    updateMutation: updateGroupMutation,
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildReservationTypeGroupCreateRequest,
    toUpdateRequest: buildReservationTypeGroupUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  const categorySave = useMasterSave<ReservationType, CategoryFormData, CreateReservationTypeRequest, UpdateReservationTypeRequest>({
    crud: categoryCrud,
    createMutation: createCategoryMutation,
    updateMutation: updateCategoryMutation,
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildReservationTypeCreateRequest,
    toUpdateRequest: buildReservationTypeUpdateRequest,
    permissions: { canCreate, canEdit },
  });

  // ── Handler wrappers ───────────────────────────────────────────
  const handleGroupEdit = useCallback((group: ReservationTypeGroup) => {
    groupCrud.handleEdit(group);
  }, [groupCrud]);

  const handleGroupAdd = useCallback(() => {
    groupCrud.handleNew();
  }, [groupCrud]);

  const handleGroupDeleteRequest = useCallback((item: ReservationTypeGroup) => {
    groupCrud.handleDeleteRequest(item);
  }, [groupCrud]);

  const handleCategoryEdit = useCallback((cat: ReservationType) => {
    categoryCrud.handleEdit(cat);
    setCategoryDefaultGroupId(undefined);
  }, [categoryCrud]);

  const handleCategoryAddInGroup = useCallback((groupId: string | undefined) => {
    categoryCrud.handleNew();
    setCategoryDefaultGroupId(groupId);
  }, [categoryCrud]);

  const handleCategoryDeleteRequest = useCallback((item: ReservationType) => {
    categoryCrud.handleDeleteRequest(item);
  }, [categoryCrud]);


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
              canCreate ? (
                <PrimaryButton onClick={() => handleCategoryAddInGroup(undefined)}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            <ReservationTypeSettingsContent
              groups={groupsRaw}
              categories={categoryCrud.filteredItems}
              activeFilters={categoryCrud.activeFilters}
              onFilterChange={categoryCrud.setActiveFilters}
              searchTerm={categoryCrud.searchTerm}
              onSearchChange={categoryCrud.setSearchTerm}
              count={categoryCrud.filteredItems.length}
              canCreate={canCreate}
              canEdit={canEdit}
              onCategoryEdit={handleCategoryEdit}
              onGroupEdit={handleGroupEdit}
              onCategoryAddInGroup={handleCategoryAddInGroup}
              onGroupAdd={handleGroupAdd}
            />
          </PageLayout>
        </div>

        <ReservationTypeSettingsSidePanels
          groupEditTarget={groupCrud.editTarget}
          groupPanelItem={groupCrud.panelItem}
          categoryEditTarget={categoryCrud.editTarget}
          categoryPanelItem={categoryCrud.panelItem}
          activeGroups={activeGroups}
          categoryDefaultGroupId={categoryDefaultGroupId}
          canDelete={canDelete}
          canEdit={canEdit}
          onGroupClose={groupCrud.handleClose}
          onGroupSave={groupSave.handleSave}
          onGroupDeleteRequest={handleGroupDeleteRequest}
          onCategoryClose={categoryCrud.handleClose}
          onCategorySave={categorySave.handleSave}
          onCategoryDeleteRequest={handleCategoryDeleteRequest}
          onDirtyChange={handleDirtyChange}
        />
      </div>

      <ReservationTypeDeleteDialogs
        pendingGroupDelete={groupCrud.pendingDelete}
        pendingCategoryDelete={categoryCrud.pendingDelete}
        onGroupDeleteCancel={groupCrud.handleDeleteCancel}
        onGroupDeleteConfirm={groupCrud.handleDeleteConfirm}
        onCategoryDeleteCancel={categoryCrud.handleDeleteCancel}
        onCategoryDeleteConfirm={categoryCrud.handleDeleteConfirm}
      />
      {dirty.discardDialog}
    </>
  );
}
