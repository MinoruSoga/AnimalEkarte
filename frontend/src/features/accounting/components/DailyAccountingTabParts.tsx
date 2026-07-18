import { C } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";

import type { CashCardAmount } from "./daily-accounting-utils";

// ── サマリーカード ────────────────────────────────────────────────
export function SummaryCard({ label, value }: { label: string; value: string }) {
  return (
    <div className={`rounded-lg border ${C.borderLight} px-4 py-3 ${C.bgWhite} min-w-[100px]`}>
      <p className={`text-xs ${C.text50} mb-0.5`}>{label}</p>
      <p className={`text-base font-semibold font-mono ${C.text}`}>{value}</p>
    </div>
  );
}

// ── 印刷エリア用セル ─────────────────────────────────────────────
interface CatCellProps {
  detail: CashCardAmount;
  isMixed: boolean;
}

export function CatCell({ detail, isMixed }: CatCellProps) {
  const total = detail.cash + detail.card;
  if (total === 0) {
    return <td className="border border-gray-300 text-right px-1 py-0.5 text-[9pt]">-</td>;
  }
  if (!isMixed) {
    return (
      <td className="border border-gray-300 text-right px-1 py-0.5 text-[9pt]">
        {formatCurrency(total)}
      </td>
    );
  }
  return (
    <td className="border border-gray-300 text-right px-1 py-0.5 text-[9pt] leading-tight">
      {detail.cash > 0 ? <div>現{detail.cash.toLocaleString()}</div> : null}
      {detail.card > 0 ? <div className="text-blue-700">カ{detail.card.toLocaleString()}</div> : null}
    </td>
  );
}
