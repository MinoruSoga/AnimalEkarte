# BUG-383: inventory_repository.CountUsageByInventoryID で clinic_id テナント分離が欠如

## 概要
`inventory_repository.go` の `CountUsageByInventoryID` は、`treatments`・`vaccines`・`medicines` テーブルをカウントする際に `clinic_id` フィルタを適用していない。Raw SQL クエリで `WHERE inventory_id = ?` のみ条件にしているため、他クリニックのレコードがカウントに混入し、誤った「使用中」判定が発生する。BUG-377（medicine_repository の同種バグ）と同一パターン。

## 脆弱性分類
- **CWE-284**: Improper Access Control
- **OWASP**: A01:2021 Broken Access Control
- **影響**: 他クリニックの treatments/vaccines/medicines が参照カウントに混入し、削除可能な在庫が「使用中」として誤検知される

## 再現手順
1. クリニック A で在庫マスタを作成（inventory_id = X）
2. クリニック B の treatment に同じ `inventory_id = X` を持つレコードが存在する
3. クリニック A の在庫削除を試みる（`DELETE /v1/inventory/{id}`）
4. **結果**: クリニック B のカウントが混入し、誤った 409 Conflict が返る

## 期待する動作
- `CountUsageByInventoryID` は同一 `clinic_id` のレコードのみカウントすること

## 現状コード

### `backend/internal/repository/inventory_repository.go:128-145`
```go
// CountUsageByInventoryID は在庫アイテムを参照している治療明細・ワクチン・薬剤の件数を返す（BUG-195）
func (r *inventoryRepository) CountUsageByInventoryID(ctx context.Context, inventoryID uint64) (int64, error) {
	var count int64
	// treatments, vaccines, medicines のいずれかから参照されていればカウント
	err := r.db.WithContext(ctx).
		Raw(`SELECT (
			SELECT COUNT(*) FROM treatments WHERE inventory_id = ? AND deleted_at IS NULL
		) + (
			SELECT COUNT(*) FROM vaccines  WHERE inventory_id = ? AND deleted_at IS NULL
		) + (
			SELECT COUNT(*) FROM medicines WHERE inventory_id = ? AND deleted_at IS NULL
		) AS total`, inventoryID, inventoryID, inventoryID).
		Scan(&count).Error
	// clinic_id フィルタが一切ない
}
```

### インターフェース定義（シグネチャに clinicID がない）
```go
// backend/internal/repository/inventory_repository.go:20
CountUsageByInventoryID(ctx context.Context, inventoryID uint64) (int64, error)
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/repository/merchandise_item_repository.go
CountUsageByMerchandiseItemID(ctx context.Context, clinicID, merchandiseItemID uint64) (int64, error)

// 実装: JOIN billings で clinic_id を JOIN 条件に明示
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

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/repository/inventory_repository.go:20` | インターフェース定義（clinicID 引数なし） | 要修正 |
| `backend/internal/repository/inventory_repository.go:129-144` | CountUsageByInventoryID の Raw SQL | 要修正 |
| `backend/internal/service/inventory_service.go` | Delete メソッドの呼び出し側（clinicID 追加が必要） | 要修正 |

## 修正方針

### 1. `backend/internal/repository/inventory_repository.go:20` — インターフェース変更
```go
// 修正前
CountUsageByInventoryID(ctx context.Context, inventoryID uint64) (int64, error)

// 修正後
CountUsageByInventoryID(ctx context.Context, clinicID, inventoryID uint64) (int64, error)
```

### 2. `backend/internal/repository/inventory_repository.go:129-144` — 実装変更
```go
func (r *inventoryRepository) CountUsageByInventoryID(ctx context.Context, clinicID, inventoryID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Raw(`SELECT (
			SELECT COUNT(*) FROM treatments
				JOIN medical_records ON medical_records.id = treatments.medical_record_id
					AND medical_records.clinic_id = ? AND medical_records.deleted_at IS NULL
				WHERE treatments.inventory_id = ? AND treatments.deleted_at IS NULL
		) + (
			SELECT COUNT(*) FROM vaccines
				JOIN vaccinations ON vaccinations.id = vaccines.vaccination_id
					AND vaccinations.clinic_id = ? AND vaccinations.deleted_at IS NULL
				WHERE vaccines.inventory_id = ? AND vaccines.deleted_at IS NULL
		) + (
			SELECT COUNT(*) FROM medicines
				WHERE inventory_id = ? AND clinic_id = ? AND deleted_at IS NULL
		) AS total`,
			clinicID, inventoryID,
			clinicID, inventoryID,
			inventoryID, clinicID).
		Scan(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "inventory_item", fmt.Sprintf("%d", inventoryID))
	}
	return count, nil
}
```

※ 各テーブルのスキーマを確認して適切な JOIN 条件を確定すること。

### 3. `backend/internal/service/inventory_service.go` — 呼び出し側更新
```go
// clinicID を渡すように変更
count, err := s.repo.Inventory.CountUsageByInventoryID(ctx, clinicID, id)
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `backend/CLAUDE.md` — JOIN を含む repository メソッドのレビューチェックリスト
> - [ ] JOIN 先テーブルの `clinic_id` フィルタが JOIN 条件に含まれているか

### プロジェクト内参照実装
`backend/internal/repository/merchandise_item_repository.go` — `CountUsageByMerchandiseItemID` で正しく clinic_id JOIN を実装

## 優先度
**Critical** — BUG-377 と同一のマルチテナント境界不備。他クリニックのデータが混入する。

## 関連チケット
- BUG-377: medicine_repository.CountUsageByMedicineID の同種バグ

## 関連ファイル
- `backend/internal/repository/inventory_repository.go:20,129-144` — 問題箇所
- `backend/internal/service/inventory_service.go` — 呼び出し側（clinicID 追加が必要）
