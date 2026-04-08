# BUG-229: useAuth の user をコールバック内のみで使用（BillingReviewSection, billing-review.ts）

## 概要
`useAuth()` から取得した `user` をコンポーネントのレンダーには使わず、コールバック（イベントハンドラ・useActionState アクション）内でのみ使用している。これは `rerender-defer-reads` 違反で、auth コンテキストが更新されるたびにコンポーネントが不要に再レンダーされる。

## 現状コード

### `features/accounting/components/BillingReviewSection.tsx`（推定）
```typescript
// ❌ user はレンダーに使わずコールバック内だけで参照
const { user } = useAuth();

const handleSubmit = useCallback(async () => {
  await submitBilling({ staffId: user?.id }); // コールバック内のみ
}, [user]);
```

### `features/accounting/hooks/billing-review.ts`（推定）
```typescript
// ❌ 同様パターン — useAuth を hook 内でサブスクライブ
const { user } = useAuth();

const [formState, formAction] = useActionState(async (prev, fd) => {
  const result = await api({ staffId: user?.id }); // アクション内のみ
  // ...
}, null);
```

## 修正方針

`useRef` に user を保持し、コールバック内で ref を読む（`rerender-defer-reads` パターン）。

```typescript
// ✅ ref でコールバック内の読み取りを遅延
import { useRef, useEffect } from "react";
import { useAuth } from "@/features/auth";

const { user } = useAuth();
const userRef = useRef(user);
useEffect(() => { userRef.current = user; }, [user]);

const handleSubmit = useCallback(async () => {
  await submitBilling({ staffId: userRef.current?.id }); // ref 経由
}, []); // user を deps から除外
```

## 準拠すべきプロジェクト規約

### `frontend/CODING_RULES.md` Section 12 — rerender-defer-reads
> useAuth の user をコールバック内のみで使用する場合は `useRef` パターンに変更し、コンテキスト更新による不要な再レンダーを防ぐ

## 優先度
**Low** — 不要な再レンダーの抑制。機能的影響なし。修正は20分（2ファイル）。

## 関連ファイル
- `frontend/src/features/accounting/components/BillingReviewSection.tsx`
- `frontend/src/features/accounting/hooks/billing-review.ts`（または類似パス）
