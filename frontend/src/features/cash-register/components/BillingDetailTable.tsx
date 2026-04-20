import { memo } from "react";
import { C, STYLE } from "@/lib/design-tokens";
import type { CloseBillingDetail } from "@/types/generated/models";
import { CATEGORY_LABELS } from "../constants";

interface BillingDetailTableProps {
  details: CloseBillingDetail[];
}

export const BillingDetailTable = memo(function BillingDetailTable({
  details,
}: BillingDetailTableProps) {
  if (details.length === 0) {
    return (
      <p className={`text-base ${C.text50} py-4 text-center`}>会計明細がありません</p>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-base">
        <thead>
          <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
            <th className={`text-left px-3 py-2 font-medium ${C.text70}`}>時刻</th>
            <th className={`text-left px-3 py-2 font-medium ${C.text70}`}>飼主 / ペット</th>
            <th className={`text-left px-3 py-2 font-medium ${C.text70}`}>部門</th>
            <th className={`text-left px-3 py-2 font-medium ${C.text70}`}>支払方法</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text70}`}>請求額</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text70}`}>返金額</th>
            <th className={`text-right px-3 py-2 font-medium ${C.text70}`}>純額</th>
          </tr>
        </thead>
        <tbody>
          {details.map((detail) => (
            <tr
              key={detail.billing_id}
              className={`border-b ${C.borderLight} ${STYLE.tableRow}`}
            >
              <td className={`px-3 py-2 ${C.text60} whitespace-nowrap`}>
                {detail.paid_at ? new Date(detail.paid_at).toLocaleTimeString("ja-JP", { hour: "2-digit", minute: "2-digit" }) : "—"}
              </td>
              <td className={`px-3 py-2 ${C.text}`}>
                <div>{detail.owner_name}</div>
                <div className={`text-sm ${C.text60}`}>
                  {detail.pet_name}
                  {detail.is_hospitalization ? " (入院)" : ""}
                </div>
              </td>
              <td className={`px-3 py-2 ${C.text}`}>
                {CATEGORY_LABELS[detail.category] ?? detail.category}
              </td>
              <td className={`px-3 py-2 ${C.text}`}>{detail.payment_method_name}</td>
              <td className={`px-3 py-2 text-right ${C.text}`}>
                ¥{detail.billing_amount.toLocaleString()}
              </td>
              <td className={`px-3 py-2 text-right ${detail.refund_amount > 0 ? C.danger : C.text50}`}>
                {detail.refund_amount > 0 ? `-¥${detail.refund_amount.toLocaleString()}` : "—"}
              </td>
              <td className={`px-3 py-2 text-right font-medium ${C.text}`}>
                ¥{detail.net_amount.toLocaleString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
});
