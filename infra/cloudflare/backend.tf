# tfstate backend notes (STG).
#
# There is intentionally no active backend block here — Terraform defaults to a
# workspace-local state file when none is set. That default is fine for
# zone/DNS/R2/notifications only.
#
# SEC-CS2-F03: Hyperdrive (which stored PlanetScale origin host/user/password
# in state) has been removed from this module. No secret-bearing Hyperdrive
# resource remains under local backend configuration. If an older local
# terraform.tfstate still contains a cloudflare_hyperdrive_config entry, treat
# destroy + state disposal + DB password rotation as USER-only ops (see
# hyperdrive.tf tombstone). Agents must not read, print, commit, or apply
# against that state.
#
# 将来の選択肢(未実装): R2 バケット(S3互換backend)への state 移行。
# 以下はイメージ例であり、実施には state の所有者・ロック・暗号化を含むレビュー済み
# 設計が必要(現行ガイドライン: docs/ops/infra/iac-guidelines.md)。切替時は人間が
# `terraform init -migrate-state` を実行し、既存の local tfstate を破棄せず R2 へ
# 移行すること（agents は migrate-state / apply を実行しない）。
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
