import { Checkbox } from "@/components/ui/checkbox";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";
import type { AggregationOwner } from "../api/get-aggregations";
import type { AggregationTab } from "./aggregation-filter-panel-model";
import { getAggregationOwnerColumns } from "./AggregationOwnerTableColumns";

interface AggregationOwnerTableProps {
  owners: AggregationOwner[];
  selectedOwnerIds: Set<string>;
  onSelectAll: (checked: boolean) => void;
  onSelectOwner: (ownerId: string, checked: boolean) => void;
  isLoading: boolean;
  activeTab: AggregationTab;
  isError?: boolean;
  errorMessage?: string;
}




export function AggregationOwnerTable({
  owners,
  selectedOwnerIds,
  onSelectAll,
  onSelectOwner,
  isLoading,
  activeTab,
  isError,
  errorMessage,
}: AggregationOwnerTableProps) {
  const columns = getAggregationOwnerColumns(activeTab);
  const colSpan = columns.length + 1; // +1 for checkbox
  const allSelected = owners.length > 0 && owners.every((o) => selectedOwnerIds.has(o.owner_id));
  const someSelected = owners.some((o) => selectedOwnerIds.has(o.owner_id)) && !allSelected;
  // 全選択チェックボックスは「行が描画されない」状態 (読み込み中 / エラー / 0件) では抑制する。
  // 行のチェックボックスは isError / isLoading では行自体が描画されないため、
  // 行内の Checkbox は disable 不要 (行は出ない)。
  const isHeaderSelectDisabled = isLoading || isError || owners.length === 0;

  return (
    <div className={STYLE.tableContainer}>
      <Table>
        <TableHeader>
          {/* DESIGN.md ex-data-table-cell: header は canvas-soft 背景（bgPage30 → bgPage）+ eyebrow 相当タイポグラフィ */}
          <TableRow className={`${STYLE.tableHeaderRow} ${C.bgPage}`}>
            <TableHead className={`${STYLE.tableHeaderCell} ${STYLE.sectionLabel} w-12 px-4`}>
              <Checkbox
                touchTarget
                checked={someSelected ? "indeterminate" : allSelected}
                onCheckedChange={(checked) => onSelectAll(!!checked)}
                aria-label="全選択"
                disabled={isHeaderSelectDisabled}
                className="-my-3"
              />
            </TableHead>
            {columns.map((col) => (
              <TableHead
                key={col.key}
                className={`${STYLE.tableHeaderCell} ${STYLE.sectionLabel} px-4 ${col.width ?? ""} ${
                  col.textAlign === "right" ? "text-right" : ""
                }`}
              >
                {col.label}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {isError ? (
            <TableRow>
              <TableCell colSpan={colSpan} className={`${STYLE.tableEmpty} ${C.danger}`}>
                {errorMessage || "エラーが発生しました"}
              </TableCell>
            </TableRow>
          ) : isLoading ? (
            <TableRow>
              <TableCell colSpan={colSpan} className={STYLE.tableEmpty}>
                読み込み中...
              </TableCell>
            </TableRow>
          ) : owners.length === 0 ? (
            <TableRow>
              <TableCell colSpan={colSpan} className={STYLE.tableEmpty}>
                データが見つかりません
              </TableCell>
            </TableRow>
          ) : (
            owners.map((owner) => (
              <TableRow
                key={owner.owner_id}
                className={`${STYLE.tableRow} ${selectedOwnerIds.has(owner.owner_id) ? C.bgBrand10 : ""}`}
              >
                <TableCell className="px-4">
                  <Checkbox
                    touchTarget
                    checked={selectedOwnerIds.has(owner.owner_id)}
                    onCheckedChange={(checked) =>
                      onSelectOwner(owner.owner_id, !!checked)
                    }
                    aria-label={`${owner.owner_name}を選択`}
                    className="-my-3"
                  />
                </TableCell>
                {columns.map((col) => (
                  <TableCell
                    key={col.key}
                    className={`${STYLE.tableCell} px-4 ${col.width ?? ""} ${
                      col.textAlign === "right" ? "text-right" : ""
                    }`}
                  >
                    {col.render(owner)}
                  </TableCell>
                ))}
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
