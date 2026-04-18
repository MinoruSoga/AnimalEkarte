import { useState, memo, useCallback, useEffect } from "react";
import { DndContext, DragOverlay, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { useSidePeekDirty } from "@/hooks/use-side-peek-dirty";
import { Plus, Building2, GripVertical } from "lucide-react";
import { Table, TableBody, TableHeader, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow, StatusToggleButton, MoneyInput, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { C, STYLE, LAYOUT, ICON } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import { useMasterCRUD } from "../hooks/use-master-crud";
import { useMasterSave } from "../hooks/use-master-save";
import { MasterCRUDPage } from "../components/MasterCRUDPage";
import { useGetAllCages, useCreateCage, useUpdateCage, useDeleteCage, useReorderCages } from "../api/cages";
import type { Cage, CageType, CageSize, CreateCageRequest, UpdateCageRequest } from "../api/cages";
import { ResourceMasterHospitalization } from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";

// ─── Constants ───
const CAGE_TYPE_LABELS: Record<CageType, string> = { icu: "ICU", dog: "犬舎", cat: "猫舎", general: "汎用" };
const CAGE_SIZE_LABELS: Record<CageSize, string> = { small: "小型", medium: "中型", large: "大型" };

const CAGE_TYPE_SELECT_ITEMS = (
  [{ value: "icu" as CageType, label: "ICU" }, { value: "dog" as CageType, label: "犬舎" },
   { value: "cat" as CageType, label: "猫舎" }, { value: "general" as CageType, label: "汎用" }] satisfies { value: CageType; label: string }[]
).map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>);

const CAGE_SIZE_SELECT_ITEMS = (
  [{ value: "small" as CageSize, label: "小型" }, { value: "medium" as CageSize, label: "中型" },
   { value: "large" as CageSize, label: "大型" }] satisfies { value: CageSize; label: string }[]
).map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>);

// ─── FormData ───
interface CageFormData { name: string; cageType: CageType; cageSize: CageSize; price: number; description: string; isActive: boolean; }

// ─── SidePanel ───
const CageSidePanel = memo(function CageSidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly, onDirtyChange,
}: { item: Cage | null; onClose: () => void; onSave: (d: CageFormData) => void; onDeleteRequest?: (i: Cage) => void; readOnly?: boolean; onDirtyChange?: (dirty: boolean) => void; }) {
  const [formData, setFormData] = useState<CageFormData>(() => ({
    name: item?.name ?? "", cageType: item?.cageType ?? "general", cageSize: item?.cageSize ?? "medium",
    price: item?.price ?? 0, description: item?.description ?? "", isActive: item?.isActive ?? true,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  // BUG-380
  useEffect(() => { onDirtyChange?.(isDirty); }, [isDirty, onDirtyChange]);
  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleTitleChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: v }));
    if (v.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleCageTypeChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, cageType: v as CageType }));
  }, [setFormDataDirty]);

  const handleCageSizeChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, cageSize: v as CageSize }));
  }, [setFormDataDirty]);

  const handlePriceChange = useCallback((v: number) => {
    setFormDataDirty((prev) => ({ ...prev, price: v }));
  }, [setFormDataDirty]);

  const handleDescriptionChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, description: v }));
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleAction = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(formData);
    setIsDirty(false);
  }, [formData, onSave]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel isNew={item === null} title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose} action={readOnly ? undefined : handleAction} onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Building2 className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}>
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="エリア">
        <Select value={formData.cageType} onValueChange={handleCageTypeChange}>
          <SelectTrigger className={STYLE.selectCompact}><SelectValue placeholder="選択" /></SelectTrigger>
          <SelectContent>{CAGE_TYPE_SELECT_ITEMS}</SelectContent>
        </Select>
      </PropertyRow>
      <PropertyRow label="サイズ">
        <Select value={formData.cageSize} onValueChange={handleCageSizeChange}>
          <SelectTrigger className={STYLE.selectCompact}><SelectValue placeholder="選択" /></SelectTrigger>
          <SelectContent>{CAGE_SIZE_SELECT_ITEMS}</SelectContent>
        </Select>
      </PropertyRow>
      <MoneyInput value={formData.price} onChange={handlePriceChange} />
      <PropertyRow label="備考">
        <PropertyInput value={formData.description} onChange={handleDescriptionChange} placeholder="補足情報など" />
      </PropertyRow>
    </MasterSidePanel>
  );
});

// ─── DragOverlay ───
function CageRowOverlay({ cage }: { cage: Cage }) {
  return (
    <div className={`flex items-center h-12 ${C.bgWhite} border ${C.borderLight} rounded-[4px] shadow-[0_4px_16px_rgba(0,0,0,0.12)] cursor-grabbing`} style={{ width: "100%" }}>
      <div className={`w-8 shrink-0 flex items-center justify-center ${C.text50}`}><GripVertical className={ICON.action} /></div>
      <div className={`flex-1 min-w-0 text-base font-medium ${C.text} px-3`}>{cage.name}</div>
      <div className={`w-[100px] shrink-0 text-base ${C.text65}`}>{CAGE_TYPE_LABELS[cage.cageType]}</div>
      <div className={`w-[90px] shrink-0 text-base ${C.text65}`}>{CAGE_SIZE_LABELS[cage.cageSize]}</div>
      <div className={`w-[120px] shrink-0 text-right pr-4 font-mono text-base ${C.text}`}>{cage.price != null ? `¥${cage.price.toLocaleString()}` : "-"}</div>
      <div className="w-[90px] shrink-0 flex justify-center"><NotionStatusPill isActive={cage.isActive} /></div>
      <div className="w-[80px] shrink-0" />
    </div>
  );
}

// ─── Columns (custom table, not DataTable) ───
const TABLE_COLUMNS = [
  { key: "grip", className: "w-8 px-0" },
  { key: "name", label: "ケージ名", className: "pl-3" },
  { key: "type", label: "エリア", className: "w-[100px]" },
  { key: "size", label: "サイズ", className: "w-[90px]" },
  { key: "price", label: "単価(税込)", className: "w-[120px] text-right pr-4" },
  { key: "status", label: "ステータス", className: "w-[90px] text-center" },
  { key: "action", label: "操作", className: "w-[80px] text-right pr-2" },
];

// ─── Page ───
export function CageSettings() {
  const { canCreate, canEdit } = usePermission(ResourceMasterHospitalization);
  const { data } = useGetAllCages();
  const createMutation = useCreateCage();
  const updateMutation = useUpdateCage();
  const deleteMutation = useDeleteCage();
  const reorderMutation = useReorderCages();

  const dirty = useSidePeekDirty();
  const crud = useMasterCRUD<Cage>({ data, deleteMutation, entityLabel: "ケージ", dirtyGuard: dirty });
  const handleDirtyChange = useCallback((d: boolean) => { if (d) dirty.markDirty(); else dirty.markClean(); }, [dirty]);

  const { orderedItems: sortedCages, sensors, activeId, handleDragStart, handleDragCancel, handleDragEnd, resetOrder } =
    useSortableList({
      items: crud.filteredItems,
      onReorder: (newIds) => { reorderMutation.mutate({ ids: newIds.map(Number) }, { onSuccess: resetOrder }); },
    });

  const { handleSave } = useMasterSave<Cage, CageFormData, CreateCageRequest, UpdateCageRequest>({
    crud, createMutation, updateMutation,
    validate: (d) => (!d.name.trim() ? "ケージ名は必須です" : null),
    toCreateRequest: (d) => ({
      name: d.name, cage_type: d.cageType, cage_size: d.cageSize, price: d.price,
      description: d.description || undefined, is_active: d.isActive,
    }),
    toUpdateRequest: (d) => ({
      name: d.name, cage_type: d.cageType, cage_size: d.cageSize, price: d.price,
      description: d.description || undefined, is_active: d.isActive,
    }),
  });

  // CageSettings uses custom table (not DataTable) for DnD + DragOverlay + bottom "add" button
  return (
    <MasterCRUDPage title="ケージマスタ" icon={<Building2 className={`${ICON.page} ${C.text}`} />} resource={ResourceMasterHospitalization}
      entityLabel="ケージ" searchPlaceholder="ケージ名で検索..." emptyMessage="ケージが登録されていません"
      crud={crud} handleSave={handleSave}
      filterProperties={[MASTER_STATUS_FILTER]}
      columns={[]} renderRow={() => null}
      renderSidePanel={({ readOnly, ...props }) => <CageSidePanel key={props.item?.id ?? "new"} {...props} readOnly={readOnly} onDirtyChange={handleDirtyChange} />}
    >
      <div className={STYLE.tableContainer}>
        <div className="flex-1 overflow-auto relative">
          <DndContext sensors={sensors} collisionDetection={closestCenter}
            onDragStart={handleDragStart} onDragEnd={handleDragEnd} onDragCancel={handleDragCancel}>
            <Table>
              <TableHeader className="sticky top-0 z-10">
                <TableRow className={STYLE.tableHeaderRow}>
                  {TABLE_COLUMNS.map((col) => (
                    <TableHead key={col.key} className={`${STYLE.tableHeaderCell} ${col.className}`}>
                      {col.label ?? ""}
                    </TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {sortedCages.length === 0 ? (
                  <TableRow><TableCell colSpan={7} className={STYLE.tableEmpty}>ケージが登録されていません</TableCell></TableRow>
                ) : null}
                <SortableContext items={sortedCages.map((m) => m.id)} strategy={verticalListSortingStrategy}>
                  {sortedCages.map((item) => (
                    <SortableDataTableRow key={item.id} id={item.id} onClick={() => crud.handleEdit(item)}>
                      <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
                      <TableCell className={`text-base ${C.text70}`}>{CAGE_TYPE_LABELS[item.cageType] || item.cageType}</TableCell>
                      <TableCell className={`text-base ${C.text70}`}>{CAGE_SIZE_LABELS[item.cageSize] || item.cageSize}</TableCell>
                      <TableCell className={`text-right font-mono text-base ${C.text} pr-4`}>{item.price != null ? `¥${item.price.toLocaleString()}` : "-"}</TableCell>
                      <TableCell className="text-center"><NotionStatusPill isActive={item.isActive} /></TableCell>
                      <TableCell className="p-0 text-right pr-2">{canEdit ? <RowActionButton onClick={() => crud.handleEdit(item)} /> : null}</TableCell>
                    </SortableDataTableRow>
                  ))}
                </SortableContext>
              </TableBody>
            </Table>
            <DragOverlay dropAnimation={null}>
              {activeId ? (() => { const m = sortedCages.find((x) => x.id === activeId); return m ? <CageRowOverlay cage={m} /> : null; })() : null}
            </DragOverlay>
          </DndContext>
        </div>
        {canCreate ? (
          <button type="button" onClick={crud.handleNew}
            className={STYLE.inlineAddBtn}>
            <Plus className={`${ICON.xs}`} />新しいケージを追加...
          </button>
        ) : null}
      </div>
    </MasterCRUDPage>
  );
}
