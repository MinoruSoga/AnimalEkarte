// React/Framework
import { memo } from "react";

// External
import { Plus, FileText } from "lucide-react";

// Internal
import { C, ICON } from "@/lib/design-tokens";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { TableCell, TableHead } from "@/components/ui/table";
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
        <table className="w-full min-w-[800px]">
          <thead className={`${C.bgPage} border-b ${C.borderMedium}`}>
            <tr>
              <TableHead className={C.text60}>治療内容</TableHead>
              <TableHead className={C.text60}>メモ</TableHead>
              <TableHead className={`text-center ${C.text60} w-16`}>保険</TableHead>
              <TableHead className={`text-right ${C.text60} w-20`}>単価(￥)</TableHead>
              <TableHead className={`text-right ${C.text60} w-16`}>数量</TableHead>
              <TableHead className={`text-right ${C.text60} w-16`}>割引(%)</TableHead>
              <TableHead className={`text-right ${C.text60} w-20`}>値引(￥)</TableHead>
              <TableHead className={`text-right ${C.text60} w-20`}>小計(￥)</TableHead>
              <TableHead className={`text-center ${C.text60} w-12`}>操作</TableHead>
            </tr>
          </thead>
          <tbody className={`divide-y ${C.divideDivider}`}>
            {treatmentPlans.map((plan, index) => (
              <tr key={plan.id} className={`${C.hoverBgLight} transition-colors h-10`}>
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
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
});
