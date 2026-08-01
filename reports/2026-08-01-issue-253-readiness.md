# Issue #253 本番環境整備 — 準備状況監査レポート

| 項目 | 値 |
|---|---|
| 対象 Issue | #253（本番環境整備） |
| 監査日 | 2026-08-01 |
| 納品日（Go-live #257 前提） | 2026-08-03 |
| 監査 HEAD（着手時） | `cdad74da8` |
| 監査範囲 | repo 内の workflow / コード / docs の **実測のみ** |
| 非範囲 | deploy 実行、`/health` 実 URL 叩き、backup/restore/rollback 実施、billing 変更、secret 発行、GitHub 書き込み |

> **未実測境界（明示）**: 本レポートは対象環境の runtime、GitHub Actions の課金状態、Environment UI、実 backup/restore を一切実行していない。runtime green を主張しない。未実施の作業を実施済みと書かない。

---

## 1. 結論（1 枚要約）

| 観点 | 実測結果 |
|---|---|
| production インフラ | docs 上 **未構築**（setup.md が事前構築手順） |
| GitHub Environment `production` + Required reviewers | workflow 上 `environment:` **0 件**。UI 作成は USER 専権（未実測） |
| `backend-deploy.yml` production トリガー | **未適用**（`branches: [staging]` のみ） |
| `frontend-deploy.yml` production | **branch トリガーあり**、Environment ゲート無し |
| `/health` | `backend/cmd/api/base_routes.go:30` — `200` + `{"status":"ok"}`（DB 非依存） |
| PROD backup / restore / rollback 実行スクリプト | **不在**（#253 最大の repo 側 gap） |
| required status check 候補 | `ci.yml` 7 job。常時 required に **安全なのは `secret-scan`（と必要なら `changes`）のみ** |
| USER 作業 | 下記 **8 段** が直列クリティカルパス。agent は 1 段も実行しない |

関連正本:

- [docs/ops/infra/production/runbook.md](../docs/ops/infra/production/runbook.md) §0 / §3.1 / §5.1 / §8
- [docs/ops/infra/production/setup.md](../docs/ops/infra/production/setup.md) §7 / §8
- [docs/ops/deploy/CI-CD-PIPELINE.md](../docs/ops/deploy/CI-CD-PIPELINE.md) §0
- [docs/delivery/GOLIVE_RUNBOOK.md](../docs/delivery/GOLIVE_RUNBOOK.md) §1

---

## 2. 実測: `.github/workflows/`（9 ファイル）

```text
actionlint.yml
backend-deploy.yml
ci.yml
e2e.yml
frontend-deploy.yml
performance-tests.yml
security-scan.yml
stg-smoke.yml
worker-secret-sync.yml
```

| ファイル | trigger | jobs | production 関連 |
|---|---|---|---|
| `ci.yml` | PR → main/staging/production; push → main | `changes`, `secret-scan`, `backend`, `frontend`, `worker`, `codegen-check`, `migration-verify` | PR base に production 含む。`environment:` 無し |
| `backend-deploy.yml` | push → **staging only** + path; workflow_dispatch | `deploy` | **production トリガー無し**。STG WORKER_URL 固定 |
| `frontend-deploy.yml` | push → staging/**production** + path; workflow_dispatch | `deploy` | `refs/heads/production` で Vercel production。GitHub Environment 無し |
| `actionlint.yml` | PR paths `.github/workflows/**` | `actionlint` | なし |
| `e2e.yml` | workflow_dispatch | `e2e` | 手動のみ |
| `security-scan.yml` | PR main/staging/production; push main/staging; dispatch | `agentshield` | fail-on-findings は base main 時 |
| `performance-tests.yml` | schedule + dispatch | `load-test`, `lighthouse`, `summary` | ローカル Docker 系 |
| `stg-smoke.yml` | workflow_dispatch | `health` | STG 既定 URL |
| `worker-secret-sync.yml` | workflow_dispatch | `sync` | STG Worker 名固定 |

### 2.1 `ci.yml` job 名（`rg -n '^  [a-z0-9_-]+:$' .github/workflows/ci.yml`）

```text
4:  pull_request:          # on: trigger — job ではない
6:  push:                  # on: trigger — job ではない
33:  changes:
75:  secret-scan:
96:  backend:
234:  frontend:
317:  worker:
349:  codegen-check:
405:  migration-verify:
```

**required status check の候補 = 上記 7 job**（trigger キーは数えない）。

| job id | 表示名 (`name:`) | 常時 required 安全性 |
|---|---|---|
| `changes` | Detect changes | **比較的安全**（常時実行）。価値は低い |
| `secret-scan` | Gitleaks Secret Scan | **最も安全**（path/`if` 無し） |
| `backend` | Backend | **不安全** — path skip + **main 向け PR では常時 skip** |
| `frontend` | Frontend | **不安全** — 同上 |
| `worker` | Worker Tests | **不安全** — path skip |
| `codegen-check` | Codegen Sync | **不安全** — path skip |
| `migration-verify` | Migration Verify | **不安全** — main PR かつ migrations path のみ |

**paths-filter 罠**: `changes` は `dorny/paths-filter@v4`。依存 job が `skipped` のとき branch protection の required check が **永久 pending** になり得る（本 repo で過去に実害あり）。最終選定は USER（Actions 実挙動の観測が必要）。

### 2.2 backend-deploy 実測（staging only）

```text
on.push.branches: staging
environment: 無し
WORKER_URL: STG workers.dev 固定
```

setup.md §8 の提案 diff（`production` branch・`environment: ${{ github.ref_name }}`・`-c wrangler.production.jsonc`）は **未適用**。

---

## 3. 実測: `/health`

| 項目 | 値 |
|---|---|
| 登録 | `backend/cmd/api/base_routes.go:30` |
| 実装 | `c.JSON(http.StatusOK, gin.H{"status": "ok"})` |
| HTTP | 200 |
| Body | `{"status":"ok"}` |
| DB / 依存 | **無し**（プロセスが応答すれば ok） |
| 実 URL 叩き | **本 unit では未実施** |

```go
router.GET("/health", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
})
```

---

## 4. 実測: backup / restore / rollback tooling

### 4.1 `scripts/` 名検索

```text
ls scripts/ | rg -i 'backup|restore|rollback|deploy'
# → 0 件（NAME_MATCH_ZERO）
```

### 4.2 内容検索

```text
rg -ln 'pg_dump|pg_restore|planetscale|backup' scripts/ Makefile
# Makefile
# scripts/local-db-reset-contract.sh   … ローカル make reset 前の pg_dumpall のみ（restore 無し）
# scripts/verify_seed_matches_stg_dump_full.sh … 既存 dump との比較（取得/復元ではない）
# 他: report / 証跡フラグのみ
# pg_restore: 0 ヒット
```

### 4.3 `infra/scripts/`

| スクリプト | 役割 |
|---|---|
| `cf-run-migrate.sh` | migrate |
| `cf-crud-smoke.sh` | smoke |
| `cf-scheduler-ops.sh` | scheduler |
| `pscale-create-stg.sh` | STG DB **作成**（dump/restore ではない） |
| backup/restore | **無し** |

### 4.4 断定

| 能力 | 実行スクリプト |
|---|---|
| PROD/STG backup | **存在しない** |
| restore | **存在しない** |
| DB/env rollback | **存在しない**（CF 再デプロイは docs + wrangler/Actions） |
| deploy | **存在する**（`backend-deploy.yml` / `frontend-deploy.yml`） |

→ runbook §3.1 / §5.1 は **手順文書のみ**。tooling を伴わない事実が #253 最大の repo 側 gap。

---

## 5. USER 実行順 8 段（逐次・判定付き）

**いずれも USER 専権。agent は実行しない。credential / 実 URL の秘密値 / メール実値は書かない。**

| 段 | 前提 | 打つコマンドまたは操作 | 判定基準（次へ進む / 止まる） | 失敗時分岐 |
|---|---|---|---|---|
| **1. Actions billing / spending 復旧** | 組織/個人の課金権限がある | GitHub Billing / Spending limits UI で Actions を実行可能にする（agent 不可） | **次へ**: 空 PR または main 上で `ci.yml` が起動し `secret-scan` が success/failure まで到達する（即 fail の billing メッセージが消える）。**止まる**: 起動直後に billing で fail し続ける | 上限・支払い手段を再確認。復旧まで 2 以降の Actions 依存段を開始しない |
| **2. Environment `production` 作成** | 段 1 完了推奨（必須ではないが後続 Actions 確認が楽） | GitHub → Settings → Environments → New environment → 名前を厳密に `production`（setup.md §7） | **次へ**: Environment 一覧に `production` がある。**止まる**: 別名（`Production` 等）や repo secrets のみで代替しようとしている | 名前を `production` に作り直す（§8 workflow が `github.ref_name` と一致させるため） |
| **3. Required reviewers 設定** | 段 2 完了 | 同 Environment で Required reviewers を 1 名以上 | **次へ**: Protection rules に reviewers が表示される。**止まる**: reviewers 0 | レビュアを追加するまで段 5 の workflow 適用を延期（無保護 Environment 自動生成を避ける） |
| **4. required status check 指定** | 段 1 完了（check 名が Actions に現れること） | Branch protection（対象 branch は運用ポリシー。通常 main / production）で checks を指定 | **次へ**: 選定した check が「skipped で永久 pending にならない」ことを **docs-only 変更 PR 等で 1 回観測**。**止まる**: Backend/Frontend 等 path job を無条件 required にして merge が詰まる | **推奨初期セット**: `Gitleaks Secret Scan`（job id `secret-scan`）。任意で `Detect changes`。他 5 job は skip 集約 job 追加まで required にしない |
| **5. production deploy トリガー適用** | 段 2+3 完了（§7 → §8 の順） | setup.md §8 提案 diff を `.github/workflows/backend-deploy.yml` に手動反映し、通常の review 経路で main へ（**本 unit は未適用・workflow 変更禁止**） | **次へ**: ファイル上 `branches` に `production`、`environment: ${{ github.ref_name }}`、`-c wrangler.production.jsonc` 分岐がある。**止まる**: §7 未完了のまま適用 → 無保護 Environment リスク | §7 を完了させてから再適用。staging 挙動が変わったら diff を見直す |
| **6. deploy 実行と `/health`** | 段 5 完了、setup.md §1〜§6（DB/R2/DNS/secrets）完了、`APP_ENV=production` | `gh workflow run backend-deploy.yml --ref production` → Actions UI で承認 → 完了後 `curl -sS -o /tmp/prod-health.json -w '%{http_code}\n' https://api.noah-karte.com/health` と `jq -r '.status' /tmp/prod-health.json` | **次へ**: 承認ゲートが挟まり、deploy/migrate/health 成功、curl が **200** かつ **ok**。**止まる**: 承認無し完走 / health 非 200 / status 非 ok | 無承認完走 → 段 3 再確認。health 失敗 → runbook §2（15 分静置）。**ECS/AWS へ戻さない** |
| **7. backup → 隔離 restore** | 段 6 完了（または同等の prod データ存在）、隔離先用意 | runbook **§5.1** のコマンド列（PlanetScale マネージド or `pg_dump`/`pg_restore` 雛形）。本番 DNS/DB をターゲットにしない | **次へ**: スナップショット ≥1、隔離 restore 終了コード 0、非 PHI 件数/合計が記録、隔離資源破棄、所要時間記録。**止まる**: スナップショット 0 / 本番を向いている / PHI 混入 | 本番向き → **即中止**。0 本 → 先に取得。PHI → 結果破棄してクエリ修正。**repo に実行スクリプトは無い** |
| **8. rollback rehearsal** | 段 6 完了、last known good SHA 特定、非ピーク or 隔離枠 | runbook **§3.1** のコマンド列（再デプロイ + `/health` + RTO 記録） | **次へ**: 判断宣言〜 `/health` ok の分が記録され、データ破壊操作が無い。**止まる**: schema 非互換で経路無し / health 未回復 | forward-fix または承認済みスナップショット。provider 障害 → §3 手順 7。DNS 旧インフラ戻しを復旧とみなさない |

### 5.1 `/health` 成功条件（段 6・8 共通）

| 観測 | 成功 | 失敗 |
|---|---|---|
| HTTP status | `200` | それ以外 |
| JSON `.status` | `ok` | 欠落・別値 |
| 契約根拠 | `base_routes.go:30-32` | 実 URL は USER が構築後に確認 |

---

## 6. repo 側 gap 全数一覧

確認コマンドと結果（監査セッション実測）:

```text
# workflows 9
ls .github/workflows/*.yml
# → 9 ファイル（上記 §2）

# backup/restore/rollback 名
ls scripts/ | rg -i 'backup|restore|rollback|deploy'
# → 0 件

rg -ln 'pg_dump|pg_restore|planetscale|backup' scripts/ Makefile
# → local reset / seed verify / Makefile コメントのみ。pg_restore 無し

ls infra/scripts/
# → cf-crud-smoke.sh cf-run-migrate.sh cf-scheduler-ops.sh pscale-create-stg.sh validate-schema.sql
# → backup/restore 無し
```

| gap | 状態 | 影響 |
|---|---|---|
| GitHub Environment `production` | 未作成（docs + workflow に environment 無し） | Required reviewers ゲートが効かない |
| Required reviewers | 未設定 | 無承認 production deploy を止められない |
| `backend-deploy.yml` production トリガー | 未適用 | 本番 API の Actions 経路が無い |
| frontend の GitHub Environment ゲート | 無し（branch だけ production 可） | backend と承認運用が揃っていない |
| PROD backup 実行スクリプト | **不在** | §5.1 は外部 CLI / マネージド UI 依存 |
| restore 実行スクリプト | **不在** | 同上 |
| rollback 実行スクリプト | **不在** | CF 再デプロイは手運用 |
| required check の skip 集約 job | **不在** | path job を required にすると pending 罠 |
| CI green on main | docs 上 billing BLOCKED（UI 未実測） | #257 前提 #2 が止まる |
| PlanetScale prod / R2 / DNS / secrets | 未構築（setup.md） | 段 6 以前にインフラが必要 |

**agent 実装候補（todo 起票はしない。A1）**: PROD 向け backup/restore ラッパ（隔離先強制・非 PHI 検証・secret 非出力）を将来タスク化するなら、本 gap 表を起点にする。

---

## 7. runbook 同期内容（本 unit の成果）

| 節 | 変更内容 |
|---|---|
| §0 | 実測根拠列を追加。frontend production 経路・tooling 不在・workflow 実測を反映 |
| §3.1 | 前提・コマンド列・次へ/止まる判定・失敗分岐を追加。`[ ]` は維持 |
| §5.1 | 同上（隔離・非 PHI・破棄）。tooling 不在を明記 |
| §8 | docs/prep を実測成果物参照に更新。8 段の索引とレポートリンク |

**変更していないもの**: `- [ ]` のチェック完了化、`.github/workflows/**`、backend/frontend/infra コード、todo.md、secret 実値。

---

## 8. 安全境界（維持）

- rollback rehearsal: 非ピークまたは隔離枠。本番データ破壊操作を含めない
- restore: 隔離環境のみ。本番 DNS/DB を直接ターゲットにしない
- 整合性: 非 PHI 指標（件数・clinic_id 別件数・金額合計）。個人名禁止
- 記録に credential・アドレス実値を載せない
- agent は支払い・spending limit・secret 実値の発行を行わない
- ECS / AWS 切り戻し経路は **再導入禁止**（2026-07-20 廃止）

---

## 9. 参照コマンド（再検証用・docs-only）

```bash
ls .github/workflows/
rg -n '^  [a-z0-9_-]+:$' .github/workflows/ci.yml
rg -n 'environment:|branches:|production' .github/workflows/backend-deploy.yml .github/workflows/frontend-deploy.yml
rg -n '"/health"' backend/cmd/api/base_routes.go
ls scripts/ | rg -i 'backup|restore|rollback|deploy' || true
rg -ln 'pg_dump|pg_restore|planetscale|backup' scripts/ Makefile
bash scripts/check-docs-symbol-drift.sh
```
