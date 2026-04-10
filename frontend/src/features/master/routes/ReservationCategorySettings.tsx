import { memo, useState, useCallback, useRef, useEffect } from "react";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { Activity, MessageCircle } from "lucide-react";
import { handleApiError } from "@/lib/handle-api-error";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { TableCell } from "@/components/ui/table";
import { DataTable } from "@/components/shared/DataTable/DataTable";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow, StatusToggleButton, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, LAYOUT, ICON, PALETTE } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import { usePermission } from "@/features/auth";
import { useGetReservationCategories, useCreateReservationCategory, useUpdateReservationCategory, useDeleteReservationCategory, useReorderReservationCategories } from "@/features/master/api/reservation-categories";
import type { ReservationCategory } from "@/features/master/api/reservation-categories";
import type { CreateReservationCategoryRequest, UpdateReservationCategoryRequest } from "@/types/reservation-category";
import { ResourceMasterReservationCategory } from "@/types/generated/models";

// rendering-hoist-jsx: 静的 SelectItem JSX をモジュール定数に巻き上げ
const RESERVATION_DAY_OPTION_ITEMS = (
  <>
    <SelectItem value="none">制限なし</SelectItem>
    <SelectItem value="weekday">平日のみ</SelectItem>
    <SelectItem value="saturday">土曜含む</SelectItem>
    <SelectItem value="anyday">毎日</SelectItem>
  </>
);

const COLUMNS = [
  { header: "", className: "w-[32px]" }, { header: "名称" },
  { header: "備考", className: "w-[240px]" },
  { header: "ステータス", className: "w-[100px]", align: "center" as const },
  { header: "操作", className: "w-[80px]", align: "right" as const },
];

interface ReservationCategoryFormData {
  name: string;
  description: string;
  color: string;
  isActive: boolean;
  // LINE予約用
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

const ReservationCategorySidePanel = memo(function ReservationCategorySidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly,
}: { item: ReservationCategory | null; onClose: () => void; onSave: (d: ReservationCategoryFormData) => void; onDeleteRequest?: (i: ReservationCategory) => void; readOnly?: boolean; }) {
  const [f, setF] = useState<ReservationCategoryFormData>(() => ({
    name: item?.name ?? "",
    description: item?.description ?? "",
    color: item?.color ?? PALETTE.pickerDefaultBlue,
    isActive: item?.isActive ?? true,
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

  const handleColorPickerChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setF((p) => ({ ...p, color: e.target.value }));
    setIsDirty(true);
  }, []);

  const handleColorInputChange = useCallback((v: string) => {
    setF((p) => ({ ...p, color: v }));
    setIsDirty(true);
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
    if (!f.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(f);
    setIsDirty(false);
  }, [f, onSave]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel isNew={item === null} title={f.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose} action={readOnly ? undefined : handleAction} onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Activity className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}>
      <StatusToggleButton isActive={f.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="カラー">
        <div className="flex items-center gap-2">
          <input type="color" value={f.color} onChange={handleColorPickerChange}
            className="w-7 h-7 rounded cursor-pointer border-0 bg-transparent p-0" />
          <PropertyInput value={f.color} onChange={handleColorInputChange} placeholder="#3B82F6" />
        </div>
      </PropertyRow>
      <PropertyRow label="備考">
        <PropertyInput value={f.description} onChange={handleDescriptionChange} placeholder="補足情報など" />
      </PropertyRow>

      {/* ── LINE予約設定 ──────────────────────────── */}
      <div className="mt-4 border-t pt-4">
        <div className="flex items-center gap-1.5 mb-3">
          <MessageCircle className="size-3.5" style={{ color: PALETTE.lineGreen }} />
          <p className="text-xs font-medium text-muted-foreground">LINE予約設定</p>
        </div>

        <PropertyRow label="LINE表示名">
          <PropertyInput
            value={f.reservationDisplayName}
            onChange={(v) => { setF((p) => ({ ...p, reservationDisplayName: v })); setIsDirty(true); }}
            placeholder={f.name || "空欄なら名称を使用"}
          />
        </PropertyRow>

        <PropertyRow label="予約ページに表示">
          <Switch
            checked={f.reservationVisible}
            onCheckedChange={(v) => { setF((p) => ({ ...p, reservationVisible: v })); setIsDirty(true); }}
          />
        </PropertyRow>

        <PropertyRow label="内部サービス">
          <Switch
            checked={f.isInternal}
            onCheckedChange={(v) => { setF((p) => ({ ...p, isInternal: v })); setIsDirty(true); }}
          />
        </PropertyRow>

        <PropertyRow label="所要時間（分）">
          <input
            type="number"
            min={5}
            max={480}
            className="w-20 rounded border px-2 py-1 text-sm"
            value={f.durationMinutes}
            onChange={(e) => { setF((p) => ({ ...p, durationMinutes: Number(e.target.value) || 15 })); setIsDirty(true); }}
          />
        </PropertyRow>

        <PropertyRow label="略称">
          <PropertyInput
            value={f.shortName}
            onChange={(v) => { setF((p) => ({ ...p, shortName: v })); setIsDirty(true); }}
            placeholder="LINE表示用の略称"
          />
        </PropertyRow>

        <PropertyRow label="略称を使用">
          <Switch
            checked={f.showShortName}
            onCheckedChange={(v) => { setF((p) => ({ ...p, showShortName: v })); setIsDirty(true); }}
          />
        </PropertyRow>

        <PropertyRow label="画像URL">
          <PropertyInput
            value={f.reservationImageUrl}
            onChange={(v) => { setF((p) => ({ ...p, reservationImageUrl: v })); setIsDirty(true); }}
            placeholder="https://..."
          />
        </PropertyRow>

        <PropertyRow label="予約可能曜日">
          <Select
            value={f.reservationDayOption}
            onValueChange={(v) => { setF((p) => ({ ...p, reservationDayOption: v })); setIsDirty(true); }}
          >
            <SelectTrigger className="h-8 text-sm">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{RESERVATION_DAY_OPTION_ITEMS}</SelectContent>
          </Select>
        </PropertyRow>

        <PropertyRow label="LINE説明文">
          <PropertyInput
            value={f.reservationComment}
            onChange={(v) => { setF((p) => ({ ...p, reservationComment: v })); setIsDirty(true); }}
            placeholder="LINE予約画面に表示する説明"
          />
        </PropertyRow>
      </div>
    </MasterSidePanel>
  );
});

export function ReservationCategorySettings() {
  const { canEdit } = usePermission(ResourceMasterReservationCategory);
  const { data } = useGetReservationCategories();
  const createMutation = useCreateReservationCategory();
  const updateMutation = useUpdateReservationCategory();
  const deleteMutation = useDeleteReservationCategory();
  const reorderMutation = useReorderReservationCategories();

  const crud = useMasterCRUD<ReservationCategory>({ data, deleteMutation, entityLabel: "予約区分" });

  const resetOrderRef = useRef<() => void>(() => {});
  const handleReorder = useCallback((newIds: string[]) => {
    reorderMutation.mutate({ ids: newIds.map(Number) }, {
      onError: (error: unknown) => { resetOrderRef.current(); handleApiError(error, "並び替え"); },
    });
  }, [reorderMutation]);

  const { orderedItems, sensors, handleDragStart, handleDragEnd, handleDragCancel, resetOrder } =
    useSortableList({ items: crud.filteredItems, onReorder: handleReorder });
  useEffect(() => { resetOrderRef.current = resetOrder; }, [resetOrder]);

  const { handleSave } = useMasterSave<ReservationCategory, ReservationCategoryFormData, CreateReservationCategoryRequest, UpdateReservationCategoryRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => (!d.name.trim() ? "名称は必須です" : null),
    toCreateRequest: (d) => ({
      name: d.name, description: d.description || undefined, color: d.color || undefined, is_active: true,
      reservation_display_name: d.reservationDisplayName || undefined,
      duration_minutes: d.durationMinutes, short_name: d.shortName || undefined,
      reservation_visible: d.reservationVisible, reservation_comment: d.reservationComment || undefined,
      reservation_image_url: d.reservationImageUrl || undefined, show_short_name: d.showShortName,
      reservation_day_option: d.reservationDayOption as "none" | "weekday" | "saturday" | "anyday",
      is_internal: d.isInternal,
    }),
    toUpdateRequest: (d) => ({
      name: d.name, description: d.description || undefined, color: d.color || undefined, is_active: d.isActive,
      reservation_display_name: d.reservationDisplayName || undefined,
      duration_minutes: d.durationMinutes, short_name: d.shortName || undefined,
      reservation_visible: d.reservationVisible, reservation_comment: d.reservationComment || undefined,
      reservation_image_url: d.reservationImageUrl || undefined, show_short_name: d.showShortName,
      reservation_day_option: d.reservationDayOption as "none" | "weekday" | "saturday" | "anyday",
      is_internal: d.isInternal,
    }),
  });

  return (
    <MasterCRUDPage title="予約区分マスタ" icon={<Activity className={`${ICON.page} ${C.text}`} />} resource={ResourceMasterReservationCategory}
      entityLabel="予約区分" searchPlaceholder="予約区分名で検索..." emptyMessage="予約区分が登録されていません"
      crud={crud} handleSave={handleSave} columns={COLUMNS}
      filterProperties={[MASTER_STATUS_FILTER]}
      renderRow={() => null}
      renderSidePanel={({ readOnly, ...props }) => <ReservationCategorySidePanel key={props.item?.id ?? "new"} {...props} readOnly={readOnly} />}
    >
      <DndContext sensors={sensors} collisionDetection={closestCenter}
        onDragStart={handleDragStart} onDragEnd={handleDragEnd} onDragCancel={handleDragCancel}>
        <SortableContext items={orderedItems.map((i) => i.id)} strategy={verticalListSortingStrategy}>
          <DataTable columns={COLUMNS} data={orderedItems} emptyMessage="予約区分が登録されていません"
            renderRow={(item) => (
              <SortableDataTableRow key={item.id} id={item.id} onClick={() => crud.handleEdit(item)}>
                <TableCell className={`font-medium text-base ${C.text}`}>
                  <div className="flex items-center gap-2">
                    <span className="size-3 rounded-full shrink-0" style={{ backgroundColor: item.color }} />
                    {item.name}
                  </div>
                </TableCell>
                <TableCell className={`text-base ${C.text70} truncate max-w-[240px]`}>{item.description || "-"}</TableCell>
                <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
                <TableCell className="p-0 text-right">{canEdit ? <RowActionButton onClick={() => crud.handleEdit(item)} /> : null}</TableCell>
              </SortableDataTableRow>
            )}
          />
        </SortableContext>
      </DndContext>
    </MasterCRUDPage>
  );
}
