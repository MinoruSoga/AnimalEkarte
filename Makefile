.PHONY: up down build logs logs-api logs-front ps db clean reset migrate seed docs-ui csv-import-preflight csv-import csv-import-verify a4-csv-import-preflight a4-csv-import a4-csv-import-verify a4-rehearsal-contract-test a4-rehearsal-config-check a4-rehearsal-up a4-rehearsal-ps a4-rehearsal-runtime-report a4-rehearsal-down f8-g4-rehearsal-contract-test f8-g4-rehearsal-config-check f8-g4-rehearsal-run f8-g4-rehearsal-down restart-api restart-front build-prod lint lint-fix test test-cover lint-front test-front build-front e2e build-go mod-download mod-tidy help codegen codegen-check sync-modules schema-check setup-hooks ci check-reset-contract check-reset-contract-test shellcheck shellcheck-test codex-security-scan

# デフォルトターゲット
.DEFAULT_GOAL := help

# $(DC) に --env-file を渡す（.env.local を変数展開の source of truth にする）
DC = docker compose --env-file .env.local

# Codex Security scan + latest artifact export
CODEX_SECURITY_MODEL ?= gpt-5.6-terra
CODEX_SECURITY_EFFORT ?= high
CODEX_SECURITY_OUTPUT_DIR ?= $(CURDIR)/codex-security-output

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

# OpenAPI 閲覧（Swagger UI :8081 / Redoc :8082）。db/backend/frontend は起動しない。
docs-ui:
	$(DC) --profile docs up -d swagger-ui redoc

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

# 完全リセット（スキーマ・シーダー含む）— local 専用・USER のみ実行
# 単一入口: scripts/local-db-reset-contract.sh
#   1) project/volume を固定値と compose 実測で照合（他環境は拒否）
#   2) umask 077 で .local-db-backups/<UTC>/ に pg_dumpall + sha256 + manifest
#   3) サービス停止後 ekarte-postgres-data のみ削除（cache 3 volume は保持）
#   4) 再起動 + missing=0 / DDL / 002_master,003_demo,004_staging /health を fail-closed 確認
# snapshot 失敗時は volume 削除へ進まない。compose の全 volume 一括削除は使わない。
# --wait の wait-set（db backend frontend, codegen 除外）は contract スクリプト側。
reset:
	@bash scripts/local-db-reset-contract.sh

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

# Codex Security の有料 scan が成功した場合だけ、全 artifact を export する。
# latest-* は codex-security-output/ 直下、原本は timestamped directory に保存される。
codex-security-scan:
	corepack pnpm exec codex-security scan . \
		--model "$(CODEX_SECURITY_MODEL)" \
		--effort "$(CODEX_SECURITY_EFFORT)"
	@bash scripts/export-codex-security-latest.sh \
		"$(CODEX_SECURITY_OUTPUT_DIR)" \
		--full-keep-files

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
# F6 CSV import: old_db's immutable 21-table CSV hand-off -> AnimalEkarte
# ============================================================================
# Required variables:
#   CSV_IMPORT_SOURCE_DIR (absolute host path), CSV_MANIFEST_SHA256,
#   CLINIC_CODE, CLINIC_ORDINAL, MIGRATION_RUN_ID,
#   TARGET_CLINIC_ID, FALLBACK_ANIMAL_SPECIES_ID, FALLBACK_EXAM_TYPE_ID,
#   TRIMMING_RESERVATION_TYPE_ID, PAYMENT_METHOD_CASH_ID,
#   PAYMENT_METHOD_CREDIT_CARD_ID.
# Apply additionally requires TARGET_DB_NAME to exactly match DB_NAME from
# .env.local. Reports contain aggregate counts and the six non-PHI seed IDs,
# and are owner-only under sensitive-local/. The source volume is read-only and
# no old_db network exists.
CSV_IMPORT_DC = $(DC) --profile csv-import
export CSV_IMPORT_SOURCE_DIR CSV_MANIFEST_SHA256 CLINIC_CODE CLINIC_ORDINAL MIGRATION_RUN_ID
export TARGET_CLINIC_ID FALLBACK_ANIMAL_SPECIES_ID FALLBACK_EXAM_TYPE_ID
export TRIMMING_RESERVATION_TYPE_ID PAYMENT_METHOD_CASH_ID
export PAYMENT_METHOD_CREDIT_CARD_ID TARGET_DB_NAME
CSV_IMPORT_COMMON_ARGS = \
	--source-dir /migration-input \
	--expected-manifest-sha256 "$${CSV_MANIFEST_SHA256}" \
	--clinic-code "$${CLINIC_CODE}" \
	--clinic-ordinal "$${CLINIC_ORDINAL}" \
	--run-id "$${MIGRATION_RUN_ID}" \
	--clinic-id "$${TARGET_CLINIC_ID}" \
	--fallback-animal-species-id "$${FALLBACK_ANIMAL_SPECIES_ID}" \
	--fallback-exam-type-id "$${FALLBACK_EXAM_TYPE_ID}" \
	--trimming-reservation-type-id "$${TRIMMING_RESERVATION_TYPE_ID}" \
	--cash-payment-method-id "$${PAYMENT_METHOD_CASH_ID}" \
	--credit-card-payment-method-id "$${PAYMENT_METHOD_CREDIT_CARD_ID}"

csv-import-preflight:
	@install -d -m 700 sensitive-local/csv-import-reports
	$(CSV_IMPORT_DC) run --rm --no-deps csv-import preflight $(CSV_IMPORT_COMMON_ARGS)

csv-import:
	@install -d -m 700 sensitive-local/csv-import-reports
	$(CSV_IMPORT_DC) run --rm --no-deps csv-import apply $(CSV_IMPORT_COMMON_ARGS) \
		--confirm-target-write --confirm-backup-ready \
		--confirm-target-host db --confirm-target-database "$${TARGET_DB_NAME}" \
		--report-path "/migration-reports/$${CLINIC_CODE}-$${MIGRATION_RUN_ID}-apply.json"

csv-import-verify:
	@install -d -m 700 sensitive-local/csv-import-reports
	$(CSV_IMPORT_DC) run --rm --no-deps csv-import verify $(CSV_IMPORT_COMMON_ARGS)

# ============================================================================
# A4 UI rehearsal: isolated, disposable, localhost-only full stack
# ============================================================================
# Required variables:
#   A4_COMPOSE_PROJECT=animalekarte-a4-<clinic/run slug>
#   A4_RUN_ID=<migration run ID>
#   A4_TARGET_RELEASE_COMMIT=<clean canonical 40-char HEAD>
# Stack startup does not import data. Use only the explicit a4-csv-import-*
# targets against this project's DB, then collect owner-only evidence in old_db.
A4_REHEARSAL_DC = COMPOSE_PROJECT_NAME="$${A4_COMPOSE_PROJECT}" docker compose \
	--env-file "$${A4_ENV_FILE}" \
	-p "$${A4_COMPOSE_PROJECT}" \
	-f docker-compose.yml -f docker-compose.a4-rehearsal.yml
A4_CSV_IMPORT_DC = COMPOSE_PROJECT_NAME="$${A4_COMPOSE_PROJECT}" docker compose \
	--env-file "$${A4_ENV_FILE}" \
	-p "$${A4_COMPOSE_PROJECT}" \
	-f docker-compose.yml -f docker-compose.a4-rehearsal.yml \
	--profile csv-import
export A4_COMPOSE_PROJECT A4_RUN_ID A4_TARGET_RELEASE_COMMIT A4_ENV_FILE

a4-rehearsal-contract-test:
	@node --test scripts/check-a4-rehearsal-compose.test.mjs \
		scripts/check-a4-env-file.test.mjs \
		scripts/check-a4-resource-boundary.test.mjs \
		scripts/write-a4-runtime-report.test.mjs

a4-rehearsal-config-check:
	@node scripts/check-a4-rehearsal-compose.mjs

a4-rehearsal-up: a4-rehearsal-config-check
	@node scripts/check-a4-resource-boundary.mjs start
	$(A4_REHEARSAL_DC) up -d --build --wait --wait-timeout 1200 db backend frontend

a4-rehearsal-ps:
	$(A4_REHEARSAL_DC) ps db backend frontend

a4-csv-import-preflight: a4-rehearsal-config-check
	@node scripts/check-a4-resource-boundary.mjs destroy
	@install -d -m 700 sensitive-local/csv-import-reports
	$(A4_CSV_IMPORT_DC) run --rm --no-deps csv-import preflight $(CSV_IMPORT_COMMON_ARGS)

a4-csv-import: a4-rehearsal-config-check
	@node scripts/check-a4-resource-boundary.mjs destroy
	@install -d -m 700 sensitive-local/csv-import-reports
	$(A4_CSV_IMPORT_DC) run --rm --no-deps csv-import apply $(CSV_IMPORT_COMMON_ARGS) \
		--confirm-target-write --confirm-backup-ready \
		--confirm-target-host db --confirm-target-database "$${TARGET_DB_NAME}" \
		--report-path "/migration-reports/$${CLINIC_CODE}-$${MIGRATION_RUN_ID}-apply.json"

a4-csv-import-verify: a4-rehearsal-config-check
	@node scripts/check-a4-resource-boundary.mjs destroy
	@install -d -m 700 sensitive-local/csv-import-reports
	$(A4_CSV_IMPORT_DC) run --rm --no-deps csv-import verify $(CSV_IMPORT_COMMON_ARGS)

a4-rehearsal-runtime-report: a4-rehearsal-config-check
	@node scripts/write-a4-runtime-report.mjs

# Explicit stop/cleanup path for the disposable project. The project name is
# mandatory, so this cannot fall back to the normal development stack.
a4-rehearsal-down:
	@node scripts/check-a4-resource-boundary.mjs destroy
	$(A4_REHEARSAL_DC) down --volumes --remove-orphans

# ============================================================================
# F8 G4 failure rehearsal: fixed synthetic rollback on a dedicated stack
# ============================================================================
F8_G4_DC = COMPOSE_PROJECT_NAME="$${F8_G4_COMPOSE_PROJECT}" docker compose \
	--env-file "$${F8_G4_ENV_FILE}" \
	-p "$${F8_G4_COMPOSE_PROJECT}" \
	-f docker-compose.f8-g4-rehearsal.yml
export F8_G4_COMPOSE_PROJECT F8_G4_RUN_ID F8_G4_TARGET_RELEASE_COMMIT
export F8_G4_ENV_FILE F8_G4_DB_PORT F8_G4_CLINIC_CODE F8_G4_CLINIC_ORDINAL

f8-g4-rehearsal-contract-test:
	@node --test scripts/lib/f8-g4-evidence.test.mjs scripts/lib/f8-g4-host-safety.test.mjs

f8-g4-rehearsal-config-check:
	@F8_G4_BUILD_CONTEXT="$(CURDIR)/backend" \
		F8_G4_BACKEND_TREE_ID=config-check-unattested \
		F8_G4_RUNNER_IMAGE=config-check-runner:unattested \
		$(F8_G4_DC) config --quiet

f8-g4-rehearsal-run:
	@node scripts/run-f8-g4-rehearsal.mjs

f8-g4-rehearsal-down:
	@node scripts/check-f8-g4-resources.mjs

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

# ヘルプ
help:
	@echo "Animal Ekarte - 開発コマンド"
	@echo ""
	@echo "使用方法: make [コマンド]"
	@echo ""
	@echo "コマンド:"
	@echo "  codex-security-scan Codex Security scan後、最新版をcodex-security-output/直下へexport（有料）"
	@echo "  up            コンテナ起動"
	@echo "  build         コンテナ起動（ビルド付き）"
	@echo "  down          コンテナ停止"
	@echo "  docs-ui       Swagger UI (:8081) / Redoc (:8082) だけ起動"
	@echo "  logs          全ログ表示"
	@echo "  logs-api      APIログ表示"
	@echo "  logs-front    フロントエンドログ表示"
	@echo "  ps            コンテナ状態確認"
	@echo "  db            DB接続（psql）"
	@echo "  clean         キャッシュクリア＆再ビルド"
	@echo "  reset         local DB 再構築（snapshot→ekarte-postgres-data のみ削除→postflight。USER のみ）"
	@echo "  migrate       差分マイグレーションのみ適用（DBは落とさない）"
	@echo "  seed              シーダーのみ適用（差分のみ・べき等）"
	@echo ""
	@echo "旧DB移行（正式経路: 21表CSV + manifest -> 本テーブル）:"
	@echo "  csv-import-preflight      source/seed/schema/空band検査（read-only）"
	@echo "  csv-import                21表CSVを単一transactionで投入（backup・target確認必須）"
	@echo "  csv-import-verify         manifest件数/clinic/sequence検証（read-only）"
	@echo "  a4-rehearsal-contract-test A4隔離構成/runtime report契約テスト（Docker起動不要）"
	@echo "  a4-rehearsal-config-check A4 Composeのlocalhost/network/volume契約検査"
	@echo "  a4-rehearsal-up          A4専用disposable stackをbuild/start"
	@echo "  a4-csv-import-*          A4専用DBへのcanonical preflight/apply/verify"
	@echo "  a4-rehearsal-runtime-report 稼働中A4 stackのowner-only証跡生成"
	@echo "  a4-rehearsal-down        指定A4 projectと専用volumeを明示破棄"
	@echo "  f8-g4-rehearsal-run      固定synthetic G4失敗を専用DBで実行しrollback証跡を生成"
	@echo "  f8-g4-rehearsal-down     labels検証後にF8 G4専用stack/volumeを削除"
	@echo "  restart-api   API再起動"
	@echo "  restart-front フロントエンド再起動"
	@echo "  build-prod    本番ビルド"
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
