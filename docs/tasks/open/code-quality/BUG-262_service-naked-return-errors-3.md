# BUG-262: Service 層 naked return（apperrors.Wrap 欠落）第3波

## 概要

BUG-249（第1波）・BUG-256（第2波）で修正されなかった service 層の naked return が 41箇所/13ファイル残存。
Repository 呼び出しの戻り値をそのまま return しており、エラーコンテキストが欠落する。

## 現状コード

### パターン: 直接 return（エラーラップなし）
```go
// ❌ 禁止: naked return
func (s *trimmingCourseService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
    return s.repo.FindAll(ctx, clinicID)
}
```

### 比較: 正しい実装（`backend/internal/service/vital_service.go:29-34`）
```go
// ✅ 正しい: apperrors.Wrap でエラーコンテキスト付加
func (s *vitalService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Vital, error) {
    vitals, err := s.repos.Vital.ListByMedicalRecordID(ctx, clinicID, medicalRecordID)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to list vitals")
    }
    return vitals, nil
}
```

## 影響範囲

| ファイル | 箇所数 | naked return メソッド |
|----------|--------|----------------------|
| `trimming_master_service.go` | 10 | List/GetByID/Create/Delete/Reorder × Course + Option |
| `merchandise_item_service.go` | 5 | List/GetByID/Update(reload×2)/Reorder |
| `medicine_service.go` | 5 | List/GetByID/Update(reload×2)/Reorder |
| `checkup_service.go` | 4 | List/ListByClinic/Create(reload)/Update(reload) |
| `treatment_plan_service.go` | 4 | ListByMedicalRecord/ListByHospitalization/Create(reload)/Update(reload) |
| `reservation_course_service.go` | 3 | Update(early return)/Update(reload)/PatchStatus(reload) |
| `billing_item_service.go` | 2 | UpdateItem(reload)/recalculateTotals(UpdateBillingTotals) |
| `hospitalization_service.go` | 2 | List/GetByID |
| `reservation_staff_service.go` | 2 | Update(reload)/PatchStatus(reload) |
| `estimate_service.go` | 2 | Create(reload)/Update(reload) |
| `reservation_service.go` | 1 | Delete |
| `reservation_customer_service.go` | 1 | Create(reload) |
| `record_image_service.go` | 1 | List |

**合計: 41箇所 / 13ファイル**

## 修正方針

全箇所を以下パターンに統一:

### 1. List/GetByID 系（読み取り）
```go
func (s *trimmingCourseService) List(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
    courses, err := s.repo.FindAll(ctx, clinicID)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to list trimming courses")
    }
    return courses, nil
}
```

### 2. Create/Update 後のリロード
```go
// Create/Update 成功後に FindByID でリロードする場合
result, err := s.repo.FindByID(ctx, clinicID, id)
if err != nil {
    return nil, apperrors.Wrap(err, "failed to reload trimming course after create")
}
return result, nil
```

### 3. Delete/Reorder パススルー
```go
func (s *trimmingCourseService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
        return apperrors.Wrap(err, "failed to reorder trimming courses")
    }
    return nil
}
```

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md` — エラーラッピング規約
> Service: 内部エラーは `apperrors.Wrap(err, "message")` でラッピング。

### `.claude/rules/go-language.md` — エラーハンドリング（Sentinel + Wrap）
> service層で Wrap

## 優先度

**High** — エラーコンテキストが欠落すると、障害調査時にどのサービスメソッドでエラーが発生したか特定困難になる。

## 関連チケット

- BUG-249: Service naked return（第1波）
- BUG-256: Service naked return（第2波）
- BUG-261: 第3回監査 親チケット
