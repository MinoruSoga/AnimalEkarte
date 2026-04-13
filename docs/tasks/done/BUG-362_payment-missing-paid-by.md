# BUG-362: Payment の書き込みフロー未実装 + 処理者（paid_by）欠落

## 優先度: HIGH

## 概要

### 問題1: Payment 書き込みフローが存在しない

フロントエンド（AccountingDetail.tsx）は会計完了時に以下を送信している:
```
payment_method, received_amount, change_amount, billing_amount,
insurance_ratio, insurance_amount, subtotal, tax_total, total_amount, completed_at
```

しかしバックエンドの `UpdateAccountingInput` にはこれらのフィールドが**存在しない**。
`accounting_service.go` の `buildBillingUpdateFields` は billings テーブルのみ更新し、
payments テーブルへの書き込みは一切行われていない。

**結果**: 会計完了時の支払情報（支払方法、受領額、おつり等）がDBに保存されない。

### 問題2: paid_by（支払処理者）カラム欠落

payments テーブルに支払処理を実行したスタッフを記録するカラムがない。

## 現状の構造

- `payments` テーブル: DB定義あり、Go Model あり、Preload で読み取りはされる
- 専用の Handler/Service/Repository: **存在しない**
- Create/Update API: **存在しない**
- フロントエンド: `updateAccounting()` にフラットに Payment フィールドを混ぜて送信しているが、バックエンドは無視している

## 修正方針

### Option A: Billing 更新時に Payment を同時作成/更新（推奨）

会計完了（status → completed）時に Payment を自動作成する方式。
フロントエンドの既存 API 呼び出しを変更せずに済む。

1. `UpdateAccountingInput` に payment 関連フィールド追加
2. `accounting_service.go` で status=completed への変更時に Payment を upsert
3. `paid_by` は handler から extractStaffID で渡す

### Option B: Payment 専用 CRUD エンドポイント追加

独立した `/accountings/:id/payment` エンドポイントを新設。
フロントエンドの呼び出しを2段階に変更する必要がある。

## 対象ファイル

- `backend/internal/handler/accounting_handler.go`
- `backend/internal/handler/accounting_request.go`
- `backend/internal/service/accounting_service.go`
- `backend/internal/repository/accounting_repository.go`
- `backend/internal/model/accounting.go` (Payment struct)
- `backend/migrations/001_init.sql` (paid_by カラム追加)
- `backend/migrations/003_seed_demo.sql`
- `backend/docs/api.yaml`
- `frontend/src/features/accounting/api/transforms.ts`
- `frontend/src/features/accounting/types/index.ts`
