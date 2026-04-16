# BUG-319: Handler 層での slog 使用 — アーキテクチャ規約違反

## 概要

`.claude/CLAUDE.md` のアーキテクチャ規約「slog は Service 層のみ」に反し、
Handler 層（3ファイル）で `slog` が直接使用されている。
ただし `response.go` の使用は例外的インフラ用途として別途評価が必要。

## 再現手順

1. 下記ファイルの該当行を参照
2. Handler 層で `slog.ErrorContext` / `slog.WarnContext` を直接呼び出している

## 期待する動作

- Handler 層では `slog` を直接使用しない
- ログ出力はすべて Service 層に委譲する
- 例外: `response.go` の内部サーバーエラーログは運用上の必要性を評価の上、判断する

## 現状コード

### `backend/internal/handler/permission_group_handler.go:86,137,174,271`
```go
// 監査ログ書き込み失敗時の slog（4箇所）
slog.ErrorContext(c.Request.Context(), "failed to log permission group creation", slog.String("error", auditErr.Error()))
slog.ErrorContext(c.Request.Context(), "failed to log permission group update", slog.String("error", auditErr.Error()))
slog.ErrorContext(c.Request.Context(), "failed to log permission group deletion", slog.String("error", auditErr.Error()))
slog.ErrorContext(c.Request.Context(), "failed to log permission rules update", slog.String("error", auditErr.Error()))
```

### `backend/internal/handler/auth_handler.go:114,133,262,289`
```go
// 監査ログ書き込み失敗時の slog（4箇所）
slog.ErrorContext(ctx, "failed to log login failure", slog.String("error", auditErr.Error()))
slog.ErrorContext(ctx, "failed to log login failure", slog.String("error", auditErr.Error()))
slog.ErrorContext(ctx, "failed to log login success", slog.String("error", auditErr.Error()))
slog.ErrorContext(ctx, "failed to log logout", slog.String("error", auditErr.Error()))
```

### `backend/internal/handler/medical_record_image_handler.go:234`
```go
// ファイルクリーンアップ失敗時の slog
slog.WarnContext(c.Request.Context(), "failed to clean up uploaded file", "key", key, "error", removeErr)
```

### `backend/internal/handler/response.go:67-70`（評価が必要な例外）
```go
// RespondError での内部サーバーエラーログ
slog.ErrorContext(c.Request.Context(), "internal server error",
    slog.String("error", err.Error()),
    slog.String("path", c.FullPath()),
    slog.String("method", c.Request.Method))
```

### 比較: 正しい実装（Service 層でのみ slog 使用）
```go
// backend/internal/service/permission_group_service.go:65
slog.Uint64("clinic_id", group.ClinicID)

// backend/internal/service/shift_entry_service.go — Service 層でのみログ
```

## 影響範囲

| 対象 | 行番号 | 内容 | 状態 |
|------|--------|------|------|
| `backend/internal/handler/permission_group_handler.go` | 86,137,174,271 | 監査ログ失敗時のエラーログ（4箇所） | 未修正 |
| `backend/internal/handler/auth_handler.go` | 114,133,262,289 | 監査ログ失敗時のエラーログ（4箇所） | 未修正 |
| `backend/internal/handler/medical_record_image_handler.go` | 234 | ファイルクリーンアップ失敗 | 未修正 |
| `backend/internal/handler/response.go` | 67-70 | 内部サーバーエラーログ（例外評価が必要） | 評価待ち |

## 修正方針

### 監査ログ失敗 (permission_group_handler.go, auth_handler.go)

監査ログ書き込みは Service 層で行われており、その失敗検知も Service 層に移動すべき。

**案A**: 監査ログ失敗をエラーとして Service から返し、Handler が `RespondError` に委譲
```go
// Service 層
if auditErr := s.audit.Log(ctx, event); auditErr != nil {
    // 監査ログ失敗はビジネス的に警告レベルなので slog はここで出す
    slog.WarnContext(ctx, "failed to write audit log", "error", auditErr)
}
```

**案B**: 監査ログ失敗を Service 層でラップして slog（現状の発火点を Service に移動）

### ファイルクリーンアップ失敗 (medical_record_image_handler.go:234)

クリーンアップロジックを Service 層に移動し、slog も Service 層で出力する。

### response.go:67-70

`RespondError` は HTTP インフラのクロスカッティング関心事であり、
Service 層から HTTP コンテキスト（`c.FullPath()`, `c.Request.Method`）にアクセスできないため、
**例外として許容するか否かを設計者が判断する**。
既存の `liff_handler.go` 例外と同様の扱いを検討すること。

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — slog ルール
> **slog rule**: 構造化ログ `log/slog` を使用し、`InfoContext`, `ErrorContext` でコンテキストを適切に伝播させる。
> （`.claude/CLAUDE.md`では「Service 層のみ」と明記）

### `.claude/rules/go-language.md` — ログは Service 層のみ
> **ログ（slog構造化ログ）**: service層のみ。handler・repositoryには記述しない。

### プロジェクト内参照実装
- `backend/internal/service/permission_group_service.go:65,80,119` — Service 層での正しい slog 使用
- `backend/internal/service/shift_entry_service.go` — Service 層での正しい slog 使用

## 優先度

**Medium** — アーキテクチャ規約違反だが、セキュリティや機能への直接的実害はない。
ただし監査ログ失敗のサイレント処理は運用上の問題になりうる。

## 関連チケット

なし

## 関連ファイル

- `backend/internal/handler/permission_group_handler.go:86,137,174,271`
- `backend/internal/handler/auth_handler.go:114,133,262,289`
- `backend/internal/handler/medical_record_image_handler.go:234`
- `backend/internal/handler/response.go:67-70`
- `backend/internal/rules/go-language.md` — slog ルール定義
