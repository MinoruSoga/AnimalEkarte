# BUG-265: マルチテナント clinic_id 欠落（第2波）

## 概要

BUG-254（第1波）で8ドメインの clinic_id を修正したが、追加の repository/handler で clinic_id テナント境界チェックが欠落している。
これにより他クリニックのデータへの不正アクセスが可能な状態。

## 脆弱性分類
- **CWE-639**: Authorization Bypass Through User-Controlled Key
- **OWASP**: A01:2021 Broken Access Control
- **影響**: 他クリニックのデータ参照・更新・削除が可能

## 影響範囲

### Repository 層（clinic_id パラメータ欠落）

| ファイル | メソッド | 現状 | 深刻度 |
|----------|---------|------|--------|
| `treatment_plan_repository.go:14-19` | 全6メソッド | `id` のみで検索。clinicID パラメータなし | CRITICAL |
| `vital_repository.go:61-86` | Update, Delete | `id` のみで WHERE。clinicID なし | HIGH |
| `vital_repository.go:31` | ListByMedicalRecordID | medical_record_id のみ。JOIN なし | HIGH |
| `billing_item_repository.go:62-84` | UpdateFields, Delete | `id` のみで WHERE。billings JOIN なし | HIGH |
| `billing_review_repository.go:48` | Update | `id` のみで WHERE。medical_records JOIN なし | HIGH |
| `inquiry_repository.go:33` | UpsertByMedicalRecordID | medical_record_id のみ。JOIN なし | HIGH |

### Handler 層（extractClinicID 欠落）

| ファイル | メソッド | 現状 | 深刻度 |
|----------|---------|------|--------|
| `staff_handler.go:199` | GetStaff | `id` のみ。extractClinicID なし | HIGH |
| `staff_handler.go:233` | GetStaffPermissionGroups | `id` のみ | HIGH |
| `staff_handler.go:248` | SetStaffPermissionGroups | `id` のみ（書き込み操作） | HIGH |
| `staff_handler.go:273` | GetStaffClinicAssignments | `id` のみ | HIGH |
| `staff_handler.go:294` | SetStaffClinicAssignments | `id` のみ（書き込み操作） | HIGH |
| `staff_handler.go:321` | GetStaffExcludedServiceTypes | `id` のみ | HIGH |
| `staff_handler.go:341` | SetStaffExcludedServiceTypes | `id` のみ（書き込み操作） | HIGH |
| `inquiry_handler.go:16` | UpdateInquiry | extractClinicID なし | HIGH |

## 現状コード

### `treatment_plan_repository.go:50-57`（全メソッドが同パターン）
```go
func (r *treatmentPlanRepository) FindByID(ctx context.Context, id uint64) (*model.TreatmentPlan, error) {
    var plan model.TreatmentPlan
    err := r.db.WithContext(ctx).First(&plan, id).Error // ← clinic_id チェックなし
    ...
}
```

### 比較: 正しい実装（`treatment_repository.go:34-46`）
```go
func (r *treatmentRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Treatment, error) {
    var treatment model.Treatment
    err := r.db.WithContext(ctx).
        Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id").
        Where("medical_records.clinic_id = ? AND treatments.id = ?", clinicID, id).
        First(&treatment).Error
    ...
}
```

### `staff_handler.go:198-211`
```go
func (h *Handler) GetStaff(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    // ← extractClinicID(c) が呼ばれていない
    staff, err := h.svc.Staff.GetByID(c.Request.Context(), id)
    ...
}
```

## 修正方針

### 1. treatment_plan_repository.go — 全メソッドに clinicID 追加（JOIN パターン）

treatment_plans は clinic_id カラムを持たないため、親テーブル（medical_records または hospitalizations）経由の JOIN が必要。

```go
// Interface
type TreatmentPlanRepository interface {
    ListByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.TreatmentPlan, error)
    FindByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error)
    // ... 他メソッドも同様
}

// Implementation
func (r *treatmentPlanRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPlan, error) {
    var plan model.TreatmentPlan
    err := r.db.WithContext(ctx).
        Joins("LEFT JOIN medical_records ON medical_records.id = treatment_plans.medical_record_id").
        Joins("LEFT JOIN hospitalizations ON hospitalizations.id = treatment_plans.hospitalization_id").
        Where("treatment_plans.id = ? AND (medical_records.clinic_id = ? OR hospitalizations.clinic_id = ?)", id, clinicID, clinicID).
        First(&plan).Error
    ...
}
```

### 2. vital_repository.go Update/Delete — clinicID 追加（JOIN パターン）

```go
func (r *vitalRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
    result := r.db.WithContext(ctx).
        Model(&model.VitalRecord{}).
        Joins("JOIN medical_records ON medical_records.id = vital_records.medical_record_id").
        Where("vital_records.id = ? AND medical_records.clinic_id = ?", id, clinicID).
        Updates(fields)
    ...
}
```

### 3. staff_handler.go — extractClinicID 追加

```go
func (h *Handler) GetStaff(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    ...
    staff, err := h.svc.Staff.GetByID(c.Request.Context(), clinicID, id)
    ...
}
```

## 準拠すべきプロジェクト規約

### `.claude/rules/database-design.md` — マルチテナント設計（clinic_id 必須）
> 全テーブルに `clinic_id`（マルチテナント）
> WHERE 句は `clinic_id` から開始

### `.claude/CLAUDE.md` — エラー処理の統一
> **Handler**: `extractClinicID(c)` でテナント境界を強制

## 優先度

**Critical** — 認可バイパスによるクロスクリニックデータアクセスが可能。treatment_plan は診療計画であり、医療情報の漏洩に直結する。

## 関連チケット

- BUG-254: マルチテナント clinic_id 欠落（第1波）
- BUG-261: 第3回監査 親チケット
