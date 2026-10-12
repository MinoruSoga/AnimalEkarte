# production 版。STG(../variables.tf)と対をなすが、値は完全に独立させる
# (同じ変数名でも STG の tfvars/環境変数を production ディレクトリで誤って再利用しないこと。
# cwdを分けても環境変数は継承される。apply前に対象account/zoneとTF_VAR_*が当該環境の承認済み設定であることを人が確認する)。

variable "account_id" {
  description = "Cloudflare Account ID。【2026-10-06 確定 EMR-298】production は専用アカウント（EMR-302 移行先）を指す。共有アカウントへの本番構築は禁止（docs/ops/infra/cloudflare-account-separation.md）。旧前提「STGと同一アカウント(migration-cloudflare.md の課金・管理の一元化)」は廃止。tfvars に書かず -var か環境変数 TF_VAR_account_id で供給する"
  type        = string
  sensitive   = true
}

variable "zone_name" {
  description = "Cloudflare ゾーン名。STGと同一ゾーン(noah-karte.com)を production でも使う(zone.tf の data source 参照)"
  type        = string
  default     = "noah-karte.com"
}

variable "environment" {
  description = "リソース名・タグに使う環境名"
  type        = string
  default     = "prod"
}

variable "r2_bucket_name" {
  description = "臨床画像用 R2 バケット名(production専用。STGバケットとは分離)"
  type        = string
  default     = "animalekarte-prod-images"
}

# SEC-CS2-F03: pscale_prod_db_* variables removed with Hyperdrive. App DB credentials
# are Worker secrets (DB_HOST/DB_USER/DB_PASSWORD), not Terraform inputs.

# 【notifications.tf 参照】production 専用の通知ポリシーは意図的に作成しない
# (STGのゾーンレベル5xxアラートが同一ゾーンをカバーするため、追加すると二重通知になる)。
# そのため notification_email 変数は本ファイルでは定義しない。
