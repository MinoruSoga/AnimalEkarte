# production 版。STG(../variables.tf)と対をなすが、値は完全に独立させる
# (同じ変数名でも STG の tfvars/環境変数を production ディレクトリで誤って再利用しないこと。
# ディレクトリが分かれているため `terraform apply` の cwd を間違えない限り混線しない)。

variable "account_id" {
  description = "Cloudflare Account ID。STGと同一アカウントを前提とする(migration-cloudflare.md の「課金・管理の一元化」方針)。tfvars に書かず -var か環境変数 TF_VAR_account_id で供給する"
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
