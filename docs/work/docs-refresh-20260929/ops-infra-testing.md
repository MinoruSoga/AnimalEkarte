# DOCS-REFRESH-20260929 — ops-infra-testing

区画: `docs/ops/infra/**` + `docs/ops/testing/*.md` 直下（scenarios/ 除外）。照合先は現行の `infra/`、`backend/wrangler*.jsonc`、`.github/workflows/`、`frontend/e2e/`、`frontend/playwright.config.ts`、`frontend/scripts/run-e2e.sh`、`backend/migrations/`、`load-tests/`。外部 runtime/account/billing 状態は一切検証していない。

## 対象と判定

| ファイル | 判定 | 主な根拠 |
|---|---|---|
| docs/ops/infra/architecture.md | corrected | `backend/worker/index.ts:54` `sleepAfter = "1h"`（旧記述 10m）；`wrangler.jsonc` STG `vars` に `DB_MAX_OPEN_CONNS=10`/`DB_MAX_IDLE_CONNS=5` |
| docs/ops/infra/iac-guidelines.md | kept | Terraform/W rangler 境界・local backend・drift 未実装は `infra/cloudflare/`・`reorg-plan.md` と整合 |
| docs/ops/infra/README.md | kept | 索引・構成記述は現行ファイル配置と一致 |
| docs/ops/infra/reorg-plan.md | kept | HISTORICAL/DO NOT EXECUTE マーカー維持。主張自体が履歴枠内 |
| docs/ops/infra/staging/runbook.md | kept | deploy trigger paths・`POST /_internal/migrate`+`MIGRATE_RUN_SECRET`（`infra/scripts/cf-run-migrate.sh:17`）・`constraints.regions=["APAC"]`（`wrangler.jsonc:151`）・`scheduling_policy="default"`（:141）・`LABEL rollout`（`Dockerfile.production:38`）を照合 |
| docs/ops/infra/production/setup.md | corrected | 実行 SoT を Linear→Plane（`setup.md:5,:19,:110`）。`wrangler.production.jsonc` が draft である表現は維持 |
| docs/ops/infra/production/runbook.md | corrected | 同上。live status → Plane、checklist の Linear→Plane |
| docs/ops/testing/CLAUDE.md | corrected | FAIL 追跡先・新規 defect intake を Linear→Plane |
| docs/ops/testing/CLINICAL-E2E-DESIGN.md | corrected | L4 範囲 S01–S13→S01–S39；Linear Done→Plane Done；e2e.yml suite 配線済み・未実行の表記は既に正確 |
| docs/ops/testing/E2E_TESTING_GUIDE.md | corrected | `inputs.suite`（auth-smoke/clinical/v04）への訂正；2026-09-23 ローカル clinical 33 PASS/7 FAIL を追記（green ではない）；`E2E_RESULTS_DIR` artifact 挙動を追記；playwright-report 不一致は維持 |
| docs/ops/testing/INTEGRATION_TEST_PLAN.md | corrected | execution state 追跡 Linear→Plane；e2e.yml suite 振り分けの訂正；記録先に Plane |
| docs/ops/testing/README.md | corrected | scenario 索引 S01–S13→S01–S39+V01–V05+UAT-254；E2E 現状を suite/clinical run に更新；Linear→Plane |
| docs/ops/testing/S09-FIXTURE-DESIGN.md | corrected | `frontend/e2e/s09-closing-time-boundaries.spec.ts` の存在・endpoint fixture 経路・明示 spec path 実行要件を追記。BLOCKED は維持。Linear Done→Plane Done |
| docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.md | corrected | `tax-breakdown.ts` パス `features/accounting/`→`features/accounting/lib/`（実在確認済み） |
| docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md | corrected | テンプレート内 Linear 欄→Plane item、Linear参照→Plane参照、冒頭に読み替え訂正 |
| docs/ops/testing/TEST_ARCHITECTURE.md | corrected | L3/L4 記述（suite 配線・S39・clinical run）、CI profile 表、FAIL/BLOCKED 追跡先 Linear→Plane |
| docs/ops/testing/UAT-DOMAIN-STATUS.md | corrected | 冒頭照合注記に Linear 閉鎖・Plane 読み替えを追記（Plane 未照会=UNKNOWN 維持）；mermaid「Linear で追跡」→Plane |
| docs/ops/testing/UAT-ENV-SETUP.md | corrected | FAIL 追跡先 Linear→Plane |
| docs/ops/testing/liff-verification.md | kept | `liff/src/hooks/use-liff-link.test.ts` は frontend container 内相対パスとして正（`frontend/liff/src/hooks/` に実在）。`use-liff.ts`/`liff_auth.go`/`make e2e`/`run-e2e.sh` 全て実在 |
| docs/ops/testing/PERFORMANCE_PROFILING.md | kept | k6 script 3 本・閾値（p95<500/p95<2000）・`LOAD_TEST_LOGIN_*` fail-closed・profile.go 削除・N+1 test 実在を照合 |

## 実施した訂正（抜粋・重要度順）

- `docs/ops/testing/TEST_ARCHITECTURE.md`: 「manual GitHub workflow は non-gating の auth smoke のまま。`--clinical` helper は実装済みだが未実行」→ `inputs.suite`（auth-smoke/clinical/v04）配線済み + ローカル clinical 1 回実行（33 PASS/7 FAIL）（根拠: `.github/workflows/e2e.yml`、`frontend/scripts/run-e2e.sh`、gitignore 済み `reports/uat-2026-09-23/clinical-e2e-emr128/` 記録）
- `docs/ops/testing/E2E_TESTING_GUIDE.md`: 「workflow が実行するのは `auth-flows.spec.ts` のみ」→ suite 振り分け 3 系（根拠: `e2e.yml` の `inputs.suite` choice + `run-e2e.sh` モード）
- `docs/ops/testing/INTEGRATION_TEST_PLAN.md`: 同上（`auth-flows.spec.ts` のみ → suite 振り分け、EMR-128・2026-09-28）
- `docs/ops/testing/README.md`: 「E2E workflow は auth smoke に配線済み・`--clinical` 未実行」→ suite 配線済み・clinical ローカル 1 回（非 green）
- `docs/ops/testing/{TEST_ARCHITECTURE,README,CLINICAL-E2E-DESIGN}.md`: L4 宣言範囲 S01–S13 → S01–S39 + V01–V05 + UAT-254（根拠: `scenarios/README.md` 索引。scenarios/ 自体は別子区画のため未編集）
- `docs/ops/infra/architecture.md`: Container `sleepAfter = "10m"` → `"1h"`（根拠: `backend/worker/index.ts:54`、`wrangler.jsonc:116` コメント）。DB pool 行を STG vars に合わせて両環境へ展開（`wrangler.jsonc` `DB_MAX_OPEN_CONNS=10`/`DB_MAX_IDLE_CONNS=5`）
- `docs/ops/testing/S09-FIXTURE-DESIGN.md`: spec 非言及 → `s09-closing-time-boundaries.spec.ts` 存在・`POST/DELETE /api/v1/uat/synthetic-closings`・`UAT_SYNTHETIC_CLOSING_PASSWORD`・明示 path 実行・allowlist 非所属を追記（根拠: spec 冒頭コメント、`backend/internal/billing/synthetic_closing_http.go`）
- `docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.md`: `features/accounting/tax-breakdown.ts` → `features/accounting/lib/tax-breakdown.ts`
- 全区画: 実行状態の正本 Linear → Plane（workspace `baritechllc` / Project `EMR`）へ統一。Linear は 2026-09-16 閉鎖・履歴参照のみ。履歴・エイリアスとしての言及は保持

## 外部依存・BLOCKED（未解消のまま残すもの）

- **S09 BLOCKED 維持**: spec は存在するが実行証跡なし。#2–#6 のブラウザ UAT・締めプレビュー集計は未実施。
- **clinical/V04 の Actions 実行証跡なし**: `e2e.yml` `inputs.suite` 配線は存在するが workflow_dispatch 実行は未検証。clinical はローカル 33 PASS/7 FAIL で green 未達。
- **auth smoke の Actions 実行・fresh DB 結果**: UNREPORTED/UNKNOWN を維持。
- **performance**: k6/Lighthouse の配線は確認済みだが実 Actions/fresh DB/k6 実行証跡は UNREPORTED/UNKNOWN。Lighthouse は `continue-on-error` で非ゲート。
- **production 構築**: `wrangler.production.jsonc`/`infra/cloudflare/production/` は draft。GitHub Environment `Production`・billing・PlanetScale・DNS・Vercel 本番設定は人間前提条件として未検証のまま。
- **STG/PROD runtime**: DB 内容・credential・provider status・稼働 instance 配置は本更新で未確認。
- **Plane 側の最新状態**: 照会していない（UNKNOWN）。docs の記述は repo 側契約の訂正のみ。

## 提案削除候補（理由・後継リンク）

- なし。`reorg-plan.md` は HISTORICAL 枠が既に正しく機能しており削除対象にしない。

## 検証

- `git diff --check -- docs/ops/infra docs/ops/testing` → exit 0（PASS）
- `bash scripts/check-docs-symbol-drift.sh` → exit 1。**FAIL 3 件は全て本区画外**（`docs/architecture/erd.md`・`docs/spec/specification.md` の「全 128 テーブル」宣言 vs 実測 130 = `backend/migrations/*.sql` 行頭 `CREATE TABLE` 合算）。main checkout でも同一 3 件 FAIL を再現（605 トークン一致）。本区画の変更による新規ドリフトなし。architecture/spec 区画の子への手続き対象。
- `node scripts/check-workflow-contracts.test.mjs` → 21/21 pass（workflow 契約記述との整合）
- `ls frontend/e2e/` → 35 spec。E2E 文書記載の spec 名（clinical-flows/clinical-smoke/examinations/vaccinations/checkups/hospitalization/estimates/auth-flows/s09/medical-records-*）は全て実在。
- `git status --porcelain` → allowlist 内 14 ファイルのみ。allowlist 外差分: なし
- `docs/ops/testing/scripts/check-uat-env.sh` → 実在確認のみ（runtime 実行は brief 禁止範囲のため未実行）

## 範囲外で観測した問題（親エージェントへの引き渡し）

- `.github/workflows/backend-deploy.yml:126` コメントに `sleepAfter 10m` の残存（現行 `1h`）。workflow ファイルは本区画外のため未訂正。
- `docs/ops/testing/scenarios/` 内に Linear 現行扱いの記述あり（UAT-254-CLOSE-CHECKLIST 等）。scenarios/ は ops-scenarios 区画のため未訂正。
- `docs/ops/README.md`・`docs/ops/deploy/*`・`docs/ops/agent-harness.md` に Linear を実行 SoT とする記述あり。ops-deploy/他区画のため未訂正。
- ERD/specification の「128 テーブル」宣言は実測 130 と不一致（architecture/spec 区画。`check-docs-symbol-drift` FAIL の原因）。
