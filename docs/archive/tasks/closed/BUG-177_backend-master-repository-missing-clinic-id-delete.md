# BUG-177: マスタ系 Repository の Delete に clinic_id フィルタ欠如（マルチテナント違反・最重大）

## 概要

15 以上のマスタ系 Repository で `Delete` 操作時に `clinic_id` フィルタが存在しない。`DELETE WHERE id = ?` のみで実行しているため、**テナントAのユーザーがテナントBのマスタデータを削除できる可能性**がある。参照（BUG-176）よりもはるかに深刻な破壊的操作の権限逸脱。

## 脆弱性分類
- **CWE-284**: Improper Access Control
- **CWE-639**: Authorization Bypass Through User-Controlled Key
- **OWASP**: A01:2021 Broken Access Control
- **影響**: 悪意あるユーザーが他テナントのマスタデータ（診察種別・薬剤・ケージ・ワクチン等）を**削除**できる。データの破壊・業務停止につながる。

## 再現手順

1. テナント B のアカウントでログインし、JWT トークンを取得
2. テナント A で作成されたマスタデータの ID を何らかの方法で取得（例: 連番推測）
3. `DELETE /v1/checkup-types/{id}` を実行（テナントBのトークンを使用）
4. **結果**: テナントBのトークンでテナントAのデータが削除される

## 期待する動作

```go
// ✅ 正しい: clinic_id を含めた削除（他テナントには影響しない）
result := r.db.WithContext(ctx).
    Delete(&model.CheckupType{}, "id = ? AND clinic_id = ?", id, clinicID)
```

## 現状コード

### `backend/internal/repository/checkup_type_repository.go:76`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.CheckupType{}, "id = ?", id)
```

### `backend/internal/repository/exam_type_repository.go:76`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.ExaminationType{}, "id = ?", id)
```

### `backend/internal/repository/procedure_repository.go:72`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.Procedure{}, "id = ?", id)
```

### `backend/internal/repository/vaccine_repository.go:76`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.Vaccine{}, "id = ?", id)
```

### `backend/internal/repository/cage_repository.go:75`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.Cage{}, "id = ?", id)
```

### `backend/internal/repository/insurance_repository.go:74`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.Insurance{}, "id = ?", id)
```

### `backend/internal/repository/occupation_repository.go:79`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.Occupation{}, "id = ?", id)
```

### `backend/internal/repository/consultation_repository.go:76`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.Consultation{}, "id = ?", id)
```

### `backend/internal/repository/chief_complaint_category_repository.go:77`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.ChiefComplaintCategory{}, "id = ?", id)
```

### `backend/internal/repository/inquiry_template_repository.go:77`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.InquiryTemplate{}, "id = ?", id)
```

### `backend/internal/repository/billing_item_repository.go:73`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.BillingItem{}, "id = ?", id)
```

### `backend/internal/repository/trimming_master_repository.go:74,175`
```go
// ❌ clinic_id なし Delete (TrimmingCourse / TrimmingOption)
result := r.db.WithContext(ctx).Delete(&model.TrimmingCourse{}, "id = ?", id)
result := r.db.WithContext(ctx).Delete(&model.TrimmingOption{}, "id = ?", id)
```

### `backend/internal/repository/treatment_plan_repository.go:81`
```go
// ❌ clinic_id なし、かつプレースホルダなし
result := r.db.WithContext(ctx).Delete(&model.TreatmentPlan{}, id)
```

### `backend/internal/repository/hospitalization_plan_repository.go:76`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.HospitalizationPlan{}, "id = ?", id)
```

### `backend/internal/repository/clinic_repository.go:96`
```go
// ❌ clinic_id なし Delete
result := r.db.WithContext(ctx).Delete(&model.Clinic{}, "id = ?", id)
```

### 比較: 正しい実装
```go
// ✅ 正しい実装パターン
func (r *CheckupTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
    result := r.db.WithContext(ctx).
        Delete(&model.CheckupType{}, "id = ? AND clinic_id = ?", id, clinicID)
    if result.Error != nil {
        return apperrors.FromGORM(result.Error, "checkup_type", fmt.Sprintf("%d", id))
    }
    if result.RowsAffected == 0 {
        return apperrors.FromGORM(gorm.ErrRecordNotFound, "checkup_type", fmt.Sprintf("%d", id))
    }
    return nil
}
```

## 影響範囲

| 対象 Repository | 行番号 | リスク |
|---|---|---|
| `checkup_type_repository.go` | 76 | 他テナントのデータ削除 |
| `exam_type_repository.go` | 76 | 他テナントのデータ削除 |
| `procedure_repository.go` | 72 | 他テナントのデータ削除 |
| `vaccine_repository.go` | 76 | 他テナントのデータ削除 |
| `cage_repository.go` | 75 | 他テナントのデータ削除 |
| `insurance_repository.go` | 74 | 他テナントのデータ削除 |
| `occupation_repository.go` | 79 | 他テナントのデータ削除 |
| `consultation_repository.go` | 76 | 他テナントのデータ削除 |
| `chief_complaint_category_repository.go` | 77 | 他テナントのデータ削除 |
| `inquiry_template_repository.go` | 77 | 他テナントのデータ削除 |
| `billing_item_repository.go` | 73 | 他テナントのデータ削除 |
| `trimming_master_repository.go` | 74, 175 | 他テナントのデータ削除 |
| `treatment_plan_repository.go` | 81 | プレースホルダなし + clinic_id なし |
| `hospitalization_plan_repository.go` | 76 | 他テナントのデータ削除 |
| `clinic_repository.go` | 96 | 任意クリニックの削除 |

## 修正方針

### 全 Delete メソッドのシグネチャに clinicID を追加

```go
// Before
func (r *CheckupTypeRepository) Delete(ctx context.Context, id uint64) error {
    result := r.db.WithContext(ctx).Delete(&model.CheckupType{}, "id = ?", id)

// After
func (r *CheckupTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
    result := r.db.WithContext(ctx).
        Delete(&model.CheckupType{}, "id = ? AND clinic_id = ?", id, clinicID)
    if result.Error != nil {
        return apperrors.FromGORM(result.Error, "checkup_type", fmt.Sprintf("%d", id))
    }
    if result.RowsAffected == 0 {
        return apperrors.FromGORM(gorm.ErrRecordNotFound, "checkup_type", fmt.Sprintf("%d", id))
    }
    return nil
}
```

### `treatment_plan_repository.go:81` — プレースホルダなし修正（最優先）

```go
// Before (危険: プレースホルダなし)
result := r.db.WithContext(ctx).Delete(&model.TreatmentPlan{}, id)

// After
result := r.db.WithContext(ctx).
    Delete(&model.TreatmentPlan{}, "id = ? AND medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ?)", id, clinicID)
```

### Service 層・Handler 層の clinicID 伝播

```go
// service/checkup_type_service.go
func (s *CheckupTypeService) Delete(ctx context.Context, clinicID, id uint64) error {
    return s.repo.Delete(ctx, clinicID, id)
}

// handler/checkup_type_handler.go
clinicID := getClinicID(c)  // JWTクレームから取得
if err := h.service.Delete(c.Request.Context(), clinicID, id); err != nil {
    RespondError(c, err)
    return
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/database-design.md` — マルチテナント設計
> **❌ 危険: clinic_id なしのクエリ（データリーク可能性）**
> **✅ 安全: 常に clinic_id を条件に含める**

### `.claude/rules/security.md` — Input Validation
> Validate all user input. Use allowlists over denylists.

### プロジェクト内参照実装
- `backend/internal/repository/owner_repository.go` — clinic_id 付き削除の正しいパターン

## 優先度
**High** — 破壊的操作（DELETE）に clinic_id フィルタがない。現時点では単一クリニック運用だが、マルチテナント展開前に必ず修正が必要。BUG-176（参照）より優先度が高い。

## 関連チケット
- BUG-176: clinic_id なし FindByID/FindAll（参照操作の同様の違反）

## 関連ファイル
- `backend/internal/repository/` 配下の全マスタ系 repository ファイル
- `.claude/rules/database-design.md`
- `.claude/rules/security.md`
