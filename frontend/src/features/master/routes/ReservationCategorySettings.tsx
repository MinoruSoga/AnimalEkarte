import { memo, useState, useCallback, useRef, useEffect, useMemo, useTransition, useDeferredValue } from "react";
import { useNavigate } from "react-router";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { Activity, Layers, MessageCircle, Pencil, Plus } from "lucide-react";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { NotionFilter } from "@/components/shared/NotionFilter/NotionFilter";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { PropertyRow, StatusToggleButton, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, ICON, LAYOUT, PALETTE } from "@/lib/design-tokens";
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

// ── 静的 SelectItem JSX (rendering-hoist-jsx) ──────────────────
const RESERVATION_DAY_OPTION_ITEMS = (
  <>
    <SelectItem value="none">制限なし</SelectItem>
    <SelectItem value="weekday">平日のみ</SelectItem>
    <SelectItem value="saturday">土曜含む</SelectItem>
    <SelectItem value="anyday">毎日</SelectItem>
  </>
);

// ── カラム定義 ─────────────────────────────────────────────────
const CATEGORY_COLUMNS_ALL = [
  { header: "", className: "w-[32px]" },
  { header: "名称" },
  { header: "グループ", className: "w-[180px]" },
  { header: "備考", className: "w-[200px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

const CATEGORY_COLUMNS_FILTERED = [
  { header: "", className: "w-[32px]" },
  { header: "名称" },
  { header: "備考", className: "w-[200px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

// ─────────────────────────────────────────────────────────────────
// GroupSidePanel
// ─────────────────────────────────────────────────────────────────

interface GroupFormData {
  name: string;
  color: string;
  isActive: boolean;
}

const GroupSidePanel = memo(function GroupSidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly,
}: {
  item: ReservationCategoryGroup | null;
  onClose: () => void;
  onSave: (d: GroupFormData) => void;
  onDeleteRequest?: (i: ReservationCategoryGroup) => void;
  readOnly?: boolean;
}) {
  const [f, setF] = useState<GroupFormData>(() => ({
    name: item?.name ?? "",
    color: item?.color ?? PALETTE.pickerDefaultBlue,
    isActive: item?.isActive ?? true,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  const handleTitleChange = useCallback((v: string) => {
    setF((p) => ({ ...p, name: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleColorPickerChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setF((p) => ({ ...p, color: e.target.value }));
    setIsDirty(true);
  }, []);

  const handleColorInputChange = useCallback((v: string) => {
    setF((p) => ({ ...p, color: v }));
    setIsDirty(true);
  }, []);

  const handleToggleActive = useCallback(() => {
    setF((p) => ({ ...p, isActive: !p.isActive }));
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(() => {
    if (!f.name.trim()) { setNameError("名称を入力してください"); return; }
    setNameError("");
    onSave(f);
    setIsDirty(false);
  }, [f, onSave]);

  const handleClose = useCallback(() => { setIsDirty(false); onClose(); }, [onClose]);

  return (
    <MasterSidePanel isNew={item === null} title={f.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={readOnly ? undefined : handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Layers className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty} titleError={nameError} titleMaxLength={100} readOnly={readOnly}>
      <StatusToggleButton isActive={f.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="カラー">
        <div className="flex items-center gap-2">
          <input type="color" value={f.color} onChange={handleColorPickerChange}
            className="w-7 h-7 rounded cursor-pointer border-0 bg-transparent p-0" />
          <PropertyInput value={f.color} onChange={handleColorInputChange} placeholder="#3B82F6" />
        </div>
      </PropertyRow>
    </MasterSidePanel>
  );
});

// ─────────────────────────────────────────────────────────────────
// CategorySidePanel（色設定なし・グループから継承）
// ─────────────────────────────────────────────────────────────────

interface CategoryFormData {
  name: string;
  description: string;
  isActive: boolean;
  groupId: string | undefined;
  // LINE予約設定
  reservationDisplayName: string;
  durationMinutes: number;
  shortName: string;
  reservationVisible: boolean;
  reservationComment: string;
  reservationImageUrl: string;
  showShortName: boolean;
  reservationDayOption: string;
  isInternal: boolean;
}

interface GroupOption { id: string; name: string; color: string; }

const CategorySidePanel = memo(function CategorySidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly, groups, defaultGroupId,
}: {
  item: ReservationCategory | null;
  onClose: () => void;
  onSave: (d: CategoryFormData) => void;
  onDeleteRequest?: (i: ReservationCategory) => void;
  readOnly?: boolean;
  groups: GroupOption[];
  defaultGroupId?: string;
}) {
  const [f, setF] = useState<CategoryFormData>(() => ({
    name: item?.name ?? "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
    groupId: item?.groupId ?? defaultGroupId,
    reservationDisplayName: item?.reservationDisplayName ?? "",
    durationMinutes: item?.durationMinutes ?? 15,
    shortName: item?.shortName ?? "",
    reservationVisible: item?.reservationVisible ?? true,
    reservationComment: item?.reservationComment ?? "",
    reservationImageUrl: item?.reservationImageUrl ?? "",
    showShortName: item?.showShortName ?? false,
    reservationDayOption: item?.reservationDayOption ?? "none",
    isInternal: item?.isInternal ?? false,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  const handleTitleChange = useCallback((v: string) => {
    setF((p) => ({ ...p, name: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleDescriptionChange = useCallback((v: string) => {
    setF((p) => ({ ...p, description: v }));
    setIsDirty(true);
  }, []);

  const handleToggleActive = useCallback(() => {
    setF((p) => ({ ...p, isActive: !p.isActive }));
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(() => {
    if (!f.name.trim()) { setNameError("名称を入力してください"); return; }
    setNameError("");
    onSave(f);
    setIsDirty(false);
  }, [f, onSave]);

  const handleClose = useCallback(() => { setIsDirty(false); onClose(); }, [onClose]);

  return (
    <MasterSidePanel isNew={item === null} title={f.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={readOnly ? undefined : handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Activity className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty} titleError={nameError} titleMaxLength={100} readOnly={readOnly}>
      <StatusToggleButton isActive={f.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="グループ">
        <Select
          value={f.groupId ?? "none"}
          onValueChange={(v) => { setF((p) => ({ ...p, groupId: v === "none" ? undefined : v })); setIsDirty(true); }}
        >
          <SelectTrigger className="h-8 text-sm">
            <SelectValue placeholder="グループを選択" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="none">未分類</SelectItem>
            {groups.map((g) => (
              <SelectItem key={g.id} value={g.id}>
                <div className="flex items-center gap-2">
                  <span className="size-2.5 rounded-full shrink-0" style={{ backgroundColor: g.color }} />
                  {g.name}
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </PropertyRow>
      <PropertyRow label="備考">
        <PropertyInput value={f.description} onChange={handleDescriptionChange} placeholder="補足情報など" />
      </PropertyRow>

      {/* ── LINE予約設定 ───────────────────────────── */}
      <div className="mt-4 border-t pt-4">
        <div className="flex items-center gap-1.5 mb-3">
          <MessageCircle className="size-3.5" style={{ color: PALETTE.lineGreen }} />
          <p className="text-xs font-medium text-muted-foreground">LINE予約設定</p>
        </div>
        <PropertyRow label="LINE表示名">
          <PropertyInput value={f.reservationDisplayName}
            onChange={(v) => { setF((p) => ({ ...p, reservationDisplayName: v })); setIsDirty(true); }}
            placeholder={f.name || "空欄なら名称を使用"} />
        </PropertyRow>
        <PropertyRow label="予約ページに表示">
          <Switch checked={f.reservationVisible}
            onCheckedChange={(v) => { setF((p) => ({ ...p, reservationVisible: v })); setIsDirty(true); }} />
        </PropertyRow>
        <PropertyRow label="内部サービス">
          <Switch checked={f.isInternal}
            onCheckedChange={(v) => { setF((p) => ({ ...p, isInternal: v })); setIsDirty(true); }} />
        </PropertyRow>
        <PropertyRow label="所要時間（分）">
          <input type="number" min={5} max={480}
            className="w-20 rounded border px-2 py-1 text-sm"
            value={f.durationMinutes}
            onChange={(e) => { setF((p) => ({ ...p, durationMinutes: Number(e.target.value) || 15 })); setIsDirty(true); }} />
        </PropertyRow>
        <PropertyRow label="略称">
          <PropertyInput value={f.shortName}
            onChange={(v) => { setF((p) => ({ ...p, shortName: v })); setIsDirty(true); }}
            placeholder="LINE表示用の略称" />
        </PropertyRow>
        <PropertyRow label="略称を使用">
          <Switch checked={f.showShortName}
            onCheckedChange={(v) => { setF((p) => ({ ...p, showShortName: v })); setIsDirty(true); }} />
        </PropertyRow>
        <PropertyRow label="画像URL">
          <PropertyInput value={f.reservationImageUrl}
            onChange={(v) => { setF((p) => ({ ...p, reservationImageUrl: v })); setIsDirty(true); }}
            placeholder="https://..." />
        </PropertyRow>
        <PropertyRow label="予約可能曜日">
          <Select value={f.reservationDayOption}
            onValueChange={(v) => { setF((p) => ({ ...p, reservationDayOption: v })); setIsDirty(true); }}>
            <SelectTrigger className="h-8 text-sm"><SelectValue /></SelectTrigger>
            <SelectContent>{RESERVATION_DAY_OPTION_ITEMS}</SelectContent>
          </Select>
        </PropertyRow>
        <PropertyRow label="LINE説明文">
          <PropertyInput value={f.reservationComment}
            onChange={(v) => { setF((p) => ({ ...p, reservationComment: v })); setIsDirty(true); }}
            placeholder="LINE予約画面に表示する説明" />
        </PropertyRow>
      </div>
    </MasterSidePanel>
  );
});

// ─────────────────────────────────────────────────────────────────
// GroupSidebar（左ペイン：グループ選択）
// ─────────────────────────────────────────────────────────────────

interface GroupSidebarProps {
  groups: ReservationCategoryGroup[];
  selectedGroupId: string | null;
  totalCount: number;
  categoryCountByGroupId: Map<string, number>;
  onSelect: (id: string | null) => void;
  onEdit: (g: ReservationCategoryGroup) => void;
  onAdd: () => void;
  canEdit: boolean;
}

const GroupSidebar = memo(function GroupSidebar({
  groups, selectedGroupId, totalCount, categoryCountByGroupId,
  onSelect, onEdit, onAdd, canEdit,
}: GroupSidebarProps) {
  return (
    <div className={`w-52 shrink-0 flex flex-col border-r ${C.borderLight}`}>
      {/* header */}
      <div className={`flex items-center justify-between px-3 h-10 border-b ${C.borderLight} shrink-0`}>
        <span className={`text-xs font-semibold ${C.text50} uppercase tracking-wide`}>グループ</span>
        {canEdit ? (
          <button
            type="button"
            onClick={onAdd}
            className={`size-6 rounded flex items-center justify-center hover:bg-accent ${C.text50} transition-colors`}
            title="グループを追加"
          >
            <Plus className="size-3.5" />
          </button>
        ) : null}
      </div>

      {/* 全て */}
      <button
        type="button"
        onClick={() => onSelect(null)}
        className={`flex items-center gap-2 w-full px-3 py-2 text-sm transition-colors
          ${selectedGroupId === null
            ? `${C.text} font-medium bg-accent`
            : `${C.text60} hover:bg-accent`
          }`}
      >
        <span className="size-2 rounded-full shrink-0 bg-gray-300" />
        <span className="flex-1 text-left">全て</span>
        <span className={`text-xs tabular-nums ${C.text40}`}>{totalCount}</span>
      </button>

      {groups.length > 0 ? (
        <div className={`mx-3 my-0.5 h-px ${C.borderLight}`} />
      ) : null}

      {/* group rows */}
      <div className="flex-1 overflow-y-auto">
        {groups.map((group) => {
          const count = categoryCountByGroupId.get(group.id) ?? 0;
          const isSelected = selectedGroupId === group.id;
          return (
            <div
              key={group.id}
              role="button"
              tabIndex={0}
              onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") onSelect(group.id); }}
              className={`group relative flex items-center gap-2 w-full px-3 py-2 cursor-pointer transition-colors select-none
                ${isSelected ? `${C.text} font-medium bg-accent` : `${C.text60} hover:bg-accent`}`}
              onClick={() => onSelect(group.id)}
            >
              <span className="size-2 rounded-full shrink-0" style={{ backgroundColor: group.color }} />
              <span className="flex-1 truncate text-sm">{group.name}</span>
              <span className={`text-xs tabular-nums ${C.text40}`}>{count}</span>
              {canEdit ? (
                <button
                  type="button"
                  onClick={(e) => { e.stopPropagation(); onEdit(group); }}
                  className={`opacity-0 group-hover:opacity-100 absolute right-1.5 top-1/2 -translate-y-1/2
                    size-5 rounded flex items-center justify-center hover:bg-background transition-opacity ${C.text50}`}
                  title="グループを編集"
                >
                  <Pencil className="size-3" />
                </button>
              ) : null}
            </div>
          );
        })}
        {groups.length === 0 ? (
          <p className={`px-3 py-3 text-xs ${C.text40}`}>グループがありません</p>
        ) : null}
      </div>
    </div>
  );
});

// ─────────────────────────────────────────────────────────────────
// CategoryList（右ペイン：予約区分一覧）
// ─────────────────────────────────────────────────────────────────

interface CategoryListProps {
  selectedGroupId: string | null;
  onEdit: (item: ReservationCategory) => void;
  canEdit: boolean;
  groupColorById: Map<string, string>;
  groupNameById: Map<string, string>;
}

function CategoryList({
  selectedGroupId, onEdit, canEdit, groupColorById, groupNameById,
}: CategoryListProps) {
  const { data: rawCategories = [] } = useGetReservationCategories();
  const reorderMutation = useReorderReservationCategories();
  const [searchTerm, setSearchTerm] = useState("");
  const [activeFilters, setActiveFilters] = useState<ActiveFilter[]>([]);
  const deferredSearch = useDeferredValue(searchTerm);
  const resetOrderRef = useRef<() => void>(() => {});

  const handleReorder = useCallback((newIds: string[]) => {
    reorderMutation.mutate({ ids: newIds.map(Number) }, {
      onError: (error: unknown) => { resetOrderRef.current(); handleApiError(error, "並び替え"); },
    });
  }, [reorderMutation]);

  const filteredRaw = useMemo(() => {
    let items = rawCategories;
    // グループフィルタ
    if (selectedGroupId !== null) {
      items = items.filter((i) => i.groupId === selectedGroupId);
    }
    // ステータスフィルタ
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
  }, [rawCategories, selectedGroupId, activeFilters, deferredSearch]);

  const { orderedItems, sensors, handleDragStart, handleDragEnd, handleDragCancel, resetOrder } =
    useSortableList({ items: filteredRaw, onReorder: handleReorder });
  useEffect(() => { resetOrderRef.current = resetOrder; }, [resetOrder]);

  // グループ選択時はグループ列を非表示
  const showGroupColumn = selectedGroupId === null;
  const columns = showGroupColumn ? CATEGORY_COLUMNS_ALL : CATEGORY_COLUMNS_FILTERED;

  return (
    <div className="flex flex-col flex-1 min-w-0 gap-4 px-4 py-4">
      <NotionFilter
        properties={[MASTER_STATUS_FILTER]}
        activeFilters={activeFilters}
        onFilterChange={setActiveFilters}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        searchPlaceholder="予約区分名で検索..."
        count={orderedItems.length}
      />
      <DndContext sensors={sensors} collisionDetection={closestCenter}
        onDragStart={handleDragStart} onDragEnd={handleDragEnd} onDragCancel={handleDragCancel}>
        <SortableContext items={orderedItems.map((i) => i.id)} strategy={verticalListSortingStrategy}>
          <DataTable columns={columns} data={orderedItems} emptyMessage="予約区分が登録されていません"
            renderRow={(item) => {
              const groupColor = item.groupId ? groupColorById.get(item.groupId) : undefined;
              const groupName = item.groupId ? groupNameById.get(item.groupId) : undefined;
              return (
                <SortableDataTableRow key={item.id} id={item.id} onClick={() => onEdit(item)}>
                  <TableCell className={`font-medium text-base ${C.text}`}>
                    <div className="flex items-center gap-2">
                      <span className="size-3 rounded-full shrink-0"
                        style={{ backgroundColor: groupColor ?? "#D1D5DB" }} />
                      {item.name}
                    </div>
                  </TableCell>
                  {showGroupColumn ? (
                    <TableCell className={`text-base ${C.text70}`}>
                      {groupName ? (
                        <div className="flex items-center gap-1.5">
                          <span className="size-2 rounded-full shrink-0" style={{ backgroundColor: groupColor }} />
                          {groupName}
                        </div>
                      ) : (
                        <span className={C.text40}>未分類</span>
                      )}
                    </TableCell>
                  ) : null}
                  <TableCell className={`text-base ${C.text70} truncate max-w-[200px]`}>
                    {item.description || "-"}
                  </TableCell>
                  <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
                  <TableCell className="p-0 text-right">
                    {canEdit ? <RowActionButton onClick={() => onEdit(item)} /> : null}
                  </TableCell>
                </SortableDataTableRow>
              );
            }}
          />
        </SortableContext>
      </DndContext>
    </div>
  );
}

// ─────────────────────────────────────────────────────────────────
// ReservationCategorySettings（統合ページ・2カラムレイアウト）
// ─────────────────────────────────────────────────────────────────

export function ReservationCategorySettings() {
  const navigate = useNavigate();
  const { canCreate, canEdit, canDelete } = usePermission(ResourceMasterReservationCategory);

  const { data: groupsRaw = [] } = useGetReservationCategoryGroups();
  const { data: categoriesRaw = [] } = useGetReservationCategories();

  // 選択グループ (null = 全て)
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);

  // グループ選択肢
  const activeGroups = useMemo(
    () => groupsRaw.filter((g) => g.isActive).map((g) => ({ id: g.id, name: g.name, color: g.color })),
    [groupsRaw],
  );

  // グループ色・名称マップ (CategoryList へ渡す)
  const groupColorById = useMemo(
    () => new Map(groupsRaw.map((g) => [g.id, g.color])),
    [groupsRaw],
  );
  const groupNameById = useMemo(
    () => new Map(groupsRaw.map((g) => [g.id, g.name])),
    [groupsRaw],
  );

  // グループごとの区分件数
  const categoryCountByGroupId = useMemo(() => {
    const map = new Map<string, number>();
    for (const cat of categoriesRaw) {
      if (cat.groupId) {
        map.set(cat.groupId, (map.get(cat.groupId) ?? 0) + 1);
      }
    }
    return map;
  }, [categoriesRaw]);

  // 編集・削除ターゲット
  const [groupEditTarget, setGroupEditTarget] = useState<ReservationCategoryGroup | "new" | null>(null);
  const [groupPendingDelete, setGroupPendingDelete] = useState<ReservationCategoryGroup | null>(null);
  const [categoryEditTarget, setCategoryEditTarget] = useState<ReservationCategory | "new" | null>(null);
  const [categoryPendingDelete, setCategoryPendingDelete] = useState<ReservationCategory | null>(null);
  const [, startTransition] = useTransition();

  // ミューテーション
  const createGroupMutation = useCreateReservationCategoryGroup();
  const updateGroupMutation = useUpdateReservationCategoryGroup();
  const deleteGroupMutation = useDeleteReservationCategoryGroup();
  const createCategoryMutation = useCreateReservationCategory();
  const updateCategoryMutation = useUpdateReservationCategory();
  const deleteCategoryMutation = useDeleteReservationCategory();

  // ── グループ操作 ─────────────────────────────────────────────
  const handleGroupSelect = useCallback((id: string | null) => {
    setSelectedGroupId(id);
    setGroupEditTarget(null);
    setCategoryEditTarget(null);
  }, []);

  const handleGroupEdit = useCallback((item: ReservationCategoryGroup) => {
    setGroupEditTarget(item);
    setCategoryEditTarget(null);
  }, []);

  const handleGroupAdd = useCallback(() => {
    setGroupEditTarget("new");
    setCategoryEditTarget(null);
  }, []);

  const handleGroupDeleteRequest = useCallback((item: ReservationCategoryGroup) => {
    setGroupPendingDelete(item);
  }, []);

  // ── 予約区分操作 ─────────────────────────────────────────────
  const handleCategoryEdit = useCallback((item: ReservationCategory) => {
    setCategoryEditTarget(item);
    setGroupEditTarget(null);
  }, []);

  const handleCategoryNew = useCallback(() => {
    setCategoryEditTarget("new");
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
          name: data.name,
          color: data.color || undefined,
          is_active: data.isActive,
        };
        updateGroupMutation.mutate({ id: groupEditTarget.id, req }, {
          onSuccess: () => { toast.success("更新しました"); setGroupEditTarget(null); },
          onError: (error) => handleApiError(error, "更新"),
        });
      } else {
        const req: CreateReservationCategoryGroupRequest = {
          name: data.name,
          color: data.color || undefined,
          is_active: data.isActive,
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
        };
        updateCategoryMutation.mutate({ id: categoryEditTarget.id, req }, {
          onSuccess: () => { toast.success("更新しました"); setCategoryEditTarget(null); },
          onError: (error) => handleApiError(error, "更新"),
        });
      } else {
        const req: CreateReservationCategoryRequest = {
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
            maxWidth="max-w-full"
            headerAction={
              canCreate ? (
                <PrimaryButton onClick={handleCategoryNew}>
                  <Plus className={`mr-1.5 ${ICON.action}`} />
                  新規登録
                </PrimaryButton>
              ) : null
            }
          >
            {/* 2カラムレイアウト: PageLayout の px-3 py-5 を打ち消して端まで広げる */}
            <div className="-mx-3 -mt-5 flex flex-1 min-h-0">
              <GroupSidebar
                groups={groupsRaw}
                selectedGroupId={selectedGroupId}
                totalCount={categoriesRaw.length}
                categoryCountByGroupId={categoryCountByGroupId}
                onSelect={handleGroupSelect}
                onEdit={handleGroupEdit}
                onAdd={handleGroupAdd}
                canEdit={canEdit}
              />
              <CategoryList
                selectedGroupId={selectedGroupId}
                onEdit={handleCategoryEdit}
                canEdit={canEdit}
                groupColorById={groupColorById}
                groupNameById={groupNameById}
              />
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
            defaultGroupId={selectedGroupId ?? undefined}
          />
        ) : null}
      </div>

      <ConfirmDialog
        open={groupPendingDelete !== null}
        onClose={() => setGroupPendingDelete(null)}
        title="グループを削除しますか？"
        description={`「${groupPendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleGroupDeleteConfirm}
      />
      <ConfirmDialog
        open={categoryPendingDelete !== null}
        onClose={() => setCategoryPendingDelete(null)}
        title="予約区分を削除しますか？"
        description={`「${categoryPendingDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={handleCategoryDeleteConfirm}
      />
    </>
  );
}
