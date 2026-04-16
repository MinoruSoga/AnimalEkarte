# BUG-264: Repository トランザクション内 Wrap → FromGORM

## 概要

Repository 層のトランザクション内部で GORM エラーを `apperrors.Wrap` で処理している箇所が残存。
GORM エラーは `apperrors.FromGORM` で変換すべき（not found → 404, unique violation → 409 の自動マッピング）。

## 影響範囲

| ファイル | 行 | メソッド | 現状 |
|----------|-----|---------|------|
| `treatment_repository.go` | 108 | BulkUpdateSortOrder(内側) | `apperrors.Wrap(result.Error, ...)` |
| `owner_repository.go` | 126 | CreateWithPets(内側 owner Create) | `apperrors.Wrap(err, "create owner")` |
| `owner_repository.go` | 133 | CreateWithPets(内側 pet Create) | `apperrors.Wrap(err, "create pet")` |
| `trimming_repository.go` | 120 | SetOptions(内側 Association.Clear) | `apperrors.Wrap(err, "failed to clear trimming options")` |
| `trimming_repository.go` | 130 | SetOptions(内側 Association.Replace) | `apperrors.Wrap(err, "failed to set trimming options")` |

**合計: 5箇所 / 3ファイル**

## 現状コード

### `treatment_repository.go:104-109`
```go
result := tx.Model(&model.Treatment{}).
    Where("id = ? AND deleted_at IS NULL", u.ID).
    Update("sort_order", u.SortOrder)
if result.Error != nil {
    return apperrors.Wrap(result.Error, fmt.Sprintf("bulk update sort_order for treatment %d", u.ID))
}
```

### `owner_repository.go:122-127`
```go
if err := tx.Create(owner).Error; err != nil {
    if isUniqueConstraintErr(err) {
        return apperrors.WrapAlreadyExists("owner", "email already registered")
    }
    return apperrors.Wrap(err, "create owner")
}
```

### 比較: 正しい実装（`consultation_repository.go:92-94`）
```go
if result.Error != nil {
    return apperrors.FromGORM(result.Error, "consultation", fmt.Sprintf("%d", id))
}
```

## 修正方針

### 1. `treatment_repository.go:108`
```go
return apperrors.FromGORM(result.Error, "treatment", fmt.Sprintf("%d", u.ID))
```

### 2. `owner_repository.go:126`
```go
return apperrors.FromGORM(err, "owner", "")
```

### 3. `owner_repository.go:133`
```go
return apperrors.FromGORM(err, "pet", fmt.Sprintf("owner_id=%d", owner.ID))
```

### 4. `trimming_repository.go:120`
```go
return apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("record_id=%d", recordID))
```

### 5. `trimming_repository.go:130`
```go
return apperrors.FromGORM(err, "trimming_option", fmt.Sprintf("record_id=%d", recordID))
```

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — エラーラッピング規約
> **Repository**: GORM エラーは原則 `apperrors.FromGORM(err, "resource", id)` を使用。

### `.claude/rules/error-handling.md`
> Repository: `apperrors.FromGORM(err, "resource", id)` 使用

## 優先度

**Medium** — FromGORM は not found / unique constraint を適切な HTTP ステータスに変換する。Wrap だとすべて 500 になる。

## 関連チケット

- BUG-248: Repository FromGORM 違反（第1波）
- BUG-255: Repository Reorder/トランザクション内 FromGORM（第2波）
- BUG-261: 第3回監査 親チケット
