# infra/cloudflare/ — Cloudflare Terraform 基盤

> STG の Cloudflare 移行(移行計画: [`../../../../docs/ops/infra/_archive/migration-cloudflare.md`](../../../../docs/ops/infra/_archive/migration-cloudflare.md))で使う Terraform 構成。
> 既存 `infra/terraform/`(AWS)とは完全に分離する。

## 安全ルール(`infra/CLAUDE.md` の AWS 向けルールをそのまま踏襲)

1. **`terraform apply` は明示承認後のみ実行する。** Claude Code は apply を自動実行しない。
   `terraform plan` までを準備し、内容をレビューした上で承認者自身が `apply` するか、
   明示的な apply 承認を得てから実行する。
2. **`-target` plan は依存切り分けのデバッグ用途のみ。** 承認判断には full plan を使う。
3. **CLOUDFLARE_API_TOKEN / CLOUDFLARE_ACCOUNT_ID はコミットしない。** 環境変数、または
   ローカル専用の `terraform.tfvars`(`.gitignore` 対象)で供給する。
4. NS 切替(レジストラ側のネームサーバ変更)は本 Terraform の管理対象外。人手で実施し、
   実施前後の状態を `../../docs/ops/infra/_archive/migration-cloudflare.md` に記録する。

## 前提ツール

| ツール | 用途 | インストール |
|---|---|---|
| `terraform` | Cloudflare リソース管理(zone/DNS/R2/Hyperdrive) | 既存(`infra/terraform/` と共用、v1.5+) |
| `wrangler` | Workers/Containers/Secrets/Cron | ルート `package.json` devDependency(`pnpm install` で導入。バージョンは固定管理) |
| `pscale` | PlanetScale DB 作成・接続・dump/restore | `brew install planetscale/tap/pscale` |
| `rclone` | S3 → R2 データ移行 | `brew install rclone` |
| `cf-terraforming` | 既存ゾーンの逆生成(取り込み用途) | `brew install cloudflare/cloudflare/cf-terraforming` |

## 認証

```bash
# 推奨: infra/cloudflare/.env.staging（gitignore 済み）を source
set -a && source infra/cloudflare/.env.staging && set +a
# 必須キー: CLOUDFLARE_API_TOKEN, TF_VAR_account_id, ZONE_ID

# または個別 export（値はコミット・ログ出力しない）
export CLOUDFLARE_API_TOKEN=...
export TF_VAR_account_id=$CLOUDFLARE_ACCOUNT_ID
export ZONE_ID=...

# wrangler 用
pnpm exec wrangler whoami   # 未ログインなら pnpm exec wrangler login

# pscale 用(サービストークンが発行済みでなければブラウザ OAuth)
pscale auth login
pscale auth check
```

## 実行フロー

```bash
cd infra/cloudflare
terraform init
terraform validate
terraform plan -out=tfplan
# ここで plan 内容をレビュー・承認を得る
terraform apply tfplan   # 承認後、承認者自身または明示承認を得てから実行
```

## ディレクトリ構成

```
infra/cloudflare/
├── providers.tf      # cloudflare provider(~> 5.21)
├── variables.tf      # account_id / zone_name / environment / r2_bucket_name / pscale_stg_db_* / notification_email
├── backend.tf        # 当面 local backend。R2 backend 切替は TODO コメント参照
├── zone.tf           # P1-1: cloudflare_zone + cloudflare_dns_record(棚卸し済み)
├── r2.tf             # P2-1: cloudflare_r2_bucket(apply 済み: animalekarte-stg-images)
├── hyperdrive.tf     # P3-4: cloudflare_hyperdrive_config(apply 済み)
└── notifications.tf  # P6-3: cloudflare_notification_policy(http_alert_edge_error。apply 未実施 — TF_VAR_notification_email 未供給かつ送信先メール事前検証(要確認)のため genuine BLOCKED)
```

### P6-3 通知ポリシー apply の前提（`notifications.tf`）

- 送信先メールアドレスは `TF_VAR_notification_email` で供給する（`terraform.tfvars` に書かない。値未供給時は `terraform plan` が変数未設定エラーで失敗する — 意図した genuine BLOCKED）。
- Cloudflareの通知メール送信先はダッシュボードの Notification Settings で確認リンク経由の事前検証が必要な可能性がある（本リポジトリでは未検証。`terraform plan` はこの制約を検出できないため、`apply` 前に運用担当者が送信先アドレスの検証状態を確認すること）。
- `http_alert_edge_error` はゾーン全体のエッジ観測5xx率に基づく代替指標であり、Worker/Containers専用のalert_typeはCloudflare通知APIに存在しない（詳細は `../../docs/ops/infra/_archive/migration-cloudflare.md` P6-3参照）。P1-2(NS切替)完了までシグナルは発生しない。

## CI デプロイ (Phase 5 / P5-2)

GitHub Actions `.github/workflows/backend-deploy.yml` が `staging` push で Cloudflare へデプロイする。
以下を **Repository Secrets** に登録する（値はコミットしない。人間タスク）:

| Secret | 用途 | スコープ目安 |
|---|---|---|
| `CLOUDFLARE_API_TOKEN` | `wrangler deploy` | Account: Workers Scripts Edit, Workers Containers Edit, Account Settings Read |
| `MIGRATE_RUN_SECRET` | `POST /_internal/migrate` 認証 | `wrangler secret put` 投入時と同一値 |
| `STG_DEMO_EMAIL` | 任意: deploy 後 CRUD smoke | demo system_admin |
| `STG_DEMO_PASSWORD` | 任意: deploy 後 CRUD smoke | 同上 |

登録コマンド例（値はプレースホルダ。実値を貼り付けてから実行し、シェル履歴に残さないよう先頭にスペースを入れるかヒアドキュメントを使うこと）:

```bash
# CLOUDFLARE_API_TOKEN: デプロイ専用の最小スコープトークンを別途発行して使うこと
# （STG検証で使った統合トークンをそのままCIに登録しない）
gh secret set CLOUDFLARE_API_TOKEN --repo <owner>/<repo>
# MIGRATE_RUN_SECRET: wrangler secret put 投入時と同一の値
gh secret set MIGRATE_RUN_SECRET --repo <owner>/<repo>
# 任意（P5-4 optional smoke 用）
gh secret set STG_DEMO_EMAIL --repo <owner>/<repo>
gh secret set STG_DEMO_PASSWORD --repo <owner>/<repo>
```

ECS ロールバック用: `.github/workflows/backend-deploy-ecs.yml`（`workflow_dispatch` のみ）。

## CI デプロイ検証 (Phase 5 / P5-5)

上記 Secrets 投入後、通常デプロイとマイグレーション込みデプロイが 2 回連続で成功することを確認する（人間承認後に実行。External write境界）:

```bash
gh workflow run backend-deploy.yml --ref staging
# 完了を待って結果確認
gh run list --workflow=backend-deploy.yml --branch=staging --limit 1
gh run view <run-id> --log-failed   # 失敗時のみ

# 直後にもう一度実行し、migrate ステップが冪等（exit 0）で成功することを確認
gh workflow run backend-deploy.yml --ref staging
```

Secrets 未設定の状態で実行すると `Verify Cloudflare credentials` / `Run database migration` ステップが
`::error::...secret is not configured` で明示的に fail する（サイレント成功はしない設計）。
P5-5 は Secrets 投入前は **BLOCKED（genuine）** として `../../docs/ops/infra/_archive/migration-cloudflare.md` に記録し、実行しない。

## 関連スクリプト(`infra/scripts/`)

| スクリプト | 用途 |
|---|---|
| `pscale-create-stg.sh` | P3-1: PlanetScale Postgres(animalekarte-stg) 作成手順 |
| `validate-schema.sql` | P3-2/P3-3: migration 適用状況・拡張機能・ENUM/text[]/jsonb/trigger 検証 |
| `migrate-images-r2.sh` | P2-4: S3→R2 データ移行(雛形。AWS+R2 双方の認証情報が必要なため未実行) |
