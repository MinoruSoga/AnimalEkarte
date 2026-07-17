# tfstate は当面 local backend(STGと同じ運用)。
# STGとは別ディレクトリのため local tfstate ファイル自体は自然に分離されるが
# (このディレクトリ配下に独自の terraform.tfstate が生成される)、将来 R2 backend へ
# 移行する際は STG の env/stg/terraform.tfstate と衝突しない key を使うこと(下記コメント参照)。
#
# TODO(P0-5 follow-up 相当): R2 バケット(S3互換backend)を用意した後にここを以下へ切替える。
# 切替時は `terraform init -migrate-state` を実行し、既存の local tfstate を破棄せず
# R2 へ移行すること。STGと同一R2バケットを使う場合は key で環境を分離する
# (bucket 自体を分ける場合は STG の tfstate 用と別に production 用 R2 バケットを用意する)。
#
# terraform {
#   backend "s3" {
#     bucket                      = "animalekarte-tfstate-cloudflare"
#     key                         = "env/prod/terraform.tfstate"
#     region                      = "auto"
#     endpoints = { s3 = "https://<ACCOUNT_ID>.r2.cloudflarestorage.com" }
#     skip_credentials_validation = true
#     skip_region_validation      = true
#     skip_requesting_account_id  = true
#     use_path_style              = true
#   }
# }
