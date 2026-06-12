import { memo } from "react";
import { C, STYLE } from "@/lib/design-tokens";
import { formatJSTTime } from "@/lib/jst-date";
import type { CloseBillingDetail } from "../api/get-cash-register-preview";
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
              key={detail.billingId}
              className={`border-b ${C.borderLight} ${STYLE.tableRow}`}
            >
              <td className={`px-3 py-2 ${C.text60} whitespace-nowrap`}>
                {detail.paidAt ? formatJSTTime(detail.paidAt) : "—"}
              </td>
              <td className={`px-3 py-2 ${C.text}`}>
                <div>{detail.ownerName}</div>
                <div className={`text-sm ${C.text60}`}>
                  {detail.petName}
                  {detail.isHospitalization ? " (入院)" : ""}
                </div>
              </td>
              <td className={`px-3 py-2 ${C.text}`}>
                {CATEGORY_LABELS[detail.category] ?? detail.category}
              </td>
              <td className={`px-3 py-2 ${C.text}`}>{detail.paymentMethodName}</td>
              <td className={`px-3 py-2 text-right ${C.text}`}>
                ¥{detail.billingAmount.toLocaleString()}
              </td>
              <td className={`px-3 py-2 text-right ${detail.refundAmount > 0 ? C.danger : C.text50}`}>
                {detail.refundAmount > 0 ? `-¥${detail.refundAmount.toLocaleString()}` : "—"}
              </td>
              <td className={`px-3 py-2 text-right font-medium ${C.text}`}>
                ¥{detail.netAmount.toLocaleString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
});
