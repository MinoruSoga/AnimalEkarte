# FE-126: 会計返金 UI 実装

**Status**: Closed
**Priority**: High
**Affects**: 会計一覧（Accounting.tsx）、会計精算画面（AccountingDetail.tsx）
**Date Created**: 2026-03-26
**Related**: TASK-031, BE-062

## Summary

BE-062 で実装される返金 API を呼び出し、
① 会計一覧に「返金あり」バッジを追加、
② 会計精算画面（AccountingDetail）に返金フォーム・返金履歴・返金可能残額を追加する。

## 現状のコード

```typescript
// frontend/src/features/accounting/types/index.ts:80
export interface Accounting {
  id: string;
  medicalRecordId?: string;
  ownerId: string;
  ownerName: string;
  petId: string;
  petName: string;
  petSpecies?: string;
  status: AccountingStatus;
  scheduledDate: string;
  completedAt?: string;
  items: AccountingItem[];
  payment?: PaymentInfo;
  memo?: string;
  // refunds は未定義
}
```

```typescript
// frontend/src/features/accounting/api/transforms.ts:44
export function transformToAccounting(data: BackendAccounting): Accounting {
  return {
    id: String(data.id ?? 0),
    // ... refunds 変換なし ...
  };
}
```

```typescript
// frontend/src/features/accounting/routes/Accounting.tsx:263-307
// renderRow: 返金バッジなし
const renderRow = useCallback(
  (r: AccountingType) => {
    const statusLabel = ACCOUNTING_STATUS_LABELS[r.status] ?? r.status;
    return (
      <DataTableRow key={r.id} onClick={() => handleEdit(r.id)}>
        {/* ... ステータスバッジはあるが返金バッジなし */}
      </DataTableRow>
    );
  },
  [handleEdit, navigate],
);
```

```typescript
// frontend/src/features/accounting/routes/AccountingDetail.tsx:1-80（冒頭のみ）
// 「返金する」ボタン・返金フォーム・返金履歴なし
```

## 必要な変更

### 1. 型定義

```typescript
// frontend/src/features/accounting/types/index.ts に追加

/** @see {@link import("@/types/generated/models").BillingRefund} */
export interface Refund {
  id: string;
  billingId: string;
  amount: number;
  reason: string;
  refundedAt: string;
  createdAt: string;
}

// Accounting 型に追加
export interface Accounting {
  // ... 既存フィールド ...
  totalRefundedAmount?: number;  // GET /accountings 一覧から取得
  refunds?: Refund[];            // GET /accountings/:id 詳細から取得
}
```

### 2. API hooks

```typescript
// frontend/src/features/accounting/api/get-refunds.ts（新規）
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES } from "@/lib/react-query";
import type { Refund } from "../types";
import type { BillingRefund } from "@/types/generated/models";

function transformToRefund(r: BillingRefund): Refund {
  return {
    id: String(r.id ?? 0),
    billingId: String(r.billing_id ?? 0),
    amount: r.amount ?? 0,
    reason: r.reason ?? "",
    refundedAt: r.refunded_at ?? "",
    createdAt: r.created_at ?? "",
  };
}

interface GetRefundsResponse {
  data: BillingRefund[];
}

export async function getRefunds(billingId: string): Promise<Refund[]> {
  const { data } = await axios.get<GetRefundsResponse>(`/v1/accountings/${billingId}/refunds`);
  return data.data.map(transformToRefund);
}

export function useGetRefunds(billingId: string) {
  return useQuery({
    queryKey: ["accountings", billingId, "refunds"],
    queryFn: () => getRefunds(billingId),
    staleTime: QUERY_STALE_TIMES.REALTIME,
    enabled: !!billingId && billingId !== "0", // 無効 ID での API 呼び出しを防ぐ
  });
}
```

```typescript
// frontend/src/features/accounting/api/create-refund.ts（新規）
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { toast } from "sonner";
import type { CreateRefundRequest } from "./types";

export async function createRefund(billingId: string, data: CreateRefundRequest): Promise<void> {
  await axios.post(`/v1/accountings/${billingId}/refunds`, data);
}

export function useCreateRefund(billingId: string, onSuccessCallback?: () => void) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateRefundRequest) => createRefund(billingId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["accountings", billingId, "refunds"] });
      queryClient.invalidateQueries({ queryKey: ["accountings", billingId] });
      queryClient.invalidateQueries({ queryKey: ["accountings"] });
      toast.success("返金処理が完了しました");
      onSuccessCallback?.(); // Dialog クローズ等の呼び出し元コールバック
    },
    onError: () => {
      toast.error("返金処理に失敗しました");
    },
  });
}
```

```typescript
// frontend/src/features/accounting/api/types.ts に追加
export interface CreateRefundRequest {
  amount: number;  // > 0
  reason?: string; // 任意
}
```

### 3. transforms.ts 変更

```typescript
// frontend/src/features/accounting/api/transforms.ts

// BackendAccounting に total_refunded_amount を追加
export type BackendAccounting = Billing & {
  total_refunded_amount?: number;
};

// transformToAccounting に追加
export function transformToAccounting(data: BackendAccounting): Accounting {
  return {
    // ... 既存フィールド ...
    totalRefundedAmount: data.total_refunded_amount ?? 0,
  };
}
```

### 4. 会計一覧（Accounting.tsx）— 返金あり バッジ追加

```typescript
// frontend/src/features/accounting/routes/Accounting.tsx

// import 追加（Badge はこのファイルに現在 import されていない）
import { Badge } from "@/components/ui/badge";

// renderRow 内の StatusBadge の後に追加
<TableCell className="py-2">
  <StatusBadge colorClass={getAccountingStatusColor(statusLabel)}>
    {statusLabel}
  </StatusBadge>
  {(r.totalRefundedAmount ?? 0) > 0 ? (
    <Badge variant="outline" className="ml-1 text-xs text-orange-600 border-orange-300">
      返金あり
    </Badge>
  ) : null}
</TableCell>
```

### 5. 会計精算画面（AccountingDetail.tsx）— 返金セクション追加

```typescript
// frontend/src/features/accounting/routes/AccountingDetail.tsx

// import 追加
import { useGetRefunds } from "../api/get-refunds";
import { useCreateRefund } from "../api/create-refund";
import { formatCurrency } from "@/utils/format/number";

// コンポーネント内に追加（completed の場合のみ表示）

// id は useParams() 由来で string | undefined。フック引数は string を要求するため ?? "" でフォールバック。
// useGetRefunds / useCreateRefund 内の enabled ガードにより id が空の場合は API 呼び出しをスキップする。
const billingId = id ?? "";

// 返金一覧取得（useGetRefunds は Rules of Hooks 準拠のため completed 条件に関わらず常に呼ぶ）
const { data: refunds = [] } = useGetRefunds(billingId);

// 返金可能残額の計算
const totalRefundedAmount = refunds.reduce((sum, r) => sum + r.amount, 0);
const refundableAmount = (accounting.payment?.totalAmount ?? 0) - totalRefundedAmount;

// 返金フォーム（useTransition で pending 管理）
const [isRefundDialogOpen, setIsRefundDialogOpen] = useState(false); // Dialog 開閉制御
const [isRefundPending, startRefundTransition] = useTransition();
const [refundAmount, setRefundAmount] = useState<number>(0);
const [refundReason, setRefundReason] = useState("");
const { mutateAsync: createRefundMutateAsync } = useCreateRefund(billingId, () => {
  // onSuccess コールバック: Dialog を閉じてフォームをリセット
  setIsRefundDialogOpen(false);
  setRefundAmount(0);
  setRefundReason("");
});

const handleRefundSubmit = useCallback(() => {
  if (refundAmount <= 0 || refundAmount > refundableAmount) return;
  startRefundTransition(async () => {
    await createRefundMutateAsync({
      amount: refundAmount,
      reason: refundReason || undefined,
    });
  });
}, [refundAmount, refundReason, refundableAmount, createRefundMutateAsync]);
```

**返金セクション JSX（completed の会計詳細画面下部に追加）:**

```tsx
{accounting.status === "completed" ? (
  <Card>
    <CardHeader>
      <CardTitle className="text-sm">返金管理</CardTitle>
    </CardHeader>
    <CardContent className="space-y-4">
      {/* 返金可能残額 */}
      <div className="flex justify-between text-sm">
        <span className="text-muted-foreground">返金可能残額</span>
        <span className={refundableAmount === 0 ? "text-muted-foreground" : "font-medium"}>
          {formatCurrency(refundableAmount)}
        </span>
      </div>

      {/* 返金フォーム（残額 > 0 の場合のみ） */}
      {refundableAmount > 0 ? (
        <Dialog open={isRefundDialogOpen} onOpenChange={setIsRefundDialogOpen}>
          <DialogTrigger asChild>
            <Button variant="outline" size="sm">返金する</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>返金処理</DialogTitle>
              <DialogDescription>
                返金可能残額: {formatCurrency(refundableAmount)}
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              <div>
                <Label>返金額（円）</Label>
                <NumberInput
                  value={refundAmount}
                  onChange={setRefundAmount}
                  min={1}
                  max={refundableAmount}
                />
              </div>
              <div>
                <Label>返金理由（任意）</Label>
                <Input
                  value={refundReason}
                  onChange={(e) => setRefundReason(e.target.value)}
                  placeholder="例: 診断ミスのため"
                />
              </div>
            </div>
            <DialogFooter>
              <Button
                onClick={handleRefundSubmit}
                disabled={isRefundPending || refundAmount <= 0 || refundAmount > refundableAmount}
              >
                {isRefundPending ? "処理中..." : "返金実行"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      ) : null}

      {/* 返金履歴リスト */}
      {refunds.length > 0 ? (
        <div className="space-y-1">
          <p className="text-xs text-muted-foreground">返金履歴</p>
          {refunds.map((refund) => (
            <div key={refund.id} className="flex justify-between text-sm py-1 border-b last:border-0">
              <div>
                <span className="font-mono text-xs text-muted-foreground">{refund.refundedAt.slice(0, 10)}</span>
                {refund.reason ? (
                  <span className="ml-2 text-muted-foreground">{refund.reason}</span>
                ) : null}
              </div>
              <span className="font-medium text-orange-600">-{formatCurrency(refund.amount)}</span>
            </div>
          ))}
        </div>
      ) : null}
    </CardContent>
  </Card>
) : null}
```

## UI 操作フロー

1. 会計一覧で返金済みの会計行に「返金あり」バッジが表示される
2. 会計詳細（completed）画面下部に「返金管理」カードが表示される
3. 「返金する」ボタンをクリック → Dialog が開く
4. 返金額（必須）・返金理由（任意）を入力して「返金実行」
5. `POST /v1/accountings/:id/refunds` が呼ばれ、成功トーストが表示される
6. 返金履歴リストに追加、返金可能残額が更新される
7. 返金可能残額が ¥0 になると「返金する」ボタン（Dialog トリガー）が非表示になる

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理
- [ ] 型は `models.ts` から導出（BillingRefund → Refund 変換）

## api/index.ts への export 追記

```typescript
// frontend/src/features/accounting/api/index.ts に追記
export { getRefunds, useGetRefunds } from "./get-refunds";
export { createRefund, useCreateRefund } from "./create-refund";
// types.ts の既存 export 行に CreateRefundRequest を追加
export type { ..., CreateRefundRequest } from "./types";
```

## 依存関係

- BE-062 が完了していること（API エンドポイント + make codegen 済み）
- `frontend/src/types/generated/models.ts` に `BillingRefund` 型が存在すること

## 完了条件

- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
- [ ] 会計一覧で `total_refunded_amount > 0` の行に「返金あり」バッジが表示される
- [ ] 会計精算画面（completed）に「返金管理」セクションが表示される
- [ ] 返金フォームで入力・送信できる
- [ ] 返金履歴リストに返金が表示される
- [ ] 返金可能残額が正しく計算・表示される
- [ ] 返金可能残額 = ¥0 で「返金する」ボタンが非表示になる
- [ ] waiting / cancelled 会計には「返金管理」セクションが表示されない
