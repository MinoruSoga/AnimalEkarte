import { useState, useCallback, useMemo } from "react";
import { useNavigate } from "react-router";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { Activity, Plus } from "lucide-react";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
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
import { GroupSidePanel } from "../components/ReservationTypeGroupSidePanel";
import type { GroupFormData } from "../components/ReservationTypeGroupSidePanel";
import { CategorySidePanel } from "../components/ReservationTypeSidePanel";
import type { CategoryFormData } from "../components/ReservationTypeSidePanel";
import { ReservationTypeGroupedTable } from "../components/ReservationTypeGroupedTable";
import {
  buildReservationTypeCreateRequest,
  buildReservationTypeGroupCreateRequest,
  buildReservationTypeGroupUpdateRequest,
  buildReservationTypeUpdateRequest,
} from "./ReservationTypeSettingsModel";

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
    () => groupsRaw.filter((g) => g.isActive).map((g) => ({ id: g.id, name: g.name, color: g.color })),
    [groupsRaw],
  );

  const filteredCategories = useMemo(() => {
    return categoriesRaw;
  }, [categoriesRaw]);

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
  });

  const categoryCrud = useMasterCRUD<ReservationType>({
    data: filteredCategories,
    deleteMutation: deleteCategoryMutation,
    entityLabel: "予約区分",
    dirtyGuard: dirty,
    searchFilter: (item, term) => item.name.toLowerCase().includes(term.toLowerCase()),
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
  });

  const categorySave = useMasterSave<ReservationType, CategoryFormData, CreateReservationTypeRequest, UpdateReservationTypeRequest>({
    crud: categoryCrud,
    createMutation: createCategoryMutation,
    updateMutation: updateCategoryMutation,
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildReservationTypeCreateRequest,
    toUpdateRequest: buildReservationTypeUpdateRequest,
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
            maxWidth="max-w-full"
            headerAction={
              canCreate ? (
                <PrimaryButton onClick={() => handleCategoryAddInGroup(undefined)}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            <div className="flex flex-col gap-4">
              <NotionFilter
                properties={[MASTER_STATUS_FILTER]}
                activeFilters={categoryCrud.activeFilters}
                onFilterChange={categoryCrud.setActiveFilters}
                searchTerm={categoryCrud.searchTerm}
                onSearchChange={categoryCrud.setSearchTerm}
                searchPlaceholder="予約区分名で検索..."
                count={categoryCrud.filteredItems.length}
              />
              <ReservationTypeGroupedTable
                groups={groupsRaw}
                categories={categoryCrud.filteredItems}
                onCategoryEdit={handleCategoryEdit}
                onGroupEdit={handleGroupEdit}
                onCategoryAddInGroup={handleCategoryAddInGroup}
                canEdit={canEdit}
              />
              {canCreate ? (
                <button type="button" onClick={handleGroupAdd}
                  className={`flex items-center gap-1.5 text-sm ${C.text45} ${C.hoverText} ${C.hoverBgLight}
                    px-2 py-1.5 rounded-[3px] transition-colors w-fit`}>
                  <Plus className={ICON.action} />
                  グループを追加
                </button>
              ) : null}
            </div>
          </PageLayout>
        </div>

        {groupCrud.editTarget !== null ? (
          <GroupSidePanel
            key={groupCrud.panelItem ? String(groupCrud.panelItem.id) : "new-group"}
            item={groupCrud.panelItem}
            onClose={groupCrud.handleClose}
            onSave={groupSave.handleSave}
            onDeleteRequest={canDelete ? handleGroupDeleteRequest : undefined}
            readOnly={!canEdit}
            onDirtyChange={handleDirtyChange}
          />
        ) : null}
        {categoryCrud.editTarget !== null ? (
          <CategorySidePanel
            key={categoryCrud.panelItem ? String(categoryCrud.panelItem.id) : "new-category"}
            item={categoryCrud.panelItem}
            onClose={categoryCrud.handleClose}
            onSave={categorySave.handleSave}
            onDeleteRequest={canDelete ? handleCategoryDeleteRequest : undefined}
            readOnly={!canEdit}
            groups={activeGroups}
            defaultGroupId={categoryDefaultGroupId}
            onDirtyChange={handleDirtyChange}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={groupCrud.pendingDelete !== null}
        onClose={groupCrud.handleDeleteCancel}
        title="グループを削除しますか？"
        description={`「${groupCrud.pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除" variant="destructive"
        onConfirm={groupCrud.handleDeleteConfirm}
      />
      <ConfirmDialog
        open={categoryCrud.pendingDelete !== null}
        onClose={categoryCrud.handleDeleteCancel}
        title="予約区分を削除しますか？"
        description={`「${categoryCrud.pendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除" variant="destructive"
        onConfirm={categoryCrud.handleDeleteConfirm}
      />
    </>
  );
}
