# BE-056: カルテ一覧API に日付範囲フィルタ追加

**Status**: Open
**Priority**: Medium
**Affects**: カルテ管理 (`/v1/medical-records`)
**Date Created**: 2026-03-25
**Related**: TASK-028, FE-119

## Summary

カルテ一覧API(`GET /v1/medical-records`)に`start_date`/`end_date`クエリパラメータを追加し、診療日（`date`カラム）での絞り込みを可能にする。

handler → service → repository の3層すべてを修正する。参照実装は`vaccination`の同パターン。

## 現状のコード

```go
// backend/internal/handler/medical_record_handler.go:37-66
func (h *Handler) ListMedicalRecords(c *gin.Context) {
    // ... petID, ownerID の抽出のみ
    records, total, err := h.svc.MedicalRecord.List(c.Request.Context(), clinicID, petID, ownerID, page, limit)
    // ← start_date / end_date の処理なし
}
```

```go
// backend/internal/service/medical_record_service.go:38
func (s *medicalRecordService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.MedicalRecord, int64, error) {
    return s.repo.FindAll(ctx, clinicID, petID, ownerID, page, limit)
    // ← startDate, endDate 引数なし
}
```

```go
// backend/internal/repository/medical_record_repository.go:31-50
func (r *medicalRecordRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.MedicalRecord, int64, error) {
    q := r.db.WithContext(ctx).Model(&model.MedicalRecord{}).Where("clinic_id = ?", clinicID)
    // ← date フィルタのwhere句なし
}
```

## 必要な変更

### 1. Handler 変更

```go
// backend/internal/handler/medical_record_handler.go
// line 64（ownerID ブロック終了直後、List呼び出しの前）に追加:

startDate, err := parseDateQuery(c, "start_date")
if err != nil {
    RespondError(c, err)
    return
}
endDate, err := parseDateQuery(c, "end_date")
if err != nil {
    RespondError(c, err)
    return
}

records, total, err := h.svc.MedicalRecord.List(c.Request.Context(), clinicID, petID, ownerID, startDate, endDate, page, limit)
```

### 2. Service Interface + 実装変更

```go
// backend/internal/service/medical_record_service.go

// Interface（同ファイルまたは medical_record_service_interface.go）:
type MedicalRecordService interface {
    List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error)
    // ... 他のメソッドは変更なし
}

// 実装:
func (s *medicalRecordService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
    return s.repo.FindAll(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}
```

### 3. Repository Interface + 実装変更

```go
// backend/internal/repository/medical_record_repository.go

// Interface:
type MedicalRecordRepository interface {
    FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error)
    // ... 他のメソッドは変更なし
}

// 実装:
func (r *medicalRecordRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
    records := make([]model.MedicalRecord, 0)
    var total int64

    q := r.db.WithContext(ctx).Model(&model.MedicalRecord{}).Where("clinic_id = ?", clinicID)
    if petID != nil {
        q = q.Where("pet_id = ?", *petID)
    }
    if ownerID != nil {
        q = q.Where("owner_id = ?", *ownerID)
    }
    // ↓ 追加
    if startDate != nil {
        q = q.Where("date >= ?", *startDate)
    }
    if endDate != nil {
        q = q.Where("date <= ?", *endDate)
    }
    // ↑ 追加
    if err := q.Count(&total).Error; err != nil {
        return nil, 0, apperrors.Wrap(err, "count medical records")
    }
    if err := q.Offset((page - 1) * limit).Limit(limit).Order("date DESC, created_at DESC").
        Preload("Owner").Preload("Pet.AnimalSpecies").Preload("Doctor").Preload("Inquiry").Preload("Billing").
        Find(&records).Error; err != nil {
        return nil, 0, apperrors.Wrap(err, "find medical records")
    }
    return records, total, nil
}
```

## APIレスポンス形式（変更なし）

既存のレスポンス形式は変わらない。フィルタは結果の件数にのみ影響する。

```
GET /v1/medical-records?start_date=2026-01-01&end_date=2026-03-31
→ 2026年1月1日〜3月31日の診療カルテのみ返す
```

## テストファイル更新（必須）

インターフェースのシグネチャが変わるため、モック実装をすべて更新する必要がある。

### `backend/internal/service/medical_record_service_test.go`

```go
// line 17: findAllFn 型定義に startDate, endDate を追加
findAllFn func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error)

// line 25-26: FindAll メソッドシグネチャを更新
func (m *mockMedicalRecordRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
    return m.findAllFn(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

// line 186 など: テストテーブル内 findAllFn ラムダに _, _ を追加
findAllFn: func(_ context.Context, _ uint64, petID *uint64, ownerID *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {

// line 194: svc.List() 呼び出しに nil, nil を追加
records, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, nil, nil, tt.page, tt.limit)
```

### `backend/internal/handler/medical_record_handler_test.go`

```go
// line 25: listFn 型定義に startDate, endDate を追加
listFn func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error)

// line 33-34: mockMedicalRecordService.List シグネチャを更新
func (m *mockMedicalRecordService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
    return m.listFn(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

// line 118, 131, 158: テストテーブル内 listFn ラムダに _, _ *string を追加
listFn: func(_ context.Context, clinicID uint64, _, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
```

## フロントエンド影響

- FE-119 で `getMedicalRecords` に日付パラメータを追加する
- `make codegen` は不要（モデル変更なし）

## 参照実装

```go
// backend/internal/handler/vaccination_handler.go:45-58
startDate, err := parseDateQuery(c, "start_date")
if err != nil {
    RespondError(c, err)
    return
}
endDate, err := parseDateQuery(c, "end_date")
// ...
vaccs, total, err := h.svc.Vaccination.List(..., startDate, endDate, page, limit)
```

```go
// backend/internal/repository/vaccination_repository.go:42-48
if startDate != nil {
    q = q.Where("vaccinations.date >= ?", *startDate)
}
if endDate != nil {
    q = q.Where("vaccinations.date <= ?", *endDate)
}
```

## 完了条件

- [ ] `GET /v1/medical-records?start_date=2026-01-01&end_date=2026-01-31` が指定期間のカルテのみ返す
- [ ] パラメータなしの場合は全件返す（既存動作を破壊しない）
- [ ] `medical_record_service_test.go` のモックシグネチャ更新済み
- [ ] `medical_record_handler_test.go` のモックシグネチャ更新済み
- [ ] `go test ./... -v` が通る
- [ ] `golangci-lint run ./...` が通る
