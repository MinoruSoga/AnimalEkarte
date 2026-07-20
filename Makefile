.PHONY: up down build logs logs-api logs-front ps db clean reset migrate seed stage-import-dry-run stage-import verify-stage-import stage-import-rollback-test restart-api restart-front build-prod lint lint-fix test test-cover lint-front test-front build-front e2e build-go mod-download mod-tidy help codegen codegen-check sync-modules schema-check setup-hooks ci dump-stg check-reset-contract check-reset-contract-test shellcheck shellcheck-test

# デフォルトターゲット
.DEFAULT_GOAL := help

# $(DC) に --env-file を渡す（.env.local を変数展開の source of truth にする）
DC = docker compose --env-file .env.local

# 起動
# --wait で db / backend / frontend の ready を待つ
# migration は backend の entrypoint 内で go run ./cmd/migrate として実行されるため、
# 個別の migrate サービスは存在しない（失敗時は backend が healthy にならず --wait がブロックする）。
up:
	$(DC) down --remove-orphans 2>/dev/null || true
	$(DC) up -d --wait --wait-timeout 1200 db backend frontend

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
# migration は backend の entrypoint 内で go run ./cmd/migrate として実行されるため、
# reset は DB 初期化 + 起動完了待ちだけに絞る
# --wait は up と同じく長寿命サービス（db backend frontend）だけを対象にする。
# codegen は一発実行で正常終了する one-shot のため、wait 対象に含めると正常終了が
# --wait の失敗扱いになり cosmetic exit 1 を起こす（必要時は make codegen で個別実行）。
reset:
	@echo "🔄 Resetting database..."
	$(DC) down -v
	$(DC) up -d --build --wait --wait-timeout 1200 db backend frontend
	@echo "✓ Reset complete — database reinitialized and services are healthy"

# reset の wait-set 契約チェック（Docker 不要・純テキスト検査・高速）
# `make reset` の `up --wait` が長寿命サービス (db backend frontend) だけを
# 待ち、one-shot codegen を含めないことを静的に保証する。これが裸の `up --wait` に
# 退行すると cosmetic exit-1 が再発するため、make ci で自動実行する。
check-reset-contract:
	@bash scripts/check-reset-wait-services.sh

# 上記契約チェック自体の回帰テスト（fixture ベース・Docker 不要）
# 正しい wait-set は通し、codegen 混入 / 必須欠落 / 裸 up --wait を reject できることを検証する。
check-reset-contract-test:
	@bash scripts/check-reset-wait-services.test.sh

# scripts/*.sh の shellcheck ゲート（severity=warning）。
# shellcheck はローカルに無ければピン留め Docker イメージ経由で実行する（再現可能）。
# 整形・行継続トリックでは欺けない AST 検査で、シェルスクリプトの退行を手動レビュー
# ではなく自動で弾く。一括/ファイル全体スコープの disable によるゲート骨抜きも reject する。
shellcheck:
	@bash scripts/shellcheck-scripts.sh

# 上記 shellcheck ゲート自体の回帰テスト（fixture ベース）。
# 実バグ・行継続で隠したバグ・disable=all / ファイル全体 disable を確実に reject し、
# clean スクリプトと正当な行内 disable は通すことを検証する。
shellcheck-test:
	@bash scripts/shellcheck-scripts.test.sh

# マイグレーション適用（差分のみ・DBは落とさない）
# 専用の migrate サービスは廃止し、backend イメージの entrypoint を go に差し替えて
# one-off 実行する（db が未起動なら depends_on 経由で自動起動される）。
migrate:
	$(DC) run --rm --entrypoint go backend run ./cmd/migrate
	@echo "✓ Migrations applied"

# シーダー適用（migrate と同一処理。SQL seed も migration ファイルとして一括適用）
seed:
	$(DC) run --rm --entrypoint go backend run ./cmd/migrate
	@echo "✓ Seed data applied"

# ============================================================================
# stage-import: animalekarte_stage -> 本テーブル (推奨経路 / replaces seed-old-db(archived))
# ============================================================================
# 検証済みの old_db 3層パイプライン (legacy_raw -> legacy_canonical ->
# animalekarte_stage) の stage スキーマを唯一の投入元として本テーブルへ取り込む。
# 旧 direct seeder (seed-old-db、backend/cmd/_archive/seed-old-db へアーカイブ済み) は
# comparison-only として deprecated。
#
# 前提:
#   - AnimalEkarte: make up でスタック起動済み (db healthy)。
#   - old_db: 別 repo で make local-postgres-up + make migration-pipeline 実行済み
#     (old-db-postgres コンテナと外部ネットワーク old_db_default が存在すること)。
#   - OLD_DB_POSTGRES_PASSWORD: old_db Postgres の TCP 接続パスワード。stage への
#     接続は read-only。未設定なら importer は SASL 認証で失敗する。
#
# Safety: importer は非ローカル TARGET DB_HOST を拒否し、stage 接続は read-only。
# apply は --apply かつ --confirm-local-destroy の両方が必須 (本テーブルの old_db 行を
# 削除して再投入する破壊的操作)。
STAGE_IMPORT_DC = $(DC) -f docker-compose.yml -f docker-compose.stage-import.yml

# dry-run: 件数のみ表示。本テーブルへの書き込みは 0。
stage-import-dry-run:
	@echo "🔎 stage-import DRY-RUN (no writes) ..."
	$(STAGE_IMPORT_DC) run --rm stage-import

# apply: 破壊的。old_db 由来行を削除し stage から再投入 (単一トランザクション)。
# demo / master / config は保持。失敗時は全ロールバック。
stage-import:
	@echo "⚠️  stage-import APPLY (destructive: delete old_db rows + reinsert) ..."
	$(STAGE_IMPORT_DC) run --rm stage-import --apply --confirm-local-destroy

# 投入後検証: 空 clinic / branch leakage / owner collision / orphan / record_no /
# blocked leakage / demo 混入 を全チェック。exit 0 で PASS。
verify-stage-import:
	@echo "🔍 Verifying stage-import results ..."
	@bash scripts/verify-stage-import.sh

# rollback / read-only 安全性の統合テスト (実 DB 必要・STAGE_IMPORT_INTEGRATION=1)。
# 注入した失敗後に本テーブル件数が不変であること、stage 接続が read-only であることを検証。
stage-import-rollback-test:
	@echo "🧪 stage-import rollback + read-only integration test ..."
	$(STAGE_IMPORT_DC) run --rm -e STAGE_IMPORT_INTEGRATION=1 --entrypoint go \
		stage-import test ./cmd/stage-import/ -run 'RollsBack|ReadOnly' -count=1 -v -timeout 300s

# バックエンドのみ再起動
restart-api:
	$(DC) restart backend

# フロントエンドのみ再起動
restart-front:
	$(DC) restart frontend

# 本番ビルド（backend のみ。FE の本番配信は Vercel(frontend-deploy.yml)が担い、
# frontend/Dockerfile は「本番」と称する dev サーバイメージという矛盾があったため
# IR-12 A案で削除済み。dev 環境のイメージビルドは frontend/Dockerfile.dev のまま）
build-prod:
	docker build -f backend/Dockerfile.production -t animal-ekarte-api:latest ./backend

# golangci-lint バージョン（CI と同一）
GOLANGCI_LINT_VERSION := v2.11.4

# ── 品質チェック分担（詳細: docs/ops/ci-policy.md）────────────────
# リモート CI 必須: path-filtered build/test/coverage、gitleaks、
#                   codegen/migration 検証、AgentShield
# ローカル必須（make ci）: inventory / guardrail / shellcheck / golangci /
#                          ESLint / type-check / knip / design CTA + design-audit + build/test
# ローカル任意: make e2e（Playwright。リモート自動 CI には含めない）
# ────────────────────────────────────────────────────────────────

# リンター実行（Go・ローカル必須）- 公式 golangci-lint イメージを使用
# module cache は backend dev コンテナと共用（ekarte-go-mod-cache）。golangci-lint 自体の
# キャッシュ（/root/.cache）はバージョン固有形式のため GOCACHE とは混ぜず専用 volume に分離する。
lint:
	docker run --rm \
		-v $(PWD)/backend:/app \
		-v ekarte-go-mod-cache:/go/pkg/mod \
		-v ekarte-golangci-cache:/root/.cache \
		-w /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) \
		golangci-lint run

# リンター実行（自動修正）
lint-fix:
	docker run --rm \
		-v $(PWD)/backend:/app \
		-v ekarte-go-mod-cache:/go/pkg/mod \
		-v ekarte-golangci-cache:/root/.cache \
		-w /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) \
		golangci-lint run --fix

# テスト実行（Go）
# -p 1: 共有 test DB (ekarte_db_test) を触る package 同士の並列 TRUNCATE を避ける。
# docker-compose の GOFLAGS=-p=4 を CLI で上書きする（repository/CLAUDE.md 共有 DB 規則）。
test:
	$(DC) exec backend go test -race -v -p 1 ./...

# テスト実行（カバレッジ付き）
test-cover:
	$(DC) exec backend go test -race -cover -p 1 ./...

# repository パッケージのみ（共有 DB・必ず serial）
test-repository:
	$(DC) exec backend go test ./internal/repository/ -count=1 -p 1 -timeout 900s

# フロント静的チェック一式（ローカル必須・CI ゲート外）
# ESLint + TypeScript type-check + knip（旧 CI Frontend の静的ステップ相当）
lint-front:
	$(DC) exec frontend pnpm run lint
	$(DC) exec frontend pnpm run type-check
	$(DC) exec frontend pnpm run unused

# テスト実行（フロントエンド）
test-front:
	$(DC) exec frontend pnpm run test:run

# Playwright E2E（ローカルのみ・リモート自動 CI 対象外）
# 前提: make up 済みで frontend が http://localhost:3003 で応答すること。
# 実体: frontend/scripts/run-e2e.sh（公式 Playwright Docker イメージ）
# 例: make e2e
#     make e2e ARGS='e2e/owners-search.spec.ts'
e2e:
	@bash frontend/scripts/run-e2e.sh $(ARGS)

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

# ローカル一括 CI（リモート CI から外した静的ゲート + build/test/lint）
# 実行前に make up でコンテナを起動しておくこと（メタゲートは Docker 不要で先に fail-fast）。
# 実体: scripts/run-local-ci.sh / 分担: docs/ops/ci-policy.md
ci:
	@bash scripts/run-local-ci.sh

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
	@echo "  seed              シーダーのみ適用（差分のみ・べき等）"
	@echo ""
	@echo "旧DB移行（推奨経路: animalekarte_stage -> 本テーブル）:"
	@echo "  stage-import-dry-run      stage 取り込みの dry-run（件数表示・書き込み0）"
	@echo "  stage-import              stage から本テーブルへ投入（破壊的・要 OLD_DB_POSTGRES_PASSWORD）"
	@echo "  verify-stage-import       stage 投入後の検証（空clinic/orphan/collision等・exit 0でPASS）"
	@echo "  stage-import-rollback-test rollback/read-only 安全性の統合テスト（要 実DB）"
	@echo "  restart-api   API再起動"
	@echo "  restart-front フロントエンド再起動"
	@echo "  build-prod    本番ビルド"
	@echo "  dump-stg      STG DBダンプ(SSM経由 pg_dump → prodData/ekarte-stg-<実行日>.sql)"
	@echo ""
	@echo "品質管理:"
	@echo "  【ローカル必須】make ci（inventory/guardrail/lint/build/test 一括）"
	@echo "  【ローカル任意】make e2e（Playwright・リモート自動 CI 外）"
	@echo "  【リモート CI】gitleaks / path-filtered build+test+coverage — docs/ops/ci-policy.md"
	@echo "  ci            ローカル一括 CI（scripts/run-local-ci.sh）"
	@echo "  lint          Goリンター実行（golangci・ローカル必須）"
	@echo "  lint-fix      Goリンター実行（自動修正）"
	@echo "  test          Goテスト実行"
	@echo "  test-cover    Goテスト実行（カバレッジ付き）"
	@echo "  lint-front    FE静的チェック（ESLint+type-check+knip・ローカル必須）"
	@echo "  test-front    フロントエンドテスト実行"
	@echo "  e2e           Playwright E2E（要 make up・ARGS= で spec 指定可）"
	@echo "  build-front   フロントエンドビルド"
	@echo "  codegen       型定義生成（Go model → TypeScript型）"
	@echo "  codegen-check 型定義の差分チェック（CI用）"
	@echo "  schema-check  GoモデルとDBスキーマの差分チェック"
	@echo "  check-reset-contract      make reset の wait-set 契約を静的検証（Docker不要）"
	@echo "  check-reset-contract-test 上記契約チェック自体の回帰テスト"
	@echo "  shellcheck       scripts/*.sh を shellcheck で検査（severity=warning・ローカル無ければDocker経由）"
	@echo "  shellcheck-test  上記 shellcheck ゲート自体の回帰テスト"
	@echo "  build-go      Goビルド（開発用）"
	@echo "  mod-download  Goモジュールダウンロード"
	@echo "  mod-tidy      Goモジュールtidy"
	@echo "  sync-modules  node_modulesをホストにコピー（IDE補完用）"
	@echo "  setup-hooks   git hooksをセットアップ（初回・新メンバー用）"
	@echo ""
	@echo "  help          このヘルプを表示"
