import { memo } from "react";
import { TableCell, TableHead } from "@/components/ui/table";
import { C, STYLE } from "@/lib/design-tokens";
import { formatJSTTime } from "@/lib/jst-date";
import { formatCurrency } from "@/lib/format/number";
import type { CloseBillingDetail } from "../api/get-cash-register-preview";
import { CATEGORY_LABELS } from "../lib/constants";

interface BillingDetailTableProps {
  details: CloseBillingDetail[];
}

export const BillingDetailTable = memo(function BillingDetailTable({
  details,
}: BillingDetailTableProps) {
  if (details.length === 0) {
    return <p className={`text-base ${C.text50} py-4 text-center`}>会計明細がありません</p>;
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-base">
        <thead>
          <tr className={`border-b ${C.borderLight} ${C.bgPage}`}>
            <TableHead className={C.text70}>時刻</TableHead>
            <TableHead className={C.text70}>飼主 / ペット</TableHead>
            <TableHead className={C.text70}>部門</TableHead>
            <TableHead className={C.text70}>支払方法</TableHead>
            <TableHead className={`text-right ${C.text70}`}>請求額</TableHead>
            <TableHead className={`text-right ${C.text70}`}>返金額</TableHead>
            <TableHead className={`text-right ${C.text70}`}>純額</TableHead>
          </tr>
        </thead>
        <tbody>
          {details.map((detail) => (
            <tr key={detail.billingId} className={`border-b ${C.borderLight} ${STYLE.tableRow}`}>
              <TableCell className={`${C.text60} whitespace-nowrap`}>
                {detail.paidAt ? formatJSTTime(detail.paidAt) : "—"}
              </TableCell>
              <TableCell className={C.text}>
                <div>{detail.ownerName}</div>
                <div className={`text-sm ${C.text60}`}>
                  {detail.petName}
                  {detail.isHospitalization ? " (入院)" : ""}
                </div>
              </TableCell>
              <TableCell className={C.text}>
                {CATEGORY_LABELS[detail.category] ?? detail.category}
              </TableCell>
              <TableCell className={C.text}>{detail.paymentMethodName}</TableCell>
              <TableCell className={`text-right ${C.text}`}>
                {formatCurrency(detail.billingAmount)}
              </TableCell>
              <TableCell className={`text-right ${detail.refundAmount > 0 ? C.danger : C.text50}`}>
                {detail.refundAmount > 0 ? `-${formatCurrency(detail.refundAmount)}` : "—"}
              </TableCell>
              <TableCell className={`text-right font-medium ${C.text}`}>
                {formatCurrency(detail.netAmount)}
              </TableCell>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
});
