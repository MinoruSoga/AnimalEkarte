# TASK-219: clinic_service.go — UpdateClinic に input nil チェックが欠落

## 優先度
Medium

## 対象ファイル
`backend/internal/service/clinic_service.go`

## 問題概要
`UpdateClinic` メソッドに `if input == nil { return WrapInvalidInput(...) }` チェックがない。
他の全 service の Update メソッドはこのガードを持っており、このファイルのみ欠落している。

nil の `*UpdateClinicInput` が渡された場合、`buildClinicUpdateFields(input)` 内でポインタデリファレンスが発生しパニックになる可能性がある（実装によって異なるが）。

## あるべき姿

```go
func (s *clinicService) UpdateClinic(ctx context.Context, id uint64, input *UpdateClinicInput) (*model.Clinic, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    // 以下既存ロジック
    ...
}
```

## 完了条件
- [ ] `UpdateClinic` 冒頭に `input == nil` チェックを追加
- [ ] `go test ./backend/internal/...` がパス
