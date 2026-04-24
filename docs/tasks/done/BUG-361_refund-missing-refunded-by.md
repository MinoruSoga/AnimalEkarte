# BUG-361: 返金レコードに処理者（refunded_by）が記録されない

## 優先度: HIGH

## 概要

`billing_refunds` テーブルに返金処理者を記録する `refunded_by` カラムが存在しない。
会計システムで金銭操作の実行者が追跡不能なのは監査上の欠陥。

## 現状

- DB: `billing_refunds` に `refunded_by` カラムなし
- Model: `BillingRefund` struct に `RefundedBy` フィールドなし
- Handler: `CreateRefund` で `extractStaffID(c)` を呼んでいない
- Service: `Create` メソッドに `staffID` パラメータなし

## 修正内容

### 1. DB スキーマ（001_init.sql）
```sql
refunded_by bigint REFERENCES staffs(id)  -- nullable（システム返金もあり得る）
```

### 2. Model（billing_refund.go）
```go
RefundedBy *uint64 `gorm:"" json:"refunded_by"`
```

### 3. Service（refund_service.go）
- `Create` シグネチャに `staffID *uint64` 追加

### 4. Handler（refund_handler.go）
- `extractStaffID(c)` で認証スタッフID取得 → Service に渡す

### 5. Response DTO（accounting_response.go）
- `refundResponse` に `RefundedBy` フィールド追加

## 副次対応

### テストファイルの誤った設計コメント削除
`refund_handler_test.go` に Update/Delete エンドポイントのテストスペックがコメントで記載されているが、
返金レコードは会計原則上 **不変（immutable）** であるべき。これらのコメントは誤った設計方針であり削除する。

## 対象ファイル

- `backend/migrations/001_init.sql`
- `backend/internal/model/billing_refund.go`
- `backend/internal/service/refund_service.go`
- `backend/internal/handler/refund_handler.go`
- `backend/internal/handler/accounting_request.go`
- `backend/internal/handler/accounting_response.go`
- `backend/internal/handler/refund_handler_test.go`
