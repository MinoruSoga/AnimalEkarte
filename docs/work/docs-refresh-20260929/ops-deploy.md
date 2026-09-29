# docs-refresh-20260929 / ops-deploy 作業レポート

- **作業日**: 2026-09-29
- **worktree**: `/private/tmp/ae-dr29-ops-deploy`（branch `docs-refresh-20260929/ops-deploy`、base `main` @ `4722f4db7`）
- **担当範囲**: `docs/ops/*.md` 直下 + `docs/ops/deploy/**`（runbooks 含む）+ 本レポート
- **照合先**: `.github/workflows/`、`Makefile`、`docker-compose*.yml`、`infra/`、`scripts/`、`backend/internal/infra/lstep/`、`backend/internal/csvimport/`、`backend/worker/`、`backend/wrangler*.jsonc`、vault `ノア案件データ移行/` カプセル
- **claim**: `claim/DOCS-REFRESH-20260929-OPS_DEPLOY` 存在を確認済み（parent 取得済み）。本 branch 上では変更していない

## 1. 対象ファイルと disposition

| ファイル | disposition | 備考 |
|---|---|---|
| `docs/ops/README.md` | **corrected** | 実行 SoT Linear → Plane（日付付き訂正×2箇所） |
| `docs/ops/CLAUDE.md` | kept | 現行と一致（AWS 退役・SSOT 分離の記述が正確） |
| `docs/ops/agent-harness.md` | **corrected** | 同上（L23） |
| `docs/ops/backlog-spreadsheet.md` | kept | 外部シートを固定値で書かない設計が現行ポリシーと一致 |
| `docs/ops/ci-policy.md` | kept | 4 backend shard / 2 frontend vitest shard を `scripts/ci_scope_plan.py:255-269` で確認。E2E manual-only・AgentShield 条件付き fail・action pin 方針も現行通り |
| `docs/ops/coverage-policy.md` | **corrected** | backend 直下 `lstep-migrate/`・`seed-old-db/` の記述を `cmd/` 配下移動へ日付付き訂正。baseline 87.3%/43.78% は現行通り（変更なし） |
| `docs/ops/deploy/README.md` | **updated** | `FIRST_SYSTEM_ADMIN.md` が index 未収録だったのを追記（同 doc は 2026-09-11 追加済み） |
| `docs/ops/deploy/A4_UI_REHEARSAL.md` | kept | `a4-rehearsal-*` Make target・`docker-compose.a4-rehearsal.yml` 実在確認 |
| `docs/ops/deploy/BREAK-HOURS-SHAPE-AUDIT.md` | kept | read-only 監査 SQL と guard 実装説明が現行と一致 |
| `docs/ops/deploy/CI-CD-PIPELINE.md` | kept | `backend-deploy.yml`・`frontend-deploy.yml` の trigger/env/secret 名と一致。checked-in config と外部状態の区別も適切 |
| `docs/ops/deploy/CLINIC_CSV_IMPORT.md` | **corrected** | `billing_items` 行の clinic 列/isolation 誤り + Linear BRT-42 → Plane EMR-46（日付付き訂正） |
| `docs/ops/deploy/CLOUDFLARE-EXTERNAL-INTEGRATIONS-AUDIT.md` | kept | LSTEP 二重 gate・UNVERIFIED スタンスが実装と一致 |
| `docs/ops/deploy/CRUD-SMOKE-TEST.md` | kept | API 契約・権限期待値が現行と一致 |
| `docs/ops/deploy/DEPLOYMENT_CHECKLIST.md` | kept | `make lint/lint-front/test/test-front/e2e`、`db_reset` 入力非存在、Production stop 条件が現行通り |
| `docs/ops/deploy/F8_G4_FAILURE_REHEARSAL.md` | kept | `f8-g4-rehearsal-*` Make target・compose 実在確認 |
| `docs/ops/deploy/FIRST_SYSTEM_ADMIN.md` | kept | `staff-provision` actor 前提・production 経路分離がコードと一致 |
| `docs/ops/deploy/LAB_DEVICE_AGENT_MACOS.md` | kept | `LAB_DEVICE_AGENT_CONSUMER_TOKEN`・503 応答・route 権限を `lab_device_agent_consumer_handler.go`・`routes_lab.go` で確認 |
| `docs/ops/deploy/LAB_DEVICE_CONNECTIVITY.md` | **corrected** | 実行状態の正本 Linear BRT-94〜100 → Plane（日付付き訂正） |
| `docs/ops/deploy/LOCAL_DB_RESET.md` | kept | `make reset`・`ekarte-postgres-data` のみ削除・snapshot・`missing=0` が script/compose と一致 |
| `docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md` | kept | `LSTEP_WRITE_API_ENABLED` exact `"true"`・`ErrWriteDisabled`・4 method・`is_sync_enabled` 二重 gate が `backend/internal/infra/lstep/` と一致。deployed 値は UNKNOWN のまま |
| `docs/ops/deploy/MIXED-PAYMENT-SMOKE-TEST.md` | kept | payment_splits 契約・legacy method 優先度と一致 |
| `docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md` | **corrected** | 実行状態 Linear → Plane（日付付き訂正） |
| `docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md` | kept | 退役 `make seed-old-db`/`003_demo`/`004_staging` が正しく履歴ラベル済み |
| `docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md` | **corrected** | PII 記入禁止先の Issue/Linear → Issue/Plane へ現行化（既存の受領記録訂正は維持） |
| `docs/ops/deploy/STG-CONTINUOUS-OPERATIONS.md` | kept | `DB_MAX_OPEN_CONNS=10`/`DB_MAX_IDLE_CONNS=5`（`wrangler.jsonc:59-60`）、`audit_write_failed`（`audit/service.go:151`）、`002_master` 4医院が一致 |
| `docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md` | kept | 未実装 cleanup automation が historical proposal と明記済み |
| `docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md` | kept | `db_reset` 入力非存在・`cf-run-migrate.sh`・`stg-uat-handoff-*` 参照が実在 |
| `docs/ops/deploy/VERCEL-FRONTEND-STAGING-TEST.md` | kept | ECS/CloudWatch/CloudFront が retired 表記。preview alias `stg.noah-karte.com` は workflow と一致 |
| `docs/ops/deploy/runbooks/README.md` | kept | 4 runbook の index 完備 |
| `docs/ops/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md` | kept | 資格情報 USER 作業境界・`STG_DEMO_*` と load-test 合成資格情報の分離が正確 |
| `docs/ops/deploy/runbooks/SCHEDULER_OPERATIONS.md` | **corrected** | STG 第4 cron（keep-alive）の日付付き追記 |
| `docs/ops/deploy/runbooks/SEC_SECRETS_5_GITLEAKS_HISTORY_INVENTORY.md` | kept | 22 findings / 9 paths は履歴棚卸し。secret 値なし |
| `docs/ops/deploy/runbooks/STG_PRE_DEPLOY_READINESS_CHECK.md` | kept | deploy/migrate/health 順序が `backend-deploy.yml` と一致 |

## 2. 主要な訂正（旧主張 → 新主張）

1. **`docs/ops/README.md` L39-40**: 「タスク台帳の実行 SoT は Linear」「Linear / GitHub で検証必須」→ **Plane**（workspace `baritechllc`）。BRT-39/BRT-40 の linear.app リンクは履歴 pointer として保持（リンク自体は閉鎖済み Linear のため文字列化）。根拠: Linear 2026-09-16 閉鎖、`todo.md` ヘッダ、`docs/work/plane-md-migration-20260923-receipt.md`。
2. **`docs/ops/agent-harness.md` L23**: 「Linearが利用できなければ」→ 実行 SoT の **Plane**。
3. **`docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md` L121**: 「実行状態は Linear」→ **Plane**。
4. **`docs/ops/deploy/LAB_DEVICE_CONNECTIVITY.md` L8**: 「実行状態の正本は Linear BRT-94〜100」→ **Plane**（BRT-94 は receipt で EMR-14 へ移行確認済み。BRT-94〜100 は履歴 ID）。
5. **`docs/ops/deploy/CLINIC_CSV_IMPORT.md` L38/L48**: `billing_items` 行 `clinic 列=no / isolation=id_band_and_parent_fk` → **`yes` / `clinic_id_column`**。根拠: `backend/internal/csvimport/cutover_contract.go:256`（`Columns` に `clinic_id`）と `:226`（placeholder `billing_items.clinic_id → {{CLINIC_ID}}`）。`CutoverIsolationClinicID` 定数と一致。
6. **`docs/ops/deploy/CLINIC_CSV_IMPORT.md` L212**: 追跡先「Linear BRT-42」→ **Plane EMR-46**（receipt L143 で BRT-42→EMR-46 を確認）。
7. **`docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md` L230**: 「PII を Issue / Linear に書く」→ 「Issue / Plane へ書く」（禁止範囲の現行化。意味は同一）。
8. **`docs/ops/coverage-policy.md` L24**: `lstep-migrate/`・`seed-old-db/` を backend 直下として記載 → `backend/cmd/lstep-migrate`・`backend/cmd/_archive/seed-old-db` へ移動済みの日付付き訂正（母集団外の結論は不変）。
9. **`docs/ops/deploy/runbooks/SCHEDULER_OPERATIONS.md` L19**: 「Workerは次の3式以外をfail-closedで拒否する」のみの記述 → STG `backend/wrangler.jsonc:123` の第4 trigger `*/4 0-9 * * *`（EMR-213 API container keep-alive、`worker/index.ts:451,539` で scheduler dispatch 前に個別処理、`jobsForCron` 非所属、production `wrangler.production.jsonc:119` には無し）を日付付き追記で区別。
10. **`docs/ops/deploy/README.md` L46**: `FIRST_SYSTEM_ADMIN.md` の index 欠落を追記（doc は 2026-09-11 から存在、README 最終更新 2026-08-31 のため未収録だった）。

上記すべて `(2026-09-29 訂正: …)` 形式の日付付きノートを併記（STAFF_ACCOUNT_PROVISIONING の禁止先変更は事実主張ではなく禁止範囲の現行化のためノートなし）。

## 3. 未解決の外部依存・BLOCKED

- **formal COMPLETE producer bundle 受領**: 未記入・受領記録なし。KNJO source 未完全で `payments.csv`/`payment_splits.csv` は header-only → apply **BLOCKED**（vault `ノア案件データ移行/` カプセルと一致）。
- **production cutover**: 本 lane では実施しない。#253/#254/#255 gate 後の USER 作業。
- **GitHub Environment reviewers / secrets / provider 実体**: `UNKNOWN`。checked-in config は証拠にならない。
- **deployed `LSTEP_WRITE_API_ENABLED` / `LAB_DEVICE_AGENT_CONSUMER_TOKEN` 実値**: `UNKNOWN`（repository default OFF/未供給のみ確認）。
- **LINE Webhook redelivery / Error statistics aggregation**: code 実装済み・Console 有効化と rehearsal は **RELEASE PENDING**。
- **八王子 `TBL_KNJO_DATA` corruption**: vault 記載の継続 migration blocker（`make stg-uat-handoff` 対象外と docs も一致）。
- **Plane 現行状態**: 本タスクは read-only 前提で live 取得していない。BRT→EMR 対応は receipt の記録値のみ使用。

## 4. 削除候補

なし。全対象は現行運用または履歴ラベル済みの証跡として残すべき内容。

## 5. 検証コマンドと結果

| コマンド | exit | 結果 |
|---|---|---|
| `git diff --check -- docs/ops` | 0 | whitespace/error なし |
| `bash scripts/check-docs-symbol-drift.sh` | **1** | `FAIL テーブル数: docs の宣言値 128 が実装の実測値 130 と不一致` ×3。**発生源は allowlist 外**（`docs/spec/specification.md` L22/L61、`docs/architecture/erd.md` L153-216 等）。実測 130 = 128 + `011_support_bug_reports.sql` + `013_support_chat_messages.sql`（`CREATE TABLE` 各1）。spec-core / arch 担当子エージェントへの引き継ぎ事項 |
| `git status --porcelain` | — | 変更 9 ファイル、すべて `docs/ops/` 配下 + 本レポート = **allowlist 内のみ** |
| 参照 sweep（make target / script / workflow 名を機械列挙して存在確認） | — | 実在しない名称の言及はすべて履歴・退役ラベル済み（`make seed-old-db`、`stg-smoke-cleanup.yml`、旧 `002_seed.sql`/`003_*.sql`）か、bundle 内ファイル（`install.sh`/`diagnose.sh`）、producer artifact（`020_canonical.sql`/`030_stage.sql`）、repo 外 secured path（`/secure/first-system-admin/create.sql`）のいずれかで誤記ではない |

## 6. allowlist-only 確認

`git status --porcelain` の変更対象: `docs/ops/README.md`、`docs/ops/agent-harness.md`、`docs/ops/coverage-policy.md`、`docs/ops/deploy/{CLINIC_CSV_IMPORT,LAB_DEVICE_CONNECTIVITY,OLD_DB_HANDOFF_LOCAL,README,STAFF_ACCOUNT_PROVISIONING}.md`、`docs/ops/deploy/runbooks/SCHEDULER_OPERATIONS.md`、本レポート `docs/work/docs-refresh-20260929/ops-deploy.md`。`docs/ops/infra/**`、`docs/ops/testing/**`、コード・migration・他領域 docs への変更なし。secret・PII・患者/飼主/スタッフ実データの記入なし。
