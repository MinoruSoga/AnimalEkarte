# FE-153: AccountingDetail — 支払フォームUI（支払方法・金額入力・課税区分）に canEdit ガードなし

## 概要

`AccountingDetail.tsx` の `PaymentCard` と `ItemListCard` において、「会計を確定する」ボタンと「返金する」ボタンは `canEdit` でガードされているが、支払フォームのUI要素に `canEdit` チェックがない。

## 影響ファイル

| コンポーネント | 問題箇所 | 詳細 |
|--------------|---------|------|
| `AccountingDetail.tsx` `PaymentCard` | 行 491-516 | 「現金」「カード」「電子マネー」ボタン — canEdit なし |
| `AccountingDetail.tsx` `PaymentCard` | 行 519-560 | お預かり金額 NumberInput / 丁度・千円単位・一万単位ボタン — canEdit なし |
| `AccountingDetail.tsx` `ItemListCard` | 行 159-178 | `TaxTypeSelector`・`TaxRateSelector` — `billingId !== undefined` チェックのみで canEdit なし |

注: 「会計を確定する」ボタン（行 572）と「返金する」ボタン（行 642）はすでに `canEdit` でガード済み ✅

## 現状の挙動（バグ）

```tsx
// PaymentCard — canEdit チェックなし
<div className="grid grid-cols-3 gap-2">
  <Button onClick={() => onPaymentMethodChange("cash")}>現金</Button>
  <Button onClick={() => onPaymentMethodChange("credit_card")}>カード</Button>
  <Button onClick={() => onPaymentMethodChange("electronic_money")}>電子マネー</Button>
</div>
<NumberInput value={receivedAmount} onChange={onReceivedAmountChange} />
<Button onClick={() => onReceivedAmountChange(billingAmount.toString())}>丁度</Button>

// ItemListCard — billingId の存在チェックのみ（canEdit なし）
{billingId !== undefined && onUpdateItemTax !== undefined ? (
  <TaxTypeSelector onChange={(v) => onUpdateItemTax(item.id, v, item.taxRate)} />
) : ...}
```

閲覧のみユーザーが会計詳細を開くと：
1. 支払方法ボタン（現金/カード/電子マネー）が操作可能
2. お預かり金額の入力フィールドが入力可能
3. 丁度・千円単位・一万単位ボタンが操作可能
4. 課税区分・税率のドロップダウンが操作可能（変更時に API が呼ばれ 403 になる）

## 期待する挙動

`canEdit` が false の場合、支払フォームUI全体を非表示にするか、disabled にする。

## 修正方針

```tsx
// PaymentCard — canEdit でフォームUIをガード
{canEdit ? (
  <div className="space-y-4">
    {/* 支払方法ボタン */}
    {/* お預かり金額入力 */}
    {/* 丁度・千円単位ボタン */}
  </div>
) : (
  // 読み取り専用表示（支払方法のみ表示）
  <div>
    <Label>支払方法</Label>
    <p>{paymentMethodLabel}</p>
  </div>
)}

// ItemListCard — canEdit をチェックに追加
{billingId !== undefined && onUpdateItemTax !== undefined && canEdit ? (
  <TaxTypeSelector ... />
) : (
  <span>{item.taxType === "excluded" ? "外税" : ...}</span>
)}
```

## 優先度

MEDIUM — 「会計確定」「返金」の実行ボタンはガードされているため実際のデータ変更は防げるが、操作UIが見えることでユーザーが混乱する。課税区分変更は API 403 エラーを発生させる。

## 関連

- `frontend/src/features/accounting/routes/AccountingDetail.tsx`
  - `PaymentCard` (行 462-590)
  - `ItemListCard` (行 88-450)
- BUG-RBAC テスト 2026-04-07
