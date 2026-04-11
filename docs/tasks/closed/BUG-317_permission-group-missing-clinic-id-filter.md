# BUG-317: PermissionGroup の FindByID / Delete に clinic_id フィルタ欠如 — テナント間アクセス可能

## 概要

`permission_group_repository.go` の `FindByID` と `Delete` に `clinic_id` フィルタが存在しない。
認証済みの Clinic A のスタッフが、他クリニック（Clinic B）の権限グループIDを知るだけで
読み取り・削除できる IDOR (Insecure Direct Object Reference) 脆弱性。

## 脆弱性分類

- **CWE-639**: Authorization Bypass Through User-Controlled Key
- **OWASP**: A01:2021 Broken Access Control
- **影響**:
  - GET: 他クリニックの権限グループ定義（リソース別権限設定）を読み取られる
  - DELETE: 他クリニックの権限グループを削除される（サービス妨害）

## 再現手順

1. Clinic A のスタッフアカウント（有効JWTトークン保持）でログイン
2. Clinic B の権限グループ ID（例: 42）を推測または列挙
3. `GET /api/clinics/{任意のclinicID}/permission-groups/42` を実行
4. **結果**: Clinic A のトークンで Clinic B のデータが返る

```bash
# 攻撃者が Clinic A のトークンで Clinic B のグループを削除
curl -X DELETE https://stg.noah-karte.com/v1/clinics/999/permission-groups/42 \
  -H "Authorization: Bearer <Clinic_A_JWT>"
```

## 期待する動作

- `FindByID` は `clinic_id` も照合し、別クリニックのグループは 404 を返す
- `Delete` は `clinic_id` も照合し、別クリニックのグループは 404 を返す

## 現状コード

### `backend/internal/repository/permission_group_repository.go:53-61`
```go
func (r *permissionGroupRepository) FindByID(ctx context.Context, id uint64) (*model.PermissionGroup, error) {
	var group model.PermissionGroup
	err := r.db.WithContext(ctx).
		Preload("Rules").
		First(&group, "id = ? AND deleted_at IS NULL", id).Error  // ← clinic_id なし
	if err != nil {
		return nil, apperrors.FromGORM(err, "permission_group", fmt.Sprintf("%d", id))
	}
	return &group, nil
}
```

### `backend/internal/repository/permission_group_repository.go:92-104`
```go
func (r *permissionGroupRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Model(&model.PermissionGroup{}).
		Where("id = ?", id).  // ← clinic_id なし
		Update("deleted_at", gorm.Expr("now()"))
	// ...
}
```

### `backend/internal/service/permission_group_service.go:27,30,51-53,88-96`
```go
// Interface にも clinicID が不在
GetByID(ctx context.Context, id uint64) (*model.PermissionGroup, error)
Delete(ctx context.Context, id uint64) error

func (s *permissionGroupService) GetByID(ctx context.Context, id uint64) (*model.PermissionGroup, error) {
	result, err := s.repo.FindByID(ctx, id)  // clinicID を渡していない
	// ...
}

func (s *permissionGroupService) Delete(ctx context.Context, id uint64) error {
	// ...
	if err := s.repo.Delete(ctx, id); err != nil {  // clinicID を渡していない
	// ...
}
```

### 比較: 正しい実装（同リポジトリ内）
```go
// permission_group_repository.go:75-90 — Update は clinicID を使用している
func (r *permissionGroupRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.PermissionGroup{}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", id, clinicID).  // ← 正しい
		Updates(fields)
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/repository/permission_group_repository.go:53` | `FindByID` — clinic_id フィルタなし | 未修正 |
| `backend/internal/repository/permission_group_repository.go:92` | `Delete` — clinic_id フィルタなし | 未修正 |
| `backend/internal/service/permission_group_service.go:27` | `PermissionGroupService` interface — `GetByID` に clinicID なし | 未修正 |
| `backend/internal/service/permission_group_service.go:30` | `PermissionGroupService` interface — `Delete` に clinicID なし | 未修正 |
| `backend/internal/service/permission_group_service.go:51` | `GetByID` 実装 — clinicID を渡していない | 未修正 |
| `backend/internal/service/permission_group_service.go:88` | `Delete` 実装 — clinicID を渡していない | 未修正 |
| `backend/internal/handler/permission_group_handler.go:20` | `GetPermissionGroup` — clinicID を service に渡していない | 未修正 |
| `backend/internal/handler/permission_group_handler.go:145` | `DeletePermissionGroup` — clinicID を service に渡していない | 未修正 |

注: `Update` の後に呼ばれる `repo.FindByID(ctx, id)` (service.go:81) も clinic_id なし → 要確認

## 修正方針

### 1. Repository Interface 修正 — `permission_group_repository.go:18`
```go
// Before
FindByID(ctx context.Context, id uint64) (*model.PermissionGroup, error)
Delete(ctx context.Context, id uint64) error

// After
FindByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
Delete(ctx context.Context, clinicID, id uint64) error
```

### 2. Repository 実装修正 — `permission_group_repository.go:53-61`
```go
func (r *permissionGroupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
	var group model.PermissionGroup
	err := r.db.WithContext(ctx).
		Preload("Rules").
		First(&group, "id = ? AND clinic_id = ? AND deleted_at IS NULL", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "permission_group", fmt.Sprintf("%d", id))
	}
	return &group, nil
}
```

### 3. Repository 実装修正 — `permission_group_repository.go:92-104`
```go
func (r *permissionGroupRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Model(&model.PermissionGroup{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Update("deleted_at", gorm.Expr("now()"))
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "permission_group", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("permission_group", fmt.Sprintf("%d", id))
	}
	return nil
}
```

### 4. Service Interface 修正 — `permission_group_service.go:27,30`
```go
GetByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
Delete(ctx context.Context, clinicID, id uint64) error
```

### 5. Service 実装修正 — `permission_group_service.go:51-53,88-96`
```go
func (s *permissionGroupService) GetByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	// ...
}

func (s *permissionGroupService) Delete(ctx context.Context, clinicID, id uint64) error {
	// ...
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
	// ...
}
```

### 6. Handler 修正 — `permission_group_handler.go:20-31,144-160`
```go
func (h *Handler) GetPermissionGroup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	pg, err := h.svc.PermissionGroup.GetByID(c.Request.Context(), clinicID, id)
	// ...
}

func (h *Handler) DeletePermissionGroup(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.PermissionGroup.Delete(c.Request.Context(), clinicID, id); err != nil {
	// ...
}
```

### 7. Update 後の FindByID も修正 — `permission_group_service.go:81`
```go
// Update 後の再取得も clinicID を渡す
result, err := s.repo.FindByID(ctx, clinicID, id)
```

## 準拠すべきプロジェクト規約

### `.claude/rules/database-design.md` — マルチテナント設計（clinic_id 必須）
> ```sql
> -- ❌ 危険: clinic_id なしのクエリ（データリーク可能性）
> SELECT * FROM owners WHERE id = 1;
>
> -- ✅ 安全: 常に clinic_id を条件に含める
> SELECT * FROM owners WHERE clinic_id = $1 AND id = $2;
> ```

### プロジェクト内参照実装
- `backend/internal/repository/permission_group_repository.go:75` — `Update` が正しく `clinicID` を使用
- `backend/internal/repository/permission_group_repository.go:220-235` — `Reorder` が正しく `clinicID` を使用
- `backend/internal/repository/vaccination_repository.go` — `FindByID(ctx, clinicID, id)` パターン

## 優先度

**Critical** — 認証済みユーザーが他クリニックの権限グループを読み取り・削除できる IDOR 脆弱性。本番環境で即時対応が必要。

## 関連チケット

- BUG-316: DB テキストフィールド命名統一（別件）

## 関連ファイル

- `backend/internal/repository/permission_group_repository.go` — Repository 実装
- `backend/internal/service/permission_group_service.go` — Service 実装
- `backend/internal/handler/permission_group_handler.go` — Handler 実装
