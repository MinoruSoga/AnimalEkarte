# production 用 R2 バケット(臨床画像)。STG の animalekarte-stg-images とは別バケット
# (テナント分離。同一バケットをSTG/productionで共有すると事故時の影響範囲がテナント跨ぎになる)。
#
# ロケーションヒントは STG と同じ apac(東京圏に最も近い。R2 はリージョン固定ではなく
# 自動複製だが、読み書きレイテンシ低減のためのヒント)。
#
# S3互換の書き込み/読み取り認証情報(Access Key ID / Secret Access Key)は、この Terraform
# リソースでは発行・管理しない(STGと同じ方針)。R2 の S3 互換 API トークンは Account API Token
# とは別体系であり、秘密値を Terraform リソース・state に載せない境界を維持するため、
# 発行・ローテーションは人間が承認済み手順(docs/ops/infra/iac-guidelines.md /
# docs/ops/infra/production/setup.md)に従い Terraform 外で行う。

resource "cloudflare_r2_bucket" "prod_images" {
  account_id = var.account_id
  name       = var.r2_bucket_name
  location   = "apac"
}

output "r2_bucket_name" {
  description = "作成された R2 バケット名(S3_ENDPOINT疎通確認・backend/wrangler.production.jsoncのr2_buckets[0].bucket_nameと一致させる)"
  value       = cloudflare_r2_bucket.prod_images.name
}
