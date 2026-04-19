# TASK-004: ReservationTypeGroup の Update を buildUpdateFields パターンに統一

## 概要

`reservation_type_group_service.go` の `Update` メソッドがインライン `map[string]any{}` を直接組み立てており、他のマスタサービス（`reservation_type_service.go`, `exam_type_service.go` 等）が採用している `buildXxxUpdateFields()` パターンと不統一になっている。

## 優先度

HIGH

## 影響ファイル

| ファイル | 問題箇所 |
|---------|---------|
| `backend/internal/service/reservation_type_group_service.go` | L85-97（Update のインライン map 組み立て） |

## 規約違反

`.claude/rules/go-language.md`:
> PATCH は ポインタ型 + `buildXxxUpdateFields()` パターンを使い統一する。

## 現状コード

```go
// service/reservation_type_group_service.go（現状）
func (s *reservationTypeGroupService) Update(...) (...) {
    fields := map[string]any{}
    if input.Name != nil { fields["name"] = *input.Name }
    if input.Color != nil { fields["color"] = *input.Color }
    // ...
    return s.repo.UpdateFields(ctx, clinicID, id, fields)
}
```

## 修正方針

専用の `buildReservationTypeGroupUpdateFields` 関数を切り出す。

```go
// service/reservation_type_group_service.go（修正後）

func buildReservationTypeGroupUpdateFields(input *UpdateReservationTypeGroupInput) map[string]any {
    fields := make(map[string]any)
    if input.Name != nil      { fields["name"] = *input.Name }
    if input.Color != nil     { fields["color"] = *input.Color }
    if input.SortOrder != nil { fields["sort_order"] = *input.SortOrder }
    if input.IsActive != nil  { fields["is_active"] = *input.IsActive }
    return fields
}

func (s *reservationTypeGroupService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationTypeGroupInput) (*model.ReservationTypeGroup, error) {
    slog.InfoContext(ctx, "updating reservation type group", slog.Uint64("id", id))
    fields := buildReservationTypeGroupUpdateFields(input)
    result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to update reservation type group")
    }
    return result, nil
}
```

## テスト

- `buildReservationTypeGroupUpdateFields` のユニットテスト（nil フィールドがスキップされること）
- Update の統合テスト（部分更新が正しく反映されること）
