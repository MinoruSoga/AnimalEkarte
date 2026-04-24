# BUG-408: permission_group_repository が GORM 自動フィルタと手動 deleted_at IS NULL を二重適用

## 概要
`permission_group_repository.go` の `FindAll` と `FindByID` で、`gorm.DeletedAt` を持つモデルに対して GORM が自動で適用する `deleted_at IS NULL` フィルタに加え、手動で `Where("deleted_at IS NULL")` を記述しており二重適用されている。実害はないが、コードが冗長になり「GORM の自動フィルタを信頼していない」という誤解を招く。他の全マスタリポジトリは手動フィルタを書かずに GORM の自動フィルタに委譲しており、不統一。

## 再現手順
コードレビューで確認可能。

## 現状コード

### `backend/internal/repository/permission_group_repository.go:40-50`（FindAll）
```go
func (r *permissionGroupRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
    groups := make([]model.PermissionGroup, 0)
    err := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).Where("deleted_at IS NULL").  // ← 冗長（GORM が自動適用する）
        Preload("Rules").
        Order("sort_order ASC, name ASC").
        Find(&groups).Error
    ...
}

func (r *permissionGroupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
    var group model.PermissionGroup
    err := r.db.WithContext(ctx).
        Preload("Rules").
        Scopes(clinicScope(clinicID)).Where("id = ? AND deleted_at IS NULL", id).First(&group).Error  // ← 冗長
    ...
}
```

### GORM の自動フィルタ（model に gorm.DeletedAt がある場合）
```go
// model/permission_group.go:20
DeletedAt gorm.DeletedAt `json:"-"`
// ↑ gorm.DeletedAt が定義されているため、Find/First 時に GORM が自動で `deleted_at IS NULL` を追加
```

### 比較: 正しい実装（animal_species_repository など）
```go
func (r *animalSpeciesRepository) FindAll(ctx context.Context) ([]model.AnimalSpecies, error) {
    species := make([]model.AnimalSpecies, 0)
    err := r.db.WithContext(ctx).
        Order("sort_order ASC, name ASC").
        Find(&species).Error  // ← 手動フィルタなし。GORM が自動処理
    ...
}
```

## 修正方針

### `permission_group_repository.go` — 手動 deleted_at フィルタを削除
```go
func (r *permissionGroupRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
    groups := make([]model.PermissionGroup, 0)
    err := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).          // ← Where("deleted_at IS NULL") を削除
        Preload("Rules").
        Order("sort_order ASC, name ASC").
        Find(&groups).Error
    ...
}

func (r *permissionGroupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
    var group model.PermissionGroup
    err := r.db.WithContext(ctx).
        Preload("Rules").
        Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&group).Error  // ← AND deleted_at IS NULL を削除
    ...
}
```

## 優先度
**Low** — 機能上の実害なし。コードの冗長性と「GORM を信頼しない書き方」という誤解の排除が目的。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/repository/permission_group_repository.go:40-57` — 修正対象
