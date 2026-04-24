# FE-127: 返金可能残額の表示とボタン無効化

**Status**: Closed
**Priority**: High
**Affects**: 会計精算画面（AccountingDetail.tsx）
**Date Created**: 2026-03-26
**Related**: TASK-031, BE-065

## Summary

`RefundSection` が元の請求金額を知らないため、返金可能残額（= 請求金額 - 返金済合計）を表示できていない。
また返金可能残額が ¥0 になっても「返金を登録」ボタンが無効化されない。
`totalAmount` prop を追加して両方を解決する。

## 現状のコード

```typescript
// frontend/src/features/accounting/routes/AccountingDetail.tsx:513-517
interface RefundSectionProps {
  billingId: string;
  isRefunding: boolean;
  onRefund: (amount: number, reason: string) => void;
  // totalAmount がない → 残額を計算できない
}
```

```typescript
// AccountingDetail.tsx:529
const totalRefunded = refunds.reduce((sum, r) => sum + r.amount, 0);
// totalRefunded は計算されるが残額は未表示・未チェック
```

```typescript
// AccountingDetail.tsx:554-558（呼び出し側）
<Button variant="outline" size="sm">
  返金を登録  {/* ← totalRefunded >= totalAmount でも常に有効 */}
</Button>
```

```typescript
// AccountingDetail.tsx:970-976（RefundSection レンダリング）
{id && accounting.status === "completed" ? (
  <RefundSection
    billingId={id}
    isRefunding={isRefunding}
    onRefund={handleRefund}
    // totalAmount が渡されていない
  />
) : null}
```

## 必要な変更

### 1. `RefundSectionProps` に `totalAmount` を追加

```typescript
// Before
interface RefundSectionProps {
  billingId: string;
  isRefunding: boolean;
  onRefund: (amount: number, reason: string) => void;
}

// After
interface RefundSectionProps {
  billingId: string;
  totalAmount: number;  // ← 追加: 元の請求金額（精算時の確定金額）
  isRefunding: boolean;
  onRefund: (amount: number, reason: string) => void;
}
```

### 2. `RefundSection` 本体に残額表示とボタン無効化を追加

```typescript
// Before（:519-529）
const RefundSection = memo(function RefundSection({
  billingId,
  isRefunding,
  onRefund,
}: RefundSectionProps) {
  const [refundDialogOpen, setRefundDialogOpen] = useState(false);
  const [refundAmount, setRefundAmount] = useState("");
  const [refundReason, setRefundReason] = useState("");
  const { data: refunds = [] } = useGetRefunds(billingId);

  const totalRefunded = refunds.reduce((sum, r) => sum + r.amount, 0);

// After
const RefundSection = memo(function RefundSection({
  billingId,
  totalAmount,   // ← 追加
  isRefunding,
  onRefund,
}: RefundSectionProps) {
  const [refundDialogOpen, setRefundDialogOpen] = useState(false);
  const [refundAmount, setRefundAmount] = useState("");
  const [refundReason, setRefundReason] = useState("");
  const { data: refunds = [] } = useGetRefunds(billingId);

  const totalRefunded = refunds.reduce((sum, r) => sum + r.amount, 0);
  const refundableAmount = totalAmount - totalRefunded;  // ← 追加
```

### 3. CardTitle に返金可能残額を表示

```tsx
// Before（:547-551）
{totalRefunded > 0 ? (
  <span className="text-xs font-normal text-orange-600 bg-orange-50 px-2 py-0.5 rounded">
    合計 ¥{totalRefunded.toLocaleString()} 返金済
  </span>
) : null}

// After
<span className="text-xs font-normal text-muted-foreground">
  返金可能残額 ¥{refundableAmount.toLocaleString()}
</span>
{totalRefunded > 0 ? (
  <span className="text-xs font-normal text-orange-600 bg-orange-50 px-2 py-0.5 rounded">
    合計 ¥{totalRefunded.toLocaleString()} 返金済
  </span>
) : null}
```

### 4. 「返金を登録」DialogTrigger ボタンを無効化

```tsx
// Before（:554-558）
<DialogTrigger asChild>
  <Button variant="outline" size="sm" className="h-8 text-xs">
    <Plus className="mr-1 h-3 w-3" />
    返金を登録
  </Button>
</DialogTrigger>

// After
<DialogTrigger asChild>
  <Button
    variant="outline"
    size="sm"
    className="h-8 text-xs"
    disabled={refundableAmount <= 0}  // ← 追加
  >
    <Plus className="mr-1 h-3 w-3" />
    返金を登録
  </Button>
</DialogTrigger>
```

### 5. 呼び出し側に `totalAmount` を渡す

```tsx
// Before（AccountingDetail.tsx:970-976）
{id && accounting.status === "completed" ? (
  <RefundSection
    billingId={id}
    isRefunding={isRefunding}
    onRefund={handleRefund}
  />
) : null}

// After
{id && accounting.status === "completed" ? (
  <RefundSection
    billingId={id}
    totalAmount={accounting.payment?.totalAmount ?? 0}  // ← 追加
    isRefunding={isRefunding}
    onRefund={handleRefund}
  />
) : null}
```

`accounting.payment` は `completed` 状態のときに必ず存在するため `?? 0` は安全なフォールバック。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useCallback` deps は primitive（`totalAmount: number` で問題なし）

## 依存関係

BE-065 と独立して実装可能。ただし BE-065 未実装時は UI 側でボタンを無効化しても BE が過剰返金を通してしまうため、理想は BE-065 先行。

## 完了条件

- [ ] 返金可能残額 `¥{refundableAmount}` が返金管理カードに表示される
- [ ] 返金可能残額が ¥0 になると「返金を登録」ボタンが `disabled` になる
- [ ] 残額がある場合は従来通り返金フォームが開ける
- [ ] `pnpm build` 型エラーなし
- [ ] `pnpm lint` エラーなし
