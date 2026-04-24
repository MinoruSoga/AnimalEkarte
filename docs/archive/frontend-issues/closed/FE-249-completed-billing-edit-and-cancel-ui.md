# FE-249: 完了後修正 UI ロック解除 + 会計キャンセル機能

**Status**: Closed (2026-04-14, commit 8fcd1382)
**Priority**: High
**Affects**: `features/accounting/routes/AccountingDetail.tsx`, `features/accounting/api/`
**Date Created**: 2026-04-14
**Related**: BUG-371, BE-111

## Summary

`status=completed` の会計詳細画面で UI 一律ロックされている修正機能を、`accounting:edit` 権限保有者に限り解放する。修正開始時に確認モーダルを表示。併せて「会計をキャンセル」ボタンを追加し、`POST /accountings/:id/cancel` を呼ぶ。

## 現状のコード

### 完了後の UI ロック箇所
```typescript
// frontend/src/features/accounting/routes/AccountingDetail.tsx:604-618
{canSubmit ? (
  <SubmitButton
    className="w-full h-14 text-lg font-bold mt-4"
    size="lg"
    disabled={
      changeAmount < 0 ||
      !receivedAmount ||
      isCompleted          // ★ 完了済みなら無条件 disabled
    }
    loadingText="処理中..."
  >
    <Save className={`mr-2 ${ICON.action}`} />
    {isCompleted ? "精算完了済み" : "会計を確定する"}  // ★ ラベルも固定
  </SubmitButton>
) : null}
```

### 削除 API クライアント
```
frontend/src/features/accounting/api/ にdelete-accounting.ts は存在しない。
削除 UI も存在しない。
```

### 権限取得（既存）
```typescript
// frontend/src/features/accounting/routes/AccountingDetail.tsx:985
const { canEdit, canCreate, canDelete } = usePermission("accounting");
```

## 必要な変更

### 1. 型定義 / API hook（新規）

```typescript
// frontend/src/features/accounting/api/cancel-accounting.ts (新規)

import { useMutation, useQueryClient } from "@tanstack/react-query";
import axios from "@/lib/axios";
import type { BackendAccounting } from "./types";
import { transformBackendAccountingToFrontend } from "./transforms";

export async function cancelAccounting(
  clinicID: number,
  id: number,
): Promise<BackendAccounting> {
  const { data } = await axios.post<BackendAccounting>(
    `/api/clinics/${clinicID}/accountings/${id}/cancel`,
  );
  return data;
}

export function useCancelAccounting(clinicID: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => cancelAccounting(clinicID, id),
    onSuccess: (data) => {
      qc.invalidateQueries({ queryKey: ["accountings"] });
      qc.invalidateQueries({ queryKey: ["accounting", data.id] });
    },
  });
}
```

### 2. UI ロック解除（PaymentCard `disabled` 修正）

```typescript
// frontend/src/features/accounting/routes/AccountingDetail.tsx:604-618
// Before:
disabled={
  changeAmount < 0 ||
  !receivedAmount ||
  isCompleted
}

// After:
disabled={
  changeAmount < 0 ||
  !receivedAmount
  // isCompleted 条件を削除 — 完了後も canEdit 権限あれば修正可能
}

// ラベル変更:
{isCompleted ? "修正を保存する" : "会計を確定する"}
```

### 3. 修正確認モーダル（保存ボタン押下時に1回）

```typescript
// frontend/src/features/accounting/routes/AccountingDetail.tsx の formAction を wrap

const [editConfirmOpen, setEditConfirmOpen] = useState(false);
const [pendingFormData, setPendingFormData] = useState<FormData | null>(null);

// SubmitButton onClick で intercept:
const handleSubmitWithConfirm = useCallback(
  (e: React.FormEvent<HTMLFormElement>) => {
    if (accounting.status === "completed") {
      e.preventDefault();
      setPendingFormData(new FormData(e.currentTarget));
      setEditConfirmOpen(true);
      return;
    }
    // 通常フロー（completed 以外はそのまま実行）
  },
  [accounting.status],
);

const handleConfirmEdit = useCallback(() => {
  if (pendingFormData) {
    formAction(pendingFormData); // useActionState の formAction 呼び出し
  }
  setEditConfirmOpen(false);
  setPendingFormData(null);
}, [pendingFormData, formAction]);

// JSX:
<ConfirmDialog
  open={editConfirmOpen}
  onOpenChange={setEditConfirmOpen}
  title="精算済みの会計を修正します"
  description="この会計は既に精算完了しています。修正してよろしいですか?"
  confirmLabel="修正する"
  cancelLabel="キャンセル"
  onConfirm={handleConfirmEdit}
  variant="warning"
/>
```

### 4. キャンセルボタン追加

```typescript
// frontend/src/features/accounting/routes/AccountingDetail.tsx の PageLayout headerAction セクション

const cancelMutation = useCancelAccounting(clinicID);
const [cancelConfirmOpen, setCancelConfirmOpen] = useState(false);

const handleCancelBilling = useCallback(async () => {
  try {
    await cancelMutation.mutateAsync(Number(id));
    setCancelConfirmOpen(false);
    navigate(paths.accounting.getHref());
  } catch (error) {
    handleApiError(error, "会計キャンセル");
  }
}, [cancelMutation, id, navigate]);

// PageLayout の headerAction に追加（既存の領収書ボタンの隣）:
{canDelete && accounting.status !== "cancelled" ? (
  <Button
    variant="outline"
    size="sm"
    onClick={() => setCancelConfirmOpen(true)}
    className="text-red-600 hover:text-red-700"
  >
    <X className={`mr-2 ${ICON.action}`} />
    会計をキャンセル
  </Button>
) : null}

// JSX:
<ConfirmDialog
  open={cancelConfirmOpen}
  onOpenChange={setCancelConfirmOpen}
  title="この会計をキャンセルします"
  description="キャンセル後は元に戻せません。よろしいですか?"
  confirmLabel="キャンセルする"
  cancelLabel="やめる"
  onConfirm={handleCancelBilling}
  variant="destructive"
  isPending={cancelMutation.isPending}
/>
```

### 5. status を保持する Update ロジック調整

完了後の修正で `status` を `waiting` に戻さないため、`updateAccounting` 呼び出し時に **`status` フィールドを送信しない**。

```typescript
// frontend/src/features/accounting/routes/AccountingDetail.tsx:840 周辺
// Before:
status: completedPayment ? "completed" : baseAccounting.status,

// After (条件分岐):
// 既に completed の場合は status を update リクエストに含めない
// （BE は nil フィールドを更新しないため completed 維持される）
...(baseAccounting.status === "completed"
  ? {} // status を送らない
  : { status: completedPayment ? "completed" : baseAccounting.status }
),
```

### 6. Feature index export

```typescript
// frontend/src/features/accounting/index.ts に追加
export { useCancelAccounting } from "./api/cancel-accounting";
```

## UI 操作フロー

### 完了済会計の修正
1. 権限保有者が `status=completed` の会計詳細を開く
2. 明細・支払方法・受領額・保険率の各入力欄が編集可能（disabled 解除）
3. ボタンラベル「**修正を保存する**」表示
4. 値を変更して保存ボタン押下
5. 確認モーダル「**精算済みの会計を修正します。よろしいですか?**」表示
6. 「修正する」押下 → API 呼び出し → status=completed のまま保存される
7. 「キャンセル」押下 → モーダル閉じる、修正は実行されない

### 会計キャンセル
1. `accounting:delete` 権限保有者の画面に「**会計をキャンセル**」ボタンが表示
2. ボタン押下 → 確認モーダル「**この会計をキャンセルします。元に戻せません**」
3. 「キャンセルする」押下 → `POST /accountings/:id/cancel` 呼び出し
4. 成功 → 会計一覧画面に遷移、トースト表示
5. 既にキャンセル済 (`status=cancelled`) ならボタン非表示

### 権限なし
- 従来通り `disabled` のまま、キャンセルボタンも非表示

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（cancel mutation の pending 状態を SubmitButton と統合）
- [ ] 型は `models.ts` 由来の `BackendAccounting` を使用（手書き interface 禁止）
- [ ] デザイントークン `C`, `STYLE` 使用
- [ ] catch ブロックで `handleApiError` 使用
- [ ] `useCallback` でハンドラ安定化

## 依存関係

- BE-111 が先に完了している必要がある（`POST /accountings/:id/cancel` API）

## 完了条件

- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
- [ ] AC-1〜AC-14（BUG-371 参照）すべて達成
- [ ] 既存の `waiting` 状態の編集動作に変更なし
- [ ] `cancelled` 状態の会計が `AccountingList` に正しく表示される
- [ ] 既存の `RefundSection` が引き続き動作
- [ ] 権限なしユーザーは従来通り disabled、キャンセルボタン非表示
