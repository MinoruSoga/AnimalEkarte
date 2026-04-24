# TASK-205: payment_method_master_service.go — Delete で FindByID 存在確認が欠落

## 優先度
Medium

## 対象ファイル
`backend/internal/service/payment_method_master_service.go`

## 問題概要
`Delete` メソッドが `FindByID` による存在確認を行わず、直接 `CountUsageByPaymentMethodID` → `Delete` を呼んでいる。

存在しない ID に対して Delete した場合、`CountUsageByPaymentMethodID` が 0 件を返し
（存在しないから使用中でもない）、そのまま Delete に進む。
Repository の `Delete` が `RowsAffected == 0` で `WrapNotFound` を返せば結果的に 404 になるが、
Service 層で先行確認するのが他ドメインとの一貫したパターンである。

## 比較（他ドメインの実装）
- `insurance_service.go:150` — Delete 前に `FindByID` で存在確認
- `cage_service.go` — Delete 前に `FindByID` で存在確認
- `vaccine_service.go` — Delete 前に `FindByID` で存在確認

## あるべき姿

```go
func (s *paymentMethodMasterService) Delete(ctx context.Context, clinicID, id uint64) error {
    // 存在確認
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return apperrors.Wrap(err, "failed to get payment method")
    }
    // FK 依存チェック
    count, err := s.repo.CountUsageByPaymentMethodID(ctx, clinicID, id)
    if err != nil {
        return apperrors.Wrap(err, "failed to count usage")
    }
    if count > 0 {
        return apperrors.WrapConflict("この支払方法は使用中のため削除できません")
    }
    ...
}
```

## 完了条件
- [ ] `Delete` 冒頭に `FindByID` 存在確認を追加
- [ ] 存在しない ID で 404 が返ることをテストで確認
- [ ] `go test ./backend/internal/...` がパス
