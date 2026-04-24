# BUG-377: medicine_repository.CountUsageByMedicineID で clinic_id テナント分離が欠如

## 概要
`medicine_repository.go` の `CountUsageByMedicineID` メソッドは、`treatments` と `care_plan_items` をカウントする際に `clinic_id` フィルタを適用していない。マルチテナント環境において、他クリニックのレコードがカウントに含まれる可能性があり、誤った「使用中」判定が発生しうる。

## 脆弱性分類
- **CWE-284**: Improper Access Control
- **OWASP**: A01:2021 Broken Access Control
- **影響**: 他クリニックの treatments/care_plan_items が参照カウントに混入し、削除可能なマスタが「使用中」として誤検知される（またはその逆）

## 再現手順
1. クリニック A で薬品マスタを作成
2. クリニック B の treatment に同じ `medicine_id` を使用
3. クリニック A の薬品削除を試みる
4. **結果**: クリニック B のカウントが混入し、誤った 409 Conflict が返る

## 期待する動作
- `CountUsageByMedicineID` は同一 `clinic_id` のレコードのみカウントすること
- 他クリニックのデータは一切参照しないこと

## 現状コード

### `backend/internal/repository/medicine_repository.go:62-76`
```go
func (r *medicineRepository) CountUsageByMedicineID(ctx context.Context, medicineID uint64) (int64, error) {
	var treatmentCount, carePlanCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Where("medicine_id = ?", medicineID). // clinic_id フィルタなし
		Count(&treatmentCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	if err := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Where("medicine_id = ?", medicineID). // clinic_id フィルタなし
		Count(&carePlanCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return treatmentCount + carePlanCount, nil
}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/repository/merchandise_item_repository.go
func (r *merchandiseItemRepository) CountUsageByMerchandiseItemID(ctx context.Context, clinicID, itemID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.BillingItem{}).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billing_items.merchandise_item_id = ?", itemID).
		Count(&count).Error
	return count, apperrors.FromGORM(err, "billing_item", "")
}
```

`backend/CLAUDE.md` に「JOIN 経由でテナント判定する場合は JOIN 条件に clinic_id を明示する」と明記されている。

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/repository/medicine_repository.go:62-76` | CountUsageByMedicineID の treatments/care_plan_items カウント | 要修正 |
| `backend/internal/service/medicine_service.go` | Delete メソッドでこの関数を呼び出し | 間接影響 |
| `/v1/masters/medicines/{id}` DELETE エンドポイント | 削除前依存チェック | 間接影響 |

## 修正方針

### `backend/internal/repository/medicine_repository.go:62-76`

インターフェースのシグネチャも `clinicID` を追加する必要がある。

```go
// インターフェース変更
CountUsageByMedicineID(ctx context.Context, clinicID, medicineID uint64) (int64, error)

// 実装変更
func (r *medicineRepository) CountUsageByMedicineID(ctx context.Context, clinicID, medicineID uint64) (int64, error) {
	var treatmentCount, carePlanCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id"+
			" AND medical_records.clinic_id = ? AND medical_records.deleted_at IS NULL", clinicID).
		Where("treatments.medicine_id = ? AND treatments.deleted_at IS NULL", medicineID).
		Count(&treatmentCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	if err := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Joins("JOIN hospitalization_plans ON hospitalization_plans.id = care_plan_items.hospitalization_plan_id"+
			" AND hospitalization_plans.clinic_id = ? AND hospitalization_plans.deleted_at IS NULL", clinicID).
		Where("care_plan_items.medicine_id = ?", medicineID).
		Count(&carePlanCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return treatmentCount + carePlanCount, nil
}
```

### `backend/internal/service/medicine_service.go` — 呼び出し側の更新
```go
// clinicID を渡すように変更
count, err := s.repo.Medicine.CountUsageByMedicineID(ctx, clinicID, id)
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `backend/CLAUDE.md` — JOIN 経由のテナント判定
> JOIN 経由でテナント判定する場合（billing_items→billings、treatments→medical_records 等）は JOIN 条件に `AND テーブル名.clinic_id = ?` を明示する。`clinicScope` は使用不可。

### `backend/CLAUDE.md` — JOIN を含む repository メソッドのレビューチェックリスト
> - [ ] JOIN 先テーブルの `clinic_id` フィルタが JOIN 条件に含まれているか
> - [ ] JOIN 先テーブルの `deleted_at IS NULL` が JOIN 条件に含まれているか

### プロジェクト内参照実装
`backend/internal/repository/merchandise_item_repository.go` — `CountUsageByMerchandiseItemID` で正しくJOIN経由のclinic_idフィルタを実装

## 優先度
**Critical** — マルチテナント境界の不備。他クリニックのデータが誤判定に影響するデータ整合性問題。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/repository/medicine_repository.go:62-76` — 問題箇所
- `backend/internal/service/medicine_service.go` — 呼び出し側（clinicID引数追加が必要）
- `backend/internal/repository/medicine_repository.go:17-25` — インターフェース定義（シグネチャ変更が必要）
