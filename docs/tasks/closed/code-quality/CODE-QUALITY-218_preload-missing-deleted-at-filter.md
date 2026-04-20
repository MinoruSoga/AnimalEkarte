# CODE-QUALITY-218: Preload 関連モデルの deleted_at フィルタ欠落

## 概要

複数の repository ファイルで、soft delete 対象の関連モデルを `Preload` する際に
`deleted_at IS NULL` フィルタが付いていない。
削除済みの関連モデルが取得されてしまう可能性がある。

## 該当箇所

### 1. `reservation_type_repository.go` — Preload("Group") に deleted_at フィルタなし

```go
// FindAll (行35)
Preload("Group")

// FindByID (行46)
Preload("Group")
```

`reservation_type_groups` テーブルは soft delete (`deleted_at`) を持つ。
削除済みグループが紐付いた予約種別を取得すると、削除済みグループ情報が返却される。

**修正案:**
```go
Preload("Group", "deleted_at IS NULL")
```

### 2. `diagnosis_repository.go` — Preload("Names") に deleted_at フィルタなし

```go
// FindAll (行42)
Preload("Names")

// FindByID (行54)
Preload("Names")
```

`diagnosis_names` テーブルは soft delete (`deleted_at`) を持つ。
削除済みの診断名が診断タイプの一覧に含まれて返却される。

**修正案:**
```go
Preload("Names", "deleted_at IS NULL")
```

## 既存の正しい実装との比較

`reservation_type_occupation_repository.go` は CODE-QUALITY-214 で修正済み:
```go
// ✅ 修正後
Preload("Occupation", "clinic_id = ? AND deleted_at IS NULL", clinicID)
```

同パターンを他の Preload にも適用すべき。

## 修正方針

1. `reservation_type_repository.go` の `Preload("Group")` を2箇所修正
2. `diagnosis_repository.go` の `Preload("Names")` を2箇所修正

なお、以下の Preload は soft delete を持たないモデルのため修正不要:
- `permission_group_repository.go` の `Preload("Rules")` — PermissionGroupRule に DeletedAt なし
- `exam_type_repository.go` の `Preload("Items")` — ExamTypeField に DeletedAt なし

## 優先度

HIGH — 削除済み関連モデルがAPIレスポンスに混入するデータ整合性問題。
