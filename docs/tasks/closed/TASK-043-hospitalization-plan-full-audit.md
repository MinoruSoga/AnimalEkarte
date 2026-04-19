# TASK-043: hospitalization_plan ドメイン全監査 — 4件

## 優先度

HIGH

---

## 問題 1: hospitalization_plan_handler が *model.HospitalizationPlan を直接構築

### ファイル
`backend/internal/handler/hospitalization_plan_handler.go:49-93`

### 問題
`CreateHospitalizationPlan` ハンドラが `&model.HospitalizationPlan{...}` を直接組み立てて service に渡している。TASK-025/031/033 で同パターンを複数ドメインで指摘済みだが hospitalization_plan は未対応。

```go
plan := &model.HospitalizationPlan{
    ClinicID:    clinicID,
    Name:        input.Name,
    // ... 多数のフィールド
}
result, err := h.svc.HospitalizationPlan.Create(c.Request.Context(), plan)
```

### 修正案
`service.CreateHospitalizationPlanInput` DTO を追加し、handler は DTO のみ構築する。

---

## 問題 2: hospitalization_plan_handler の parseIDParam / extractClinicID 順序が逆

### ファイル
`backend/internal/handler/hospitalization_plan_handler.go:96-104`

### 問題
```go
id, ok := parseIDParam(c, "id")     // ← 先に parseIDParam
if !ok { return }
clinicID, ok := extractClinicID(c)  // ← 後で extractClinicID
```

TASK-027 で `trimming_master_handler.go` の同問題を指摘済みだが、hospitalization_plan_handler の Update/Delete でも同パターンが存在する。`extractClinicID` を先に呼び、JWT クレーム不正の場合は ID パースより前に返すべき。

### 修正案
```go
clinicID, ok := extractClinicID(c)  // 先に clinic 認証
if !ok { return }
id, ok := parseIDParam(c, "id")
if !ok { return }
```

---

## 問題 3: hospitalization_plan_service の Reorder に slog.InfoContext なし

### ファイル
`backend/internal/service/hospitalization_plan_service.go:90-98`

### 問題
```go
func (s *hospitalizationPlanService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    if len(ids) == 0 {
        return apperrors.WrapInvalidInput("ids must not be empty")
    }
    if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
        return apperrors.Wrap(err, "failed to reorder hospitalization plan")
    }
    return nil  // slog なし
}
```

Update slog も clinic_id が欠落（L70: `slog.InfoContext(ctx, "hospitalization plan updated", slog.Uint64("hospitalization_plan_id", id))`）。

### 修正案
```go
// Reorder
slog.InfoContext(ctx, "hospitalization plans reordered",
    slog.Uint64("clinic_id", clinicID),
    slog.Int("count", len(ids)))

// Update
slog.InfoContext(ctx, "hospitalization plan updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("hospitalization_plan_id", id))
```

---

## 問題 4: hospitalization_plan_repository の CountCarePlanItemsByPlanID が deleted_at IS NULL 欠落

### ファイル
`backend/internal/repository/hospitalization_plan_repository.go:87-96`

### 問題
```go
func (r *hospitalizationPlanRepository) CountCarePlanItemsByPlanID(ctx context.Context, planID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.CarePlanItem{}).
        Where("hospitalization_plan_id = ?", planID).  // deleted_at IS NULL なし
        Count(&count).Error; err != nil {
```

論理削除済みの CarePlanItem もカウントに含まれ、削除可能な Plan が削除不可と誤判定される（TASK-011 と同根の問題）。

### 修正案
```go
Where("hospitalization_plan_id = ? AND deleted_at IS NULL", planID).
```
