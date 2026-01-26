.PHONY: up down build logs logs-api logs-front ps db clean reset restart-api restart-front build-prod lint lint-fix test test-cover swagger build-go mod-download mod-tidy help

# デフォルトターゲット
.DEFAULT_GOAL := help

# 起動
up:
	rm -rf frontend/node_modules
	docker compose up -d
	docker compose cp frontend:/app/node_modules ./frontend/
	docker compose cp frontend:/app/package-lock.json ./frontend/

# 起動（ビルド付き）
build:
	docker compose up -d --build

# 停止
down:
	docker compose down

# ログ表示（全体）
logs:
	docker compose logs -f

# ログ表示（API）
logs-api:
	docker compose logs -f backend

# ログ表示（フロントエンド）
logs-front:
	docker compose logs -f frontend

# コンテナ状態確認
ps:
	docker compose ps

# DB接続
db:
	docker compose exec db psql -U ekarte_user -d ekarte_db

# キャッシュクリア＆再ビルド
clean:
	docker compose down --rmi local --volumes --remove-orphans
	docker compose build --no-cache

# 完全リセット（データ含む）
reset:
	docker compose down -v
	docker compose up -d --build

# バックエンドのみ再起動
restart-api:
	docker compose restart backend

# フロントエンドのみ再起動
restart-front:
	docker compose restart frontend

# 本番ビルド
build-prod:
	docker build -t animal-ekarte-api:latest ./backend
	docker build -t animal-ekarte-front:latest ./frontend

# リンター実行（Go）
lint:
	docker compose exec backend golangci-lint run

# リンター実行（自動修正）
lint-fix:
	docker compose exec backend golangci-lint run --fix

# テスト実行（Go）
test:
	docker compose exec backend go test -v ./...

# テスト実行（カバレッジ付き）
test-cover:
	docker compose exec backend go test -cover ./...

# Swagger ドキュメント生成
swagger:
	docker compose exec backend swag init -g cmd/api/main.go -o docs

# Goビルド（開発用）
build-go:
	docker compose exec backend go build ./cmd/api

# Goモジュールダウンロード
mod-download:
	docker compose exec backend go mod download

# Goモジュールtidy
mod-tidy:
	docker compose exec backend go mod tidy

# ヘルプ
help:
	@echo "Animal Ekarte - 開発コマンド"
	@echo ""
	@echo "使用方法: make [コマンド]"
	@echo ""
	@echo "コマンド:"
	@echo "  up            コンテナ起動"
	@echo "  build         コンテナ起動（ビルド付き）"
	@echo "  down          コンテナ停止"
	@echo "  logs          全ログ表示"
	@echo "  logs-api      APIログ表示"
	@echo "  logs-front    フロントエンドログ表示"
	@echo "  ps            コンテナ状態確認"
	@echo "  db            DB接続（psql）"
	@echo "  clean         キャッシュクリア＆再ビルド"
	@echo "  reset         完全リセット（データ削除）"
	@echo "  restart-api   API再起動"
	@echo "  restart-front フロントエンド再起動"
	@echo "  build-prod    本番ビルド"
	@echo ""
	@echo "品質管理:"
	@echo "  lint          Goリンター実行"
	@echo "  lint-fix      Goリンター実行（自動修正）"
	@echo "  test          Goテスト実行"
	@echo "  test-cover    Goテスト実行（カバレッジ付き）"
	@echo "  swagger       Swaggerドキュメント生成"
	@echo "  build-go      Goビルド（開発用）"
	@echo "  mod-download  Goモジュールダウンロード"
	@echo "  mod-tidy      Goモジュールtidy"
	@echo ""
	@echo "  help          このヘルプを表示"
