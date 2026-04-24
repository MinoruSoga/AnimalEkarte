# 011: カルテ会計医師確認ワークフロー実装

## 概要
カルテ詳細画面に医師確認ステータスを表示し、確認・差し戻し操作を可能にする。差し戻し時は返却理由の入力モーダルを表示する。会計一覧でも確認ステータスを参照できるようにする。

## 優先度
high

## 関連APIエンドポイント
- GET `/v1/medical-records/{id}/billing-review`
- POST `/v1/medical-records/{id}/billing-review/confirm`
- POST `/v1/medical-records/{id}/billing-review/return`

## 関連バックエンドチケット
backend/issues/open/007_billing_review.md（実装済み）

## 実装内容

### API層 (`features/medical-records/api/`)
`billing-review.ts` を新規作成し以下の TanStack Query hooks を実装する。
- `useBillingReview(medicalRecordId: string)` — GET 確認ステータス取得
- `useConfirmBillingReview(medicalRecordId: string)` — POST 確認操作
- `useReturnBillingReview(medicalRecordId: string)` — POST 差し戻し操作（`return_reason` を body で送信）

### コンポーネント (`features/medical-records/components/`)
`BillingReviewSection/` ディレクトリを新規作成する。
- `BillingReviewSection.tsx` — ステータスバッジ + 確認ボタン + 差し戻しボタン
- `ReturnReasonDialog.tsx` — 差し戻し理由入力モーダル（`Textarea` + 送信ボタン）
- `BillingReviewSection/index.ts` — named export

ステータスバッジの表示仕様:
- `pending` — 「確認待ち」（yellow）
- `confirmed` — 「確認済み」（green）
- `returned` — 「差し戻し」（red）

ボタンの活性/非活性:
- `confirmed` 状態では「確認」ボタンを `disabled` にする
- `returned` 状態では「確認」ボタンを再度活性化する（再確認フロー）

### ページ/ルート (`features/medical-records/routes/`)
カルテ詳細ルートコンポーネントの診察タブまたはヘッダー領域に `BillingReviewSection` を組み込む。

### 型定義 (`features/medical-records/types/`)
`index.ts` に以下を追加する。
```typescript
export type BillingReviewStatus = 'pending' | 'confirmed' | 'returned';

export interface BillingReview {
  id: string;
  medical_record_id: string;
  status: BillingReviewStatus;
  confirmed_by?: string | null;
  confirmed_at?: string | null;
  return_reason?: string | null;
  returned_by?: string | null;
  returned_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface ReturnBillingReviewInput {
  return_reason: string;
}
```

## 完了条件
- [ ] 確認ステータスバッジがカルテ詳細画面に常時表示される
- [ ] 「確認」ボタンで `confirmed` に変更できる
- [ ] 「差し戻し」ボタンクリックで返却理由入力ダイアログが開く
- [ ] 返却理由を入力して送信すると `returned` に変更される
- [ ] `confirmed` 状態では「確認」ボタンが `disabled` になる
- [ ] ミューテーション成功後にステータスが即時反映される（キャッシュ invalidate）
- [ ] ESLint エラー 0 件
- [ ] `docker compose exec frontend pnpm build` が通る

## 備考
- バックエンドチケット 007 は実装済みのためフロントエンド単独で進められる
- `return_reason` は必須入力とし、空文字では送信ボタンを `disabled` にする
- 会計一覧への確認ステータス表示は別チケットとして切り出してもよい
