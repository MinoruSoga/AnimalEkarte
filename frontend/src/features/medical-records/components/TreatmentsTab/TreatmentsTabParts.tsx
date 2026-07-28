import { Plus, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TableCell, TableHead } from "@/components/ui/table";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";
import type { MedicineDoseContext } from "../../api/medicine-dose-lookup";
import type { Treatment, TreatmentItemType, UpdateTreatmentInput } from "../../types";
import { TreatmentRow } from "./TreatmentRow";

// DESIGN.md ex-data-table-cell: header は canvas-soft 背景 + eyebrow 相当タイポグラフィ（STYLE.sectionLabel）。
const TABLE_HEADER = (
  <thead>
    <tr className={`border-b ${C.borderLight} ${C.bgPage} h-11`}>
      <TableHead className="w-10" />
      <TableHead className="w-24">種別</TableHead>
      <TableHead>内容</TableHead>
      <TableHead className="w-16 text-center">保険</TableHead>
      <TableHead className="w-28 text-right">単価</TableHead>
      <TableHead className="w-20 text-right">数量</TableHead>
      <TableHead className="w-28 text-right">値引き</TableHead>
      <TableHead className="w-28 text-right">小計</TableHead>
      <TableHead>メモ</TableHead>
      <TableHead className="w-28 text-right" />
    </tr>
  </thead>
);

const ITEM_TYPE_OPTIONS: { value: TreatmentItemType; label: string }[] = [
  { value: "consultation", label: "診療" },
  { value: "procedure", label: "処置" },
  { value: "medicine", label: "薬品" },
  { value: "other", label: "その他" },
];

const ADMIN_ROUTE_OPTIONS = [
  { value: "", label: "投与方法を選択" },
  { value: "経口", label: "経口" },
  { value: "注射", label: "注射" },
  { value: "外用", label: "外用" },
  { value: "点眼", label: "点眼" },
  { value: "点耳", label: "点耳" },
  { value: "吸入", label: "吸入" },
  { value: "その他", label: "その他" },
] as const;

interface TreatmentsTableProps {
  treatments: Treatment[];
  isUpdating: boolean;
  canDelete: boolean;
  canEditDiscount: boolean;
  focusLastRow: boolean;
  onUpdate: (treatmentId: string, input: UpdateTreatmentInput) => void;
  onDelete: (treatmentId: string) => void;
  onMoveUp: (treatmentId: string) => void;
  onMoveDown: (treatmentId: string) => void;
  onAutoFocusDone: () => void;
  /** #201: 投与量自動計算プレビューに必要なコンテキスト */
  doseContext: MedicineDoseContext;
}

export function TreatmentsTable({
  treatments,
  isUpdating,
  canDelete,
  canEditDiscount,
  focusLastRow,
  onUpdate,
  onDelete,
  onMoveUp,
  onMoveDown,
  onAutoFocusDone,
  doseContext,
}: TreatmentsTableProps) {
  return (
    <table className="w-full">
      {TABLE_HEADER}
      <tbody>
        {treatments.length === 0 ? (
          <tr>
            <TableCell data-empty-state colSpan={10} className={STYLE.tableEmptySm}>
              治療明細がありません。下の「明細を追加」ボタンから追加してください。
            </TableCell>
          </tr>
        ) : (
          treatments.map((treatment, index) => {
            const autoFocusQuantity = focusLastRow && index === treatments.length - 1;
            return (
              <TreatmentRow
                key={treatment.id}
                treatment={treatment}
                isFirst={index === 0}
                isLast={index === treatments.length - 1}
                onUpdate={onUpdate}
                onDelete={onDelete}
                onMoveUp={onMoveUp}
                onMoveDown={onMoveDown}
                isUpdating={isUpdating}
                canDelete={canDelete}
                canEditDiscount={canEditDiscount}
                autoFocusQuantity={autoFocusQuantity}
                onAutoFocusDone={autoFocusQuantity ? onAutoFocusDone : undefined}
                doseContext={doseContext}
              />
            );
          })
        )}
      </tbody>
    </table>
  );
}

interface TreatmentAddControlsProps {
  canCreate: boolean;
  isAdding: boolean;
  isPending: boolean;
  addItemType: TreatmentItemType;
  addContent: string;
  addAdminRoute: string;
  onItemTypeChange: (value: TreatmentItemType) => void;
  onContentChange: (value: string) => void;
  onAdminRouteChange: (value: string) => void;
  onSubmit: () => void;
  onCancel: () => void;
  onOpenSearch: () => void;
  onStartAdding: () => void;
}

export function TreatmentAddControls({
  canCreate,
  isAdding,
  isPending,
  addItemType,
  addContent,
  addAdminRoute,
  onItemTypeChange,
  onContentChange,
  onAdminRouteChange,
  onSubmit,
  onCancel,
  onOpenSearch,
  onStartAdding,
}: TreatmentAddControlsProps) {
  if (isAdding) {
    return (
      <div className={`flex items-center gap-2 px-3 py-2 border-t ${C.borderLight} ${C.bgPage30}`}>
        <select
          value={addItemType}
          onChange={(event) => onItemTypeChange(event.target.value as TreatmentItemType)}
          className={`h-8 text-sm rounded-xxs border ${C.borderMedium} ${C.bgWhite} px-2 ${C.text}`}
        >
          {ITEM_TYPE_OPTIONS.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
        {addItemType === "medicine" ? (
          <select
            value={addAdminRoute}
            onChange={(event) => onAdminRouteChange(event.target.value)}
            className={`h-8 text-sm rounded-xxs border ${C.borderMedium} ${C.bgWhite} px-2 ${C.text}`}
          >
            {ADMIN_ROUTE_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        ) : null}
        <input
          autoFocus
          type="text"
          placeholder="内容を入力..."
          value={addContent}
          onChange={(event) => onContentChange(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter") onSubmit();
            if (event.key === "Escape") onCancel();
          }}
          aria-label="治療内容"
          className={`flex-1 h-8 text-sm border ${C.borderMedium} rounded-xxs px-2 ${C.bgWhite} ${C.text} outline-none ${C.focusBorderAccent}`}
        />
        <Button
          size="sm"
          className={`${C.bgBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand} ${C.textOnBrand} rounded-full border-transparent transition-colors h-8 text-xs px-3`}
          onClick={onSubmit}
          disabled={isPending || !addContent.trim()}
        >
          追加
        </Button>
        <Button
          size="sm"
          variant="outline"
          className={`h-8 text-xs px-3 ${C.borderMedium}`}
          onClick={onCancel}
        >
          キャンセル
        </Button>
      </div>
    );
  }

  if (!canCreate) return null;

  return (
    <div className="flex items-center gap-0">
      <button type="button" className={STYLE.inlineAddBtn} onClick={onOpenSearch}>
        <Search className={ICON.xs} />
        <span>マスタから追加</span>
      </button>
      <button type="button" className={STYLE.inlineAddBtn} onClick={onStartAdding}>
        <Plus className={ICON.xs} />
        <span>手入力で追加</span>
      </button>
    </div>
  );
}

interface TreatmentTotalsProps {
  totalCount: number;
  totalSubtotal: number;
  selectedCount: number;
  selectedSubtotal: number;
  finalTotal: number;
  ownerDiscountRate: number;
}

export function TreatmentTotals({
  totalCount,
  totalSubtotal,
  selectedCount,
  selectedSubtotal,
  finalTotal,
  ownerDiscountRate,
}: TreatmentTotalsProps) {
  return (
    <div className={`${C.bgWhite} border ${C.borderLight} rounded-xs px-4 py-3`}>
      <div className="flex flex-col gap-1.5">
        <div className={`flex items-center justify-between text-sm ${C.text60}`}>
          <span>全明細合計 ({totalCount}件)</span>
          <span className="font-mono">{formatCurrency(totalSubtotal)}</span>
        </div>
        <div className={`flex items-center justify-between text-sm font-medium ${C.text}`}>
          <span>選択済み合計 ({selectedCount}件)</span>
          <span className="font-mono text-base">
            {formatCurrency(selectedSubtotal)}
          </span>
        </div>
        <div className={`flex items-center justify-between text-sm pt-1.5 border-t ${C.borderLight}`}>
          <span className={C.text60}>
            税込合計 (10% {ownerDiscountRate > 0 ? `飼主割引${ownerDiscountRate}%適用後` : ""})
          </span>
          <span className={`font-mono font-semibold ${C.textCostBlue}`}>
            {formatCurrency(finalTotal)}
          </span>
        </div>
      </div>
    </div>
  );
}
