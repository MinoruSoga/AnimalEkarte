variable "account_id" {
  description = "Cloudflare Account ID (CLOUDFLARE_ACCOUNT_ID と同値。tfvars に書かず -var か環境変数 TF_VAR_account_id で供給する)"
  type        = string
  sensitive   = true
}

variable "zone_name" {
  description = "Cloudflare に追加するゾーン名"
  type        = string
  default     = "noah-karte.com"
}

variable "environment" {
  description = "リソース名・タグに使う環境名"
  type        = string
  default     = "stg"
}

variable "r2_bucket_name" {
  description = "臨床画像用 R2 バケット名"
  type        = string
  default     = "animalekarte-stg-images"
}

# P3-4: Hyperdrive 接続先(PlanetScale Postgres animalekarte-stg / main ブランチ)。
# 値は tfvars に書かず環境変数(TF_VAR_pscale_stg_db_host 等)または `pscale role` で
# 都度発行したクレデンシャルを都度供給する運用とする(state に残るため rotation 運用が必要)。
variable "pscale_stg_db_host" {
  description = "PlanetScale Postgres(animalekarte-stg/main) の接続ホスト(pscale role reset-default 等で取得)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "pscale_stg_db_user" {
  description = "PlanetScale Postgres の接続ユーザー"
  type        = string
  sensitive   = true
  default     = ""
}

variable "pscale_stg_db_password" {
  description = "PlanetScale Postgres の接続パスワード"
  type        = string
  sensitive   = true
  default     = ""
}
