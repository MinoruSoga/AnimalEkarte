# BUG-310: RefreshToken モデルが完全未使用（デッドコード）

## 概要

`model/auth.go` の `RefreshToken` 構造体は DB モデルとして定義されているが、
リポジトリ・サービス・GORM 操作が一切存在しない。
実際の refresh token 処理は JWT Cookie ベース（`auth_handler.go:324`）で実装されており、
DB テーブル `refresh_tokens` もマイグレーションに存在しない。

## 現状コード

### `model/auth.go:5-16`
```go
// RefreshToken はリフレッシュトークンのDBモデル
type RefreshToken struct {
    ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID    uint64     `gorm:"not null"                 json:"user_id"`
    ClinicID  uint64     `gorm:"not null"                 json:"clinic_id"`
    TokenHash string     `gorm:"not null;uniqueIndex"     json:"-"`
    ExpiresAt time.Time  `gorm:"not null"                 json:"expires_at"`
    RevokedAt *time.Time `                                json:"revoked_at,omitempty"`
    CreatedAt time.Time  `gorm:"autoCreateTime"           json:"created_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
```

## 未使用の根拠

| チェック項目 | 結果 |
|-------------|------|
| Repository 実装 | なし（`refresh_token_repository.go` 不存在） |
| Service 参照 | ゼロ |
| Handler での GORM 操作 | ゼロ（`h.RefreshToken` は JWT Cookie ハンドラ） |
| マイグレーション | `refresh_tokens` テーブル定義なし |
| AutoMigrate | 使用なし |

## 修正方針

`model/auth.go` から `RefreshToken` 構造体と `TableName()` メソッドを削除する。

```go
// 削除対象: model/auth.go:5-16
```

注意: `auth_handler.go` の `func (h *Handler) RefreshToken(c *gin.Context)` はハンドラメソッド名であり、モデルとは無関係。削除による影響なし。

## 優先度

**Low** — 機能的影響なし。デッドコードの除去。

## 関連チケット

- BUG-279: デッドコード第1回（PasswordResetToken 削除済み、同ファイル）
- BUG-286: デッドコード第2回（RefreshToken.IsExpired 等のメソッドは削除済み。構造体本体が残存）
