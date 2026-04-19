# TASK-069: consultation_repository — CountUsageByConsultationID にテナント分離なし

## 優先度

HIGH

---

## 概要

`consultation_repository.go` の `CountUsageByConsultationID` メソッドが `clinic_id` でスコープされていない。
DELETE 前の依存チェックで**他クリニックの Treatment レコードをカウントしてしまう**ため、
正当なデータが「使用中」と誤判定されて削除できなくなる、またはセキュリティ上の情報漏洩につながる。

---

## 問題箇所

### backend/internal/repository/consultation_repository.go（L91-100 概算）

```go
// ❌ clinicID パラメータなし → 全テナントの Treatment をカウント
func (r *consultationRepository) CountUsageByConsultationID(ctx context.Context, consultationID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Treatment{}).
        Where("consultation_id = ?", consultationID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "treatment", "")
    }
    return count, nil
}
```

**問題点:**
1. `clinicID` を引数に取らないため、WHERE 句が `consultation_id` のみ
2. `Treatment` テーブルには直接 `clinic_id` がなく、`medical_records.clinic_id` で判断する必要がある
3. 他クリニックで同じ `consultation_id` を持つ Treatment（あり得ないが）や、テナント分離が崩れた場合にカウントが漏れる

---

## 修正方針

```go
// ✅ 修正後: clinicID でテナント分離
func (r *consultationRepository) CountUsageByConsultationID(ctx context.Context, clinicID, consultationID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Treatment{}).
        Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id"+
            " AND medical_records.clinic_id = ? AND medical_records.deleted_at IS NULL", clinicID).
        Where("treatments.consultation_id = ?", consultationID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "treatment", "")
    }
    return count, nil
}
```

### Interface 変更

```go
CountUsageByConsultationID(ctx context.Context, clinicID, consultationID uint64) (int64, error)
```

### Service 呼び出し側も修正

```go
// consultation_service.go Delete メソッド
count, err := s.repo.CountUsageByConsultationID(ctx, clinicID, id) // clinicID 追加
```

---

## 参照実装

`insurance_repository.go` の `CountPetsInsuredByInsuranceID` — JOIN でテナント分離しているパターン。
