---
name: docker-patterns
description: Docker Compose開発環境パターン。このプロジェクトのDocker運用ルール・トラブルシューティング・最適化。Docker設定変更・開発環境問題解決時に使用。
origin: ECC (adapted for AnimalEkarte)
---

# Docker パターン

このプロジェクトの Docker Compose 開発環境の標準パターン。

## When to Activate

- Docker Compose 設定の変更
- コンテナ起動・停止の問題解決
- ボリューム・ネットワーク問題
- Dockerfile の最適化
- 開発環境のセットアップ

## ⚠️ 最重要ルール: Docker 経由で実行

```bash
# ❌ ローカル直接実行（禁止）
pnpm build
go test ./internal/service/...

# ✅ Docker Compose 経由（必須）
docker compose exec frontend pnpm build   # ※全体 build 自体も自動実行禁止 — ユーザーに依頼
docker compose exec backend go test ./internal/service/...
```

## よく使うコマンド

```bash
# 環境起動・停止
make up          # docker compose up -d
make down        # docker compose down
make logs        # docker compose logs -f
# ⚠️ 上3つは CLAUDE.md の自動実行禁止コマンド（docker compose up/down/logs）。ユーザーに手動実行を依頼する

# Frontend
docker compose exec frontend pnpm lint
docker compose exec frontend pnpm test:run
docker compose exec frontend pnpm build
docker compose exec frontend npx tsc --noEmit
# ⚠️ 上4つは CLAUDE.md の自動実行禁止コマンド。ユーザーに手動実行を依頼するか、
#    スコープ限定版（`npx eslint <ファイル>` / `npx vitest run <spec>`）を使う

# Backend
docker compose exec backend go test ./...
# ⚠️ CLAUDE.md の自動実行禁止コマンド。ユーザーに手動実行を依頼するか、スコープ限定版（例: `go test ./internal/service/...`）を使う
docker compose exec backend go vet ./...
docker compose exec backend golangci-lint run ./...
# ⚠️ CLAUDE.md の自動実行禁止コマンド。ユーザーに手動実行を依頼するか、スコープ限定版を使う

# Database
make db  # = docker compose exec db sh -c 'psql -U $POSTGRES_USER -d $POSTGRES_DB'
# ⚠️ CLAUDE.md の自動実行禁止コマンド。ユーザーに手動実行を依頼するか、スコープ限定版を使う

# CodeGen（Go モデル → TypeScript 型）
make codegen
# ⚠️ CLAUDE.md の自動実行禁止コマンド。ユーザーに手動実行を依頼する
```

## ヘルスチェック・トラブルシューティング

```bash
# コンテナ状態確認
docker compose ps

# ログ確認
docker compose logs frontend --tail=50
docker compose logs backend --tail=50
docker compose logs db --tail=50

# コンテナに入って調査
docker compose exec frontend sh
docker compose exec backend sh

# ネットワーク確認
docker network ls
docker network inspect ekarte-network

# ボリューム確認
docker volume ls
```

## よくある問題と解決方法

### node_modules がコンテナにない

まず `docker compose ps` と frontend の mount 設定を確認し、依存定義と lockfile が整合しているかを調べる。通常の調査・復旧で volume を削除したり、stack を停止したりしない。依存の再取得が必要なら、CLAUDE.md の自動実行禁止に従い、ユーザーに対象コンテナでの手動実行を依頼する。

volume を消す操作は node_modules だけでなく DB を含む永続データも失わせ得る。どうしても再現検証が必要な場合は、既存・共有環境ではなく disposable な隔離 Compose project / volume を用意し、対象とデータ消失影響を示したユーザーの明示承認後にだけ実施する。

### ポート競合
```bash
# 使用中のポートを確認
lsof -i :3003
lsof -i :8080
lsof -i :5434  # db はホスト 5434 → コンテナ 5432 にマップ

# 停止後に再起動
make down
make up
```

### DB 接続エラー
```bash
# DB コンテナのログを確認
docker compose logs db

# DB が起動しているか確認
docker compose exec db pg_isready

# 接続テスト
docker compose exec db pg_isready -U "$DB_USER"
```

### 変更が反映されない
```bash
# フロントエンドは Hot Reload が有効なはず
# バックエンドは Air で自動リロード

# まず対象サービスの設定・mount・build context を確認する
docker compose config --services
# build cache が原因と特定できた場合だけ、隔離環境で no-cache build を検討する
```

`make clean` は Compose の volume を含む環境を削除してから再ビルドする破壊的な target であり、通常の cache/node_modules 復旧には使わない。実施するなら disposable な隔離環境に限定し、削除対象とデータ消失影響を説明したうえでユーザーの明示承認を得る。

## Dockerfile ベストプラクティス

### マルチステージビルド

```dockerfile
# Backend
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download      # 依存を先にキャッシュ
COPY . .
RUN CGO_ENABLED=0 go build -o app ./cmd/api

FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/app .
# Non-root ユーザー
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
USER appuser
EXPOSE 8080
CMD ["./app"]
```

```dockerfile
# Frontend（本番）
FROM node:24-alpine AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile                # 依存を先にキャッシュ
COPY . .
RUN pnpm build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
```

### レイヤーキャッシュ最適化
```dockerfile
# ✅ 変更頻度が低いものを先に
COPY go.mod go.sum ./    # 依存（変更少ない）
RUN go mod download
COPY . .                 # ソースコード（変更多い）
RUN go build ...

# ❌ 変更頻度が高いものを先にすると毎回フルビルド
COPY . .
RUN go mod download
RUN go build ...
```

## セキュリティ

```dockerfile
# ✅ Non-root ユーザー（本番必須）
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
USER appuser

# ✅ 最小イメージ（attack surface 削減）
FROM alpine:3.21  # debian/ubuntu より小さい

# ✅ .dockerignore で機密ファイル除外
.env
.env.*
*.pem
*.key
node_modules
.git
```

## make コマンド一覧

```bash
make up       # 全サービス起動
make down     # 全サービス停止
make logs     # ログ表示（follow）
make db       # DB接続（psql）。マイグレーションは make migrate
make codegen  # Go モデル → TS 型生成
make clean    # ⚠️ volume を含む環境を削除する破壊的 target。通常復旧には使わず、隔離環境・影響説明・ユーザー明示承認が必要
make test-front  # フロントエンドテスト
make lint-front  # フロントエンド lint
```

⚠️ `make db` は CLAUDE.md の自動実行禁止コマンドに該当する（DB リセット等の高影響操作）。ユーザーに手動実行を依頼するか、スコープ限定版を使う
