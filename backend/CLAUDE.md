# Backend - Go/Gin API

## ⚠️ コマンド実行ルール

**goコマンドはローカル実行禁止。必ずDocker経由で実行すること。**

```bash
# ❌ NG
go test ./...
go mod tidy
golangci-lint run

# ✅ OK
docker compose exec backend go test ./...
docker compose exec backend go mod tidy
docker compose exec backend golangci-lint run ./...
```

## よく使うコマンド

| タスク | コマンド |
|--------|---------|
| テスト実行 | `docker compose exec backend go test ./... -v` |
| カバレッジ | `docker compose exec backend go test ./... -cover` |
| Lint | `docker compose exec backend golangci-lint run ./...` |
| モジュール更新 | `docker compose exec backend go mod tidy` |
| Swagger生成 | `docker compose exec backend swag init -g cmd/api/main.go` |

## アーキテクチャ

```
internal/
├── config/      # 設定
├── errors/      # エラー定義（センチネルエラー）
├── handler/     # HTTPハンドラ（Gin）
├── logger/      # slog構造化ログ
├── model/       # ドメインモデル
├── repository/  # データアクセス（GORM）
└── service/     # ビジネスロジック
```

## コーディング規約

### エラーハンドリング
```go
// センチネルエラー + ラッピング
if errors.Is(err, apperrors.ErrNotFound) {
    return nil, apperrors.WrapNotFound("pet", id)
}
```

### Context伝播
```go
// 全関数で第一引数にcontext.Context
func (s *Service) GetPet(ctx context.Context, id string) (*Pet, error)
```

### 構造化ログ
```go
slog.InfoContext(ctx, "pet created", slog.String("pet_id", id))
```

## テスト

- testify/assert でアサーション
- testify/mock でモック
- インターフェース経由で依存性注入

## 禁止事項

- ❌ panic の乱用
- ❌ エラー握りつぶし `_ = err`
- ❌ グローバル変数
- ❌ 未使用インポート
