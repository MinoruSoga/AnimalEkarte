# infra/cloudflare/ — Cloudflare Terraform 基盤

> 現行 STG の Cloudflare Terraform 構成。構成の正本は
> [`docs/ops/infra/architecture.md`](../../docs/ops/infra/architecture.md)、運用規約は
> [`docs/ops/infra/iac-guidelines.md`](../../docs/ops/infra/iac-guidelines.md)。
> AWS 基盤は 2026-07-20 に廃止済みで、旧 `infra/terraform/` は存在しない。
> 移行経緯は [`migration-cloudflare.md`](../../docs/ops/infra/_archive/migration-cloudflare.md) の凍結履歴を参照する。

## 安全ルール（`infra/CLAUDE.md` / `docs/ops/infra/iac-guidelines.md`）

1. **`terraform apply` は明示承認後のみ実行する。** Claude Code は apply を自動実行しない。
   `terraform plan` までを準備し、内容をレビューした上で承認者自身が `apply` するか、
   明示的な apply 承認を得てから実行する。
2. **`-target` plan は依存切り分けのデバッグ用途のみ。** 承認判断には full plan を使う。
3. **CLOUDFLARE_API_TOKEN / CLOUDFLARE_ACCOUNT_ID はコミットしない。** 環境変数、または
   ローカル専用の `terraform.tfvars`(`.gitignore` 対象)で供給する。
4. NS 切替(レジストラ側のネームサーバ変更)は本 Terraform の管理対象外。人手で実施し、
   実施前後の状態を運用記録へ残す。

## 前提ツール

| ツール | 用途 | インストール |
|---|---|---|
| `terraform` | Cloudflare リソース管理(zone/DNS/R2/notifications) | Terraform CLI v1.5+ |
| `wrangler` | Workers/Containers/Secrets/Cron | ルート `package.json` devDependency(`pnpm install` で導入。バージョンは固定管理) |
| `pscale` | PlanetScale DB 作成・接続・dump/restore | `brew install planetscale/tap/pscale` |
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
├── variables.tf      # account_id / zone_name / environment / r2_bucket_name / notification_email
├── backend.tf        # backend block なし(Terraform 既定の local state)。R2 backend 切替は TODO コメント参照
├── zone.tf           # P1-1: cloudflare_zone + cloudflare_dns_record(棚卸し済み)
├── r2.tf             # P2-1: cloudflare_r2_bucket(apply 済み: animalekarte-stg-images)
├── hyperdrive.tf     # SEC-CS2-F03: resource 削除済み(tombstone + USER ops のみ)
└── notifications.tf  # P6-3: cloudflare_notification_policy(http_alert_edge_error。apply 未実施 — TF_VAR_notification_email 未供給かつ送信先メール事前検証(要確認)のため genuine BLOCKED)
```

### Hyperdrive 削除 (SEC-CS2-F03) — USER-only 運用

Runtime(Container Go API)は PlanetScale へ `DB_*` secrets で直結する。Hyperdrive は
Container 非対応のため未使用であり、origin に DB 資格情報を載せる Terraform 定義を削除した。
`backend/wrangler.jsonc` の `hyperdrive` バインディングも除去済み。

**エージェントは以下を実行しない。** 既存リソース/旧 state が残っている場合の片付けは人間のみ:

1. このディレクトリで `terraform plan` を確認し、state に `cloudflare_hyperdrive_config.stg_planetscale` が残っていれば destroy 差分が出る。
2. 明示承認後のみ `terraform apply`（または Cloudflare ダッシュボード / `wrangler hyperdrive delete` で当該 config を削除）。
3. Hyperdrive origin パスワードを一度でも載せた local `terraform.tfstate` はコミットせず、破棄時は慎重に処分する（`cat`/ログ貼付禁止）。
4. Hyperdrive 専用に使った PlanetScale パスワードがあれば `pscale role reset-default` 等でローテーションする。
5. App 本体の `DB_*` Worker secrets とは別経路。誤って本番アプリ用パスワードを巻き込まないこと。

### P6-3 通知ポリシー apply の前提（`notifications.tf`）

- 送信先メールアドレスは `TF_VAR_notification_email` で供給する（`terraform.tfvars` に書かない。値未供給時は `terraform plan` が変数未設定エラーで失敗する — 意図した genuine BLOCKED）。
- Cloudflareの通知メール送信先はダッシュボードの Notification Settings で確認リンク経由の事前検証が必要な可能性がある（本リポジトリでは未検証。`terraform plan` はこの制約を検出できないため、`apply` 前に運用担当者が送信先アドレスの検証状態を確認すること）。
- `http_alert_edge_error` はゾーン全体のエッジ観測5xx率に基づく代替指標であり、Worker/Containers専用のalert_typeはCloudflare通知APIに存在しない（判断経緯は [`migration-cloudflare.md`](../../docs/ops/infra/_archive/migration-cloudflare.md) P6-3参照）。

## CI デプロイ

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

AWS への切り戻し workflow は存在しない。障害時は
[`docs/ops/infra/staging/runbook.md`](../../docs/ops/infra/staging/runbook.md) に従う。

## CI デプロイ検証

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
Secrets 投入前は **BLOCKED（genuine）** として扱い、実行しない。

## 関連スクリプト(`infra/scripts/`)

| スクリプト | 用途 |
|---|---|
| `pscale-create-stg.sh` | P3-1: PlanetScale Postgres(animalekarte-stg) 作成手順 |
| `validate-schema.sql` | P3-2/P3-3: migration 適用状況・拡張機能・ENUM/text[]/jsonb/trigger 検証 |
