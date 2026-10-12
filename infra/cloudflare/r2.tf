# P2-1: R2 バケット作成(臨床画像用)
#
# バケット名は variables.tf の r2_bucket_name デフォルト(`animalekarte-stg-images`)を使用。
# ロケーションヒントは東京圏に最も近い apac を指定
# (R2 はリージョン固定ではなく自動複製だが、読み書きレイテンシ低減のためのヒント)。
# リソース自体は 2026-07 当時の記録で apply 済み。
#
# S3互換の読み書き認証情報(Access Key ID / Secret Access Key)は Terraform では発行・
# 管理しない。R2 の S3 互換 API トークンは Account API Token とは別体系であり、
# 秘密値を Terraform リソース・state に載せない境界を維持するため、発行・ローテーションは
# 人間が承認済み手順(docs/ops/infra/iac-guidelines.md / docs/ops/infra/staging/runbook.md)
# に従い Terraform 外で行う。

resource "cloudflare_r2_bucket" "stg_images" {
  account_id = var.account_id
  name       = var.r2_bucket_name
  location   = "apac"
}

output "r2_bucket_name" {
  description = "作成された R2 バケット名(S3_ENDPOINT 疎通確認・rclone 設定で使用)"
  value       = cloudflare_r2_bucket.stg_images.name
}
