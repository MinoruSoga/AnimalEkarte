# BE-064: 入院・ホテル管理 API 日付範囲フィルタ追加

**Status**: Open
**Priority**: Medium
**Affects**: `GET /v1/hospitalizations`
**Date Created**: 2026-03-26
**Related**: TASK-031, FE-127

## Summary

`ListHospitalizations` ハンドラに `start_date` / `end_date` クエリパラメータを追加し、handler → service → repository の全3層で入院開始日の日付範囲フィルタを実装する。

## 現状のコード

```go
// backend/internal/handler/hospitalization_handler.go:49
hospitalizations, total, err := h.svc.Hospitalization.List(
    c.Request.Context(), clinicID, petID, ownerID, status, page, limit)
// ↑ startDate/endDate なし

// backend/internal/service/hospitalization_service.go:11
List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, page, limit int) ([]model.Hospitalization, int64, error)

// backend/internal/repository/hospitalization_repository.go:15
FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, page, limit int) ([]model.Hospitalization, int64, error)
```

DB カラム: `hospitalizations.start_date` (`gorm:"type:date;not null"`)

## 必要な変更

### 1. Handler 変更

```go
// backend/internal/handler/hospitalization_handler.go
func (h *Handler) ListHospitalizations(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok { return }
    page, limit, err := parsePagination(c)
    if err != nil { /* ... */ }

    // 追加: 日付フィルタ
    startDate, err := parseDateQuery(c, "start_date")
    if err != nil {
        RespondError(c, fmt.Errorf("invalid start_date: %w", apperrors.ErrInvalidInput))
        return
    }
    endDate, err := parseDateQuery(c, "end_date")
    if err != nil {
        RespondError(c, fmt.Errorf("invalid end_date: %w", apperrors.ErrInvalidInput))
        return
    }

    // ... 既存の petID / ownerID / status パース ...

    hospitalizations, total, err := h.svc.Hospitalization.List(
        c.Request.Context(), clinicID, petID, ownerID, status, startDate, endDate, page, limit)
    // ...
}
```

### 2. Service 変更

```go
// backend/internal/service/hospitalization_service.go
// Before
List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status *string, page, limit int) ([]model.Hospitalization, int64, error)

// After
List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)

func (s *hospitalizationService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error) {
    return s.repo.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}
```

### 3. Repository 変更

```go
// backend/internal/repository/hospitalization_repository.go
// Interface
FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)

// 実装（既存 WHERE 句に追記）
if startDate != nil {
    q = q.Where("hospitalizations.start_date >= ?", *startDate)
}
if endDate != nil {
    q = q.Where("hospitalizations.start_date <= ?", *endDate)
}
```

> `end_date` カラムではなく `start_date` カラムで絞り込む。「入院開始日が指定期間内」が自然な検索意図のため。

## 完了条件

- [ ] `GET /v1/hospitalizations?start_date=2026-01-01&end_date=2026-03-31` が期間内の入院開始日のレコードのみ返す
- [ ] `start_date` / `end_date` なしの場合は従来通り全件返す
- [ ] 既存の `status` フィルタと共存して動作する
- [ ] `docker compose exec backend go test ./... -v` パス
