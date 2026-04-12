# BUG-331: Repositories.Transaction() が context.Context を受け取らずキャンセル不可

## 概要
`backend/internal/repository/repositories.go:144` の `Transaction()` メソッドがシグネチャに `context.Context` を持たないため、呼び出し元のコンテキスト（タイムアウト・キャンセル）が GORM トランザクションに伝播しない。長時間実行クエリが HTTP リクエストキャンセル後もブロックし続けるリスクがある。

## 再現手順
1. `hospitalization_service.go:178` から `s.repos.Transaction(func(txRepos) error { ... })` を呼び出す
2. HTTP リクエストがタイムアウトまたはクライアントキャンセルされる
3. **結果**: `Transaction` 内の DB 操作はキャンセルされず、接続を保持し続ける

## 期待する動作
- `Transaction(ctx context.Context, fn func(*Repositories) error) error` でコンテキストを受け取る
- 内部で `r.db.WithContext(ctx).Transaction(...)` を使用する

## 現状コード

### `backend/internal/repository/repositories.go:144-155`
```go
func (r *Repositories) Transaction(fn func(repos *Repositories) error) error {
	if r.TransactionFn != nil {
		return r.TransactionFn(fn)
	}
	if err := r.db.Transaction(func(tx *gorm.DB) error {  // ← WithContext なし
		txRepos := NewRepositories(tx)
		return fn(txRepos)
	}); err != nil {
		return apperrors.Wrap(err, "transaction failed")
	}
	return nil
}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/service/appointment_service.go:58
err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    // ctx が正しく伝播している
})
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/repository/repositories.go:144` | `Transaction()` シグネチャ | 要修正 |
| `backend/internal/service/hospitalization_service.go:178` | 呼び出し元 | シグネチャ変更に追従 |
| `backend/internal/service/treatment_service.go:124` | 呼び出し元 | シグネチャ変更に追従 |

## 修正方針

### 1. Transaction シグネチャに ctx を追加 — `backend/internal/repository/repositories.go:144`
```go
func (r *Repositories) Transaction(ctx context.Context, fn func(repos *Repositories) error) error {
	if r.TransactionFn != nil {
		return r.TransactionFn(fn)
	}
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepos := NewRepositories(tx)
		return fn(txRepos)
	}); err != nil {
		return apperrors.Wrap(err, "transaction failed")
	}
	return nil
}
```

### 2. 呼び出し元を更新 — `backend/internal/service/hospitalization_service.go:178`
```go
err = s.repos.Transaction(ctx, func(txRepos *repository.Repositories) error {
    // ...
})
```

### 3. 呼び出し元を更新 — `backend/internal/service/treatment_service.go:124`
```go
err := s.repos.Transaction(ctx, func(txRepos *repository.Repositories) error {
    // ...
})
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `backend/CLAUDE.md` — Context伝播（必須）
> **全ての関数・メソッドの第一引数に `context.Context` を渡す。**

### `.claude/rules/go-language.md` — Context伝播
> ```go
> // Repository
> func (r *ownerRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Owner, error) {
>     return &owner, r.db.WithContext(ctx).First(&owner, "id = ?", id).Error
> }
> ```

### プロジェクト内参照実装
- `backend/internal/service/appointment_service.go:58` — `s.db.WithContext(ctx).Transaction(...)` が正しいパターン

## 優先度
**High** — コンテキストキャンセルが伝播しないため、サーバー負荷・接続枯渇リスクがある

## 関連チケット
なし

## 関連ファイル
- `backend/internal/repository/repositories.go:144-155` — 修正対象
- `backend/internal/service/hospitalization_service.go:178` — 修正対象
- `backend/internal/service/treatment_service.go:124` — 修正対象
