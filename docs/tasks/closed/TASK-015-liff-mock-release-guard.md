# TASK-015: LIFF_MOCK が release モードでも認証をバイパスできる — 本番流出リスク

## 概要

`liff_auth.go` の `LIFF_MOCK=true` チェックが `GIN_MODE=release` を無視しているため、本番環境で環境変数を設定すれば LINE ID Token 検証を完全スキップできる。

## 優先度

CRITICAL（本番認証バイパスリスク）

## 影響ファイル

| ファイル | 行 |
|---------|-----|
| `backend/internal/middleware/liff_auth.go` | L50-62 |
| `backend/internal/config/config.go` | `Validate()` |

## 現状コード

```go
// liff_auth.go L50-62（現状）
if os.Getenv("LIFF_MOCK") == "true" {
    // LINE ID Token 検証を完全スキップ
    c.Set("line_user_id", "mock-user-id")
    c.Next()
    return
}
```

## 修正方針

```go
// liff_auth.go（修正後）
if os.Getenv("LIFF_MOCK") == "true" && os.Getenv("GIN_MODE") != "release" {
    c.Set("line_user_id", "mock-user-id")
    c.Next()
    return
}
```

```go
// config.go の Validate() に追加
if c.GinMode == "release" && os.Getenv("LIFF_MOCK") == "true" {
    return fmt.Errorf("LIFF_MOCK must not be set in release mode")
}
```

## テスト

- `GIN_MODE=release` かつ `LIFF_MOCK=true` の状態でサーバー起動が失敗することを確認
- release モードで `LIFF_MOCK` なしの場合、LINE ID Token 検証が正常に動作することを確認
