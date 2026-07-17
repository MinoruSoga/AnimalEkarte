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

# Hyperdrive 接続先(PlanetScale Postgres animalekarte-prod / main ブランチ)。
# 値は tfvars に書かず環境変数(TF_VAR_pscale_prod_db_host 等)または `pscale role` で
# 都度発行したクレデンシャルを都度供給する運用とする(STGと同じガード。docs/infra/deploy/
# PRODUCTION_CF_SETUP.md 参照)。未供給(空文字)のまま plan/apply すると origin の password等が
# クリアされる差分が出るため、Hyperdrive を touch する回だけ資格情報を供給すること。
variable "pscale_prod_db_host" {
  description = "PlanetScale Postgres(animalekarte-prod/main) の接続ホスト"
  type        = string
  sensitive   = true
  default     = ""
}

variable "pscale_prod_db_user" {
  description = "PlanetScale Postgres の接続ユーザー"
  type        = string
  sensitive   = true
  default     = ""
}

variable "pscale_prod_db_password" {
  description = "PlanetScale Postgres の接続パスワード"
  type        = string
  sensitive   = true
  default     = ""
}

# 【notifications.tf 参照】production 専用の通知ポリシーは意図的に作成しない
# (STGのゾーンレベル5xxアラートが同一ゾーンをカバーするため、追加すると二重通知になる)。
# そのため notification_email 変数は本ファイルでは定義しない。
