# TASK-181: UpsertPayment の ErrRecordNotFound 未チェックによる誤 INSERT

## 優先度: High

## 概要
`accountingRepository.UpsertPayment` で、`First` のエラーを `err != nil` で一括判定し、
`gorm.ErrRecordNotFound` 以外のエラー（DB接続エラー・タイムアウト等）でも CREATE を試みる。
これはデータ破壊・無限リトライの温床であり、High 優先度の違反。

## 対象ファイル
`backend/internal/repository/accounting_repository.go`

## 現状コード

```go
// L212〜222
var existing model.Payment
err := r.db.WithContext(ctx).
    Where("billing_id = ?", payment.BillingID).
    First(&existing).Error

if err != nil {
    // レコードなし → 新規作成
    // ❌ gorm.ErrRecordNotFound 以外（DB エラー等）も CREATE してしまう
    if err := r.db.WithContext(ctx).Create(payment).Error; err != nil {
        return apperrors.FromGORM(err, "payment", fmt.Sprintf("billing_id=%d", payment.BillingID))
    }
    return nil
}
```

## 問題の詳細
- `gorm.ErrRecordNotFound` の場合のみ INSERT すべきだが、DB 接続エラー等でも INSERT を試みる
- 規約: Repository のエラーは `apperrors.FromGORM` で変換してから上位に返す必要がある
- 誤った条件分岐により、`errors.Is(err, gorm.ErrRecordNotFound)` の直接比較も発生している
  （これは規約では `apperrors.FromGORM` で変換してから上位で判定すべき）

## 修正後コード

```go
import "errors"
import "gorm.io/gorm"

var existing model.Payment
err := r.db.WithContext(ctx).
    Where("billing_id = ?", payment.BillingID).
    First(&existing).Error

if err != nil {
    if !errors.Is(err, gorm.ErrRecordNotFound) {
        // DB エラー → そのまま変換して返す
        return apperrors.FromGORM(err, "payment", fmt.Sprintf("billing_id=%d", payment.BillingID))
    }
    // レコードなし → 新規作成
    if err := r.db.WithContext(ctx).Create(payment).Error; err != nil {
        return apperrors.FromGORM(err, "payment", fmt.Sprintf("billing_id=%d", payment.BillingID))
    }
    return nil
}

// 既存レコード → map で更新（ゼロ値も反映）
if err := r.db.WithContext(ctx).
    Model(&model.Payment{}).
    Where("billing_id = ?", payment.BillingID).
    Updates(fields).Error; err != nil {
    return apperrors.FromGORM(err, "payment", fmt.Sprintf("billing_id=%d", payment.BillingID))
}
payment.ID = existing.ID
return nil
```

## 影響範囲
- 会計更新時 (`UpdateAccounting`) の Payment upsert 処理全体
- DB 障害時に不正な Payment レコードが INSERT される可能性がある
