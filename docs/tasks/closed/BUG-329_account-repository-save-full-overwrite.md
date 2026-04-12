# BUG-329: account_repository の Save() が全フィールド上書きでゼロ値問題を引き起こす

## 概要
`account_repository.go` の `Update()` が `db.Save()` を使用しており、ゼロ値フィールドも含めて全カラムを上書きする。GORM の `Save()` はゼロ値を「変更なし」ではなく「値を空にする」として扱うため、予期しないデータ破損が発生しうる。プロジェクト規約では PATCH パターン（pointer 型 + `Updates(fields)`）が義務付けられている。

## 再現手順
1. account レコードの `is_active = true` を維持したまま `password_hash` だけ更新する
2. `UpdatePasswordHash` 呼び出し時にモデルの boolean フィールドが `false` ゼロ値のまま `Save()` に渡った場合、`is_active` が `false` に上書きされる
3. **結果**: アカウントが無効化され、ログインできなくなる

## 期待する動作
- `UpdatePasswordHash` は `password_hash` カラムのみを更新する
- 他のカラム（`is_active`, `is_system_admin` 等）は変更しない

## 現状コード

### `backend/internal/repository/account_repository.go:65-68`
```go
// Update はアカウント情報を更新する
func (r *accountRepository) Update(ctx context.Context, account *model.Account) error {
	if err := r.db.WithContext(ctx).Save(account).Error; err != nil {
		return apperrors.FromGORM(err, "account", fmt.Sprintf("%d", account.ID))
	}
	return nil
}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/repository/vital_repository.go:62-78
func (r *vitalRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.VitalRecord{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "vital", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vital", fmt.Sprintf("%d", id))
	}
	return nil
}
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/repository/account_repository.go:66` | `Save()` 使用箇所 | 要修正 |
| `backend/internal/service/account_service.go:50` | `repo.Update(ctx, account)` 呼び出し側 | シグネチャ変更が必要 |
| `backend/internal/handler/auth_handler.go:507-511` | `ChangeMyPassword` からの呼び出し経路 | シグネチャ変更に合わせて修正 |

## 修正方針

### 1. repository シグネチャを `UpdateFields` に変更 — `backend/internal/repository/account_repository.go:64-70`
```go
func (r *accountRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.Account{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "account", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("account", fmt.Sprintf("%d", id))
	}
	return nil
}
```

### 2. service 呼び出しを fields map に変更 — `backend/internal/service/account_service.go:44-56`
```go
func (s *accountService) UpdatePasswordHash(ctx context.Context, accountID uint64, newHash string) error {
	if err := s.repo.Update(ctx, accountID, map[string]any{"password_hash": newHash}); err != nil {
		return apperrors.Wrap(err, "failed to update password hash")
	}
	slog.InfoContext(ctx, "password hash updated",
		slog.Uint64("account_id", accountID))
	return nil
}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — GORM PATCH パターン (MANDATORY)
> PATCH は pointer 型 + `buildXxxUpdateFields()` パターンを使用
> `.Save()` で全フィールド上書きは禁止

### `.claude/rules/go-language.md` — GORM PATCH（ポインタ型 + buildUpdateFields）
> `repository/owner_repository.go` の `UpdateFields(ctx, id, fields)` が参照実装

### プロジェクト内参照実装
- `backend/internal/repository/vital_repository.go:62-78` — `Updates(fields)` パターン
- `backend/internal/repository/billing_item_repository.go:62-78` — `UpdateFields` パターン

## 優先度
**High** — ゼロ値フィールドが Save に渡るとアカウントが意図せず無効化される恐れがある

## 関連チケット
なし

## 関連ファイル
- `backend/internal/repository/account_repository.go:64-70` — 修正対象
- `backend/internal/service/account_service.go:44-56` — 修正対象
- `backend/internal/repository/repositories.go` — AccountRepository インターフェース定義
