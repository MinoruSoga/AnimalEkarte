import { useState, useCallback, useRef, useEffect, useMemo, useTransition, useDeferredValue, Fragment } from "react";
import { useNavigate } from "react-router";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { Activity, ChevronDown, Plus } from "lucide-react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { TableCell } from "@/components/ui/table";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { C, ICON, PALETTE, STYLE } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";
import { paths } from "@/config/paths";
import { usePermission } from "@/features/auth";
import { ResourceMasterReservationCategory } from "@/types/generated/models";
import {
  useGetReservationCategories,
  useCreateReservationCategory,
  useUpdateReservationCategory,
  useDeleteReservationCategory,
  useReorderReservationCategories,
} from "@/features/master/api/reservation-categories";
import type { ReservationCategory } from "@/features/master/api/reservation-categories";
import {
  useGetReservationCategoryGroups,
  useCreateReservationCategoryGroup,
  useUpdateReservationCategoryGroup,
  useDeleteReservationCategoryGroup,
} from "@/features/master/api/reservation-category-groups";
import type { ReservationCategoryGroup } from "@/features/master/api/reservation-category-groups";
import type { CreateReservationCategoryGroupRequest, UpdateReservationCategoryGroupRequest } from "@/features/master/api/reservation-category-groups";
import type { CreateReservationCategoryRequest, UpdateReservationCategoryRequest } from "@/types/reservation-category";
import { GroupSidePanel } from "./ReservationCategoryGroupSidePanel";
import type { GroupFormData } from "./ReservationCategoryGroupSidePanel";
import { CategorySidePanel } from "./ReservationCategorySidePanel";
import type { CategoryFormData } from "./ReservationCategorySidePanel";

// ─────────────────────────────────────────────────────────────────
// GroupedTable
// Notionライク：グループヘッダー（折りたたみ）＋インデントされた区分行
// ─────────────────────────────────────────────────────────────────

const UNCATEGORIZED_ID = "__uncategorized__";

interface GroupedTableProps {
  groups: ReservationCategoryGroup[];
  categories: ReservationCategory[];
  onCategoryEdit: (cat: ReservationCategory) => void;
  onGroupEdit: (group: ReservationCategoryGroup) => void;
  onCategoryAddInGroup: (groupId: string | undefined) => void;
  canEdit: boolean;
}

function GroupedTable({
  groups, categories, onCategoryEdit, onGroupEdit, onCategoryAddInGroup, canEdit,
}: GroupedTableProps) {
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set());
  const reorderMutation = useReorderReservationCategories();
  const resetOrderRef = useRef<() => void>(() => {});

  const handleReorder = useCallback((newIds: string[]) => {
    reorderMutation.mutate({ ids: newIds.map(Number) }, {
      onError: (error: unknown) => { resetOrderRef.current(); handleApiError(error, "並び替え"); },
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
    const map = new Map<string, ReservationCategory[]>();
    const uncat: ReservationCategory[] = [];
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
        <div className={`rounded-[4px] border ${C.borderLight} overflow-hidden bg-white`}>
          <table className="w-full border-collapse">
            <thead>
              <tr className={STYLE.tableHeaderRow}>
                <th className="w-8" />
                <th className={`text-left ${STYLE.tableHeaderCell} px-3`}>名称</th>
                <th className={`text-left ${STYLE.tableHeaderCell} px-3 w-56`}>備考</th>
                <th className={`text-center ${STYLE.tableHeaderCell} px-3 w-24`}>ステータス</th>
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
                    <tr className={`border-b ${C.borderLight} ${C.bgPage} group/grouprow`}>
                      <td colSpan={5} className="px-2 py-0">
                        <div className="flex items-center gap-1 h-8">
                          <button type="button" onClick={() => toggleCollapse(group.id)}
                            className={`size-5 flex items-center justify-center rounded ${C.text35} ${C.hoverBgMedium} transition-colors shrink-0`}>
                            <ChevronDown className="size-3.5 transition-transform duration-150"
                              style={{ transform: isCollapsed ? "rotate(-90deg)" : "rotate(0deg)" }} />
                          </button>
                          <span className="size-2.5 rounded-full shrink-0" style={{ backgroundColor: group.color }} />
                          <button type="button" onClick={() => onGroupEdit(group)}
                            className={`text-sm font-medium ${C.text} ${C.hoverBgLight} px-1 rounded-[3px] transition-colors`}>
                            {group.name}
                          </button>
                          <span className={`text-xs ${C.text35} tabular-nums`}>{groupCats.length}</span>
                          {canEdit ? (
                            <button type="button" onClick={() => onCategoryAddInGroup(group.id)}
                              className={`ml-auto flex items-center gap-1 text-xs ${C.text45} opacity-0 group-hover/grouprow:opacity-100
                                px-2 py-0.5 rounded-[3px] ${C.hoverBgMedium} transition-all`}>
                              <Plus className="size-3" />追加
                            </button>
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
                  <tr className={`border-b ${C.borderLight} ${C.bgPage} group/grouprow`}>
                    <td colSpan={5} className="px-2 py-0">
                      <div className="flex items-center gap-1 h-8">
                        <button type="button" onClick={() => toggleCollapse(UNCATEGORIZED_ID)}
                          className={`size-5 flex items-center justify-center rounded ${C.text35} ${C.hoverBgMedium} transition-colors shrink-0`}>
                          <ChevronDown className="size-3.5 transition-transform duration-150"
                            style={{ transform: uncatCollapsed ? "rotate(-90deg)" : "rotate(0deg)" }} />
                        </button>
                        <span className="size-2.5 rounded-full shrink-0" style={{ backgroundColor: PALETTE.grayMedium }} />
                        <span className={`text-sm font-medium ${C.text55}`}>未分類</span>
                        <span className={`text-xs ${C.text35} tabular-nums`}>{uncatCats.length}</span>
                        {canEdit ? (
                          <button type="button" onClick={() => onCategoryAddInGroup(undefined)}
                            className={`ml-auto flex items-center gap-1 text-xs ${C.text45} opacity-0 group-hover/grouprow:opacity-100
                              px-2 py-0.5 rounded-[3px] ${C.hoverBgMedium} transition-all`}>
                            <Plus className="size-3" />追加
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
// ReservationCategorySettings
// ─────────────────────────────────────────────────────────────────

export function ReservationCategorySettings() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterReservationCategory);

  const { data: groupsRaw = [] } = useGetReservationCategoryGroups();
  const { data: categoriesRaw = [] } = useGetReservationCategories();

  // グループ選択肢（有効のみ）
  const activeGroups = useMemo(
    () => groupsRaw.filter((g) => g.isActive).map((g) => ({ id: g.id, name: g.name, color: g.color })),
    [groupsRaw],
  );

  // 検索・フィルタ
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);

  const filteredCategories = useMemo(() => {
    let items = categoriesRaw;
    for (const f of activeFilters) {
      if (f.key === "status" && typeof f.value === "string") {
        const want = f.value === "active";
        items = items.filter((i) => (f.condition === "is" ? i.isActive === want : i.isActive !== want));
      }
    }
    if (deferredSearch) {
      const lower = deferredSearch.toLowerCase();
      items = items.filter((i) => i.name.toLowerCase().includes(lower));
    }
    return items;
  }, [categoriesRaw, activeFilters, deferredSearch]);

  // 編集・削除ターゲット
  const [groupEditTarget, setGroupEditTarget] = useState<ReservationCategoryGroup | "new" | null>(null);
  const [groupPendingDelete, setGroupPendingDelete] = useState<ReservationCategoryGroup | null>(null);
  const [categoryEditTarget, setCategoryEditTarget] = useState<ReservationCategory | "new" | null>(null);
  const [categoryDefaultGroupId, setCategoryDefaultGroupId] = useState<string | undefined>(undefined);
  const [categoryPendingDelete, setCategoryPendingDelete] = useState<ReservationCategory | null>(null);
  const [, startTransition] = useTransition();

  // ミューテーション
  const createGroupMutation = useCreateReservationCategoryGroup();
  const updateGroupMutation = useUpdateReservationCategoryGroup();
  const deleteGroupMutation = useDeleteReservationCategoryGroup();
  const createCategoryMutation = useCreateReservationCategory();
  const updateCategoryMutation = useUpdateReservationCategory();
  const deleteCategoryMutation = useDeleteReservationCategory();

  // ── ハンドラ ─────────────────────────────────────────────────
  const handleGroupEdit = useCallback((group: ReservationCategoryGroup) => {
    setGroupEditTarget(group);
    setCategoryEditTarget(null);
  }, []);

  const handleGroupAdd = useCallback(() => {
    setGroupEditTarget("new");
    setCategoryEditTarget(null);
  }, []);

  const handleGroupDeleteRequest = useCallback((item: ReservationCategoryGroup) => {
    setGroupPendingDelete(item);
  }, []);

  const handleCategoryEdit = useCallback((cat: ReservationCategory) => {
    setCategoryEditTarget(cat);
    setGroupEditTarget(null);
    setCategoryDefaultGroupId(undefined);
  }, []);

  const handleCategoryAddInGroup = useCallback((groupId: string | undefined) => {
    setCategoryEditTarget("new");
    setCategoryDefaultGroupId(groupId);
    setGroupEditTarget(null);
  }, []);

  const handleCategoryDeleteRequest = useCallback((item: ReservationCategory) => {
    setCategoryPendingDelete(item);
  }, []);

  // ── グループ保存 ─────────────────────────────────────────────
  const handleGroupSave = useCallback((data: GroupFormData) => {
    startTransition(() => {
      if (groupEditTarget !== null && groupEditTarget !== "new") {
        const req: UpdateReservationCategoryGroupRequest = {
          name: data.name, color: data.color || undefined, is_active: data.isActive,
        };
        updateGroupMutation.mutate({ id: groupEditTarget.id, req }, {
          onSuccess: () => { toast.success("更新しました"); setGroupEditTarget(null); },
          onError: (error) => handleApiError(error, "更新"),
        });
      } else {
        const req: CreateReservationCategoryGroupRequest = {
          name: data.name, color: data.color || undefined, is_active: data.isActive,
        };
        createGroupMutation.mutate(req, {
          onSuccess: () => { toast.success("登録しました"); setGroupEditTarget(null); },
          onError: (error) => handleApiError(error, "登録"),
        });
      }
    });
  }, [groupEditTarget, updateGroupMutation, createGroupMutation]);

  const handleGroupDeleteConfirm = useCallback(() => {
    if (!groupPendingDelete) return;
    startTransition(() => {
      deleteGroupMutation.mutate(groupPendingDelete.id, {
        onSuccess: () => { toast.success("削除しました"); setGroupPendingDelete(null); },
        onError: (error) => handleApiError(error, "削除"),
      });
    });
  }, [groupPendingDelete, deleteGroupMutation]);

  // ── 予約区分保存 ─────────────────────────────────────────────
  const handleCategorySave = useCallback((data: CategoryFormData) => {
    startTransition(() => {
      if (categoryEditTarget !== null && categoryEditTarget !== "new") {
        const req: UpdateReservationCategoryRequest = {
          name: data.name, description: data.description || undefined,
          is_active: data.isActive, group_id: data.groupId ? Number(data.groupId) : undefined,
          reservation_display_name: data.reservationDisplayName || undefined,
          duration_minutes: data.durationMinutes, short_name: data.shortName || undefined,
          reservation_visible: data.reservationVisible,
          reservation_comment: data.reservationComment || undefined,
          reservation_image_url: data.reservationImageUrl || undefined,
          show_short_name: data.showShortName,
          reservation_day_option: data.reservationDayOption as "none" | "weekday" | "saturday" | "anyday",
          is_internal: data.isInternal,
        };
        updateCategoryMutation.mutate({ id: categoryEditTarget.id, req }, {
          onSuccess: () => { toast.success("更新しました"); setCategoryEditTarget(null); },
          onError: (error) => handleApiError(error, "更新"),
        });
      } else {
        const req: CreateReservationCategoryRequest = {
          name: data.name, description: data.description || undefined,
          is_active: true, group_id: data.groupId ? Number(data.groupId) : undefined,
          reservation_display_name: data.reservationDisplayName || undefined,
          duration_minutes: data.durationMinutes, short_name: data.shortName || undefined,
          reservation_visible: data.reservationVisible,
          reservation_comment: data.reservationComment || undefined,
          reservation_image_url: data.reservationImageUrl || undefined,
          show_short_name: data.showShortName,
          reservation_day_option: data.reservationDayOption as "none" | "weekday" | "saturday" | "anyday",
          is_internal: data.isInternal,
        };
        createCategoryMutation.mutate(req, {
          onSuccess: () => { toast.success("登録しました"); setCategoryEditTarget(null); },
          onError: (error) => handleApiError(error, "登録"),
        });
      }
    });
  }, [categoryEditTarget, updateCategoryMutation, createCategoryMutation]);

  const handleCategoryDeleteConfirm = useCallback(() => {
    if (!categoryPendingDelete) return;
    startTransition(() => {
      deleteCategoryMutation.mutate(categoryPendingDelete.id, {
        onSuccess: () => { toast.success("削除しました"); setCategoryPendingDelete(null); },
        onError: (error) => handleApiError(error, "削除"),
      });
    });
  }, [categoryPendingDelete, deleteCategoryMutation]);

  const groupPanelItem = groupEditTarget === "new" ? null : (groupEditTarget ?? null);
  const categoryPanelItem = categoryEditTarget === "new" ? null : (categoryEditTarget ?? null);

  return (
    <>
      <div className="flex h-full">
        <div className="flex-1 min-w-0">
          <PageLayout
            title="予約区分マスタ"
            icon={<Activity className={`${ICON.page} ${C.text}`} />}
            resource={ResourceMasterReservationCategory}
            onBack={() => navigate(paths.settings.getHref())}
            maxWidth="max-w-4xl"
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
                activeFilters={activeFilters}
                onFilterChange={setActiveFilters}
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                searchPlaceholder="予約区分名で検索..."
                count={filteredCategories.length}
              />
              <GroupedTable
                groups={groupsRaw}
                categories={filteredCategories}
                onCategoryEdit={handleCategoryEdit}
                onGroupEdit={handleGroupEdit}
                onCategoryAddInGroup={handleCategoryAddInGroup}
                canEdit={canEdit}
              />
              {canCreate ? (
                <button type="button" onClick={handleGroupAdd}
                  className={`flex items-center gap-1.5 text-sm ${C.text45} ${C.hoverText} ${C.hoverBgLight}
                    px-2 py-1.5 rounded-[3px] transition-colors w-fit`}>
                  <Plus className="size-3.5" />
                  グループを追加
                </button>
              ) : null}
            </div>
          </PageLayout>
        </div>

        {groupEditTarget !== null ? (
          <GroupSidePanel
            key={groupPanelItem ? String(groupPanelItem.id) : "new-group"}
            item={groupPanelItem}
            onClose={() => setGroupEditTarget(null)}
            onSave={handleGroupSave}
            onDeleteRequest={canDelete ? handleGroupDeleteRequest : undefined}
            readOnly={!canEdit}
          />
        ) : null}
        {categoryEditTarget !== null ? (
          <CategorySidePanel
            key={categoryPanelItem ? String(categoryPanelItem.id) : "new-category"}
            item={categoryPanelItem}
            onClose={() => setCategoryEditTarget(null)}
            onSave={handleCategorySave}
            onDeleteRequest={canDelete ? handleCategoryDeleteRequest : undefined}
            readOnly={!canEdit}
            groups={activeGroups}
            defaultGroupId={categoryDefaultGroupId}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={groupPendingDelete !== null}
        onClose={() => setGroupPendingDelete(null)}
        title="グループを削除しますか？"
        description={`「${groupPendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除" variant="destructive"
        onConfirm={handleGroupDeleteConfirm}
      />
      <ConfirmDialog
        open={categoryPendingDelete !== null}
        onClose={() => setCategoryPendingDelete(null)}
        title="予約区分を削除しますか？"
        description={`「${categoryPendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除" variant="destructive"
        onConfirm={handleCategoryDeleteConfirm}
      />
    </>
  );
}
