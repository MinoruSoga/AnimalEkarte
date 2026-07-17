# P0-5: tfstate は当面 local backend。
# TODO(P0-5 follow-up): migration-cloudflare.md §運用原則 のとおり、R2 バケット(S3互換backend)を
# P2-1 で作成した後にここを以下へ切替える。切替時は `terraform init -migrate-state` を実行し、
# 既存の local tfstate を破棄せず R2 へ移行すること。
#
# terraform {
#   backend "s3" {
#     bucket                      = "animalekarte-tfstate-cloudflare"
#     key                         = "env/stg/terraform.tfstate"
#     region                      = "auto"
#     endpoints = { s3 = "https://<ACCOUNT_ID>.r2.cloudflarestorage.com" }
#     skip_credentials_validation = true
#     skip_region_validation      = true
#     skip_requesting_account_id  = true
#     use_path_style              = true
#   }
# }
