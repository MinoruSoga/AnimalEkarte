# STG Pre-Deploy Readiness Check

| 項目 | 内容 |
|---|---|
| Type | Ops Runbook |
| 対象環境 | Staging (`stg.noah-karte.com` / `api.stg.noah-karte.com`) |
| 作成日 | 2026-05-20 |
| 対象 commit | `f9810356` (migration 統合 + lstep fire hour 削除) |
| 関連 | [CRUD-SMOKE-TEST.md](../CRUD-SMOKE-TEST.md) / [DEPLOYMENT-CHECKLIST.md](../DEPLOYMENT-CHECKLIST.md) / [CI-CD-PIPELINE.md](../CI-CD-PIPELINE.md) |

main → staging への大規模反映時の事前チェック手順。本 runbook は **再利用可能な恒久版** であり、特定の commit に縛られない (対象 commit 欄は初回反映時の記録)。

---

## 1. 前提

| # | 前提条件 |
|---|---|
| 1 | migrations は `001_init.sql`〜`004_seed_staging.sql` の **4 ファイルに統合済み** |
| 2 | **STG 初回反映は DB reset 前提**。schema_migrations checksum mismatch 回避のため空 DB から 001〜004 を順次適用 |
| 3 | `LstepFireHourJST` (configurable fire hour) は削除済み。仕様 §6.4 で `deliveryTriggerHourJST = 10` (JST) hardcoded 実装 |
| 4 | Lステップ Write API は **noop のまま** (`backend/internal/infra/lstep/tag.go` `user.go` の `[DISABLED]` コメント参照) |
| 5 | `.env.staging` の内容監査はユーザーが手動実施 (Claude 閲覧不可ディレクトリ) |
| 6 | 本番 (`production` ブランチ) には直接 push しない。staging で smoke OK 後に別 PR で昇格 |

---

## 2. 必須チェック

### 2.1 `.env.staging` 監査

- [ ] シークレット混入なし (`JWT_SECRET` / DB password / API トークン類 / `INTEGRATION_ENCRYPTION_KEY`)
  - 望ましくは SSM Parameter Store 経由が理想。現状 `backend-deploy.yml` L72-87 で `.env.staging` 全変数を平文展開している点に注意。
- [ ] 必須 env 欠落なし
  - `JWT_SECRET` (本番起動拒否対象) / `CORS_ALLOWED_ORIGIN` / `DB_HOST` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` / `INTEGRATION_ENCRYPTION_KEY` / `LSTEP_*`
  - `CORS_ALLOWED_ORIGIN` は `https://stg.noah-karte.com,https://api.stg.noah-karte.com` を含むこと (env var 名が `ALLOWED_ORIGINS` ではないことに注意 — `backend/internal/middleware/cors.go` 参照)
- [ ] `DB_RESET=true` は **初回 STG デプロイ時のみ** 設定する
- [ ] デプロイ後 `DB_RESET=false` に戻す (本番影響回避)

### 2.2 main → staging PR 作成

- [ ] PR 作成: `gh pr create --base staging --head main --title "..." --body "..."`
- [ ] PR body に **対象 commits の要約** + 主要変更点を記載
- [ ] CI workflow `ci.yml` の `conclusion: success` を確認
- [ ] **`gh run view --json jobs <RUN_ID>` で job 単位の conclusion を確認**
  - 教訓: workflow conclusion=success でも `paths-filter` で skipped job があると当該層未検証 (memory: `feedback_paths_filter_silent_green`)
  - 期待: `backend-test` / `frontend-test` / `lint` / `codegen-check` の全 conclusion が `success`。`skipped` は NG 扱い。
- [ ] CI green を確認するまで merge しない

### 2.3 STG deploy with DB reset

**実行方式の選択:**

| 方式 | コマンド | 利点 | 欠点 |
|---|---|---|---|
| **A. workflow_dispatch (推奨)** | `gh workflow run backend-deploy.yml --ref staging -f db_reset=true` | `.env.staging` を書き換えずに済む。戻し忘れ防止。 | 手動起動が必要 |
| B. staging push | `.env.staging` の `DB_RESET=true` を一時設定 → `git push origin staging` | 自動起動 | 戻し忘れリスクあり |

- [ ] migrate task logs を `aws logs tail /ecs/animalekarte-stg --follow --region us-east-1` で監視
- [ ] 期待ログ (DB reset 時):
  1. `DROP DATABASE ekarte_db` → `CREATE DATABASE ekarte_db`
  2. `applying 001_init.sql`
  3. `applying 002_seed_master.sql`
  4. `applying 003_seed_demo.sql`
  5. `applying 004_seed_staging.sql`
- [ ] **checksum mismatch エラーが出ないこと** (空 DB からの開始なら原理的に発生しない)
- [ ] migrate task exit code = 0 (`backend-deploy.yml` L227-238 のガード)
- [ ] seed 適用確認 (例):
  - `SELECT COUNT(*) FROM clinics` ≥ 1
  - `SELECT COUNT(*) FROM staffs` ≥ 1
  - `SELECT COUNT(*) FROM permission_groups` ≥ 1
  - `SELECT COUNT(*) FROM lstep_auto_managed_prefixes` = 19

### 2.4 STG health check

- [ ] `curl -sf http://animalekarte-stg-alb-1915768826.us-east-1.elb.amazonaws.com/health` → **HTTP 200**
- [ ] `curl -sf https://api.stg.noah-karte.com/health` → **HTTP 200** (CloudFront 経由)
- [ ] `aws ecs describe-services --query 'services[0].{status,runningCount,desiredCount}'` で `runningCount == desiredCount`
- [ ] backend logs に起動エラーなし: `aws logs tail /ecs/animalekarte-stg --since 10m --region us-east-1`
- [ ] CloudWatch Logs に **`ERROR` レベルログがゼロ**: `aws logs filter-log-events --log-group-name /ecs/animalekarte-stg --filter-pattern "ERROR"`
- [ ] ALB Target Health が `healthy`: `aws elbv2 describe-target-health --target-group-arn <TG_ARN>`

### 2.5 CRUD smoke test

[CRUD-SMOKE-TEST.md](../CRUD-SMOKE-TEST.md) に厳密に従う。

- [ ] **医院 / クリニック CRUD**
  - [ ] API: GET 一覧 (所属のみ / scope=all) / GET 詳細 / POST 作成 / PATCH 更新 / DELETE 削除 (FK あれば 409)
  - [ ] FE: `ClinicMasterSettings` 画面で一覧・追加・編集・削除
  - [ ] system_admin の `scope=all` が全クリニック返却 / 通常スタッフは 403
- [ ] **権限グループ CRUD**
  - [ ] API: GET 一覧/詳細 / POST 作成 / PATCH 更新 / PUT ルール更新 / DELETE 削除
  - [ ] FE: `PermissionGroupSettings` 画面で全操作
  - [ ] ルール変更後の認可反映 (middleware diff 検出, FEAT-374 Phase 2)
- [ ] **スタッフ CRUD**
  - [ ] API: GET 一覧/詳細 / POST 作成 / PATCH 更新 / DELETE 削除 / PUT 権限グループ・所属クリニック設定
  - [ ] FE: `StaffSettings` 画面で全操作
  - [ ] 削除時の使用中チェック 409 (BUG-030 仕様)

### 2.6 rollback 判定基準

以下のいずれか 1 つでも該当すれば即座にロールバック発動。

- [ ] API エラーレート > **1%**
- [ ] ページロード > **5 秒**
- [ ] DB エラー (connection refused / migration failure / FK violation 異常頻度)
- [ ] CRUD smoke test の **NG が 1 件でも発生**
- [ ] security / env leak (CloudWatch logs にシークレット文字列出力)

ロールバック手順:

```bash
export AWS_PROFILE=AnimalEkarte

# 利用可能な Task Definition バージョン確認
aws ecs list-task-definitions \
  --family-prefix animalekarte-stg-api \
  --region us-east-1

# 前バージョンにロールバック (例: revision 4)
aws ecs update-service \
  --cluster animalekarte-stg-cluster \
  --service animalekarte-stg-service \
  --task-definition animalekarte-stg-api:4 \
  --region us-east-1
```

詳細: [DEPLOYMENT-CHECKLIST.md](../DEPLOYMENT-CHECKLIST.md) のロールバックセクション参照。

### 2.7 デプロイ後後始末

- [ ] **`DB_RESET=false` 戻し確認** (workflow_dispatch 方式なら不要 / `.env.staging` 方式なら必須)
- [ ] PR URL と commit hash を「結果記録」セクションに記録 (チケット/ Issue 管理側でも可)
- [ ] CRUD smoke test 結果を記録 (○/△/× + 備考)
- [ ] 問題発生時は memory `stg_deploy_<YYYYMMDD>.md` などで運用記録を残す

---

## 3. 完了条件

| # | 条件 |
|---|---|
| 1 | STG deploy 成功 (`backend-deploy.yml` workflow run conclusion=success かつ全 jobs success) |
| 2 | `/health` 200 + CloudWatch ERROR ゼロ (デプロイ後 10 分間監視) |
| 3 | CRUD smoke test 全 PASS (3 系統 18 API + 3 画面) |
| 4 | `DB_RESET=false` に戻した (workflow_dispatch 方式なら自動的に false 維持) |
| 5 | PR merge 完了 (main → staging) |

---

## 4. 結果記録テンプレート

デプロイのたびに以下を記録する (Issue / チケット / memory のいずれかに残す)。

```markdown
### STG デプロイ記録 — <YYYY-MM-DD>

**PR / commit:**
- PR URL: <url>
- merge commit: <sha>
- 対象 commits 範囲: <main の最初の commit> .. <main の最後の commit>
- backend-deploy.yml run ID: <run_id>
- デプロイ実行時刻 (JST): <timestamp>

**smoke test 結果:**
| 系統 | API | FE | 備考 |
|---|---|---|---|
| 医院 / クリニック | ○/△/× | ○/△/× | ... |
| 権限グループ | ○/△/× | ○/△/× | ... |
| スタッフ | ○/△/× | ○/△/× | ... |

**CloudWatch ERROR 監視 (デプロイ後 10 分):** ゼロ / N 件

**rollback 発動:** あり / なし
（あった場合は理由と発動時刻、戻し先 task definition revision を記録）

**問題点・runbook 追記事項:**
- ...
```

---

## 5. 関連リソース

| 種別 | 参照先 |
|---|---|
| smoke test 手順 | [docs/infra/deploy/CRUD-SMOKE-TEST.md](../CRUD-SMOKE-TEST.md) |
| デプロイ手順全般 | [docs/infra/deploy/DEPLOYMENT-CHECKLIST.md](../DEPLOYMENT-CHECKLIST.md) |
| CI/CD パイプライン | [docs/infra/deploy/CI-CD-PIPELINE.md](../CI-CD-PIPELINE.md) |
| インフラ概要 | [docs/infra/architecture.md](../../architecture.md) |
| backend-deploy workflow | `.github/workflows/backend-deploy.yml` |
| memory: checksum mismatch 復旧 | `ops_migration_checksum_mismatch.md` |
| memory: paths-filter 盲点 | `feedback_paths_filter_silent_green.md` |
| memory: 設定変更時の CI 波及 | `feedback_config_change_ci_propagation.md` |
| memory: Lstep Write API noop | `lstep_write_pause_20260515.md` |
