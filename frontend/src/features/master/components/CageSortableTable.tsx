import { type ComponentProps, useMemo } from "react";
import { DndContext, DragOverlay, closestCenter } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { Unlink } from "lucide-react";
import { RowActionButton } from "@/components/shared/RowActionButton";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { SortableDataTableRow } from "@/components/shared/DataTable/SortableDataTableRow";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { C, STYLE } from "@/lib/design-tokens";
import { resolveCageNameById } from "@/lib/cage-name";
import { cageMembersByRepresentative, isCageMergeEligible } from "@/lib/cage-merge";

import type { Cage } from "../api/cages";
import { CAGE_SIZE_LABELS, CAGE_TYPE_LABELS, formatCagePrice } from "../lib/cage-side-panel-model";
import { CageRowOverlay } from "./CageRowOverlay";

const TABLE_COLUMNS = [
  { key: "grip", className: "w-11 min-w-11 px-0" },
  { key: "select", className: "w-11 min-w-11 px-0" },
  { key: "name", label: "ケージ名", className: "pl-3 min-w-[140px]" },
  { key: "type", label: "エリア", className: "w-[100px] min-w-[100px] whitespace-nowrap" },
  { key: "size", label: "サイズ", className: "w-[90px] min-w-[90px] whitespace-nowrap" },
  {
    key: "price",
    label: "単価(税込)",
    className: "w-[120px] min-w-[120px] whitespace-nowrap text-right pr-4",
  },
  {
    key: "status",
    label: "ステータス",
    className: "w-[100px] min-w-[100px] whitespace-nowrap text-center",
  },
  {
    key: "action",
    label: "操作",
    className: "w-[80px] min-w-[80px] whitespace-nowrap text-right pr-2",
  },
];

interface CageSortableTableProps {
  items: Cage[];
  sensors: ComponentProps<typeof DndContext>["sensors"];
  activeId: string | null;
  onDragStart: NonNullable<ComponentProps<typeof DndContext>["onDragStart"]>;
  onDragEnd: NonNullable<ComponentProps<typeof DndContext>["onDragEnd"]>;
  onDragCancel: NonNullable<ComponentProps<typeof DndContext>["onDragCancel"]>;
  canEdit: boolean;
  onEdit: (item: Cage) => void;
  /** EMR-306: 連結対象として選択中の cage id 集合 */
  selectedIds: ReadonlySet<string>;
  onToggleSelect: (id: string, checked: boolean) => void;
  /**
   * EMR-306: 占有中（admitted/reserved 入院が存在）の cage id 集合。
   * undefined = 占有状況が不明 — spec「占有中でないケージのみ選択可」に従い
   * fail-closed で選択不可とする（最終可否は BE の 409 が担保）。
   */
  occupiedCageIds: ReadonlySet<string> | undefined;
  /** EMR-306: 連結解除要求。メンバー行からは代表 cage id に解決済みで渡される */
  onRequestUnmerge: (representativeId: string) => void;
}

export function CageSortableTable({
  items,
  sensors,
  activeId,
  onDragStart,
  onDragEnd,
  onDragCancel,
  canEdit,
  onEdit,
  selectedIds,
  onToggleSelect,
  occupiedCageIds,
  onRequestUnmerge,
}: CageSortableTableProps) {
  const activeItem = useMemo(
    () => items.find((item) => item.id === activeId) ?? null,
    [items, activeId],
  );

  // EMR-306: 代表 cageId → 構成メンバー索引（shared predicate に委譲）。
  // キーが存在する = そのケージは連結グループの代表。
  const membersByRepId = useMemo(() => cageMembersByRepresentative(items), [items]);

  return (
    <div className={STYLE.tableContainer}>
      <div className="flex-1 overflow-auto relative">
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragStart={onDragStart}
          onDragEnd={onDragEnd}
          onDragCancel={onDragCancel}
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
              {items.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className={STYLE.tableEmpty}>
                    ケージが登録されていません
                  </TableCell>
                </TableRow>
              ) : null}
              <SortableContext
                items={items.map((item) => item.id)}
                strategy={verticalListSortingStrategy}
              >
                {items.map((item) => {
                  // EMR-306: メンバー行は mergedIntoCageId が指す代表名を badge 表示し、
                  // 代表行・メンバー行ともグループ全体の解除アクションを持つ。
                  // 連結済み・無効行は merge の選択対象外（checkbox は disabled 表示）。
                  const mergedRepName =
                    item.mergedIntoCageId != null
                      ? resolveCageNameById(items, item.mergedIntoCageId)
                      : null;
                  const isMergedMember = item.mergedIntoCageId != null;
                  const mergedMembers = membersByRepId.get(item.id);
                  const isMergedRepresentative = mergedMembers != null;
                  const isMerged = isMergedMember || isMergedRepresentative;
                  const unmergeRepId = item.mergedIntoCageId ?? item.id;
                  // EMR-306 spec: 選択できるのは未連結・有効・占有中でないケージのみ。
                  // 占有状況が不明（undefined）な間も選択不可とする fail-closed。
                  return (
                    <SortableDataTableRow
                      key={item.id}
                      id={item.id}
                      dragLabel={`並べ替え: ケージ ${item.name} (ID ${item.id})`}
                      dragDisabled={!canEdit}
                    >
                      <TableCell className="w-11 min-w-11 px-0 text-center">
                        <Checkbox
                          aria-label={`連結選択: ケージ ${item.name} (ID ${item.id})`}
                          checked={selectedIds.has(item.id)}
                          disabled={
                            !canEdit || !isCageMergeEligible(item, occupiedCageIds, membersByRepId)
                          }
                          onCheckedChange={(checked) => onToggleSelect(item.id, checked === true)}
                        />
                      </TableCell>
                      <TableCell className={`min-w-[140px] font-medium ${C.text}`}>
                        <div className="flex items-center gap-2">
                          <DataTableRowButton
                            aria-label={`詳細: ケージ ${item.name} (ID ${item.id})`}
                            onClick={() => onEdit(item)}
                          >
                            {item.name}
                          </DataTableRowButton>
                          {isMergedMember ? (
                            <Badge variant="secondary">{`連結中: ${mergedRepName}`}</Badge>
                          ) : isMergedRepresentative ? (
                            <Badge variant="secondary">
                              {`連結中（代表）: ${mergedMembers.map((m) => m.name).join(" + ")}`}
                            </Badge>
                          ) : null}
                        </div>
                      </TableCell>
                      <TableCell className={`min-w-[100px] whitespace-nowrap ${C.text70}`}>
                        {CAGE_TYPE_LABELS[item.cageType] || item.cageType}
                      </TableCell>
                      <TableCell className={`min-w-[90px] whitespace-nowrap ${C.text70}`}>
                        {CAGE_SIZE_LABELS[item.cageSize] || item.cageSize}
                      </TableCell>
                      <TableCell
                        className={`min-w-[120px] whitespace-nowrap text-right font-mono ${C.text} pr-4`}
                      >
                        {formatCagePrice(item.price)}
                      </TableCell>
                      <TableCell className="min-w-[100px] whitespace-nowrap text-center">
                        <StatusPill isActive={item.isActive} />
                      </TableCell>
                      <TableCell className="min-w-[80px] whitespace-nowrap text-right pr-2">
                        <div className="flex items-center justify-end gap-1">
                          {canEdit && isMerged ? (
                            <RowActionButton
                              icon={Unlink}
                              aria-label={`連結解除: ケージ ${item.name} (ID ${item.id})`}
                              onClick={() => onRequestUnmerge(unmergeRepId)}
                            />
                          ) : null}
                          {canEdit ? (
                            <RowActionButton
                              aria-label={`編集: ケージ ${item.name} (ID ${item.id})`}
                              onClick={() => onEdit(item)}
                            />
                          ) : null}
                        </div>
                      </TableCell>
                    </SortableDataTableRow>
                  );
                })}
              </SortableContext>
            </TableBody>
          </Table>
          <DragOverlay dropAnimation={null}>
            {activeItem ? <CageRowOverlay cage={activeItem} /> : null}
          </DragOverlay>
        </DndContext>
      </div>
    </div>
  );
}
