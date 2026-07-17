# Animal Ekarte Backend

動物病院 電子カルテシステムのバックエンドAPI

## 技術スタック

- **言語**: Go 1.25
- **フレームワーク**: Gin v1.12
- **ORM**: GORM v1.30
- **データベース**: PostgreSQL 18
- **API仕様**: OpenAPI 3.0（手動管理: `docs/api.yaml`）
- **ホットリロード**: Air
- **ロギング**: slog（構造化ログ）
- **リンター**: golangci-lint

## ディレクトリ構成

```
backend/
├── cmd/
│   ├── api/               # メインAPIサーバー エントリーポイント
│   ├── migrate/           # SQLマイグレーション適用
│   ├── coverage-ratchet/  # カバレッジ低下防止（BE-refactor.md R3-5）
│   ├── seed-export/       # seed CSV エクスポート（migrations/seeds/ 用）
│   ├── seed-old-db/       # 旧DB TSV ローカル投入（開発専用）
│   ├── stage-import/      # 旧DB移行データの本テーブル取り込み
│   └── lstep-migrate/     # Lステップ連携データ移行
├── internal/              # 内部パッケージ（外部からimport不可）
│   ├── apicontract/       # OpenAPI (docs/api.yaml) とルート実装の整合性テスト
│   ├── config/            # 環境変数・設定読み込み
│   ├── dbconn/            # DB接続確立
│   ├── errors/            # apperrors（センチネルエラー・FromGORM・Wrap）
│   ├── handler/           # HTTPハンドラ（プレゼンテーション層）
│   ├── infra/             # 外部インフラ連携（ファイルストレージ、LINE、暗号化等）
│   ├── logger/            # slog構造化ロガー
│   ├── middleware/        # ミドルウェア（認証、CORS等）
│   ├── model/             # データモデル・リクエスト型
│   ├── repository/        # データアクセス層
│   ├── seedbundle/        # seedデータのバンドル定義
│   └── service/           # ビジネスロジック層
├── migrations/            # DBマイグレーション（SQL、番号付き）
├── docs/                  # APIドキュメント（api.yaml、手動管理）
├── .golangci.yml          # リンター設定
├── Dockerfile.dev         # 開発用（docker compose が参照）
├── Dockerfile.production  # 本番用（ECS + Cloudflare Container デプロイが参照）
├── entrypoint.sh          # コンテナ起動スクリプト
├── .air.toml              # ホットリロード設定
├── go.mod                 # 依存関係
└── go.sum                 # 依存関係ロック
```

各層の詳細な実装ルールは `internal/handler/CLAUDE.md`、`internal/service/CLAUDE.md`、`internal/repository/CLAUDE.md` を参照。

## 開発環境セットアップ

### 前提条件

- Go 1.25以上
- PostgreSQL 18（またはDocker）

### ローカル開発（Docker使用）

プロジェクトルートから:

```bash
# コンテナ起動
make build

# ログ確認
make logs-api

# API再起動
make restart-api
```

### ローカル開発（Docker不使用）

```bash
# 依存関係インストール
go mod download

# 環境変数設定
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=ekarte_user
export DB_PASSWORD=ekarte_password
export DB_NAME=ekarte_db

# 実行
go run cmd/api/main.go

# または Air でホットリロード
go install github.com/air-verse/air@latest
air -c .air.toml
```

## API エンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | /health | ヘルスチェック |
| GET | /api/v1/ | ウェルカムメッセージ |
| GET | /api/v1/pets | ペット一覧取得 |
| GET | /api/v1/pets/:id | ペット詳細取得 |
| POST | /api/v1/pets | ペット新規登録 |
| PATCH | /api/v1/pets/:id | ペット情報更新 |
| DELETE | /api/v1/pets/:id | ペット削除 |

上記は代表例。全エンドポイントは `docs/api.yaml`（OpenAPI 3.0、手動管理）を参照。

## API ドキュメント

API 仕様は `docs/api.yaml`（OpenAPI 3.0）で手動管理。

## 新しいCRUD機能の追加手順

Handler → Service → Repository の各層の実装規約（エラーハンドリング、レスポンス形式、
`RequirePermission` によるルート保護、命名規則等）は `CODING_RULES.md` と各層の
`internal/handler/CLAUDE.md` / `internal/service/CLAUDE.md` / `internal/repository/CLAUDE.md`
（P1–P18 準拠チェックリスト）を参照。手順のコード例をここに重複記載しない
（重複した手順ドキュメントは二重管理となり `docs/product-philosophy.md` の原則に反する）。

概略の流れ:

1. `internal/model/` にモデルを追加（主キーは `uint64` autoincrement、`ClinicID` を含める）
2. `migrations/` に SQL マイグレーションファイルを追加（`AutoMigrate` は使用しない）
3. `internal/repository/` にリポジトリを追加（`apperrors.FromGORM` でエラー変換）
4. `internal/service/` にビジネスロジックを追加（`apperrors.Wrap` でエラーラップ）
5. `internal/handler/` にハンドラーを追加（`RespondError` でエラー応答、`toXxxResponse()` でレスポンス変換）
6. ルート登録時に全ルートへ `RequirePermission` を付与
7. `docs/api.yaml` を更新し、`make codegen` で型生成

## テスト

```bash
# 全テスト実行
go test ./...

# カバレッジ付き
go test -cover ./...

# 特定パッケージ
go test ./internal/service/...
```

## 本番ビルド

```bash
# Dockerイメージビルド
docker build -t animal-ekarte-api:latest .

# バイナリビルド
CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api
```

## 環境変数

| 変数名 | 説明 | デフォルト値 |
|-------|------|-------------|
| PORT | APIポート | 8080 |
| DB_HOST | DBホスト | localhost |
| DB_PORT | DBポート | 5432 |
| DB_USER | DBユーザー | ekarte_user |
| DB_PASSWORD | DBパスワード | ekarte_password |
| DB_NAME | DB名 | ekarte_db |
| GIN_MODE | Ginモード | debug |
| LOG_LEVEL | ログレベル | info |

## コーディングパターン

### エラーハンドリング

センチネルエラーとエラーラッピングを使用:

```go
import apperrors "github.com/animal-ekarte/backend/internal/errors"

// エラー発生時はラッピング
if err != nil {
    return nil, apperrors.Wrap(err, "failed to get pet")
}

// NotFound エラー（Repository層は apperrors.FromGORM を使用）
return nil, apperrors.WrapNotFound("pet", fmt.Sprintf("%d", id))

// バリデーションエラー
return nil, apperrors.WrapInvalidInput("pet name is required")

// Handler層はエラーの種類を問わず RespondError に委譲する
RespondError(c, err)
```

### 構造化ログ（slog）

```go
import "log/slog"

// 基本的なログ
slog.Info("server starting", slog.String("port", cfg.Port))

// コンテキスト付きログ
slog.InfoContext(ctx, "pet created", slog.Uint64("pet_id", pet.ID))
slog.ErrorContext(ctx, "failed to create pet", slog.String("error", err.Error()))
```

### Context伝播

全てのレイヤーで `context.Context` を第一引数に:

```go
// Service層
func (s *Service) GetPetByID(ctx context.Context, id string) (*model.Pet, error)

// Repository層
func (r *Repository) GetPetByID(ctx context.Context, id uint64) (*model.Pet, error) {
    var pet model.Pet
    result := r.db.WithContext(ctx).First(&pet, "id = ?", id)
    return &pet, apperrors.FromGORM(result.Error, "pet", fmt.Sprintf("%d", id))
}
```

## リンター

golangci-lint でコード品質をチェック:

```bash
# インストール
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 実行
golangci-lint run

# 自動修正
golangci-lint run --fix
```
