# BE-042: 予防接種 API に日付範囲フィルタ追加

**Status**: Open
**Priority**: Medium
**Affects**: vaccinations feature — 一覧 API
**Date Created**: 2026-03-17
**Related**: TASK-002, FE-016

## Summary

予防接種一覧 API（`GET /v1/vaccinations`）に `start_date` / `end_date` クエリパラメータを追加し、`vaccinations.date` カラムで日付範囲フィルタを可能にする。BE-041（健康診断）と同一パターン。

## 現状のコード

### Handler — ListVaccinations

```go
// backend/internal/handler/vaccination_handler.go:13-51
func (h *Handler) ListVaccinations(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    // ...
    // 既存フィルタ: pet_id, owner_id（status なし）
    var petID *uint64
    if s := c.Query("pet_id"); s != "" { /* ... */ }
    var ownerID *uint64
    if s := c.Query("owner_id"); s != "" { /* ... */ }

    // 行50: service 呼び出し（日付パラメータなし）
    vaccinations, total, err := h.svc.Vaccination.List(
        c.Request.Context(), clinicID, petID, ownerID, page, limit,
    )
}
```

### Service — List

```go
// backend/internal/service/vaccination_service.go:26-28
func (s *vaccinationService) List(
    ctx context.Context, clinicID uint64,
    petID, ownerID *uint64,
    page, limit int,
) ([]model.Vaccination, int64, error) {
    return s.repo.FindAll(ctx, clinicID, petID, ownerID, page, limit)
}
```

### Repository — FindAll

```go
// backend/internal/repository/vaccination_repository.go:30-58
func (r *vaccinationRepository) FindAll(
    ctx context.Context, clinicID uint64,
    petID, ownerID *uint64,
    page, limit int,
) ([]model.Vaccination, int64, error) {
    buildBase := func() *gorm.DB {
        q := r.db.WithContext(ctx).Model(&model.Vaccination{}).
            Where("vaccinations.clinic_id = ?", clinicID)
        if petID != nil { q = q.Where("vaccinations.pet_id = ?", *petID) }
        if ownerID != nil {
            q = q.Joins("JOIN pets ON pets.id = vaccinations.pet_id").
                Where("pets.owner_id = ?", *ownerID)
        }
        // ← 日付フィルタなし
        return q
    }
    // ...
    Order("vaccinations.date DESC, vaccinations.created_at DESC")
}
```

### DB — vaccinations テーブル

```sql
-- backend/migrations/001_init.sql:683-684
date      date NOT NULL,   -- 実施日
next_date date,            -- 次回予定日
```

## 必要な変更

### 1. Handler — start_date / end_date パラメータ追加

```go
// backend/internal/handler/vaccination_handler.go
// 既存の owner_id フィルタの後に追加:

var startDate, endDate *string
if s := c.Query("start_date"); s != "" {
    startDate = &s
}
if s := c.Query("end_date"); s != "" {
    endDate = &s
}

vaccinations, total, err := h.svc.Vaccination.List(
    c.Request.Context(), clinicID, petID, ownerID,
    startDate, endDate,  // ← 追加
    page, limit,
)
```

### 2. Service — シグネチャ変更

```go
// backend/internal/service/vaccination_service.go
type VaccinationService interface {
    List(ctx context.Context, clinicID uint64, petID, ownerID *uint64,
         startDate, endDate *string,  // ← 追加
         page, limit int) ([]model.Vaccination, int64, error)
    // ...
}

func (s *vaccinationService) List(
    ctx context.Context, clinicID uint64,
    petID, ownerID *uint64,
    startDate, endDate *string,  // ← 追加
    page, limit int,
) ([]model.Vaccination, int64, error) {
    return s.repo.FindAll(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}
```

### 3. Repository — WHERE 日付範囲追加

```go
// backend/internal/repository/vaccination_repository.go
// buildBase() 内に追加:

if startDate != nil {
    q = q.Where("vaccinations.date >= ?", *startDate)
}
if endDate != nil {
    q = q.Where("vaccinations.date <= ?", *endDate)
}
```

## API 使用例

```
GET /v1/vaccinations?start_date=2026-01-01&end_date=2026-03-31
GET /v1/vaccinations?start_date=2026-03-01
GET /v1/vaccinations?end_date=2026-03-17
```

## フロントエンド影響

- FE-016 で `get-vaccinations.ts` にパラメータ追加

## 完了条件

- [ ] Handler に `start_date` / `end_date` クエリパラメータ追加
- [ ] Service インターフェース・実装のシグネチャ変更
- [ ] Repository の buildBase() に日付範囲 WHERE 追加
- [ ] `?start_date=2026-01-01&end_date=2026-03-31` で期間内のみ返る
- [ ] 既存テストが通る（`docker compose exec backend go test ./... -v`）
