# BE-063: トリミング API 日付範囲フィルタ追加

**Status**: Open
**Priority**: Medium
**Affects**: `GET /v1/trimmings`
**Date Created**: 2026-03-26
**Related**: TASK-031, FE-126

## Summary

`ListTrimmings` ハンドラに `start_date` / `end_date` クエリパラメータを追加し、handler → service → repository の全3層で日付範囲フィルタを実装する。他のエンドポイント（vaccination、examination）と同パターン。

## 現状のコード

```go
// backend/internal/handler/trimming_handler.go:46
trimmings, total, err := h.svc.Trimming.List(c.Request.Context(), clinicID, petID, ownerID, page, limit)
// ↑ startDate/endDate なし

// backend/internal/service/trimming_service.go:52
List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.TrimmingRecord, int64, error)

// backend/internal/repository/trimming_repository.go:15
FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.TrimmingRecord, int64, error)
```

DB カラム: `trimmings.date` (`gorm:"type:date;not null"`)

## 必要な変更

### 1. Handler 変更

```go
// backend/internal/handler/trimming_handler.go
func (h *Handler) ListTrimmings(c *gin.Context) {
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

    var petID *uint64
    // ... (既存の pet_id / owner_id パース)

    trimmings, total, err := h.svc.Trimming.List(c.Request.Context(), clinicID, petID, ownerID, startDate, endDate, page, limit)
    // ...
}
```

### 2. Service 変更

```go
// backend/internal/service/trimming_service.go
// Before
List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.TrimmingRecord, int64, error)

// After
List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.TrimmingRecord, int64, error)

// 実装
func (s *trimmingService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.TrimmingRecord, int64, error) {
    return s.repo.FindAll(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}
```

### 3. Repository 変更

```go
// backend/internal/repository/trimming_repository.go
// Interface
FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.TrimmingRecord, int64, error)

// 実装（既存の WHERE 句に追記）
if startDate != nil {
    q = q.Where("trimmings.date >= ?", *startDate)
}
if endDate != nil {
    q = q.Where("trimmings.date <= ?", *endDate)
}
```

## 完了条件

- [ ] `GET /v1/trimmings?start_date=2026-01-01&end_date=2026-03-31` が期間内のレコードのみ返す
- [ ] `start_date` / `end_date` なしの場合は従来通り全件返す
- [ ] 既存テストがパスする
- [ ] `docker compose exec backend go test ./... -v` パス
