# TASK-214: reservation_staff_service.go — Create 時の staff 作成と exclusion 設定が独立トランザクション

## 優先度
High

## 対象ファイル
`backend/internal/service/reservation_staff_service.go`

## 問題概要
`Create` メソッドで以下の2操作が**独立した別トランザクション**で実行されている。

1. `s.repo.Create(ctx, staff, clinicID)` — tx1: staff + staff_clinic_assignment を生成
2. `s.repo.ReplaceExcludedReservationTypes(ctx, staff.ID, ...)` — tx2: 除外予約種別を設定

tx1 が成功し tx2 が失敗した場合、スタッフだけが残り除外設定が欠落した不整合状態になる。
データ整合性の観点から2操作はひとつのトランザクション内で実行されるべき。

## 現状コード（行91-119付近）

```go
// tx1（独立）
staff, err := s.repo.Create(ctx, ...)
if err != nil {
    return nil, apperrors.Wrap(err, "failed to create reservation staff")
}

// tx2（独立 ← NG）
if err := s.repo.ReplaceExcludedReservationTypes(ctx, staff.ID, input.ExcludedReservationTypeIDs); err != nil {
    return nil, apperrors.Wrap(err, "failed to set excluded reservation types")
}
```

## あるべき姿

```go
// service 層に Transactor を注入し、両操作をひとつのトランザクション内で実行
var created *model.Staff
if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
    var err error
    created, err = s.repo.CreateInTx(txCtx, ...)
    if err != nil {
        return apperrors.Wrap(err, "failed to create reservation staff")
    }
    if err := s.repo.ReplaceExcludedReservationTypesInTx(txCtx, created.ID, input.ExcludedReservationTypeIDs); err != nil {
        return apperrors.Wrap(err, "failed to set excluded reservation types")
    }
    return nil
}); err != nil {
    return nil, err
}
return created, nil
```

## 完了条件
- [ ] service 層に `Transactor` を注入（既存の `medicine_service.go` や `staff_service.go` の実装を参考に）
- [ ] `Create` の2操作をひとつの `WithTx` でラップ
- [ ] tx1 成功・tx2 失敗のロールバックをテストで確認
- [ ] `go test ./backend/internal/...` がパス
