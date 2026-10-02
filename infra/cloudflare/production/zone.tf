# production は STG と同一ゾーン(noah-karte.com)を使う。ゾーンそのもの
# (`cloudflare_zone` リソース)は STG 側(infra/cloudflare/zone.tf)が既に管理しているため、
# ここで同じリソースを再度 `resource "cloudflare_zone"` として宣言してはならない
# (同一ゾーンを2つの tfstate が管理しようとすると、2回目の apply で "already exists" 相当の
# 衝突になる)。production 側は既存ゾーンを `data` source として参照し、新規に必要な
# DNS レコードのみを追加する。
#
# 【前提】このデータソース参照には Cloudflare API Token に Zone:Read(zone resources =
# noah-karte.com)スコープが必要。production 用トークンの発行時に付与すること
# (providers.tf のコメント参照)。
#
# 【apply の前提条件】本ディレクトリの apply は、critical path 上 STG Phase 7(NS切替)完了後に
# 行う想定(migration-cloudflare.md「現況サマリ」2026-07-15 決定の含意 #2 参照)。NS切替前に
# apply しても Terraform 自体は成功するが、ゾーンが Cloudflare 上で active になっていない間は
# 実トラフィックがこのレコード経由でWorkerへ届かない(STGのnotifications.tf/zone.tfが記録した
# 「dormant」と同じ状態)。

data "cloudflare_zone" "noah_karte" {
  filter = {
    name = var.zone_name
  }
}

# production Backend API 用の新規 DNS レコード。STG の api_stg_backend
# (infra/cloudflare/zone.tf)とは異なり、最初から proxied = true で作成する。
# 【実測 2026-10-02】api.noah-karte.com は個別レコードを持たず、wildcard
# `*.noah-karte.com` CNAME -> cname.vercel-dns-016.com(Vercel)が被覆していた。
# 明示的な A レコードは wildcard より優先されるため削除/import 不要で作成でき、
# 作成と同時にこのホスト名だけが CF エッジ(proxied)へ切り替わる。
# apex(noah-karte.com)と wildcard 自体は Vercel のまま維持(frontend 切替時まで)。
# (Workers Route は "pattern" + "/*" 形式の場合、マッチ対象ホスト名のDNSレコードが
# proxied=trueでないとWorkerへルーティングされない。proxied=falseのまま
# workers_dev=falseで初回デプロイすると、CIのヘルスチェックが到達できるURLが
# どこにも存在しなくなるデッドロックになるため、STGの「後で切替」パターンをそのまま
# 複製しない)。
#
# content は Workers Route 経由でCloudflareエッジが横取りするため実質未使用
# (Workerが常にリクエストを先取りする限り、この値へのプロキシは発生しない)。
# RFC 5737 のドキュメント用アドレス(TEST-NET-1)をプレースホルダとして使う
# (実在するIPを書かない。Cloudflareダッシュボード上の表示値としてのみ意味を持つ)。
resource "cloudflare_dns_record" "api_prod_backend" {
  zone_id = data.cloudflare_zone.noah_karte.id
  name    = "api.${var.zone_name}"
  type    = "A"
  content = "192.0.2.1"
  ttl     = 1 # proxied=true の場合 ttl は自動("1"=auto)扱い
  proxied = true
  comment = "production Backend API(#253)。Workers Route宛のプレースホルダレコード。実トラフィックはWorkerが横取りするためcontentは未使用"
}

# EMR-255: PROD frontend(animalekarte-prod-frontend Worker)用 DNS レコード。
# frontend/wrangler.production.jsonc の route "www.noah-karte.com/*" が機能するには
# proxied=true のレコードが必要(wildcard CNAME→Vercel では route は発火しない)。
# 【実測 2026-10-02】www も個別レコードではなく wildcard CNAME が被覆していたが、
# 明示 A レコードが優先されるため作成のみで切替済み(apply 済み)。
# content は Worker が横取りするため TEST-NET-1 プレースホルダ(api_prod_backend 同規則)。
resource "cloudflare_dns_record" "frontend_www" {
  zone_id = data.cloudflare_zone.noah_karte.id
  name    = "www.${var.zone_name}"
  type    = "A"
  content = "192.0.2.1"
  ttl     = 1
  proxied = true
  comment = "EMR-255: PROD frontend www=animalekarte-prod-frontend Worker(Workers Route宛プレースホルダ)"
}

# apex(noah-karte.com)は infra/cloudflare/zone.tf(STG state)の apex_flatten が
# CNAME→Vercel として管理中。同名レコードは同一ゾーンで両立できないため、PROD
# frontend 切替時に apex_flatten を A プレースホルダ(proxied)へ置き換える。
# 本ファイルには宣言しない(別 tfstate の管理領域を二重宣言すると apply が衝突する)。

output "zone_id" {
  description = "noah-karte.com ゾーンのID(STGとの共有ゾーン。参照用)"
  value       = data.cloudflare_zone.noah_karte.id
}
