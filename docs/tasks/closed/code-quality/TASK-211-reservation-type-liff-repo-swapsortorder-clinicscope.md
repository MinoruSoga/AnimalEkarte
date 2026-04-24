# TASK-211: reservation_type_liff_repository.go — SwapSortOrder の Update に clinicScope が欠落

## 優先度
Low

## 対象ファイル
`backend/internal/repository/reservation_type_liff_repository.go`

## 問題概要
`SwapSortOrder`（行91-127付近）でターゲットと隣接レコードを取得する際は `Scopes(clinicScope(clinicID))` を適用しているが、
実際の `Update` 操作には `clinicScope` が適用されていない。

```go
// 取得（clinicScope あり）
target := tx.Scopes(clinicScope(clinicID)).Where("id = ?", targetID).First(&t)

// 更新（clinicScope なし ← NG）
tx.Model(&model.ReservationTypeLiff{}).Where("id = ?", t.ID).Update("sort_order", neighbor.SortOrder)
```

`id` が主キーであるため他テナントのレコードを誤って更新するリスクは低いが、
マルチテナント設計規約「全クエリで clinic_id を条件に含める」に違反している。
取得と更新で条件の一貫性がないため、将来の変更でリグレッションが起きやすい。

## あるべき姿

```go
// 更新にも clinicScope または clinic_id 条件を追加
tx.Model(&model.ReservationTypeLiff{}).
    Scopes(clinicScope(clinicID)).
    Where("id = ?", t.ID).
    Update("sort_order", neighbor.SortOrder)
```

## 完了条件
- [ ] `SwapSortOrder` の2箇所の `Update` に `clinicScope` または `WHERE clinic_id = ?` を追加
- [ ] `go test ./backend/internal/...` がパス
