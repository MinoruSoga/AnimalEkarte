/**
 * 明細行アイテムの共通型・計算ユーティリティ
 *
 * TreatmentTable（medical-records）と EstimateLineItems（estimates）の
 * 金額計算ロジックを共有するために使用する。UI 構造は統合しない。
 */

/**
 * 明細行アイテムの共通最小構造
 * TreatmentItem / EstimateLineItem 両方に対応できる
 */
interface LineItemBase {
  unitPrice: number;
  quantity: number;
  discountAmount?: number;
  discountRate?: number;
}

/**
 * 明細行アイテム1件の金額計算
 *
 * 割引額の優先順位:
 *   1. discountAmount が指定されていれば直接使用
 *   2. discountRate がある場合は小計に乗じて計算
 *   3. 上記いずれもない場合は 0
 */
export function calcLineItemAmount(item: LineItemBase): {
  subtotal: number;
  discount: number;
  total: number;
} {
  const subtotal = item.unitPrice * item.quantity;
  const discount =
    item.discountAmount != null
      ? item.discountAmount
      : item.discountRate
        ? Math.floor(subtotal * item.discountRate)
        : 0;
  return {
    subtotal,
    discount,
    total: subtotal - discount,
  };
}
