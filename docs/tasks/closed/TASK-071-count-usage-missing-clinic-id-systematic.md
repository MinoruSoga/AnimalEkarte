# TASK-071: CountUsageBy* メソッド — clinic_id フィルタ欠落（組織的バグ）

## 優先度

HIGH

---

## 概要

マスタ削除前の FK 依存チェック（`CountUsageBy*` / `CountRecordsBy*`）が、
複数リポジトリで `clinic_id` によるテナント分離なしに実装されている。

DELETE 前のチェックが**全テナントのレコードをカウント**するため：
1. 他クリニックのデータが「使用中」と誤判定されて正当な削除が拒否される
2. 他テナントのレコード件数が実質的に参照できてしまう（情報漏洩）

TASK-069（consultation_repository）と同一パターンの、組織的な実装ミス。

---

## 対象リポジトリ・メソッド一覧

| ファイル | メソッド | カウント対象 | 状態 |
|---------|---------|------------|------|
| `exam_type_repository.go` | `CountUsageByExamTypeID(ctx, examTypeID)` | `examinations` | ❌ clinic_id なし |
| `checkup_type_repository.go` | `CountUsageByCheckupTypeID(ctx, checkupTypeID)` | `checkup_records` | ❌ clinic_id なし |
| `procedure_repository.go` | `CountUsageByProcedureID(ctx, procedureID)` | `treatments` + `care_plan_items` | ❌ clinic_id なし |
| `vaccine_repository.go` | `CountUsageByVaccineID(ctx, vaccineID)` | `vaccinations` | ❌ clinic_id なし |
| `merchandise_item_repository.go` | `CountUsageByMerchandiseItemID(ctx, merchandiseItemID)` | `billing_items` (ok) + `estimate_items` | ⚠️ estimate_items のみ欠落 |
| `trimming_master_repository.go` | `CountRecordsByCourseID(ctx, courseID)` | `trimming_records` | ❌ clinic_id なし |
| `trimming_master_repository.go` | `CountRecordsByOptionID(ctx, optionID)` | `trimming_records` | ❌ clinic_id なし |
| `diagnosis_repository.go` | `CountNamesByCategoryID(ctx, categoryID)` | `diagnosis_names` | ❌ clinic_id なし |
| `diagnosis_repository.go` | `CountClinicalPlansByDiagnosisNameID(ctx, nameID)` | `clinical_plans` | ❌ clinic_id なし |
| `insurance_repository.go` | `CountPetsByInsuranceID(ctx, insuranceID)` | `pets` | ❌ clinic_id なし |
| `occupation_repository.go` | `CountStaffsByOccupationID(ctx, occupationID)` | `staffs` | ❓ staffs がクロステナント設計か要確認 |
| `permission_group_repository.go` | `CountStaffsByGroupID(ctx, groupID)` | `staffs` | ❓ 同上 |

---

## 修正パターン

### パターン A: カウント対象テーブルに clinic_id が直接ある場合

```go
// ❌ 修正前
func (r *examTypeRepository) CountUsageByExamTypeID(ctx context.Context, examTypeID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Examination{}).
        Where("exam_type_id = ?", examTypeID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "examination", "")
    }
    return count, nil
}

// ✅ 修正後
func (r *examTypeRepository) CountUsageByExamTypeID(ctx context.Context, clinicID, examTypeID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Examination{}).
        Where("exam_type_id = ? AND clinic_id = ?", examTypeID, clinicID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "examination", "")
    }
    return count, nil
}
```

### パターン B: clinic_id が親テーブル経由の場合（JOIN が必要）

```go
// ✅ JOIN でテナント分離（consultation TASK-069 参照実装）
func (r *procedureRepository) CountUsageByProcedureID(ctx context.Context, clinicID, procedureID uint64) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.Treatment{}).
        Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id"+
            " AND medical_records.clinic_id = ? AND medical_records.deleted_at IS NULL", clinicID).
        Where("treatments.procedure_id = ?", procedureID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "treatment", "")
    }
    return count, nil
}
```

---

## Interface 変更が必要なメソッド

各リポジトリの Interface 定義も合わせて修正する：

```go
// ❌ 修正前
CountUsageByExamTypeID(ctx context.Context, examTypeID uint64) (int64, error)

// ✅ 修正後
CountUsageByExamTypeID(ctx context.Context, clinicID, examTypeID uint64) (int64, error)
```

---

## Service 呼び出し側の修正

各サービスの Delete メソッドも `clinicID` を渡すよう修正する：

```go
// ❌ 修正前
count, err := s.repo.CountUsageByExamTypeID(ctx, id)

// ✅ 修正後
count, err := s.repo.CountUsageByExamTypeID(ctx, clinicID, id)
```

---

## 参照実装

- `consultation_repository.go` TASK-069 修正後の実装
- `insurance_repository.go` の `CountPetsInsuredByInsuranceID`（JOIN パターン参照）

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `exam_type_repository.go` | clinicID パラメータ追加・WHERE 句修正 |
| `checkup_type_repository.go` | 同上 |
| `procedure_repository.go` | clinicID パラメータ追加・JOIN 修正（treatments / care_plan_items） |
| `vaccine_repository.go` | clinicID パラメータ追加・WHERE 句修正 |
| `merchandise_item_repository.go` | estimate_items クエリに clinic_id JOIN 追加 |
| `trimming_master_repository.go` | clinicID パラメータ追加・WHERE 句修正（2メソッド） |
| `diagnosis_repository.go` | clinicID パラメータ追加・WHERE 句修正（2メソッド） |
| `insurance_repository.go` | clinicID パラメータ追加・WHERE 句修正 |
| `occupation_repository.go` | Staff がクロステナント設計か確認後対応 |
| `permission_group_repository.go` | 同上 |
| 各 service ファイル | Delete メソッドで clinicID を CountUsage に渡す |
| 各 repository interface | メソッドシグネチャ更新 |
