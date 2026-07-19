# P3-4: Hyperdrive 設定 — PlanetScale Postgres(animalekarte-stg / main ブランチ)への接続プール。
#
# 【試行6で apply 済み】ID: 45ae9b2a018a4c0fa84c1744c0f12efa(名前: animalekarte-stg-planetscale)。
# pscale_stg_db_* は `pscale role reset-default animalekarte-stg main --org noah-animalekarte --force`
# で都度発行し TF_VAR_pscale_stg_db_* として供給する運用(検証後は同コマンドで失効させる)。
# 未供給(空文字)のまま plan/apply すると origin の password 等がクリアされる差分が出るため、
# Hyperdrive を touch する回だけ資格情報を供給すること(意図的ガード。variables.tf 参照)。
#
# 【P3-5 確定(試行7・2026-07-05)】wrangler dev --remote(エフェメラルな edge プレビュー。
# 永続 deploy は行っていない)+ postgres.js から env.HYPERDRIVE.connectionString 経由で
# SELECT 1 / スクラッチテーブルへの INSERT・UPDATE・SELECT・DELETE / sql.begin() での
# BEGIN→COMMIT および BEGIN→例外→ROLLBACK(ロールバック後に行が存在しないことまで確認)を
# 全て実行し PASS。inet_server_addr() が Worker から見て Hyperdrive のプール内部アドレス
# (10.x系)であることを確認済み — PlanetScale への直接コネクションではなく Hyperdrive 経由で
# あることの証拠。詳細ログは migration-cloudflare.md 試行7参照。
#
# 【GORM × Hyperdrive 互換性判断(静的解析。上記は Node/postgres.js での確認であり GORM 自体の
#  実行確認ではない点に留意)】
#   - backend/internal/repository/db.go は `gorm.Open(postgres.Open(dsn), &gorm.Config{})` — 
#     PrepareStmt を明示していないため既定値 false。GORM はクエリ毎に名前付き prepared
#     statement をキャッシュ・再利用しない(SQLレベルの PREPARE/DEALLOCATE 文を発行しない)。
#     → Hyperdrive の「SQLレベル prepared statement 非対応」制約に対し、通常の
#       通常のGo APIトラフィックは抵触しないと判断する。
#   - `backend/internal/repository/*_test.go`(テストのみ)と `backend/cmd/migrate/main.go`
#     (本番 one-shot マイグレーションタスク)は `pg_advisory_lock` を使用する。advisory lock は
#     Hyperdrive 非対応機能に明記されているため、migrate 実行は Hyperdrive を経由せず
#     PlanetScale へ直結する接続文字列を使うこと(P4-5 の exec() 一shot 実行でも同様)。
#     LISTEN/NOTIFY は runtime/migration いずれのコードにも使用なし(grep 確認済み)。
#   - 結論: API トラフィックは Hyperdrive 経由、migrate は直結。GORM(Go)での実接続確認は
#     Phase 4 で実際の Worker/Container からの呼び出しが組まれた時点で追加検証すること
#     (Cloudflare Hyperdrive の connectionString は Worker 実行コンテキスト外からは取得不可のため、
#     Go の単体バイナリ/ローカル psql から直接検証する手段は存在しない — Cloudflare 公式ドキュメント
#     ("Getting started"/"Connect to PostgreSQL")準拠の制約)。
#
# キャッシュは臨床/会計データの陳腐化読み取り防止のため無効(disabled = true)で開始する。
#
# 【既知の provider バグ】apply 時に `modified_on` の計算値差分で
# "Provider produced inconsistent result after apply" が出ることがある(cloudflare/terraform-provider-cloudflare
# 側の既知動作)。実際のリソース更新自体は成功しているため、エラー後に
# `terraform apply -refresh-only -auto-approve` → `terraform plan` で「No changes」に
# 収束することを都度確認すること(試行6・試行7で再現・回避手順を確立済み)。

resource "cloudflare_hyperdrive_config" "stg_planetscale" {
  account_id = var.account_id
  name       = "animalekarte-${var.environment}-planetscale"

  origin = {
    scheme   = "postgres"
    host     = var.pscale_stg_db_host
    port     = 5432
    database = "postgres" # PlanetScale Postgres は branch あたり単一DB(既定名 postgres)
    user     = var.pscale_stg_db_user
    password = var.pscale_stg_db_password
  }

  caching = {
    disabled = true
  }
}

output "hyperdrive_config_id" {
  description = "wrangler.jsonc の hyperdrive バインディングで参照する Hyperdrive 設定 ID"
  value       = cloudflare_hyperdrive_config.stg_planetscale.id
}
