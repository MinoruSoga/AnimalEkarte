import { useState, memo } from "react";
import { DndContext, DragOverlay, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useSortableList } from "@/hooks/use-sortable-list";
import { Plus, ShoppingBag, GripVertical } from "lucide-react";
import { Table, TableBody, TableHeader, TableRow, TableHead, TableCell } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { MoneyInput } from "@/components/shared/SidePeek/MoneyInput";
import { PropInput } from "@/components/shared/SidePeek/PropInput";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { C, STYLE, LAYOUT } from "@/lib/design-tokens";
import { MASTER_STATUS_FILTER } from "@/features/master/constants/styles";
import { useMasterCRUD } from "@/features/master/hooks/use-master-crud";
import { useMasterSave } from "@/features/master/hooks/use-master-save";
import { MasterCRUDPage } from "@/features/master/components/MasterCRUDPage";
import {
  useGetAllMerchandiseItems,
  useCreateMerchandiseItem,
  useUpdateMerchandiseItem,
  useDeleteMerchandiseItem,
  useReorderMerchandiseItems,
} from "../api/merchandise-items";
import type {
  FrontendMerchandiseItem,
  CreateMerchandiseItemRequest,
  UpdateMerchandiseItemRequest,
} from "../api/merchandise-items";

// ─── Constants ───

const CATEGORY_LABELS: Record<string, string> = {
  food: "フード",
  goods: "物販",
  other: "その他",
};

const CATEGORY_SELECT_ITEMS = [
  { value: "food", label: "フード" },
  { value: "goods", label: "物販" },
  { value: "other", label: "その他" },
].map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>);

const TAX_RATE_SELECT_ITEMS = [
  { value: "0.1", label: "10%" },
  { value: "0.08", label: "8%（軽減税率）" },
  { value: "0", label: "非課税" },
].map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>);

// ─── FormData ───

interface MerchandiseFormData {
  name: string;
  category: string;
  unitPrice: number;
  taxRate: number;
  isActive: boolean;
}

// ─── SidePanel ───

const MerchandiseSidePanel = memo(function MerchandiseSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
}: {
  item: FrontendMerchandiseItem | null;
  onClose: () => void;
  onSave: (d: MerchandiseFormData) => void;
  onDeleteRequest: (i: FrontendMerchandiseItem) => void;
}) {
  const [f, setF] = useState<MerchandiseFormData>(() => ({
    name: item?.name ?? "",
    category: item?.category ?? "goods",
    unitPrice: item?.unitPrice ?? 0,
    taxRate: item?.taxRate ?? 0.1,
    isActive: item?.isActive ?? true,
  }));

  return (
    <MasterSidePanel
      isNew={item === null}
      title={f.name}
      onTitleChange={(v) => setF((p) => ({ ...p, name: v }))}
      onClose={onClose}
      onSave={() => onSave(f)}
      onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
      icon={<ShoppingBag className={LAYOUT.pageIcon.innerIcon} />}
      titlePlaceholder="品目名"
    >
      <StatusToggleButton isActive={f.isActive} onToggle={() => setF((p) => ({ ...p, isActive: !p.isActive }))} />
      <PropertyRow label="カテゴリ">
        <Select value={f.category} onValueChange={(v) => setF((p) => ({ ...p, category: v }))}>
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>{CATEGORY_SELECT_ITEMS}</SelectContent>
        </Select>
      </PropertyRow>
      <MoneyInput value={f.unitPrice} onChange={(v) => setF((p) => ({ ...p, unitPrice: v }))} />
      <PropertyRow label="税率">
        <Select value={String(f.taxRate)} onValueChange={(v) => setF((p) => ({ ...p, taxRate: Number(v) }))}>
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>{TAX_RATE_SELECT_ITEMS}</SelectContent>
        </Select>
      </PropertyRow>
    </MasterSidePanel>
  );
});

// ─── DragOverlay ───

function MerchandiseRowOverlay({ item }: { item: FrontendMerchandiseItem }) {
  return (
    <div
      className={`flex items-center h-12 bg-white border ${C.borderLight} rounded-[4px] shadow-[0_4px_16px_rgba(0,0,0,0.12)] cursor-grabbing`}
      style={{ width: "100%" }}
    >
      <div className="w-8 shrink-0 flex items-center justify-center text-[#37352F]/50">
        <GripVertical className="size-4" />
      </div>
      <div className={`flex-1 min-w-0 text-base font-medium ${C.text} px-3`}>{item.name}</div>
      <div className={`w-[90px] shrink-0 text-base ${C.text70} text-center`}>
        {CATEGORY_LABELS[item.category] ?? item.category}
      </div>
      <div className={`w-[120px] shrink-0 text-right pr-4 font-mono text-base ${C.text}`}>
        ¥{item.unitPrice.toLocaleString()}
      </div>
      <div className={`w-[80px] shrink-0 text-base ${C.text70} text-center`}>
        {item.taxRate === 0.1 ? "10%" : item.taxRate === 0.08 ? "8%" : `${item.taxRate * 100}%`}
      </div>
      <div className="w-[90px] shrink-0 flex justify-center">
        <NotionStatusPill isActive={item.isActive} />
      </div>
      <div className="w-[80px] shrink-0" />
    </div>
  );
}

// ─── Columns ───

const TABLE_COLUMNS = [
  { key: "grip", className: "w-8 px-0" },
  { key: "name", label: "品目名", className: "pl-3" },
  { key: "category", label: "カテゴリ", className: "w-[90px] text-center" },
  { key: "price", label: "単価(税込)", className: "w-[120px] text-right pr-4" },
  { key: "taxRate", label: "税率", className: "w-[80px] text-center" },
  { key: "status", label: "ステータス", className: "w-[90px] text-center" },
  { key: "action", label: "操作", className: "w-[80px] text-right pr-2" },
];

// ─── Page ───

export function MerchandiseItemSettings() {
  const { data } = useGetAllMerchandiseItems();
  const createMutation = useCreateMerchandiseItem();
  const updateMutation = useUpdateMerchandiseItem();
  const deleteMutation = useDeleteMerchandiseItem();
  const reorderMutation = useReorderMerchandiseItems();

  const crud = useMasterCRUD<FrontendMerchandiseItem>({
    data,
    deleteMutation,
    entityLabel: "品目",
  });

  const {
    orderedItems: sortedItems,
    sensors,
    activeId,
    handleDragStart,
    handleDragCancel,
    handleDragEnd,
    resetOrder,
  } = useSortableList({
    items: crud.filteredItems,
    onReorder: (newIds) => {
      reorderMutation.mutate(
        { ids: newIds.map(Number) },
        { onSuccess: resetOrder },
      );
    },
  });

  const { handleSave } = useMasterSave<
    FrontendMerchandiseItem,
    MerchandiseFormData,
    CreateMerchandiseItemRequest,
    UpdateMerchandiseItemRequest
  >({
    crud,
    createMutation,
    updateMutation,
    validate: (d) => (!d.name.trim() ? "品目名は必須です" : null),
    toCreateRequest: (d) => ({
      name: d.name,
      category: d.category,
      unit_price: d.unitPrice,
      tax_rate: d.taxRate,
      is_active: d.isActive,
    }),
    toUpdateRequest: (d) => ({
      name: d.name,
      category: d.category,
      unit_price: d.unitPrice,
      tax_rate: d.taxRate,
      is_active: d.isActive,
    }),
  });

  return (
    <MasterCRUDPage
      title="物販・その他マスタ"
      icon={<ShoppingBag className="size-5 text-[#37352F]" />}
      entityLabel="品目"
      searchPlaceholder="品目名で検索..."
      emptyMessage="品目が登録されていません"
      crud={crud}
      handleSave={handleSave}
      filterProperties={[MASTER_STATUS_FILTER]}
      columns={[]}
      renderRow={() => null}
      renderSidePanel={(props) => (
        <MerchandiseSidePanel key={props.item?.id ?? "new"} {...props} />
      )}
    >
      <div className={STYLE.tableContainer}>
        <div className="flex-1 overflow-auto relative">
          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
            onDragCancel={handleDragCancel}
          >
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
                {sortedItems.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className={STYLE.tableEmpty}>
                      品目が登録されていません
                    </TableCell>
                  </TableRow>
                ) : null}
                <SortableContext items={sortedItems.map((m) => m.id)} strategy={verticalListSortingStrategy}>
                  {sortedItems.map((item) => (
                    <SortableDataTableRow key={item.id} id={item.id} onClick={() => crud.handleEdit(item)}>
                      <TableCell className={`font-medium text-base ${C.text}`}>{item.name}</TableCell>
                      <TableCell className={`text-base ${C.text70} text-center`}>
                        {CATEGORY_LABELS[item.category] ?? item.category}
                      </TableCell>
                      <TableCell className={`text-right font-mono text-base ${C.text} pr-4`}>
                        ¥{item.unitPrice.toLocaleString()}
                      </TableCell>
                      <TableCell className={`text-base ${C.text70} text-center`}>
                        {item.taxRate === 0.1 ? "10%" : item.taxRate === 0.08 ? "8%" : `${item.taxRate * 100}%`}
                      </TableCell>
                      <TableCell className="text-center">
                        <NotionStatusPill isActive={item.isActive} />
                      </TableCell>
                      <TableCell className="p-0 text-right pr-2">
                        <RowActionButton onClick={() => crud.handleEdit(item)} />
                      </TableCell>
                    </SortableDataTableRow>
                  ))}
                </SortableContext>
              </TableBody>
            </Table>
            <DragOverlay dropAnimation={null}>
              {activeId
                ? (() => {
                    const m = sortedItems.find((x) => x.id === activeId);
                    return m ? <MerchandiseRowOverlay item={m} /> : null;
                  })()
                : null}
            </DragOverlay>
          </DndContext>
        </div>
        <button
          type="button"
          onClick={crud.handleNew}
          className="flex items-center gap-1.5 w-full px-3 py-2.5 text-base text-[#37352F]/40 hover:text-[#37352F]/65 hover:bg-[#F7F6F3]/50 transition-colors rounded-b-[4px]"
        >
          <Plus className="size-3.5" />
          新しい品目を追加...
        </button>
      </div>
    </MasterCRUDPage>
  );
}
