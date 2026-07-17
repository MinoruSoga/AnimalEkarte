# infra/cloudflare/production/ — Cloudflare Terraform 基盤(Production)

> STG 版は [`../README.md`](../README.md)。安全ルール・認証手順・実行フローは同一のため
> ここでは再掲しない(重複管理を避ける)。人間の実施手順の正本は
> [`docs/ops/deploy/PRODUCTION_CF_SETUP.md`](../../../docs/ops/deploy/PRODUCTION_CF_SETUP.md)。
> 追跡Issue: #253。PO決定: migration-cloudflare.md「現況サマリ」2026-07-15 ブロック参照。

## STGとの違い

- **ゾーンは新規作成しない**。`noah-karte.com` は STG 側(`infra/cloudflare/zone.tf`)が
  既に管理しているため、本ディレクトリでは `data "cloudflare_zone"` で参照するのみ
  (`zone.tf` 参照)。
- **通知ポリシーは新規作成しない**。ゾーンレベルの5xxアラートはSTG側の既存ポリシーが
  同一ゾーンを既にカバーするため、production専用ポリシーを追加すると二重通知になる
  (`notifications.tf` 参照)。
- **tfstate は分離**。このディレクトリで `terraform init` すると、STGとは別の
  local tfstate(`infra/cloudflare/production/terraform.tfstate`)が作られる。
- **R2バケット・Hyperdrive・PlanetScale DB はすべて新規**(STGと共有しない)。

## ディレクトリ構成

```
infra/cloudflare/production/
├── providers.tf      # cloudflare provider(~> 5.21、STGと同一バージョン)
├── variables.tf       # account_id / zone_name / environment / r2_bucket_name / pscale_prod_db_*
├── backend.tf          # 当面 local backend
├── zone.tf             # data "cloudflare_zone"(既存ゾーン参照)+ api.noah-karte.com DNSレコード新規
├── r2.tf                # cloudflare_r2_bucket(production専用: animalekarte-prod-images)
├── hyperdrive.tf         # cloudflare_hyperdrive_config(production専用。未使用予約。要フォローアップ判断)
└── notifications.tf       # リソース定義なし(理由をコメントで記録。上記参照)
```

## 検証について

`terraform init -backend=false` / `terraform validate` はこのドラフト作成時に実行済み
(認証不要)。`terraform plan` / `terraform apply` は Cloudflare認証情報が必要なため未実行。
実行手順は `docs/ops/deploy/PRODUCTION_CF_SETUP.md` を参照すること。
