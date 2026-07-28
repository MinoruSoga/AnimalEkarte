import { memo, useMemo } from 'react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { formatCurrency } from '@/lib/format/number';
import { calcLineItemAmount } from '@/lib/line-item-helpers';
import type { EstimateLineItem } from '../../types';
import { C } from '@/lib/design-tokens';
// FE5-26: 会計側の訳語(処方/フード/物販等)へ統一。旧ローカル定義(薬剤/食事/物品)は廃止。
import { CATEGORY_LABELS } from '@/constants/item-category';

interface EstimateLineItemsProps {
  items: EstimateLineItem[];
  subtotal: number;
  taxTotal: number;
  insuranceAmount: number;
  discountAmount: number;
  totalAmount: number;
}

export const EstimateLineItems = memo(function EstimateLineItems({
  items,
  subtotal,
  taxTotal,
  insuranceAmount,
  discountAmount,
  totalAmount,
}: EstimateLineItemsProps) {
  const sortedRows = useMemo(
    () =>
      items
        .slice()
        .sort((a, b) => a.sortOrder - b.sortOrder)
        .map((item, idx) => {
          const { total: lineTotal } = calcLineItemAmount(item);
          return (
            <TableRow key={item.id} className={`text-sm ${C.text}`}>
              <TableCell className={C.text40}>{idx + 1}</TableCell>
              <TableCell>{item.name}</TableCell>
              <TableCell className={C.text60}>
                {CATEGORY_LABELS[item.category] ?? item.category}
              </TableCell>
              <TableCell className="text-right font-mono">{formatCurrency(item.unitPrice)}</TableCell>
              <TableCell className="text-right">{item.quantity}</TableCell>
              <TableCell className="text-right">{Math.round(item.taxRate * 100)}%</TableCell>
              <TableCell className="text-right">
                {item.discountAmount > 0 ? `-${formatCurrency(item.discountAmount)}` : '-'}
              </TableCell>
              <TableCell className="text-right font-mono font-medium">
                {formatCurrency(lineTotal)}
              </TableCell>
            </TableRow>
          );
        }),
    [items],
  );

  return (
    <div className="space-y-3">
      <Table>
        <TableHeader>
          <TableRow className={`text-xs ${C.text60}`}>
            <TableHead className="w-[40px]">#</TableHead>
            <TableHead>品目名</TableHead>
            <TableHead className="w-[90px]">カテゴリ</TableHead>
            <TableHead className="w-[90px] text-right">単価</TableHead>
            <TableHead className="w-[60px] text-right">数量</TableHead>
            <TableHead className="w-[60px] text-right">税率</TableHead>
            <TableHead className="w-[90px] text-right">割引</TableHead>
            <TableHead className="w-[100px] text-right">金額</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.length === 0 ? (
            <TableRow>
              <TableCell data-empty-state colSpan={8} className={`text-center text-sm ${C.text40} py-6`}>
                明細がありません
              </TableCell>
            </TableRow>
          ) : (
            sortedRows
          )}
        </TableBody>
      </Table>

      {/* 合計 */}
      <div className="flex justify-end">
        <div className="w-[280px] space-y-1 text-sm">
          <div className={`flex justify-between ${C.text60}`}>
            <span>小計</span>
            <span className="font-mono">{formatCurrency(subtotal)}</span>
          </div>
          <div className={`flex justify-between ${C.text60}`}>
            <span>消費税</span>
            <span className="font-mono">{formatCurrency(taxTotal)}</span>
          </div>
          {insuranceAmount > 0 ? (
            <div className={`flex justify-between ${C.text60}`}>
              <span>保険適用額</span>
              <span className="font-mono">-{formatCurrency(insuranceAmount)}</span>
            </div>
          ) : null}
          {discountAmount > 0 ? (
            <div className={`flex justify-between ${C.text60}`}>
              <span>割引</span>
              <span className="font-mono">-{formatCurrency(discountAmount)}</span>
            </div>
          ) : null}
          <div className={`flex justify-between border-t ${C.borderLight} pt-2 font-semibold ${C.text}`}>
            <span>合計</span>
            <span className="font-mono">{formatCurrency(totalAmount)}</span>
          </div>
        </div>
      </div>
    </div>
  );
});
