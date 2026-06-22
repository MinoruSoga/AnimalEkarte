.PHONY: up down build logs logs-api logs-front ps db clean reset migrate seed restart-api restart-front build-prod lint lint-fix test test-cover lint-front test-front build-front build-go mod-download mod-tidy help codegen codegen-check sync-modules schema-check setup-hooks ci-local dump-stg

# デフォルトターゲット
.DEFAULT_GOAL := help

# $(DC) に --env-file を渡す（.env.local を変数展開の source of truth にする）
DC = docker compose --env-file .env.local

# 起動
# --wait で db / migrate / backend / frontend の ready を待つ
up:
	$(DC) down --remove-orphans 2>/dev/null || true
	$(DC) up -d --wait --wait-timeout 1200 db migrate backend frontend

# node_modules をホストにコピー（IDE補完用・初回 or package.json 変更時のみ実行）
sync-modules:
	$(DC) exec -T frontend pnpm install
	$(DC) cp frontend:/app/node_modules ./frontend/
	$(DC) cp frontend:/app/pnpm-lock.yaml ./frontend/

# 起動（ビルド付き）
build:
	$(DC) up -d --build

# 停止
down:
	$(DC) down

# ログ表示（全体）
logs:
	$(DC) logs -f

# ログ表示（API）
logs-api:
	$(DC) logs -f backend

# ログ表示（フロントエンド）
logs-front:
	$(DC) logs -f frontend

# コンテナ状態確認
ps:
	$(DC) ps

# DB接続
db:
	$(DC) exec db sh -c 'psql -U $$POSTGRES_USER -d $$POSTGRES_DB'

# キャッシュクリア＆再ビルド
clean:
	$(DC) down --rmi local --volumes --remove-orphans
	$(DC) build --no-cache

# 完全リセット（スキーマ・シーダー含む）
# migration は compose の one-shot service に任せ、reset は DB 初期化 + 起動完了待ちだけに絞る
reset:
	@echo "🔄 Resetting database..."
	$(DC) down -v
	$(DC) up -d --build --wait --wait-timeout 1200
	@echo "✓ Reset complete — database reinitialized and services are healthy"

# マイグレーション適用（差分のみ・DBは落とさない）
# 使うのは backend ではなく one-shot の migrate サービス
migrate:
	$(DC) run --rm migrate
	@echo "✓ Migrations applied"

# シーダー適用（migrate と同一処理。SQL seed も migration サービスで一括適用）
seed:
	$(DC) run --rm migrate
	@echo "✓ Seed data applied"

# バックエンドのみ再起動
restart-api:
	$(DC) restart backend

# フロントエンドのみ再起動
restart-front:
	$(DC) restart frontend

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
	$(DC) exec backend go test -race -v ./...

# テスト実行（カバレッジ付き）
test-cover:
	$(DC) exec backend go test -race -cover ./...

# リンター実行（フロントエンド）
lint-front:
	$(DC) exec frontend pnpm run lint

# テスト実行（フロントエンド）
test-front:
	$(DC) exec frontend pnpm run test:run

# フロントエンドビルド
build-front:
	$(DC) exec frontend pnpm run build

# 型定義生成（Go model → TypeScript型）
# backend/internal/model/*.go が single source of truth
codegen:
	mkdir -p frontend/src/types/generated
	$(DC) run --rm codegen

# 型定義の差分チェック（CI用）
codegen-check: codegen
	git diff --exit-code frontend/src/types/generated/

# スキーマ差分チェック（GoモデルとDBの整合性検証）
schema-check:
	$(DC) exec backend go test ./internal/model/ -run TestSchemaDrift -v

# Goビルド（開発用）
build-go:
	$(DC) exec backend go build ./cmd/api

# Goモジュールダウンロード
mod-download:
	$(DC) exec backend go mod download

# Goモジュールtidy
mod-tidy:
	$(DC) exec backend go mod tidy

# CI と同等のチェックをローカル Docker で実行
# 実行前に make up でコンテナを起動しておくこと
ci-local:
	@echo "=== [1/7] Backend: build ==="
	$(DC) exec backend go build ./...
	@echo "=== [2/7] Backend: test ==="
	$(DC) exec backend go test ./... -count=1 -race -timeout 120s
	@echo "=== [3/7] Backend: lint ==="
	docker run --rm \
		-v $(PWD)/backend:/app \
		-w /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) \
		golangci-lint run
	@echo "=== [4/7] Backend: schema drift ==="
	$(DC) exec backend go test ./internal/model/ -run TestSchemaDrift -v
	@echo "=== [5/7] Codegen: sync check ==="
	$(MAKE) codegen
	git diff --exit-code frontend/src/types/generated/ || (echo "ERROR: models.ts is out of sync. Commit the updated file." && exit 1)
	@echo "=== [6/7] Frontend: lint ==="
	$(DC) exec frontend pnpm run lint
	@echo "=== [7/7] Frontend: build ==="
	$(DC) exec frontend pnpm run build
	@echo ""
	@echo "✓ All CI checks passed"

# git hooks セットアップ（初回・新メンバーオンボーディング時に実行）
setup-hooks:
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-commit
	@echo "Git hooks を .githooks に設定しました（pre-commit: lint + 型チェック）"

# STG DB ダンプ（SSM ポートフォワード経由 pg_dump → prodData/ekarte-stg-<実行日>.sql）
# 前提: AWS プロファイル(AnimalEkarte)認証済み。DB 認証は既定で .env.staging の
#       DB_USER / DB_NAME / DB_PASSWORD を使用（PGPASSWORD 等の env で上書き可）。
# prodData/ は .gitignore 済。
dump-stg:
	@bash scripts/dump-stg.sh

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
	@echo "  dump-stg      STG DBダンプ(SSM経由 pg_dump → prodData/ekarte-stg-<実行日>.sql)"
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
