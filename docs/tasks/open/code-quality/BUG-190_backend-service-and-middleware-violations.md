# BUG-190: Backend service/middleware 規約違反（5件）

| 項目 | 内容 |
|------|------|
| 優先度 | **High** |
| カテゴリ | エラーハンドリング / 並行処理 / アーキテクチャ |

## 1. permission_group_service.go — 裸の return err（2箇所）

### `backend/internal/service/permission_group_service.go:85`
```go
func (s *permissionGroupService) Delete(ctx context.Context, id uint64) error {
    // ...
    if err := s.repo.Delete(ctx, id); err != nil {
        return err  // apperrors.Wrap なし
    }
```

### `backend/internal/service/permission_group_service.go:104`
```go
func (s *permissionGroupService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
    if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
        return err  // apperrors.Wrap なし
    }
```

修正:
```go
return apperrors.Wrap(err, "failed to delete permission group")
return apperrors.Wrap(err, "failed to reorder permission groups")
```

### 参照実装: 同ファイル :79
```go
return apperrors.Wrap(err, "failed to find permission group")
```

## 2. rate_limit.go — cleanupLoop に context キャンセルなし

### `backend/internal/middleware/rate_limit.go:30-41`
```go
go s.cleanupLoop()  // context なしで goroutine 起動

func (s *RateLimitStore) cleanupLoop() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {  // 永遠に止まらない
        s.evict(10 * time.Minute)
    }
}
```

修正:
```go
func NewRateLimitStore(ctx context.Context) *RateLimitStore {
    s := &RateLimitStore{limiters: make(map[string]*limiterEntry)}
    go s.cleanupLoop(ctx)
    return s
}

func (s *RateLimitStore) cleanupLoop(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            s.evict(10 * time.Minute)
        }
    }
}
```

## 3. errors.go — WrapConflict が ErrAlreadyExists を使用

### `backend/internal/errors/errors.go:88-94`
```go
func WrapConflict(message string) error {
    return &AppError{
        Code:    "CONFLICT",
        Message: message,
        Err:     ErrAlreadyExists,  // ← ErrConflict が存在しない
    }
}
```

「重複エラー」と「FK参照によるブロック」が同一センチネルに混在し、handler 側での分岐が不能。

## 4. auth_handler.go — handler 層に slog 11箇所

規約「slog は service 層のみ」に違反。

| 行 | 内容 |
|-----|------|
| 127 | `slog.WarnContext` (password mismatch) |
| 133-136 | `slog.InfoContext` (account found) |
| 140 | `slog.ErrorContext` (failed to find staff) |
| 147 | `slog.WarnContext` (inactive staff) |
| 155 | `slog.ErrorContext` (clinic assignments) |
| 248 | `slog.WarnContext` (failed to list clinics) |
| 259 | `slog.InfoContext` (login successful) |
| 305 | `slog.InfoContext` (logout) |
| 428 | `slog.InfoContext` (token refreshed) |
| 489 | `slog.InfoContext` (password changed) |
| 583-585 | `slog.ErrorContext` (permission calculation) |

## 5. DB スキーマ — billing_items に updated_at なし / payments に deleted_at なし

### `backend/migrations/001_init.sql`
- `billing_items`: `created_at` と `deleted_at` はあるが `updated_at` がない。GORM の UpdatedAt が機能しない。
- `payments`: `deleted_at` がない。財務テーブルの物理削除が可能。

## 準拠すべきプロジェクト規約

- `.claude/rules/go-language.md`: エラーラッピング（`apperrors.Wrap`）
- `.claude/rules/go-language.md` §5: 並行処理（errgroup / context）
- `.claude/CLAUDE.md`: slog は service 層のみ
- `.claude/rules/database-design.md`: 全テーブルに updated_at, deleted_at

## 優先度
**High** — cleanupLoop の goroutine リークとエラーハンドリングの不統一。
