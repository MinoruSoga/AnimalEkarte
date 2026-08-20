# infra/cloudflare/production/ — Cloudflare Terraform 基盤(Production)

> STG 版は [`../README.md`](../README.md)。安全ルール・認証手順・実行フローは同一のため
> ここでは再掲しない(重複管理を避ける)。人間の実施手順の正本は
> [`../../../docs/ops/infra/production/setup.md`](../../../docs/ops/infra/production/setup.md)。
> 追跡Issue: #253。現行構成は [`docs/ops/infra/architecture.md`](../../../docs/ops/infra/architecture.md)。

## STGとの違い

- **ゾーンは新規作成しない**。`noah-karte.com` は STG 側(`infra/cloudflare/zone.tf`)が
  既に管理しているため、本ディレクトリでは `data "cloudflare_zone"` で参照するのみ
  (`zone.tf` 参照)。
- **通知ポリシーは新規作成しない**。ゾーンレベルの5xxアラートはSTG側の既存ポリシーが
  同一ゾーンを既にカバーするため、production専用ポリシーを追加すると二重通知になる
  (`notifications.tf` 参照)。
- **tfstate は分離**。このディレクトリで `terraform init` すると、STGとは別の
  local tfstate(`infra/cloudflare/production/terraform.tfstate`)が作られる。
- **R2バケット・PlanetScale DB はすべて新規**(STGと共有しない)。Hyperdrive は
  SEC-CS2-F03 で Terraform 定義ごと削除（未使用予約を廃止）。

## ディレクトリ構成

```
infra/cloudflare/production/
├── providers.tf      # cloudflare provider(~> 5.21、STGと同一バージョン)
├── variables.tf       # account_id / zone_name / environment / r2_bucket_name
├── backend.tf          # backend block なし(Terraform 既定の local state)
├── zone.tf             # data "cloudflare_zone"(既存ゾーン参照)+ api.noah-karte.com DNSレコード新規
├── r2.tf                # cloudflare_r2_bucket(production専用: animalekarte-prod-images)
├── hyperdrive.tf         # SEC-CS2-F03: resource 削除済み(tombstone + USER ops のみ)
└── notifications.tf       # リソース定義なし(理由をコメントで記録。上記参照)
```

### Hyperdrive 削除 (SEC-CS2-F03) — USER-only 運用

Production でも Container は Hyperdrive 非対応・DB 直結。未使用の
`cloudflare_hyperdrive_config` と `pscale_prod_db_*` 変数、
`backend/wrangler.production.jsonc` の `hyperdrive` バインディングを削除した。

**エージェントは以下を実行しない。** 万一 prior apply で作成済みなら人間のみ:

1. このディレクトリで `terraform plan` を確認し、state に
   `cloudflare_hyperdrive_config.prod_planetscale` が残っていれば destroy 差分が出る。
2. 明示承認後のみ `terraform apply`（またはダッシュボード / `wrangler hyperdrive delete`）。
3. origin 資格情報を含み得る local `terraform.tfstate` はコミットせず慎重に処分する。
4. Hyperdrive 専用 PlanetScale パスワードがあればローテーションする。

## 検証について

`terraform init -backend=false` / `terraform validate` はこのドラフト作成時に実行済み
(認証不要)。`terraform plan` / `terraform apply` は Cloudflare認証情報が必要なため未実行。
実行手順は `../../../docs/ops/infra/production/setup.md` を参照すること。
