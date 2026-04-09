# BUG-247: clinical_plan に clinic_id マルチテナント境界なし

## 概要

`clinical_plan_handler.go` の `GetClinicalPlan` / `UpdateClinicalPlan` / `DeleteClinicalPlan` は
`extractClinicID(c)` を呼んでおらず、`medical_record_id` のみで操作を行う。
`clinical_plan_repository.go` の `FindByMedicalRecordID` も `clinic_id` でのフィルタリングがない。

あるクリニックの staff が他クリニックのカルテの `medical_record_id` を推測・取得した場合、
clinical_plan を参照・変更・削除できるマルチテナント境界の欠如。

## 現状コード

### `backend/internal/handler/clinical_plan_handler.go:17-25` — clinicID 未使用
```go
func (h *Handler) GetClinicalPlan(c *gin.Context) {
    medicalRecordID, err := strconv.ParseUint(c.Param("medicalRecordId"), 10, 64)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("invalid medical record ID"))
        return
    }
    // ← extractClinicID が呼ばれていない
    plan, err := h.svc.ClinicalPlan.GetByMedicalRecordID(c.Request.Context(), medicalRecordID)
    ...
}
```

### `backend/internal/repository/clinical_plan_repository.go:30-40` — clinic_id フィルタなし
```go
func (r *clinicalPlanRepository) FindByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (*model.ClinicalPlan, error) {
    var plan model.ClinicalPlan
    if err := r.db.WithContext(ctx).
        Where("medical_record_id = ?", medicalRecordID).  // ← clinic_id なし
        First(&plan).Error; err != nil {
        return nil, apperrors.FromGORM(err, "clinical_plan", fmt.Sprintf("medical_record_id=%d", medicalRecordID))
    }
    return &plan, nil
}
```

### 比較: 正しい実装（他ドメイン）
```go
// examination_handler.go — clinicID を取得して service に渡す
func (h *Handler) GetExamination(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok { return }
    // ...
    exam, err := h.svc.Examination.GetByID(c.Request.Context(), clinicID, id)
}
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/handler/clinical_plan_handler.go:17` | GetClinicalPlan | clinic_id 未検証 |
| `backend/internal/handler/clinical_plan_handler.go:35` | UpdateClinicalPlan | clinic_id 未検証 |
| `backend/internal/handler/clinical_plan_handler.go:55` | DeleteClinicalPlan | clinic_id 未検証 |
| `backend/internal/service/clinical_plan_service.go` | 全メソッド | clinicID パラメータなし |
| `backend/internal/repository/clinical_plan_repository.go:30` | FindByMedicalRecordID | clinic_id フィルタなし |
| `backend/internal/repository/clinical_plan_repository.go` | Upsert / Delete | clinic_id フィルタなし |

## 修正方針

### 1. Handler に `extractClinicID` を追加
```go
func (h *Handler) GetClinicalPlan(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    medicalRecordID, err := strconv.ParseUint(c.Param("medicalRecordId"), 10, 64)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("invalid medical record ID"))
        return
    }
    plan, err := h.svc.ClinicalPlan.GetByMedicalRecordID(c.Request.Context(), clinicID, medicalRecordID)
    ...
}
```

### 2. Service / Repository に clinicID を伝播
```go
// repository — medical_records JOIN で clinic_id を検証
func (r *clinicalPlanRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error) {
    var plan model.ClinicalPlan
    if err := r.db.WithContext(ctx).
        Joins("JOIN medical_records ON medical_records.id = clinical_plans.medical_record_id").
        Where("medical_records.clinic_id = ? AND clinical_plans.medical_record_id = ?", clinicID, medicalRecordID).
        First(&plan).Error; err != nil {
        return nil, apperrors.FromGORM(err, "clinical_plan", fmt.Sprintf("medical_record_id=%d", medicalRecordID))
    }
    return &plan, nil
}
```

## 準拠すべきプロジェクト規約

### `.claude/rules/database-design.md` — マルチテナント設計
> 常に clinic_id を条件に含める。clinic_id なしのクエリはデータリーク可能性がある。

### `.claude/CLAUDE.md` — マルチテナント
> WHERE 句は `clinic_id` から開始。

## 優先度
**Critical** — 他クリニックの clinical_plan への不正アクセスが可能。マルチテナントのデータ分離境界の欠如。

## 関連チケット
- BUG-244: バックエンド Go コード規約準拠監査（親チケット）

## 関連ファイル
- `backend/internal/handler/clinical_plan_handler.go:17-73` — Handler 修正対象
- `backend/internal/service/clinical_plan_service.go` — Service シグネチャ変更
- `backend/internal/repository/clinical_plan_repository.go` — Repository クエリ修正
