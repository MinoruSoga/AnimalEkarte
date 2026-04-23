.PHONY: up down build logs logs-api logs-front ps db clean reset migrate seed restart-api restart-front build-prod lint lint-fix test test-cover lint-front test-front build-front build-go mod-download mod-tidy help codegen codegen-check sync-modules schema-check setup-hooks ci-local

# デフォルトターゲット
.DEFAULT_GOAL := help

# 起動
up:
	docker compose down --remove-orphans 2>/dev/null || true
	docker compose up -d

# node_modules をホストにコピー（IDE補完用・初回 or package.json 変更時のみ実行）
sync-modules:
	docker compose exec -T frontend pnpm install
	docker compose cp frontend:/app/node_modules ./frontend/
	docker compose cp frontend:/app/pnpm-lock.yaml ./frontend/

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
	docker compose exec db sh -c 'psql -U $$POSTGRES_USER -d $$POSTGRES_DB'

# キャッシュクリア＆再ビルド
clean:
	docker compose down --rmi local --volumes --remove-orphans
	docker compose build --no-cache

# 完全リセット（スキーマ・シーダー含む）
# 環境ファイルから DB_USER を読み込み（デフォルト: postgres）
reset:
	@echo "🔄 Resetting database..."
	docker compose down -v
	docker compose up -d --build
	@echo "⏳ Waiting for database to be ready..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		if docker compose exec -T db pg_isready > /dev/null 2>&1; then \
			echo "✓ Database is ready"; \
			break; \
		fi; \
		if [ $$i -eq 10 ]; then echo "✗ Database startup timeout"; exit 1; fi; \
		echo "Retrying... ($$i/10)"; \
		sleep 2; \
	done
	@echo "⏳ Waiting for backend to be ready..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do \
		if docker compose exec -T backend wget -qO- http://localhost:8080/health > /dev/null 2>&1; then \
			echo "✓ Backend is ready"; \
			break; \
		fi; \
		if [ $$i -eq 15 ]; then echo "✗ Backend startup timeout"; exit 1; fi; \
		echo "Retrying... ($$i/15)"; \
		sleep 2; \
	done
	@echo "⏳ Running migrations and seeding..."
	docker compose exec -T backend go run ./cmd/migrate
	@echo "✓ Reset complete — all migrations and seed data applied"

# マイグレーション適用（差分のみ・DBは落とさない）
migrate:
	docker compose exec -T backend sh -c "go run ./cmd/migrate"
	@echo "✓ Migrations applied"

# シーダー適用（マイグレーションと同じコマンド — seed は SQL ファイルとして管理されているため差分のみ適用）
seed:
	docker compose exec -T backend sh -c "go run ./cmd/migrate"
	@echo "✓ Seed data applied"

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

# golangci-lint バージョン（CI と同一）
GOLANGCI_LINT_VERSION := v2.11.4

# リンター実行（Go）- CI と同一の公式イメージを使用
lint:
	docker run --rm \
		-v $(PWD)/backend:/app \
		-w /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) \
		golangci-lint run

# リンター実行（自動修正）
lint-fix:
	docker run --rm \
		-v $(PWD)/backend:/app \
		-w /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) \
		golangci-lint run --fix

# テスト実行（Go）
test:
	docker compose exec backend go test -race -v ./...

# テスト実行（カバレッジ付き）
test-cover:
	docker compose exec backend go test -race -cover ./...

# リンター実行（フロントエンド）
lint-front:
	docker compose exec frontend pnpm run lint

# テスト実行（フロントエンド）
test-front:
	docker compose exec frontend pnpm run test:run

# フロントエンドビルド
build-front:
	docker compose exec frontend pnpm run build

# 型定義生成（Go model → TypeScript型）
# backend/internal/model/*.go が single source of truth
codegen:
	mkdir -p frontend/src/types/generated
	docker compose run --rm codegen

# 型定義の差分チェック（CI用）
codegen-check: codegen
	git diff --exit-code frontend/src/types/generated/

# スキーマ差分チェック（GoモデルとDBの整合性検証）
schema-check:
	docker compose exec backend go test ./internal/model/ -run TestSchemaDrift -v

# Goビルド（開発用）
build-go:
	docker compose exec backend go build ./cmd/api

# Goモジュールダウンロード
mod-download:
	docker compose exec backend go mod download

# Goモジュールtidy
mod-tidy:
	docker compose exec backend go mod tidy

# CI と同等のチェックをローカル Docker で実行
# 実行前に make up でコンテナを起動しておくこと
ci-local:
	@echo "=== [1/7] Backend: build ==="
	docker compose exec backend go build ./...
	@echo "=== [2/7] Backend: test ==="
	docker compose exec backend go test ./... -count=1 -race -timeout 120s
	@echo "=== [3/7] Backend: lint ==="
	docker run --rm \
		-v $(PWD)/backend:/app \
		-w /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) \
		golangci-lint run
	@echo "=== [4/7] Backend: schema drift ==="
	docker compose exec backend go test ./internal/model/ -run TestSchemaDrift -v
	@echo "=== [5/7] Codegen: sync check ==="
	$(MAKE) codegen
	git diff --exit-code frontend/src/types/generated/ || (echo "ERROR: models.ts is out of sync. Commit the updated file." && exit 1)
	@echo "=== [6/7] Frontend: lint ==="
	docker compose exec frontend pnpm run lint
	@echo "=== [7/7] Frontend: build ==="
	docker compose exec frontend pnpm run build
	@echo ""
	@echo "✓ All CI checks passed"

# git hooks セットアップ（初回・新メンバーオンボーディング時に実行）
setup-hooks:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-commit
	@echo "Git hooks を .githooks に設定しました（pre-commit: lint + 型チェック）"

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
	@echo "  reset         完全リセット（ボリューム削除→マイグレーション＋シーダー全適用）"
	@echo "  migrate       差分マイグレーションのみ適用（DBは落とさない）"
	@echo "  seed          シーダーのみ適用（差分のみ・べき等）"
	@echo "  restart-api   API再起動"
	@echo "  restart-front フロントエンド再起動"
	@echo "  build-prod    本番ビルド"
	@echo ""
	@echo "品質管理:"
	@echo "  lint          Goリンター実行"
	@echo "  lint-fix      Goリンター実行（自動修正）"
	@echo "  test          Goテスト実行"
	@echo "  test-cover    Goテスト実行（カバレッジ付き）"
	@echo "  lint-front    フロントエンドリンター実行"
	@echo "  test-front    フロントエンドテスト実行"
	@echo "  build-front   フロントエンドビルド"
	@echo "  codegen       型定義生成（Go model → TypeScript型）"
	@echo "  codegen-check 型定義の差分チェック（CI用）"
	@echo "  schema-check  GoモデルとDBスキーマの差分チェック"
	@echo "  ci-local      CI と同等のチェックをローカル Docker で実行"
	@echo "  build-go      Goビルド（開発用）"
	@echo "  mod-download  Goモジュールダウンロード"
	@echo "  mod-tidy      Goモジュールtidy"
	@echo "  sync-modules  node_modulesをホストにコピー（IDE補完用）"
	@echo "  setup-hooks   git hooksをセットアップ（初回・新メンバー用）"
	@echo ""
	@echo "  help          このヘルプを表示"
