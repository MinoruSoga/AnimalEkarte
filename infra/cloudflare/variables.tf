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

# SEC-CS2-F03: pscale_stg_db_* variables removed with Hyperdrive. App DB credentials
# are Worker secrets (DB_HOST/DB_USER/DB_PASSWORD), not Terraform inputs.

# P6-3: 通知ポリシーの送信先。既存のリポジトリ内に汎用の運用アラート受信メールアドレスの
# 前例が無い(SMTP_FROM=noreply@noah-karte.com は送信専用アドレスであり受信先ではないため転用不可)。
# 値は tfvars に書かず環境変数 TF_VAR_notification_email で供給する(default 未設定=必須値。
# 供給されるまで plan/apply は失敗する。これは意図した genuine BLOCKED — 実在する運用者の
# メールアドレスを人間が決定するまで、Claude Code が推測で埋めるべきではない)。
variable "notification_email" {
  description = "P6-3 Cloudflare 通知ポリシーの送信先メールアドレス(運用担当者が決定。TF_VAR_notification_email で供給)"
  type        = string
  sensitive   = true
}
