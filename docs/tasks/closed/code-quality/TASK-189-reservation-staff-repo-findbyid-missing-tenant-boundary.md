# TASK-189: reservation_staff_repository.go — FindByID にテナント境界なし

## 優先度
High

## 対象ファイル
`backend/internal/repository/reservation_staff_repository.go`

## 問題概要
`FindByID` メソッドが `clinic_id` を条件に含まず、全テナントのスタッフを ID 検索できる状態になっている。
マルチテナント設計規約「全クエリで clinic_id を条件に含めること」に違反する。

`FindByIDAndClinicID` も Interface に存在しているが、`FindByID` が公開 Interface に残っている以上、
Service 層での誤用リスクが排除できない。

## 現状コード（行48-50）

```go
func (r *reservationStaffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
    var staff model.Staff
    err := r.db.WithContext(ctx).First(&staff, "id = ?", id).Error
    // clinic_id チェックなし ← NG
```

## 修正方針（優先順位順）

### 推奨: Interface から `FindByID` を除去
```go
// Interface（NG: clinic_id なし版を公開）
type ReservationStaffRepository interface {
    FindByID(ctx context.Context, id uint64) (*model.Staff, error)
    FindByIDAndClinicID(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
    ...
}

// 修正後（FindByID を除去または unexported に）
type ReservationStaffRepository interface {
    FindByIDAndClinicID(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
    ...
}
```

### 代替: internal 利用のみ許可するコメントで明示
`FindByID` が内部専用（トランザクション等）で必要な場合は
その旨をコメントで明記し、Service 層からの直接呼び出しを禁止する。

## 完了条件
- [ ] `FindByID`（テナント境界なし）の Interface からの除去、または内部専用明示
- [ ] Service 層が `FindByIDAndClinicID` のみ使用していることを確認
- [ ] `go test ./backend/internal/...` がパス
