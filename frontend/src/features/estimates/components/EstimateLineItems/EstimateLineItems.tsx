import { memo, useMemo } from 'react';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { formatCurrency } from '@/utils/format/number';
import { calcLineItemAmount } from '@/utils/line-item-helpers';
import type { EstimateLineItem } from '@/features/estimates/types';
import { C } from '@/lib/design-tokens';

const CATEGORY_LABELS: Record<string, string> = {
  examination: '診察',
  test: '検査',
  procedure: '処置',
  surgery: '手術',
  medicine: '薬剤',
  food: '食事',
  goods: '物品',
  other: 'その他',
};

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
              <TableCell className={`${C.text40} py-2`}>{idx + 1}</TableCell>
              <TableCell className="py-2">{item.name}</TableCell>
              <TableCell className={`py-2 ${C.text60}`}>
                {CATEGORY_LABELS[item.category] ?? item.category}
              </TableCell>
              <TableCell className="text-right font-mono py-2">{formatCurrency(item.unitPrice)}</TableCell>
              <TableCell className="text-right py-2">{item.quantity}</TableCell>
              <TableCell className="text-right py-2">{Math.round(item.taxRate * 100)}%</TableCell>
              <TableCell className="text-right py-2">
                {item.discountAmount > 0 ? `-${formatCurrency(item.discountAmount)}` : '-'}
              </TableCell>
              <TableCell className="text-right font-mono font-medium py-2">
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
              <TableCell colSpan={8} className={`text-center text-sm ${C.text40} py-6`}>
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
