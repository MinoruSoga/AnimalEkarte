terraform {
  required_version = ">= 1.5"
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.21"
    }
  }
}

# CLOUDFLARE_API_TOKEN は環境変数から供給する(このファイル・tfvars に書かない)。
# STG(../providers.tf)と同じ発行方針だが、production 用のトークンは STG のトークンを
# 使い回さないこと(スコープ分離。README.md 参照)。
# 発行スコープ: Workers Scripts/R2/Hyperdrive/Account Rulesets の Edit + Zone/DNS/Zone Settings の Edit
# (noah-karte.com に限定) + Zone: Zone Read(下記 zone.tf の data source lookup に必要)。
provider "cloudflare" {}
