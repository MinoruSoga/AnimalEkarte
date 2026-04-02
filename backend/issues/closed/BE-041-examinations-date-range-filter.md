# BE-041: 健康診断 API に日付範囲フィルタ追加

**Status**: Open
**Priority**: Medium
**Affects**: examinations feature — 一覧 API
**Date Created**: 2026-03-17
**Related**: TASK-002, FE-015

## Summary

健康診断一覧 API（`GET /v1/examinations`）に `start_date` / `end_date` クエリパラメータを追加し、`exams.date` カラムで日付範囲フィルタを可能にする。

## 現状のコード

### Handler — ListExaminations

```go
// backend/internal/handler/examination_handler.go:13-56
func (h *Handler) ListExaminations(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    // ...
    // 既存フィルタ: pet_id, owner_id, status
    var petID *uint64
    if s := c.Query("pet_id"); s != "" { /* ... */ }
    var ownerID *uint64
    if s := c.Query("owner_id"); s != "" { /* ... */ }
    var status *string
    if s := c.Query("status"); s != "" { status = &s }

    // 行50: service 呼び出し（日付パラメータなし）
    exams, total, err := h.svc.Examination.List(
        c.Request.Context(), clinicID, petID, ownerID, status, page, limit,
    )
}
```

### Service — List

```go
// backend/internal/service/examination_service.go:26-28
func (s *examinationService) List(
    ctx context.Context, clinicID uint64,
    petID, ownerID *uint64, status *string,
    page, limit int,
) ([]model.Examination, int64, error) {
    return s.repo.FindAll(ctx, clinicID, petID, ownerID, status, page, limit)
}
```

### Repository — FindAll

```go
// backend/internal/repository/examination_repository.go:30-59
func (r *examinationRepository) FindAll(
    ctx context.Context, clinicID uint64,
    petID, ownerID *uint64, status *string,
    page, limit int,
) ([]model.Examination, int64, error) {
    buildBase := func() *gorm.DB {
        q := r.db.WithContext(ctx).Model(&model.Examination{}).
            Joins("JOIN medical_records ON medical_records.id = exams.medical_record_id").
            Where("medical_records.clinic_id = ?", clinicID)
        if petID != nil { q = q.Where("exams.pet_id = ?", *petID) }
        if ownerID != nil {
            q = q.Joins("JOIN pets ON pets.id = exams.pet_id").Where("pets.owner_id = ?", *ownerID)
        }
        if status != nil { q = q.Where("exams.status = ?", *status) }
        // ← 日付フィルタなし
        return q
    }
    // ...
    Order("exams.date DESC, exams.created_at DESC")
}
```

### DB — exams テーブル

```sql
-- backend/migrations/001_init.sql:722
date date NOT NULL  -- 実施日
```

## 必要な変更

### 1. Handler — start_date / end_date パラメータ追加

```go
// backend/internal/handler/examination_handler.go
// 既存の status フィルタの後に追加:

var startDate, endDate *string
if s := c.Query("start_date"); s != "" {
    startDate = &s
}
if s := c.Query("end_date"); s != "" {
    endDate = &s
}

exams, total, err := h.svc.Examination.List(
    c.Request.Context(), clinicID, petID, ownerID, status,
    startDate, endDate,  // ← 追加
    page, limit,
)
```

### 2. Service — シグネチャ変更

```go
// backend/internal/service/examination_service.go
// インターフェース:
type ExaminationService interface {
    List(ctx context.Context, clinicID uint64, petID, ownerID *uint64,
         status *string, startDate, endDate *string,  // ← 追加
         page, limit int) ([]model.Examination, int64, error)
    // ...
}

// 実装:
func (s *examinationService) List(
    ctx context.Context, clinicID uint64,
    petID, ownerID *uint64, status *string,
    startDate, endDate *string,  // ← 追加
    page, limit int,
) ([]model.Examination, int64, error) {
    return s.repo.FindAll(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}
```

### 3. Repository — WHERE 日付範囲追加

```go
// backend/internal/repository/examination_repository.go
// buildBase() 内に追加:

if startDate != nil {
    q = q.Where("exams.date >= ?", *startDate)
}
if endDate != nil {
    q = q.Where("exams.date <= ?", *endDate)
}
```

パラメータ形式: `YYYY-MM-DD`（Go の `time.Parse` 不要、PostgreSQL が date 文字列を直接比較可能）

## API 使用例

```
GET /v1/examinations?start_date=2026-01-01&end_date=2026-03-31
GET /v1/examinations?start_date=2026-03-01
GET /v1/examinations?end_date=2026-03-17
GET /v1/examinations?start_date=2026-03-01&end_date=2026-03-31&status=completed
```

## フロントエンド影響

- FE-015 で `get-examinations.ts` にパラメータ追加
- 既存の `useGetExaminations()` のシグネチャ変更

## 完了条件

- [ ] Handler に `start_date` / `end_date` クエリパラメータ追加
- [ ] Service インターフェース・実装のシグネチャ変更
- [ ] Repository の buildBase() に日付範囲 WHERE 追加
- [ ] `?start_date=2026-01-01&end_date=2026-03-31` で期間内のみ返る
- [ ] `start_date` のみ指定で以降の全件返る
- [ ] `end_date` のみ指定で以前の全件返る
- [ ] 既存テストが通る（`docker compose exec backend go test ./... -v`）
