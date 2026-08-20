# CI/CD パイプライン構成書

> **目的**: 自動デプロイ、手動トリガー、障害時の確認経路、本番承認ゲートを定義する。
> **読者**: 運用者・新規参加者。
> **最新更新**: 2026-07-31（#253 production delivery contract）
>
> Backend の正系統は Cloudflare Workers + Containers、Frontend は Vercel。
> 旧 AWS ECS/RDS 経路と関連 workflow は 2026-07-20 に廃止済み。**ロールバック先は Cloudflare のみ**（#99 residual: ECS へ戻さない）。

## 0. Deployment contract（#253 正本）

| 経路 | トリガー | 承認 | 到達先 |
|---|---|---|---|
| **STG 自動** | `main` の変更が `staging` へ PR マージされた結果の `staging` push（対象 path） | 不要（自動） | STG Cloudflare Worker/Container + Vercel preview |
| **Production** | `production` ブランチへの対象 path push、または `workflow_dispatch` で production を指定 | **GitHub Environment `production` の Required reviewers 必須** | 本番 Cloudflare (`api.noah-karte.com`) + Vercel production |
| **main 単独 push** | CI（`ci.yml` 等）のみ | n/a | **本番へはデプロイしない**（acceptance 禁止） |

### 0.1 契約の不変条件

1. **STG は `main` 由来のコードが `staging` に載った時点で自動デプロイされる**（人間の deploy 承認は不要）。日常開発は `main`、検証は `staging`（project branch 規約）。
2. **production は Required reviewers 無しでは開始できない**。main push から無承認 production deployment を acceptance にしない。
3. **rollback は last known good の Cloudflare 再デプロイ + migration 互換確認**のみ。AWS/ECS への切り戻し・hot standby は存在しない（#99）。
4. **secret / credential / PHI を log・artifact・Issue に出さない**。値の生成・登録は人間のみ（#89 依存）。
5. **CI green は GitHub Actions billing/spending 復旧が前提**。agent は支払い・上限変更を実行しない（§7）。

### 0.2 現状（実装 vs 契約）

| 項目 | 現状 | 契約上の次アクション |
|---|---|---|
| STG backend auto-deploy | ✅ `backend-deploy.yml` が `staging` push で稼働 | 維持 |
| STG frontend auto-deploy | ✅ `frontend-deploy.yml` が `staging` push で稼働 | 維持 |
| Production backend workflow | ⚠ 未適用（`setup.md` §8 提案 diff） | Environment 作成後に workflow 適用 |
| Production Environment + Required reviewers | ⚠ 未作成（`setup.md` §7） | USER が GitHub 設定 |
| Production frontend | ⚠ `production` push で Vercel デプロイ可能だが Environment ゲート無し | Environment 保護を backend と揃える |
| ECS workflow | ✅ repository に残存なし | 再導入禁止 |
| CI green on latest main | ❌ billing/spending limit で job が即 failure | USER billing 復旧（§7） |

本番構築の人間手順は [`../infra/production/setup.md`](../infra/production/setup.md)、稼働後の日常運用は [`../infra/production/runbook.md`](../infra/production/runbook.md) を正本とする。

---

## 1. 全体フロー

| コンポーネント | 実行環境 | デプロイ方式 | 自動トリガー | Workflow |
|---|---|---|---|---|
| Backend API (STG) | Cloudflare Workers + Containers | `wrangler deploy` + migrate one-shot + `/health` polling | `staging` への対象 path push | `.github/workflows/backend-deploy.yml` |
| Backend API (PROD) | Cloudflare Workers + Containers | `wrangler deploy -c wrangler.production.jsonc` + migrate + `/health` | `production` push（**Environment 承認後**・workflow 適用後） | 同上（`setup.md` §8） |
| Frontend (STG/PROD) | Vercel | Vercel CLI | `staging` / `production` への対象 path push | `.github/workflows/frontend-deploy.yml` |

現在のインフラ構成と運用手順は [`../infra/README.md`](../infra/README.md) を正本とする。STG 移行の実施記録は git 履歴。

---

## 2. Backend パイプライン

### 2.1 実行ステップ（STG / 将来 PROD 共通）

1. Checkout
2. pnpm / Node.js setup と依存関係の取得
3. `CLOUDFLARE_API_TOKEN` の存在確認と `wrangler whoami`
4. `backend/` を working directory として `npx wrangler deploy`（PROD は `-c wrangler.production.jsonc`）
5. `MIGRATE_RUN_SECRET` を使った `infra/scripts/cf-run-migrate.sh`
6. `/health` が HTTP 200 かつ `status: ok` になるまで polling
7. 認証情報が設定されている場合だけ `infra/scripts/cf-crud-smoke.sh`

deploy 直後から migration 完了まで（最大 `MIGRATE_TIMEOUT=150s`）新 binary が旧 schema へ到達し得る制約は、
現行 Cloudflare 構成の既知・受容済み制約である。workflow は migration を health check より前へ置き、
この区間を最小化する。詳細は `.github/workflows/backend-deploy.yml` の契約コメントを正本とする。

### 2.2 手動実行と障害対応

```bash
# STG
gh workflow run backend-deploy.yml --ref staging

# PROD（workflow 適用 + Environment secrets 登録後）
# Required reviewers 承認待ちでジョブが一時停止する
gh workflow run backend-deploy.yml --ref production
```

- 失敗した job を成功扱いにせず、deploy / migration / health / smoke のどこで失敗したかを切り分ける
- DB reset、credential 変更、production 操作は別途明示承認を得る
- STG の停止・復旧は [`../infra/staging/runbook.md`](../infra/staging/runbook.md)
- PROD の停止・復旧・rollback は [`../infra/production/runbook.md`](../infra/production/runbook.md)（**CF-only**）

### 2.3 Production 承認ゲート（GitHub Environment）

1. Repository → Settings → Environments → `production` を作成（名前一致必須）
2. **Required reviewers** を 1 名以上設定する（production-impacting action の物理ゲート）
3. Environment secrets に production 専用値のみ登録する（STG の Repository secrets を流用しない）
4. `backend-deploy.yml` の job に `environment: ${{ github.ref_name }}` を付与する（`setup.md` §8）

Environment 未作成のまま production トリガーだけを足すと、保護ルール無しの自動生成 Environment になり得る。
**workflow 適用順序は setup.md §7 → §8 を厳守**する。

---

## 3. Frontend パイプライン

`.github/workflows/frontend-deploy.yml` が Vercel CLI を GitHub Actions 上で実行する。
Vercel の native Git 連携 hook は使用しない。

1. `staging` / `production` への対象 path push、または `workflow_dispatch` を検知
2. `vercel pull` で対象 environment の情報を取得
3. `vercel build` で artifact を生成
4. `vercel deploy --prebuilt` で配布

手動実行時は workflow の `environment` input で `preview` または `production` を選ぶ。
production 実行は外部変更として明示承認を必要とする（backend の Required reviewers と運用を揃えること）。

---

## 4. Rollback（CF-only・#99 一本化）

| やってよい | やってはいけない |
|---|---|
| 直前 green commit を特定し、schema 互換を確認したうえで Cloudflare へ再 `wrangler deploy` | AWS ECS/RDS への切り戻し・旧 workflow の復活 |
| migration 非互換なら expand 済み前提で forward-fix、または承認済みスナップショットから再建 | DNS/NS を「旧インフラ」へ戻して復旧したとみなす |
| Terraform 差分は `plan` レビュー後に明示承認して `apply` | secret を log / artifact に出しての緊急再投入 |
| 現場を Access 等の旧業務へ一時退避（業務継続）し、CF 正系統を復旧 | 「ECS hot standby がある」前提の判定 |

手順の詳細は:

- STG: [`../infra/staging/runbook.md`](../infra/staging/runbook.md)
- PROD: [`../infra/production/runbook.md`](../infra/production/runbook.md)
- Go-live 当日の切り戻し判断: [`../../delivery/GOLIVE_RUNBOOK.md`](../../delivery/GOLIVE_RUNBOOK.md) §4

---

## 5. Monitoring / failure notification（秘密値なし）

| 監視対象 | 手段 | 備考 |
|---|---|---|
| API 生存 | `GET /health` → 200 `{"status":"ok"}` | STG: `api.stg.noah-karte.com` / PROD: `api.noah-karte.com` |
| 5xx 率 | Cloudflare Notification Policy（ゾーン全体） | STG 側ポリシーが `noah-karte.com` をカバー。PROD 専用ポリシーは二重通知になるため追加しない（`setup.md` §10.2） |
| Deploy 失敗 | GitHub Actions run failure | 失敗 step（deploy / migrate / health / smoke）を切り分け |
| Workers / Containers | Cloudflare Observability / Workers Logs | PHI・credential をログに出さない |
| Frontend | Vercel deployment status | production は承認後のみ |

通知先メール等の実値は文書に書かない。登録・検証は人間がダッシュボードで行う。

---

## 6. Backup / restore gate（秘密値なし）

本番受け入れ（#253）に必要な backup 契約。**値・接続文字列は書かない**。

| 項目 | 契約 | 検証方法（非 PHI） |
|---|---|---|
| RPO 目標 | 当日 Go-live 前スナップショット + 以降の運用ポリシー（確定待ち） | スナップショット取得時刻の記録 |
| RTO 目標 | Go-live 当日の rollback rehearsal で実測して記録 | 復旧開始〜`/health` ok までの分 |
| 取得 | PlanetScale 側バックアップ / 明示 `pg_dump`（当日ランブック） | ファイル存在・サイズ・取得時刻のみ記録 |
| 復元 rehearsal | **隔離環境**へ restore（本番上書き禁止） | テーブル件数・clinic_id 別件数など非 PHI 指標 |
| 失敗時 | restore 失敗を成功扱いにせず、原因をインシデント記録 | credential は記録しない |

詳細チェックリストは [`../infra/production/runbook.md`](../infra/production/runbook.md) §監視・バックアップ。

---

## 7. Security

- Cloudflare の credential、DB 接続情報、migration secret は Cloudflare Secrets または GitHub Encrypted Secrets / Environment secrets で管理する
- Vercel token、org ID、project ID は GitHub Encrypted Secrets で管理する
- secret を workflow YAML、repository、log へ直接記載しない
- workflow の権限は job に必要な最小権限へ限定する
- production 用 `JWT_SECRET` / `MIGRATE_RUN_SECRET` / `INTEGRATION_ENCRYPTION_KEY` は STG と共有しない（`wrangler.production.jsonc` / `setup.md` §5）

---

## 8. BLOCKED residual — GitHub Actions billing（USER only）

#253 受け入れ条件「latest main の required CI が green」は、**コード修正だけでは達成できない**。

| 観測 | 内容 |
|---|---|
| 現象 | 最新 main の CI / Security Scan が job 開始直後に failure（steps 空） |
| 原因区分 | GitHub account の payment failure または spending limit（#253 本文 P0） |
| agent がやらないこと | 支払い手段変更、spending limit 変更、課金操作 |
| USER 復旧手順 | 1. GitHub Billing / Spending limits を安全に確認・復旧 2. 最新 `main` で Backend / Frontend / Migration / Security を含む必須 job green を確認 3. その後 STG deploy・health・failure notification を実地確認 |

本ドキュメントと production runbook の整備は **delivery surface prep** であり、billing 復旧後の green CI と実地 rehearsal が残る。status は **PARTIAL**（docs prep COMPLETE + CI green BLOCKED）として扱う。
