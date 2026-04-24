# TASK-032: 臨床系 MEDIUM 問題 4件

## 優先度

MEDIUM

---

## 問題 1: vital_repository の Update/Delete が非アトミックな 2-query パターン

### ファイル
`backend/internal/repository/vital_repository.go:62-95`

### 問題
`Update` / `Delete` ともに「FindByID で所有確認 → 別クエリで更新/削除」の 2-query パターン。FindByID と Update の間のレース条件に加え、Update 文に clinic_id が含まれていないため所有確認が有名無実になっている。他の repository（treatment, vaccination）は JOIN で clinic_id を UPDATE/DELETE に直接含めており整合性がある。

### 修正案
```go
// vital_repository.go Update
result := r.db.WithContext(ctx).
    Model(&model.VitalRecord{}).
    Joins("JOIN medical_records ON medical_records.id = vital_records.medical_record_id"+
        " AND medical_records.clinic_id = ?"+
        " AND medical_records.deleted_at IS NULL", clinicID).
    Where("vital_records.id = ?", id).
    Updates(fields)
if result.Error != nil {
    return nil, apperrors.FromGORM(result.Error, "vital_record", fmt.Sprintf("%d", id))
}
if result.RowsAffected == 0 {
    return nil, apperrors.WrapNotFound("vital_record", fmt.Sprintf("%d", id))
}
```

Delete も同様のパターンで実装する。

---

## 問題 2: hospitalization Delete で DailyRecords / TreatmentPlans の FK チェック欠落

### ファイル
`backend/internal/service/hospitalization_service.go:140-157`

### 問題
`Delete` は `CarePlanItems` の依存チェックは行うが、`DailyRecords`・`TreatmentPlans` の存在チェックがない。DB の FK 制約設定（CASCADE vs RESTRICT）によっては実行時エラーになる。

### 修正案
```go
// 削除前に全子テーブルをチェック
dailyCount, _ := s.repo.CountDailyRecordsByHospitalizationID(ctx, id)
if dailyCount > 0 {
    return apperrors.WrapConflict("入院記録に日々の記録が登録されているため削除できません")
}
planCount, _ := s.repo.CountTreatmentPlansByHospitalizationID(ctx, id)
if planCount > 0 {
    return apperrors.WrapConflict("入院記録に治療計画が登録されているため削除できません")
}
```

---

## 問題 3: hospitalization_plan Delete で CarePlanItems の FK チェック欠落

### ファイル
`backend/internal/handler/hospitalization_plan_handler.go:147-162`

### 問題
`HospitalizationPlan` は `care_plan_items.hospitalization_plan_id` から参照されている。`DeleteHospitalizationPlan` の service/repository に依存チェックが存在しない可能性がある。`CountUsageByHospitalizationPlanID` → `WrapConflict` パターンを実装すること。

---

## 問題 4: trimming_handler の CreateTrimming で ReservationTypeID=0 を未検証

### ファイル
`backend/internal/handler/trimming_handler.go:101-134`

### 問題
`ReservationTypeID` が 0 のまま渡されても service/repository でエラーになるまで検出されない。service 層または handler で事前チェックが必要。

### 修正案
```go
// service または handler でチェック
if input.ReservationTypeID == 0 {
    return apperrors.WrapInvalidInput("予約種別を選択してください")
}
```
