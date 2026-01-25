import { Plus, Trash2, FileText } from "lucide-react";
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import { TreatmentPlan } from "../../../types";
import { H_STYLES } from "../styles";

interface HospitalizationTreatmentTableProps {
  treatmentPlans: TreatmentPlan[];
  onAdd: () => void;
  onRemove: (id: string) => void;
  onUpdate: (id: string, field: keyof TreatmentPlan, value: string | number | boolean) => void;
}

export const HospitalizationTreatmentTable = ({ 
    treatmentPlans, 
    onAdd, 
    onRemove, 
    onUpdate 
}: HospitalizationTreatmentTableProps) => {
  return (
    <div className={`bg-white rounded-lg shadow-sm border border-[rgba(55,53,47,0.16)] ${H_STYLES.padding.box} mb-3`}>
      <div className="flex items-center justify-between mb-3">
        <h2 className={`${H_STYLES.text.base} font-bold flex items-center gap-2 text-[#37352F]`}>
          <FileText className="h-4 w-4 text-[#37352F]/60" />
          治療プラン
        </h2>
        <Button
          onClick={onAdd}
          variant="outline"
          size="sm"
          className={`gap-1.5 ${H_STYLES.button.action} text-[#37352F] border-[rgba(55,53,47,0.16)] hover:bg-[rgba(55,53,47,0.06)]`}
        >
          <Plus className={H_STYLES.button.icon} />
          追加
        </Button>
      </div>

      {/* Table */}
      <div className="border border-[rgba(55,53,47,0.16)] rounded-md overflow-hidden overflow-x-auto">
        <table className="w-full min-w-[800px]">
          <thead className="bg-[#F7F6F3] border-b border-[rgba(55,53,47,0.16)]">
            <tr>
              <th className={`text-left px-3 py-2 ${H_STYLES.text.sm} font-medium text-[#37352F]/60`}>治療内容</th>
              <th className={`text-left px-3 py-2 ${H_STYLES.text.sm} font-medium text-[#37352F]/60`}>メモ</th>
              <th className={`text-center px-3 py-2 ${H_STYLES.text.sm} font-medium text-[#37352F]/60 w-16`}>保険</th>
              <th className={`text-right px-3 py-2 ${H_STYLES.text.sm} font-medium text-[#37352F]/60 w-20`}>単価(￥)</th>
              <th className={`text-right px-3 py-2 ${H_STYLES.text.sm} font-medium text-[#37352F]/60 w-16`}>数量</th>
              <th className={`text-right px-3 py-2 ${H_STYLES.text.sm} font-medium text-[#37352F]/60 w-16`}>割引(%)</th>
              <th className={`text-right px-3 py-2 ${H_STYLES.text.sm} font-medium text-[#37352F]/60 w-20`}>値引(￥)</th>
              <th className={`text-right px-3 py-2 ${H_STYLES.text.sm} font-medium text-[#37352F]/60 w-20`}>小計(￥)</th>
              <th className={`text-center px-3 py-2 ${H_STYLES.text.sm} font-medium text-[#37352F]/60 w-12`}>操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-[rgba(55,53,47,0.09)]">
            {treatmentPlans.map((plan) => (
              <tr key={plan.id} className="hover:bg-[rgba(55,53,47,0.06)] transition-colors h-10">
                <td className="px-3 py-1">
                  <Input
                    value={plan.treatmentContent}
                    onChange={(e) =>
                      onUpdate(plan.id, "treatmentContent", e.target.value)
                    }
                    className={`${H_STYLES.text.base} h-10 border-none shadow-none focus-visible:ring-1 focus-visible:ring-[#2EAADC] bg-transparent text-[#37352F]`}
                    placeholder="治療内容を入力..."
                  />
                </td>
                <td className="px-3 py-1">
                  <Input
                    value={plan.memo}
                    onChange={(e) =>
                      onUpdate(plan.id, "memo", e.target.value)
                    }
                    className={`${H_STYLES.text.base} h-10 border-none shadow-none focus-visible:ring-1 focus-visible:ring-[#2EAADC] bg-transparent text-[#37352F]`}
                    placeholder="メモ..."
                  />
                </td>
                <td className={`px-3 py-1 text-center ${H_STYLES.text.base} text-[#37352F]`}>
                  {plan.insurance ? "◯" : "×"}
                </td>
                <td className={`px-3 py-1 text-right ${H_STYLES.text.base} tabular-nums text-[#37352F]`}>
                  {plan.unitPrice.toLocaleString()}
                </td>
                <td className={`px-3 py-1 text-right ${H_STYLES.text.base} tabular-nums text-[#37352F]`}>
                  {plan.quantity}
                </td>
                <td className={`px-3 py-1 text-right ${H_STYLES.text.base} tabular-nums text-[#37352F]`}>
                  {plan.discount}
                </td>
                <td className={`px-3 py-1 text-right ${H_STYLES.text.base} tabular-nums text-[#37352F]`}>
                  {plan.discountAmount.toLocaleString()}
                </td>
                <td className={`px-3 py-1 text-right ${H_STYLES.text.base} tabular-nums font-medium text-[#37352F]`}>
                  {plan.subtotal.toLocaleString()}
                </td>
                <td className="px-3 py-1 text-center">
                  <button
                    onClick={() => onRemove(plan.id)}
                    className="text-[#37352F]/40 hover:text-[#E03E3E] transition-colors"
                  >
                    <Trash2 className={H_STYLES.button.icon} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
