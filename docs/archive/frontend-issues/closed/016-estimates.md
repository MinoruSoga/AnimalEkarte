# 016: 見積書 CRUD 実装

## 概要
見積書の作成・閲覧・編集・削除画面を実装する。見積明細の管理も含み、合計金額・税額・保険適用額を表示する。

## 優先度
medium

## 関連APIエンドポイント
- GET `/v1/estimates`
- POST `/v1/estimates`
- GET `/v1/estimates/{id}`
- PATCH `/v1/estimates/{id}`
- DELETE `/v1/estimates/{id}`

## 関連バックエンドチケット
なし（バックエンド実装状況を確認してから着手すること）

## 実装内容

### API層 (`features/estimates/api/`)
以下のファイルを新規作成する。
- `get-estimates.ts` — `useEstimates()` hook（一覧取得、ページネーション対応）
- `get-estimate.ts` — `useEstimate(id: string)` hook（詳細取得）
- `create-estimate.ts` — `useCreateEstimate()` mutation hook
- `update-estimate.ts` — `useUpdateEstimate()` mutation hook
- `delete-estimate.ts` — `useDeleteEstimate()` mutation hook
- `types.ts` — API リクエスト/レスポンス型
- `transforms.ts` — バックエンド ↔ フロントエンド変換

### コンポーネント (`features/estimates/components/`)
- `EstimateStatusBadge/` — ステータスバッジ（draft/sent/accepted/rejected/expired）
- `EstimateLineItems/` — 見積明細テーブル（追加・編集・削除対応）

ステータスバッジの色分け:
- `draft` — 「下書き」（gray）
- `sent` — 「送付済み」（blue）
- `accepted` — 「承認済み」（green）
- `rejected` — 「却下」（red）
- `expired` — 「期限切れ」（orange）

### ページ/ルート (`features/estimates/routes/`)
- `EstimateList.tsx` — 見積一覧（ページネーション・ステータスフィルタ）
- `EstimateDetail.tsx` — 見積詳細（見積明細一覧・合計金額・税額・保険適用額）
- `EstimateForm.tsx` — 見積作成・編集フォーム

### 型定義 (`features/estimates/types/`)
`index.ts` を新規作成し以下を定義する。
```typescript
export type EstimateStatus = 'draft' | 'sent' | 'accepted' | 'rejected' | 'expired';

export interface Estimate {
  id: string;
  estimate_no: string;
  owner_id: string;
  pet_id: string;
  status: EstimateStatus;
  subtotal: number;
  tax_amount: number;
  insurance_amount: number;
  total_amount: number;
  valid_until?: string | null;  // YYYY-MM-DD
  note?: string | null;
  created_at: string;
  updated_at: string;
}

export interface EstimateLineItem {
  id: string;
  estimate_id: string;
  name: string;
  unit_price: number;
  quantity: number;
  discount_amount: number;
  insurance: boolean;
  sort_order: number;
}
```

## 完了条件
- [ ] 見積一覧ページ（`/estimates`）でステータスフィルタ・ページネーションが動作する
- [ ] 見積詳細ページ（`/estimates/:id`）で見積番号・明細・金額サマリが表示される
- [ ] 見積作成フォームで新規見積を作成できる
- [ ] 見積編集フォームで既存見積を更新できる
- [ ] 見積を削除できる（確認ダイアログあり）
- [ ] `valid_until`（有効期限）が表示される
- [ ] ESLint エラー 0 件
- [ ] `docker compose exec frontend pnpm build` が通る

## 備考
- バックエンド API の実装状況を事前に `backend/docs/api.yaml` で確認すること
- `features/estimates/` は新規 feature ディレクトリとして作成する
- ルーターへの追加は `app/router.tsx` で行うこと
