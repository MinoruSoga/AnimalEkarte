# REMOVED (SEC-CS2-F03): Hyperdrive is unused by the Worker/Container runtime.
#
# Containers (Go/Gin) connect to PlanetScale directly via DB_* secrets
# (Hyperdrive is not supported inside Containers:
# https://github.com/cloudflare/containers/issues/97). The former
# cloudflare_hyperdrive_config.stg_planetscale resource and hyperdrive_config_id
# output are deleted from this module so no DB origin credentials remain in
# Terraform config (and therefore no new secret-bearing Hyperdrive state is
# created under the local backend).
#
# ── USER-only operational follow-up (agents must NOT run these) ────────────
# If a prior apply still left Hyperdrive in remote Cloudflare / local state:
#   1. Review `terraform plan` in this directory; expect destroy of
#      cloudflare_hyperdrive_config.stg_planetscale if still in state.
#   2. After explicit human approval only: `terraform apply` (or destroy that
#      resource in the Cloudflare dashboard / `wrangler hyperdrive delete`).
#   3. Dispose any local terraform.tfstate that ever held Hyperdrive origin
#      passwords carefully (do not commit; prefer secure delete after remote
#      backend migration or confirmed destroy).
#   4. Rotate any PlanetScale password that was only used for Hyperdrive origin
#      (`pscale role reset-default` for the STG role that was wired in).
# Do not re-introduce Hyperdrive config unless a non-Container Worker path
# actually needs it and a threat-model review accepts origin secrets in state.
