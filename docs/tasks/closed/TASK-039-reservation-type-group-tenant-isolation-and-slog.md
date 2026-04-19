# TASK-039: reservation_type_group — CountCategories テナント分離欠落 & Reorder slog欠落

## 優先度

HIGH（CountCategories テナント分離）/ MEDIUM（Reorder slog・reservation_type Reorder slog）

---

## 問題 1: reservation_type_group_repository の CountCategories に clinicScope なし（テナント分離欠落）

### ファイル
`backend/internal/repository/reservation_type_group_repository.go:48-55`

### 問題
```go
func (r *reservationTypeGroupRepository) CountCategories(ctx context.Context, groupID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).Model(&model.ReservationType{}).
        Where("group_id = ?", groupID).Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "reservation_type", "")
    }
    return count, nil
}
```

`group_id` のみで絞り込み、`clinic_id` フィルタが存在しない。`reservation_type_groups.id` は連番 BIGSERIAL のため、クリニック A の groupID がクリニック B の `reservation_types.group_id` に一致する可能性がある。Delete 前の依存チェックが他テナントのカウントを含む可能性があり、誤って削除が阻止される（またはその逆）。

インターフェースシグネチャも `clinicID` を取らない設計になっており、呼び出し元 service でも `clinicID` を渡せない。

### 修正案
```go
// インターフェース
CountCategories(ctx context.Context, clinicID, groupID uint64) (int64, error)

// 実装
func (r *reservationTypeGroupRepository) CountCategories(ctx context.Context, clinicID, groupID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).Model(&model.ReservationType{}).
        Scopes(clinicScope(clinicID)).
        Where("group_id = ?", groupID).Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "reservation_type", "")
    }
    return count, nil
}

// service 側（reservation_type_group_service.go Delete）
count, err := s.repo.CountCategories(ctx, clinicID, id)
```

---

## 問題 2: reservation_type_group_service の Reorder に slog.InfoContext なし

### ファイル
`backend/internal/service/reservation_type_group_service.go`（Reorder メソッド）

### 問題
TASK-023 で `Update` の slog 欠落・ベア文字列キーを指摘したが、`Reorder` の slog 欠落は別途未起票。Create/Delete には slog があるが Reorder のみ欠落している。

### 修正案
```go
slog.InfoContext(ctx, "reservation_type_groups reordered",
    slog.Uint64("clinic_id", clinicID),
    slog.Int("count", len(ids)))
return nil
```

---

## 問題 3: reservation_type_service の Reorder に slog.InfoContext なし

### ファイル
`backend/internal/service/reservation_type_service.go:275-282`

### 問題
```go
func (s *reservationTypeService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    if len(ids) == 0 {
        return apperrors.WrapInvalidInput("ids must not be empty")
    }
    if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
        return apperrors.Wrap(err, "failed to reorder service types")
    }
    return nil  // slog なし
}
```

TASK-029 では exam_type/checkup_type/diagnosis の Reorder slog 欠落を対象としたが reservation_type は含まれていなかった。

### 修正案
```go
slog.InfoContext(ctx, "reservation_types reordered",
    slog.Uint64("clinic_id", clinicID),
    slog.Int("count", len(ids)))
return nil
```
