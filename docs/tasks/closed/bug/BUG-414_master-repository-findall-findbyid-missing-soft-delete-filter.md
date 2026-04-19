# BUG-414: 全マスタリポジトリの FindAll/FindByID に deleted_at フィルタ欠落

## 概要

マスタ系リポジトリ14ファイルの `FindAll` / `FindByID` メソッドで、
`deleted_at IS NULL` による論理削除フィルタが実装されていない。
削除済みレコードが API レスポンスに混入するリスクがある。

> **注意**: GORM の `gorm:"softDelete"` タグ or `gorm.Model` 埋め込みが
> モデルに設定されている場合は自動フィルタが効くため、確認してから修正すること。

## 問題箇所（全14ファイル共通）

代表例:
```go
// checkup_type_repository.go:33-39
err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).
    Order("sort_order ASC, name ASC").Find(&checkupTypes).Error
// ↑ deleted_at フィルタなし

// medicine_repository.go:31-48
err := buildBase().Offset(offset).Limit(limit).
    Order("sort_order ASC, name ASC").Find(&medicines).Error
// ↑ deleted_at フィルタなし
```

## 対象ファイル

| ファイル | 対象メソッド |
|---------|------------|
| animal_species_repository.go | FindAll, FindByID |
| checkup_type_repository.go | FindAll, FindByID |
| exam_type_repository.go | FindAll, FindByID |
| medicine_repository.go | FindAll, FindByID |
| procedure_repository.go | FindAll, FindByID |
| vaccine_repository.go | FindAll, FindByID |
| chief_complaint_repository.go | FindAll, FindByID |
| consultation_repository.go | FindAll, FindByID |
| diagnosis_repository.go | FindAll, FindByID（DiagnosisType/DiagnosisName両方） |
| reservation_type_repository.go | FindAll, FindByID |
| reservation_type_group_repository.go | FindAll, FindByID |
| reservation_type_liff_repository.go | FindAll, FindByID |
| reservation_type_occupation_repository.go | FindAll, FindByID |
| reservation_type_unavailable_time_repository.go | FindAll, FindByID |

## 確認手順

まず各モデルファイルで GORM soft delete の設定を確認する。

```bash
grep -rn "DeletedAt\|softDelete\|gorm.Model" backend/internal/model/ | grep -v "_test.go"
```

**GORM soft delete が有効な場合**: `gorm.Model` 埋め込みまたは `DeletedAt gorm.DeletedAt` フィールドがある場合、GORM が自動で `deleted_at IS NULL` を付加するため修正不要。

**GORM soft delete が無効な場合**: 手動で WHERE 句に追加が必要。

## 修正方針（GORM soft delete が無効な場合）

```go
// 修正例
err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).
    Where("deleted_at IS NULL").   // 追加
    Order("sort_order ASC, name ASC").Find(&checkupTypes).Error
```

または `clinicScope` 内で `deleted_at IS NULL` を統合する。

```go
// base.go の clinicScope に論理削除を統合
func clinicScope(clinicID uint64) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("clinic_id = ? AND deleted_at IS NULL", clinicID)
    }
}
```

## 優先度

**High** — 削除済みレコードが API レスポンスに混入する可能性。ただし GORM 設定確認次第で対応範囲が変わる。

## 関連規約

- `.claude/rules/database-design.md` — 「論理削除対応」セクション
- `docs/ERD.md` — 各テーブルの `deleted_at` カラム定義
