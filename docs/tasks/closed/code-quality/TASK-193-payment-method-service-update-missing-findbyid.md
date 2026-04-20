# TASK-193: payment_method_master_service.go — Update で FindByID による存在確認が欠落

## 優先度
Medium

## 対象ファイル
`backend/internal/service/payment_method_master_service.go`

## 問題概要
`Update` メソッドが `FindByID` による存在確認を行わずに `UpdateFields` を呼んでいる。
存在しない ID を指定した場合に、404 ではなく予期しないエラーが返る可能性がある。

他のすべての service（cage, exam_type, insurance, vaccine, reservation_type 等）は
Update 冒頭で `FindByID` を呼んでいるため、このファイルのみ挙動が不一致。

## 現状コード（行99-118）

```go
func (s *paymentMethodMasterService) Update(ctx context.Context, ...) (*model.PaymentMethodMaster, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    // ← FindByID による存在確認なし
    fields := buildPaymentMethodUpdateFields(input)
    ...
}
```

## あるべき姿

```go
func (s *paymentMethodMasterService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePaymentMethodInput) (*model.PaymentMethodMaster, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get payment method")
    }
    fields := buildPaymentMethodUpdateFields(input)
    ...
}
```

## 完了条件
- [ ] `Update` 冒頭に `FindByID` 存在確認を追加
- [ ] 存在しない ID で 404 が返ることを確認（テストまたは手動確認）
- [ ] `go test ./backend/internal/...` がパス
