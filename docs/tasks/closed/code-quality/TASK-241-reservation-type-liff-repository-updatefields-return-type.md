# TASK-241: reservation_type_liff_repository.go — UpdateFields の戻り値型が他リポジトリと不統一

## 優先度
Medium

## 対象ファイル
- `backend/internal/repository/reservation_type_liff_repository.go`

## 問題概要
`ReservationTypeLiffRepository` インターフェースの `UpdateFields` は `error` のみを返すが、
`ReservationTypeRepository` および他マスタリポジトリの `UpdateFields` は
`(*model.{Entity}, error)` を返す設計に統一されている。

呼び出し元 service が更新後エンティティを取得するために別途 `FindByID` を呼ぶ必要が生じ、
クエリ数が増加する。

## 現状コード

```go
// ReservationTypeLiffRepository インターフェース（行19）
UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) error

// 実装
func (r *reservationTypeLiffRepository) UpdateFields(...) error {
    return r.db.WithContext(ctx).Model(&model.ReservationType{}).
        Scopes(clinicScope(clinicID)).
        Where("id = ?", id).
        Updates(fields).Error
}
```

## 比較（ReservationTypeRepository）

```go
// ReservationTypeRepository インターフェース（行21）
UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error)

// 実装: Updates → First でエンティティを返す
```

## あるべき姿

```go
// インターフェース
UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error)

// 実装
func (r *reservationTypeLiffRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error) {
    var rt model.ReservationType
    err := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Model(&rt).
        Where("id = ?", id).
        Updates(fields).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "reservation_type_liff", fmt.Sprintf("%d", id))
    }
    return r.FindByID(ctx, clinicID, id)
}
```

呼び出し元 service（`reservation_type_liff_service.go`）も戻り値を受け取るよう修正する。

## 完了条件
- [ ] `ReservationTypeLiffRepository.UpdateFields` の戻り値を `(*model.ReservationType, error)` に変更
- [ ] 実装を Updates + FindByID パターンに変更
- [ ] 呼び出し元 service を修正
- [ ] `go test ./backend/internal/...` がパス
