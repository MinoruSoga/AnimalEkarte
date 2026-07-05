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
# 発行スコープ: P0-5 実施記録参照(Workers Scripts/R2/KV/Account Rulesets + Zone/DNS/Zone Settings/WAF の Edit)。
provider "cloudflare" {}
