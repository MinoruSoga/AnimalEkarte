# BUG-229: useAuth の値をコールバック内でのみ使用している（2箇所）

## 概要

`useAuth()` から取得した `user` をレンダー出力（JSX）で使わず、
コールバック関数の中でのみ使っている箇所が 2件ある。
これにより `user` が変わるたびに（ログイン/ログアウト等）コンポーネントが不要なレンダーを行う。
`rerender-defer-reads` ルールでは、コールバックのみで使う状態は
`useRef` で保持してレンダーをトリガーしないことを推奨している。

## 現状コード（2箇所 — 実コード確認済み）

### 1. `features/medical-records/components/BillingReviewSection/BillingReviewSection.tsx:49,57`

```typescript
// ❌ user をコンポーネント state として購読しているが、JSX では使わない
const { user } = useAuth();

const handleConfirm = useCallback(() => {
  confirmMutation.mutate(
    { confirmed_by: Number(user?.id ?? 0) },  // ← コールバック内でのみ使用
    { ... }
  );
}, [confirmMutation, user?.id]);  // user?.id が deps に入る → user 変更でコールバック再生成
```

JSX 出力（lines 102-end）に `user` は一切現れない。
`useAuth()` の購読がコンポーネントを余分にレンダーさせる原因となっている。

### 2. `features/medical-records/api/billing-review.ts:56,63`

```typescript
// ❌ カスタムフック内で useAuth() を購読しているが mutationFn 内でのみ使用
export function useReturnBillingReview() {
  const { user } = useAuth();

  return useMutation({
    mutationFn: async (reason: string) => {
      await reviewApi.returnReview(user?.id ?? 0, reason);  // ← mutationFn 内でのみ
    },
  });
}
```

`mutationFn` はミューテーション実行時に呼ばれるため、`user` をクロージャで捕捉しても
`user` 変更時に古い値が使われる可能性がある（ただし auth user はほぼ変わらない）。

## 比較: 正しい実装（プロジェクト内参照実装）

```typescript
// rerender-defer-reads の推奨パターン — useRef でトランジェント値を保持
const { user } = useAuth();
const userRef = useRef(user);
// レンダー中に ref を更新（useEffect は不要）
userRef.current = user;

const handleConfirm = useCallback(() => {
  confirmMutation.mutate(
    { confirmed_by: Number(userRef.current?.id ?? 0) },
  );
}, [confirmMutation]);  // user?.id が deps から外れてコールバックが安定
```

## 修正方針

### 1. BillingReviewSection.tsx

```typescript
// Before:
const { user } = useAuth();
const handleConfirm = useCallback(() => {
  confirmMutation.mutate({ confirmed_by: Number(user?.id ?? 0) }, { ... });
}, [confirmMutation, user?.id]);

// After:
const { user } = useAuth();
const userIdRef = useRef<number>(Number(user?.id ?? 0));
userIdRef.current = Number(user?.id ?? 0);  // 毎レンダーで最新値に更新（副作用なし）

const handleConfirm = useCallback(() => {
  confirmMutation.mutate({ confirmed_by: userIdRef.current }, { ... });
}, [confirmMutation]);  // ✅ deps から user を除去、コールバックが安定
```

### 2. billing-review.ts の useReturnBillingReview

このケースは `mutationFn` が呼ばれる時点で最新の `user?.id` が必要。
`useCallback` ではないため ref パターンよりも、上流コンポーネントから `userId` を渡す設計が清潔:

```typescript
// Before:
export function useReturnBillingReview() {
  const { user } = useAuth();
  return useMutation({
    mutationFn: async (reason: string) => {
      await reviewApi.returnReview(user?.id ?? 0, reason);
    },
  });
}

// After: userId を引数として受け取る（hook からの useAuth 依存を排除）
export function useReturnBillingReview(userId: number) {
  return useMutation({
    mutationFn: async (reason: string) => {
      await reviewApi.returnReview(userId, reason);
    },
  });
}

// 呼び出し側 (BillingReviewSection.tsx) で:
const { user } = useAuth();
const userId = user?.id ?? 0;
const returnMutation = useReturnBillingReview(userId);
// → useAuth の購読はコンポーネント側に集約、hook は純粋
```

## 影響範囲

| ファイル | 行 | 問題 |
|---------|-----|------|
| `features/medical-records/components/BillingReviewSection/BillingReviewSection.tsx` | 49,57,64 | `user` がコールバックのみで使用 |
| `features/medical-records/api/billing-review.ts` | 56,63 | `user` が mutationFn のみで使用 |

## 準拠すべきプロジェクト規約・ベストプラクティス

### Vercel React Best Practices — `rerender-defer-reads`
> Don't subscribe to state only used in callbacks.
> Use refs to read state in callbacks without triggering re-renders.

### Vercel React Best Practices — `rerender-use-ref-transient-values`
> Use refs for transient frequent values to avoid re-renders.

## 優先度

**Low** — `user` が変わるのはログイン/ログアウト時のみで頻度は極めて低い。
実害（不正な `confirmed_by` 等）はない。コールバックの安定性向上が主な効果。
