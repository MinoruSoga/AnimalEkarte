# REMOVED (SEC-CS2-F03): Hyperdrive is unused by the Worker/Container runtime.
#
# Production Containers connect to PlanetScale directly via DB_* secrets
# (Hyperdrive is not supported inside Containers:
# https://github.com/cloudflare/containers/issues/97). The former
# cloudflare_hyperdrive_config.prod_planetscale resource and hyperdrive_config_id
# output are deleted from this module so no DB origin credentials remain in
# Terraform config (and therefore no new secret-bearing Hyperdrive state is
# created under the local backend).
#
# ── USER-only operational follow-up (agents must NOT run these) ────────────
# Production Hyperdrive was a reserved/unused resource. If any prior apply
# created it in Cloudflare or left it in local state:
#   1. Review `terraform plan` in this directory; expect destroy of
#      cloudflare_hyperdrive_config.prod_planetscale if still in state.
#   2. After explicit human approval only: `terraform apply` (or destroy that
#      resource in the Cloudflare dashboard / `wrangler hyperdrive delete`).
#   3. Dispose any local terraform.tfstate that ever held Hyperdrive origin
#      passwords carefully (do not commit).
#   4. Rotate any PlanetScale password that was only used for Hyperdrive origin.
# Do not re-introduce Hyperdrive config for "maybe later" reasons — origin
# secrets in tfstate are not justified without an actual consumer.
