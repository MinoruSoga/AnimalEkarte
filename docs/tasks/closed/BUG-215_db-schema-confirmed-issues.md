# BUG-215: DB スキーマ — 確認済みの問題

| 項目 | 内容 |
|------|------|
| 優先度 | **High** |
| カテゴリ | データベース設計 |

## 確認済み

### 1. payments テーブルに deleted_at なし

`backend/migrations/001_init.sql` — `CREATE TABLE payments` に `deleted_at` カラムが存在しない。
支払記録が物理削除のみ可能。財務監査の観点でリスクあり。

### 2. billing_items テーブルに updated_at なし

`backend/migrations/001_init.sql` — `CREATE TABLE billing_items` に `updated_at` カラムが存在しない。
`created_at` と `deleted_at` のみ。GORM の `UpdatedAt` 自動更新が機能しない。

### 3. vital_repository.FindByID に clinicID なし

`backend/internal/repository/vital_repository.go:42-49`
```go
func (r *vitalRepository) FindByID(ctx context.Context, id uint64) (*model.VitalRecord, error) {
    var vital model.VitalRecord
    err := r.db.WithContext(ctx).First(&vital, "id = ?", id).Error
```

### 4. billing_item_repository.FindByID に clinicID なし

`backend/internal/repository/billing_item_repository.go:31-38`
```go
func (r *billingItemRepository) FindByID(ctx context.Context, id uint64) (*model.BillingItem, error) {
    var item model.BillingItem
    err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
```

## 判断保留（設計意図の確認が必要）

以下は「違反」ではなく設計判断として確認が必要：
- `animal_species` に clinic_id なし → システム共通マスタとして意図的な可能性
- `treatments`, `vital_records` テーブルに clinic_id なし → FK チェーンで間接分離
- 30+テーブルに deleted_at なし → append-only やマスタとして意図的な可能性が混在

これらは設計意図を確認してからイシュー化すること。
