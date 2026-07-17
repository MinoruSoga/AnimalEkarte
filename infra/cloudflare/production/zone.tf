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
# (infra/cloudflare/zone.tf)とは異なり、このホスト名(api.noah-karte.com)には移行前の
# 「置き換えるべき既存レコード」が存在しない(docs/ops/deploy/README.md: 「本番向け
# バックエンド自動デプロイワークフローは未整備」)。そのため STG が踏んだ
# 「まずproxied=falseで作成→Full(strict)SSL確認後にproxied=trueへ切替」という2段階を
# 踏襲する必要が薄いと判断し、最初から proxied = true で作成する
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

output "zone_id" {
  description = "noah-karte.com ゾーンのID(STGとの共有ゾーン。参照用)"
  value       = data.cloudflare_zone.noah_karte.id
}
