# production Hyperdrive 設定 — PlanetScale Postgres(animalekarte-prod / main ブランチ、
# 未作成。docs/infra/deploy/PRODUCTION_CF_SETUP.md の pscale コマンド手順で先に作成すること)
# への接続プール。STGのhyperdrive.tfと同じ構成パターンだが、接続先DBは完全に別物。
#
# 【STGからの既知の制約の継承】Container(Go/Ginバイナリ)は Hyperdrive に非対応
# (素のLinuxプロセスであり、Workers runtime固有のHyperdriveバインディングを利用できない:
# https://github.com/cloudflare/containers/issues/97)。そのためAPIトラフィックは
# PlanetScaleへ直結し、本リソースは「将来Worker自身が直接DBを触るユースケースが増えた場合」
# のために予約するだけで、Phase 4相当の構成では未使用のまま残る(STGのhyperdrive.tf冒頭
# コメントと同じ判断)。
#
# 【正直な評価・要フォローアップ】このリソースは現状使われない予約枠でありながら、
# origin にDB接続情報(host/user/password)をtfstateへ平文保持する(sensitive指定でも
# tfstate自体は暗号化されない。ローカルbackend運用中は特に注意)。STG運用時と同じく
# credential は都度発行・都度失効(pscale role reset-default)する短命運用を前提とする。
# 「使っていないのにcredentialローテ運用の対象になる」というコストは、production開始時点で
# 一度棚卸しし、当面本当に不要と判断されればこのファイル自体の削除を検討すること
# (①要件を疑う: 「将来使うかもしれない」だけでは維持理由として弱い)。
#
# キャッシュは臨床/会計データの陳腐化読み取り防止のため無効(disabled = true)で開始する(STGと同じ)。

resource "cloudflare_hyperdrive_config" "prod_planetscale" {
  account_id = var.account_id
  name       = "animalekarte-${var.environment}-planetscale"

  origin = {
    scheme   = "postgres"
    host     = var.pscale_prod_db_host
    port     = 5432
    database = "postgres" # PlanetScale Postgres は branch あたり単一DB(既定名 postgres)
    user     = var.pscale_prod_db_user
    password = var.pscale_prod_db_password
  }

  caching = {
    disabled = true
  }
}

output "hyperdrive_config_id" {
  description = "backend/wrangler.production.jsonc の hyperdrive バインディング(<PLACEHOLDER-HYPERDRIVE-ID>)へ投入するID"
  value       = cloudflare_hyperdrive_config.prod_planetscale.id
}
