// React/Framework
import { memo } from "react";

// External
import { Plus, FileText } from "lucide-react";

// Internal
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { HospitalizationTreatmentPlan } from "@/types";

// Relative
import { H_STYLES } from "../lib/styles";

interface HospitalizationTreatmentTableProps {
  treatmentPlans: HospitalizationTreatmentPlan[];
  onAdd: () => void;
  onRemove?: (id: string) => void;
  onUpdate: (
    id: string,
    field: keyof HospitalizationTreatmentPlan,
    value: string | number | boolean,
  ) => void;
  readOnly?: boolean;
}

export const HospitalizationTreatmentTable = memo(function HospitalizationTreatmentTable({
  treatmentPlans,
  onAdd,
  onRemove,
  onUpdate,
  readOnly = false,
}: HospitalizationTreatmentTableProps) {
  return (
    <div
      className={`${C.bgWhite} rounded-lg border ${C.borderMedium} ${H_STYLES.padding.box} mb-3`}
    >
      <div className="flex items-center justify-between mb-3">
        <h2 className={`${H_STYLES.text.base} font-bold flex items-center gap-2 ${C.text}`}>
          <FileText className={`${ICON.action} ${C.text60}`} />
          治療プラン
        </h2>
        <Button
          onClick={onAdd}
          disabled={readOnly}
          variant="outline"
          size="sm"
          className={`gap-1.5 ${H_STYLES.button.action} ${C.text} ${C.borderMedium} ${C.bgSkeleton}`}
        >
          <Plus className={H_STYLES.button.icon} />
          追加
        </Button>
      </div>

      {/* Table */}
      <div
        role="region"
        aria-label="治療プラン"
        tabIndex={0}
        className={`border ${C.borderMedium} rounded-md overflow-hidden overflow-x-auto outline-none focus-visible:ring-2 focus-visible:ring-inset ${C.focusRingAccent40}`}
      >
        <Table className="min-w-[800px]">
          <TableHeader>
            <TableRow className={STYLE.tableHeaderRow}>
              <TableHead className={STYLE.tableHeaderCell}>治療内容</TableHead>
              <TableHead className={STYLE.tableHeaderCell}>メモ</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} text-center w-16`}>保険</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} text-right w-20`}>単価(￥)</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} text-right w-16`}>数量</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} text-right w-16`}>割引(%)</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} text-right w-20`}>値引(￥)</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} text-right w-20`}>小計(￥)</TableHead>
              <TableHead className={`${STYLE.tableHeaderCell} text-center w-12`}>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {treatmentPlans.map((plan, index) => (
              <TableRow key={plan.id} className={`${C.hoverBgPageHalf} h-10`}>
                <TableCell>
                  <Input
                    aria-label={`治療内容 ${index + 1}`}
                    value={plan.treatmentContent}
                    disabled={readOnly}
                    onChange={(e) => onUpdate(plan.id, "treatmentContent", e.target.value)}
                    className={`${H_STYLES.text.base} h-11 border-none shadow-none focus-visible:ring-1 ${C.focusVisibleRingActionPrimary} bg-transparent ${C.text}`}
                    placeholder="治療内容を入力..."
                  />
                </TableCell>
                <TableCell>
                  <Input
                    aria-label={`治療メモ ${index + 1}`}
                    value={plan.memo}
                    disabled={readOnly}
                    onChange={(e) => onUpdate(plan.id, "memo", e.target.value)}
                    className={`${H_STYLES.text.base} h-11 border-none shadow-none focus-visible:ring-1 ${C.focusVisibleRingActionPrimary} bg-transparent ${C.text}`}
                    placeholder="メモ..."
                  />
                </TableCell>
                <TableCell className={`text-center ${C.text}`}>
                  {plan.is_insurance ? "◯" : "×"}
                </TableCell>
                <TableCell className={`text-right tabular-nums ${C.text}`}>
                  {plan.unitPrice.toLocaleString()}
                </TableCell>
                <TableCell className={`text-right tabular-nums ${C.text}`}>
                  {plan.quantity}
                </TableCell>
                <TableCell className={`text-right tabular-nums ${C.text}`}>
                  {plan.discount}
                </TableCell>
                <TableCell className={`text-right tabular-nums ${C.text}`}>
                  {plan.discountAmount.toLocaleString()}
                </TableCell>
                <TableCell className={`text-right tabular-nums font-medium ${C.text}`}>
                  {plan.subtotal.toLocaleString()}
                </TableCell>
                <TableCell className="text-center">
                  {onRemove !== undefined && !readOnly ? (
                    <DeleteIconButton onClick={() => onRemove(plan.id)} />
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
});
