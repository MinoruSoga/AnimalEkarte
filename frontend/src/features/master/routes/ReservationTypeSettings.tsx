import { useState, useCallback, useRef, useEffect, useMemo, Fragment } from "react";
import { useNavigate } from "react-router";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { Activity, ChevronDown, Pencil, Plus } from "lucide-react";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { TableCell } from "@/components/ui/table";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, ICON, PALETTE, STYLE } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { paths } from "@/config/paths";
import { usePermission } from "@/hooks/use-permission";
import { ResourceMasterReservationType } from "@/types/generated/models";
import {
  useGetReservationTypes,
  useCreateReservationType,
  useUpdateReservationType,
  useDeleteReservationType,
  useReorderReservationTypes,
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
import type { CreateReservationTypeRequest, UpdateReservationTypeRequest } from "@/types/reservation-type";
import { GroupSidePanel } from "./ReservationTypeGroupSidePanel";
import type { GroupFormData } from "./ReservationTypeGroupSidePanel";
import { CategorySidePanel } from "./ReservationTypeSidePanel";
import type { CategoryFormData } from "./ReservationTypeSidePanel";

// ─────────────────────────────────────────────────────────────────
// GroupedTable
// Notionライク：グループヘッダー（折りたたみ）＋インデントされた区分行
// ─────────────────────────────────────────────────────────────────

const UNCATEGORIZED_ID = "__uncategorized__";

interface GroupedTableProps {
  groups: ReservationTypeGroup[];
  categories: ReservationType[];
  onCategoryEdit: (cat: ReservationType) => void;
  onGroupEdit: (group: ReservationTypeGroup) => void;
  onCategoryAddInGroup: (groupId: string | undefined) => void;
  canEdit: boolean;
}

function GroupedTable({
  groups, categories, onCategoryEdit, onGroupEdit, onCategoryAddInGroup, canEdit,
}: GroupedTableProps) {
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set());
  const reorderMutation = useReorderReservationTypes();
  const resetOrderRef = useRef<() => void>(() => {});

  const handleReorder = useCallback((newIds: string[]) => {
    reorderMutation.mutate({ ids: newIds.map(Number) }, {
      onError: () => { resetOrderRef.current(); },
    });
  }, [reorderMutation]);

  const toggleCollapse = useCallback((id: string) => {
    setCollapsed((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id); else next.add(id);
      return next;
    });
  }, []);

  const { orderedItems, sensors, handleDragStart, handleDragEnd, handleDragCancel, resetOrder } =
    useSortableList({ items: categories, onReorder: handleReorder });
  useEffect(() => { resetOrderRef.current = resetOrder; }, [resetOrder]);

  // orderedItems をグループIDで振り分け
  const categoriesByGroupId = useMemo(() => {
    const map = new Map<string, ReservationType[]>();
    const uncat: ReservationType[] = [];
    for (const cat of orderedItems) {
      if (cat.groupId) {
        const arr = map.get(cat.groupId) ?? [];
        arr.push(cat);
        map.set(cat.groupId, arr);
      } else {
        uncat.push(cat);
      }
    }
    return { map, uncat };
  }, [orderedItems]);

  const uncatCats = categoriesByGroupId.uncat;
  const uncatCollapsed = collapsed.has(UNCATEGORIZED_ID);

  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter}
      onDragStart={handleDragStart} onDragEnd={handleDragEnd} onDragCancel={handleDragCancel}>
      <SortableContext items={orderedItems.map((i) => i.id)} strategy={verticalListSortingStrategy}>
        <div className={`rounded-[4px] border ${C.borderLight} overflow-hidden ${C.bgWhite}`}>
          <table className="w-full border-collapse">
            <thead>
              <tr className={STYLE.tableHeaderRow}>
                <th className="w-8" />
                <th className={`text-left ${STYLE.tableHeaderCell} px-3`}>名称</th>
                <th className={`text-left ${STYLE.tableHeaderCell} px-3 w-56`}>備考</th>
                <th className={`text-center ${STYLE.tableHeaderCell} px-3 w-24 whitespace-nowrap`}>ステータス</th>
                <th className="w-20" />
              </tr>
            </thead>
            <tbody>
              {/* グループ別セクション */}
              {groups.map((group) => {
                const groupCats = categoriesByGroupId.map.get(group.id) ?? [];
                const isCollapsed = collapsed.has(group.id);
                return (
                  <Fragment key={group.id}>
                    {/* グループヘッダー行 */}
                    <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
                      <td colSpan={5} className="px-2 py-0">
                        <div className="flex items-center gap-1 h-11">
                          <button type="button" onClick={() => toggleCollapse(group.id)}
                            className={`${STYLE.iconBtn20} ${C.text35} ${C.hoverBgMedium} shrink-0`}>
                            <ChevronDown className={`${ICON.smXs} transition-transform duration-150`}
                              style={{ transform: isCollapsed ? "rotate(-90deg)" : "rotate(0deg)" }} />
                          </button>
                          <span className={`${ICON.dotMd} rounded-full shrink-0`} style={{ backgroundColor: group.color }} />
                          <button type="button" onClick={() => onGroupEdit(group)}
                            className={`text-sm font-medium ${C.text} ${C.hoverBgLight} px-1 rounded-[3px] transition-colors`}>
                            {group.name}
                          </button>
                          <span className={`text-xs ${C.text35} tabular-nums`}>{groupCats.length}</span>
                          {canEdit ? (
                            <div className="ml-auto flex items-center gap-1">
                              <button type="button" onClick={() => onGroupEdit(group)}
                                className={`flex items-center gap-1 text-xs ${C.text45}
                                  ${LAYOUT.inputCompact} ${C.hoverBgMedium} transition-colors`}>
                                <Pencil className={ICON.action} />編集
                              </button>
                              <button type="button" onClick={() => onCategoryAddInGroup(group.id)}
                                className={`flex items-center gap-1 text-xs ${C.text45}
                                  ${LAYOUT.inputCompact} ${C.hoverBgMedium} transition-colors`}>
                                <Plus className={ICON.action} />追加
                              </button>
                            </div>
                          ) : null}
                        </div>
                      </td>
                    </tr>
                    {/* 区分行 */}
                    {!isCollapsed ? (
                      groupCats.length > 0 ? (
                        groupCats.map((cat) => (
                          <SortableDataTableRow key={cat.id} id={cat.id} onClick={() => onCategoryEdit(cat)}>
                            <TableCell className={`font-medium text-sm ${C.text} pl-7`}>
                              {cat.name}
                            </TableCell>
                            <TableCell className={`text-sm ${C.text60} max-w-[220px] truncate`}>
                              {cat.description || "—"}
                            </TableCell>
                            <TableCell className="text-center">
                              <NotionStatusPill isActive={cat.isActive} />
                            </TableCell>
                            <TableCell className="p-0 text-right">
                              {canEdit ? <RowActionButton onClick={() => onCategoryEdit(cat)} /> : null}
                            </TableCell>
                          </SortableDataTableRow>
                        ))
                      ) : (
                        <tr className={`border-b ${C.borderLight}`}>
                          <td colSpan={5} className={`pl-10 py-2 text-sm ${C.text35} italic`}>
                            予約区分がありません
                          </td>
                        </tr>
                      )
                    ) : null}
                  </Fragment>
                );
              })}

              {/* 未分類セクション */}
              {uncatCats.length > 0 || groups.length === 0 ? (
                <Fragment key={UNCATEGORIZED_ID}>
                  <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
                    <td colSpan={5} className="px-2 py-0">
                      <div className="flex items-center gap-1 h-11">
                        <button type="button" onClick={() => toggleCollapse(UNCATEGORIZED_ID)}
                          className={`${STYLE.iconBtn20} ${C.text35} ${C.hoverBgMedium} shrink-0`}>
                          <ChevronDown className={`${ICON.smXs} transition-transform duration-150`}
                            style={{ transform: uncatCollapsed ? "rotate(-90deg)" : "rotate(0deg)" }} />
                        </button>
                        <span className={`${ICON.dotMd} rounded-full shrink-0`} style={{ backgroundColor: PALETTE.grayMedium }} />
                        <span className={`text-sm font-medium ${C.text55}`}>未分類</span>
                        <span className={`text-xs ${C.text35} tabular-nums`}>{uncatCats.length}</span>
                        {canEdit ? (
                          <button type="button" onClick={() => onCategoryAddInGroup(undefined)}
                            className={`ml-auto flex items-center gap-1 text-xs ${C.text45}
                              ${LAYOUT.inputCompact} ${C.hoverBgMedium} transition-colors`}>
                            <Plus className={ICON.action} />追加
                          </button>
                        ) : null}
                      </div>
                    </td>
                  </tr>
                  {!uncatCollapsed ? (
                    uncatCats.map((cat) => (
                      <SortableDataTableRow key={cat.id} id={cat.id} onClick={() => onCategoryEdit(cat)}>
                        <TableCell className={`font-medium text-sm ${C.text} pl-7`}>
                          {cat.name}
                        </TableCell>
                        <TableCell className={`text-sm ${C.text60} max-w-[220px] truncate`}>
                          {cat.description || "—"}
                        </TableCell>
                        <TableCell className="text-center">
                          <NotionStatusPill isActive={cat.isActive} />
                        </TableCell>
                        <TableCell className="p-0 text-right">
                          {canEdit ? <RowActionButton onClick={() => onCategoryEdit(cat)} /> : null}
                        </TableCell>
                      </SortableDataTableRow>
                    ))
                  ) : null}
                </Fragment>
              ) : null}
            </tbody>
          </table>
        </div>
      </SortableContext>
    </DndContext>
  );
}

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
    toCreateRequest: (data) => ({
      name: data.name,
      color: data.color || undefined,
      is_active: data.isActive,
    }),
    toUpdateRequest: (data) => ({
      name: data.name,
      color: data.color || undefined,
      is_active: data.isActive,
    }),
  });

  const categorySave = useMasterSave<ReservationType, CategoryFormData, CreateReservationTypeRequest, UpdateReservationTypeRequest>({
    crud: categoryCrud,
    createMutation: createCategoryMutation,
    updateMutation: updateCategoryMutation,
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: (data) => ({
      name: data.name,
      description: data.description || undefined,
      is_active: true,
      group_id: data.groupId ? Number(data.groupId) : undefined,
      reservation_display_name: data.reservationDisplayName || undefined,
      duration_minutes: data.durationMinutes,
      short_name: data.shortName || undefined,
      reservation_visible: data.reservationVisible,
      reservation_comment: data.reservationComment || undefined,
      reservation_image_url: data.reservationImageUrl || undefined,
      show_short_name: data.showShortName,
      reservation_day_option: data.reservationDayOption as "none" | "weekday" | "saturday" | "anyday",
      is_internal: data.isInternal,
    }),
    toUpdateRequest: (data) => ({
      name: data.name,
      description: data.description || undefined,
      is_active: data.isActive,
      group_id: data.groupId ? Number(data.groupId) : undefined,
      reservation_display_name: data.reservationDisplayName || undefined,
      duration_minutes: data.durationMinutes,
      short_name: data.shortName || undefined,
      reservation_visible: data.reservationVisible,
      reservation_comment: data.reservationComment || undefined,
      reservation_image_url: data.reservationImageUrl || undefined,
      show_short_name: data.showShortName,
      reservation_day_option: data.reservationDayOption as "none" | "weekday" | "saturday" | "anyday",
      is_internal: data.isInternal,
    }),
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
              <GroupedTable
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
