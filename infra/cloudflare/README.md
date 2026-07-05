# infra/cloudflare/ — Cloudflare Terraform 基盤

> STG の Cloudflare 移行(移行計画: [`../../migration-cloudflare.md`](../../migration-cloudflare.md))で使う Terraform 構成。
> 既存 `infra/terraform/`(AWS)とは完全に分離する。

## 安全ルール(`infra/CLAUDE.md` の AWS 向けルールをそのまま踏襲)

1. **`terraform apply` は明示承認後のみ実行する。** Claude Code は apply を自動実行しない。
   `terraform plan` までを準備し、内容をレビューした上で承認者自身が `apply` するか、
   明示的な apply 承認を得てから実行する。
2. **`-target` plan は依存切り分けのデバッグ用途のみ。** 承認判断には full plan を使う。
3. **CLOUDFLARE_API_TOKEN / CLOUDFLARE_ACCOUNT_ID はコミットしない。** 環境変数、または
   ローカル専用の `terraform.tfvars`(`.gitignore` 対象)で供給する。
4. NS 切替(レジストラ側のネームサーバ変更)は本 Terraform の管理対象外。人手で実施し、
   実施前後の状態を `migration-cloudflare.md` に記録する。

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
├── providers.tf   # cloudflare provider(~> 5.21)
├── variables.tf    # account_id / zone_name / environment / r2_bucket_name / pscale_stg_db_*
├── backend.tf      # 当面 local backend。R2 backend 切替は TODO コメント参照
├── zone.tf          # P1-1: cloudflare_zone + cloudflare_dns_record(棚卸し済み)
├── r2.tf            # P2-1: cloudflare_r2_bucket(BLOCKED — CLOUDFLARE_API_TOKEN 未設定のため未apply)
└── hyperdrive.tf    # P3-4: cloudflare_hyperdrive_config(BLOCKED — 同上 + pscale_stg_db_* 未供給)
```

## 関連スクリプト(`infra/scripts/`)

| スクリプト | 用途 |
|---|---|
| `pscale-create-stg.sh` | P3-1: PlanetScale Postgres(animalekarte-stg) 作成手順 |
| `validate-schema.sql` | P3-2/P3-3: migration 適用状況・拡張機能・ENUM/text[]/jsonb/trigger 検証 |
| `migrate-images-r2.sh` | P2-4: S3→R2 データ移行(雛形。AWS+R2 双方の認証情報が必要なため未実行) |
