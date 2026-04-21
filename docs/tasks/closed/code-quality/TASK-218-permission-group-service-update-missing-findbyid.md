# TASK-218: permission_group_service.go — Update に FindByID 存在確認が欠落

## 優先度
Medium

## 対象ファイル
`backend/internal/service/permission_group_service.go`

## 問題概要
`Update` メソッドが `FindByID` による存在確認を行わず、直接 `buildPermissionGroupUpdateFields` → `repo.UpdateFields` を呼んでいる。

存在しない ID を指定した場合、`repo.UpdateFields` の `RowsAffected == 0` チェックで NotFound が返ることがあるが、Service 層で先行確認するのが他の全ドメインとの一貫したパターンである。

## 比較（正しい実装パターン）
- `animal_species_service.go` — Update 冒頭で `FindByID` 呼び出し
- `cage_service.go` — Update 冒頭で `FindByID` 呼び出し
- `insurance_service.go` — Update 冒頭で `FindByID` 呼び出し
- `medicine_service.go` — Update 冒頭で `FindByID` 呼び出し

## あるべき姿

```go
func (s *permissionGroupService) Update(ctx context.Context, clinicID, id uint64, input *UpdatePermissionGroupInput) (*model.PermissionGroup, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    // 追加: 存在確認
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get permission group")
    }
    fields := buildPermissionGroupUpdateFields(input)
    if len(fields) == 0 {
        return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
    }
    ...
}
```

## 完了条件
- [ ] `Update` 冒頭（`input == nil` チェックの直後）に `FindByID` 存在確認を追加
- [ ] 存在しない ID で 404 が返ることをテストで確認
- [ ] `go test ./backend/internal/...` がパス
