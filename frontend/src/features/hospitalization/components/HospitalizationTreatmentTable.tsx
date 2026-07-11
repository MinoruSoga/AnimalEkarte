import { C, ICON } from "@/lib/design-tokens";
import { memo } from "react";
import { Plus, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { HospitalizationTreatmentPlan } from "@/types";
import { H_STYLES } from "../styles";

interface HospitalizationTreatmentTableProps {
  treatmentPlans: HospitalizationTreatmentPlan[];
  onAdd: () => void;
  onRemove?: (id: string) => void;
  onUpdate: (id: string, field: keyof HospitalizationTreatmentPlan, value: string | number | boolean) => void;
}

export const HospitalizationTreatmentTable = memo(function HospitalizationTreatmentTable({
    treatmentPlans,
    onAdd,
    onRemove,
    onUpdate
}: HospitalizationTreatmentTableProps) {
  return (
    <div className={`${C.bgWhite} rounded-lg shadow-sm border ${C.borderMedium} ${H_STYLES.padding.box} mb-3`}>
      <div className="flex items-center justify-between mb-3">
        <h2 className={`${H_STYLES.text.base} font-bold flex items-center gap-2 ${C.text}`}>
          <FileText className={`${ICON.action} ${C.text60}`} />
          治療プラン
        </h2>
        <Button
          onClick={onAdd}
          variant="outline"
          size="sm"
          className={`gap-1.5 ${H_STYLES.button.action} ${C.text} ${C.borderMedium} ${C.bgSkeleton}`}
        >
          <Plus className={H_STYLES.button.icon} />
          追加
        </Button>
      </div>

      {/* Table */}
      <div className={`border ${C.borderMedium} rounded-md overflow-hidden overflow-x-auto`}>
        <table className="w-full min-w-[800px]">
          <thead className={`${C.bgPage} border-b ${C.borderMedium}`}>
            <tr>
              <th className={`text-left px-3 py-2 ${H_STYLES.text.sm} font-medium ${C.text60}`}>治療内容</th>
              <th className={`text-left px-3 py-2 ${H_STYLES.text.sm} font-medium ${C.text60}`}>メモ</th>
              <th className={`text-center px-3 py-2 ${H_STYLES.text.sm} font-medium ${C.text60} w-16`}>保険</th>
              <th className={`text-right px-3 py-2 ${H_STYLES.text.sm} font-medium ${C.text60} w-20`}>単価(￥)</th>
              <th className={`text-right px-3 py-2 ${H_STYLES.text.sm} font-medium ${C.text60} w-16`}>数量</th>
              <th className={`text-right px-3 py-2 ${H_STYLES.text.sm} font-medium ${C.text60} w-16`}>割引(%)</th>
              <th className={`text-right px-3 py-2 ${H_STYLES.text.sm} font-medium ${C.text60} w-20`}>値引(￥)</th>
              <th className={`text-right px-3 py-2 ${H_STYLES.text.sm} font-medium ${C.text60} w-20`}>小計(￥)</th>
              <th className={`text-center px-3 py-2 ${H_STYLES.text.sm} font-medium ${C.text60} w-12`}>操作</th>
            </tr>
          </thead>
          <tbody className={`divide-y ${C.divideDivider}`}>
            {treatmentPlans.map((plan) => (
              <tr key={plan.id} className={`${C.hoverBgLight} transition-colors h-10`}>
                <td className="px-3 py-1">
                  <Input
                    value={plan.treatmentContent}
                    onChange={(e) =>
                      onUpdate(plan.id, "treatmentContent", e.target.value)
                    }
                    className={`${H_STYLES.text.base} h-10 border-none shadow-none focus-visible:ring-1 ${C.focusRingMedicalBlue} bg-transparent ${C.text}`}
                    placeholder="治療内容を入力..."
                  />
                </td>
                <td className="px-3 py-1">
                  <Input
                    value={plan.memo}
                    onChange={(e) =>
                      onUpdate(plan.id, "memo", e.target.value)
                    }
                    className={`${H_STYLES.text.base} h-10 border-none shadow-none focus-visible:ring-1 ${C.focusRingMedicalBlue} bg-transparent ${C.text}`}
                    placeholder="メモ..."
                  />
                </td>
                <td className={`px-3 py-1 text-center ${H_STYLES.text.base} ${C.text}`}>
                  {plan.is_insurance ? "◯" : "×"}
                </td>
                <td className={`px-3 py-1 text-right ${H_STYLES.text.base} tabular-nums ${C.text}`}>
                  {plan.unitPrice.toLocaleString()}
                </td>
                <td className={`px-3 py-1 text-right ${H_STYLES.text.base} tabular-nums ${C.text}`}>
                  {plan.quantity}
                </td>
                <td className={`px-3 py-1 text-right ${H_STYLES.text.base} tabular-nums ${C.text}`}>
                  {plan.discount}
                </td>
                <td className={`px-3 py-1 text-right ${H_STYLES.text.base} tabular-nums ${C.text}`}>
                  {plan.discountAmount.toLocaleString()}
                </td>
                <td className={`px-3 py-1 text-right ${H_STYLES.text.base} tabular-nums font-medium ${C.text}`}>
                  {plan.subtotal.toLocaleString()}
                </td>
                <td className="px-3 py-1 text-center">
                  {onRemove !== undefined ? (
                    <DeleteIconButton onClick={() => onRemove(plan.id)} />
                  ) : null}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
});
