# P2-1: R2 バケット作成(臨床画像用)
#
# 命名は migration-cloudflare.md の指示どおり `animalekarte-stg-images` 相当(variables.tf の
# r2_bucket_name デフォルト値を使用)。ロケーションヒントは東京圏に最も近い apac を指定
# (R2 はリージョン固定ではなく自動複製だが、読み書きレイテンシ低減のためのヒント)。
#
# 【試行4で apply 済み】バケット `animalekarte-stg-images`(apac)。wrangler CLI 経由の
# put/get/delete は試行4・試行7で再確認済み(migration-cloudflare.md 参照)。
#
# S3互換の書き込み/読み取り認証情報(Access Key ID / Secret Access Key)は、この Terraform
# リソースでは発行しない。R2 の S3 互換 API トークンは Cloudflare が「Account API Token」とは
# 別体系(R2 token)で発行するため、`cloudflare_api_token` リソース(R2 スコープ付き)を
# 別途 P0-5 のトークン管理方針に沿って追加する必要がある。値は Terraform state に残さず
# `wrangler r2 bucket create` 経路 or ダッシュボードでの発行(例外操作)を検討すること。
#
# 【P2-3 PASS — 2026-07-05 試行8】Account API Token に `Account API Tokens Write` を追加後、
# `POST /accounts/{account_id}/tokens` で R2 スコープ子トークンを CLI 発行。Access Key ID =
# 子トークン id、Secret = token value の SHA-256。`.env.staging` に保存し
# `TestR2S3Live`(backend/internal/infra/s3_r2_live_test.go) で S3 SDK 経路を検証済み。
# 再発行手順(ダッシュボード or API)は migration-cloudflare.md 試行7・8 参照。

resource "cloudflare_r2_bucket" "stg_images" {
  account_id = var.account_id
  name       = var.r2_bucket_name
  location   = "apac"
}

output "r2_bucket_name" {
  description = "作成された R2 バケット名(S3_ENDPOINT 疎通確認・rclone 設定で使用)"
  value       = cloudflare_r2_bucket.stg_images.name
}
