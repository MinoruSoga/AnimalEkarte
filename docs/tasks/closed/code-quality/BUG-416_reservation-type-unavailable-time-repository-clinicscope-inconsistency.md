# BUG-416: reservation_type_unavailable_time_repository が clinicScope ヘルパーを使用していない

## 概要

`reservation_type_unavailable_time_repository.go` の FindAll で、他の全マスタリポジトリが
`clinicScope(clinicID)` ヘルパーを使用しているのに対し、直接 `WHERE clinic_id = ?` を記述している。
実装パターンの不統一であり、将来 clinicScope に共通ロジック（例: deleted_at フィルタ）を
追加した場合にこのファイルだけ取り残されるリスクがある。

## 問題箇所

```go
// reservation_type_unavailable_time_repository.go:31-42
func (r *reservationTypeUnavailableTimeRepository) FindAll(
    ctx context.Context, clinicID, reservationTypeID uint64,
) ([]model.ReservationTypeUnavailableTime, error) {
    var results []model.ReservationTypeUnavailableTime
    err := r.db.WithContext(ctx).
        Where("clinic_id = ? AND reservation_type_id = ?", clinicID, reservationTypeID).  // ← 直接 WHERE
        Order("id ASC").
        Find(&results).Error
    // ...
}
```

## 他ファイルとの比較

```go
// checkup_type_repository.go:33-35（標準パターン）
err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).
    Order("sort_order ASC, name ASC").Find(&checkupTypes).Error

// reservation_type_occupation_repository.go:49-57（同様に clinicScope 未使用）
// ← こちらも同様の問題
```

## 修正方針

```go
// 修正後
err := r.db.WithContext(ctx).
    Scopes(clinicScope(clinicID)).                                 // clinicScope 使用
    Where("reservation_type_id = ?", reservationTypeID).           // 追加条件のみ
    Order("id ASC").
    Find(&results).Error
```

同様に `reservation_type_occupation_repository.go` の FindAll も clinicScope を使用するよう統一する。

## 影響ファイル

- `backend/internal/repository/reservation_type_unavailable_time_repository.go` — 行 31-42
- `backend/internal/repository/reservation_type_occupation_repository.go` — 行 49-57（同様の問題あり）

## 優先度

**Low** — 実装パターン不統一。動作への影響はないが、将来の clinicScope 変更時にリスクが生じる。
