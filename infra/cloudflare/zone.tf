# P1-1: ゾーン作成 + 既存DNSレコードの Terraform化
#
# 現行棚卸し結果(2026-07-05 訂正版)。
#
# 【重要・訂正履歴】初回棚卸し(2026-07-05 早朝)は `dig` のみに依拠し、apex(@)の A レコードを
# 静的な2値として記録した。しかし再検証で `dig` を複数回・複数リゾルバ(1.1.1.1 / 8.8.8.8 /
# 権威 ns1.vercel-dns.com)に投げたところ、apex の A レスポンスは 216.150.1.{1,65,129,193} /
# 216.150.16.{1,65,129,193} の間でクエリ毎に変動することが判明(Vercel 側のAnycast/ALIAS
# レコードの動的解決によるローテーションであり、固定の A レコードではない)。
# 静的な IP を Terraform に複製すると「dig 結果と完全一致」の制約を満たせず、将来 Vercel が
# Anycast プールを変更した際にドリフトする。そこで `vercel domains inspect` /
# `vercel dns ls noah-karte.com`(ドメインは Vercel が registrar かつ DNS provider。
# `noah-karte.com` の Nameservers は Intended/Current 共に ns1/ns2.vercel-dns.com)で
# Vercel 側の一次情報(実際に登録されているレコード種別・値)を取得し、それを正として
# 以下に書き直した。dig の値は「ALIAS/wildcard ALIAS の動的解決結果のサンプル」であり
# レコードそのものではないため転記しない。
#
# 棚卸し結果(Vercel API `GET /v4/domains/noah-karte.com/records` 実測、2026-07-05):
#   - name="stg"  CNAME → cname.vercel-dns.com.                                            (ttl 60)
#   - name="_a1ecec492bd7059488c176bca7348f1a.stg" CNAME →
#       _75c08600d067c8a35ee081e0735fcefe.jkddzztszm.acm-validations.aws.                    (ttl 300)
#     → ACM証明書のDNS検証用レコード(命名から stg 関連の ACM 証明書の所有権検証と推測)。
#       【要人間確認】どの ACM 証明書(api.stg のCloudFront用か、他か)の検証か不明なため、
#       削除するとその証明書の自動更新が失敗するリスクがある。移行時も削除せず維持すること。
#   - name=""(apex) CAA → 0 issue "amazon.com"                                              (ttl 60・カスタム登録)
#   - name="api.stg" CNAME → dcqico6azu5w2.cloudfront.net.                                   (ttl 300)
#   - name=""(apex) CAA → 0 issue "pki.goog" / "sectigo.com" / "letsencrypt.org"             (ttl 60・Vercel既定)
#   - name="*"(wildcard) ALIAS → cname.vercel-dns-016.com.                                   (ttl 60・Vercel既定)
#   - name=""(apex)      ALIAS → cname.vercel-dns-016.com.                                   (ttl 60・Vercel既定)
#     → apex の実体はこの ALIAS(Vercel側の動的Anycast解決)。Cloudflare には ALIAS 型が無いが、
#       zone apex への CNAME は Cloudflare が自動で CNAME flattening するため CNAME で代替する
#       (proxied=false でも flattening は機能する。Cloudflare のゾーン外部の挙動なので
#       "現行と異なる値を作らない" の趣旨に反しない)。
#   - www は個別レコードなし → "*" ワイルドカードでカバーされている(現行動作を維持)。
#
# P0-2 決定: サブドメイン委任は Cloudflare Enterprise 限定のため不採用。
# noah-karte.com ゾーン全体を Cloudflare へフル移管する(NS切替自体は P1-2 で人手作業・本ファイルの管轄外)。
#
# 【P1-2 に向けた追加の重要事実】`noah-karte.com` は registrar も Vercel(Vercel Domains経由で
# 2026-03-21 取得・自動更新)。一般的な「外部レジストラでNS変更」ではなく、Vercel の
# ドメイン管理画面(またはAPI)側でカスタムネームサーバー(Cloudflareが割り当てるNS)に
# 変更する操作になる。レジストラ移管(トランスファー)は不要(ネームサーバー変更のみ)。
# 実施は P1-2 で人間が行う(本 Terraform の管轄外)。
#
# proxied はすべて false(DNS only)で作成する。NS切替後の Full (strict) SSL 検証・動作確認が
# 済んでから、必要なレコードのみ proxied = true に切替える(切替は別コミットで実施しコメントを更新すること)。

resource "cloudflare_zone" "noah_karte" {
  account = {
    id = var.account_id
  }
  name = var.zone_name
  type = "full"
}

# 【試行7 修正】content の末尾ドット(FQDN表記)は Cloudflare API が正規化して除去して
# state に保存する(2026-07-05 実測: `GET /zones/{id}/dns_records/{id}` で末尾ドット無しの
# 値が返ることを確認済み)。末尾ドット付きで宣言すると apply 毎に in-place update の
# 差分が発生し続ける(値としての意味は同じだが plan が収束しない)ため、以下は末尾ドット無しで統一する。

# apex(@) — Vercel 側の実体は ALIAS(cname.vercel-dns-016.com への動的解決)。
# Cloudflare は zone apex の CNAME を自動フラット化するため CNAME で代替する。
resource "cloudflare_dns_record" "apex_flatten" {
  zone_id = cloudflare_zone.noah_karte.id
  name    = var.zone_name
  type    = "CNAME"
  content = "cname.vercel-dns-016.com"
  ttl     = 60
  proxied = false
  comment = "P1-1 棚卸し複製。Vercel apex ALIAS の代替(CNAME flattening)。apex 自体は現状デプロイ未割当(本番未使用)"
}

# ワイルドカード — www 含む未定義サブドメインすべてを Vercel へ向ける現行動作を維持。
resource "cloudflare_dns_record" "wildcard" {
  zone_id = cloudflare_zone.noah_karte.id
  name    = "*.${var.zone_name}"
  type    = "CNAME"
  content = "cname.vercel-dns-016.com"
  ttl     = 60
  proxied = false
  comment = "P1-1 棚卸し複製。Vercel 既定ワイルドカード ALIAS の代替(www 等をカバー)"
}

resource "cloudflare_dns_record" "stg_frontend" {
  zone_id = cloudflare_zone.noah_karte.id
  name    = "stg.${var.zone_name}"
  type    = "CNAME"
  content = "cname.vercel-dns.com"
  ttl     = 60
  proxied = false
  comment = "P1-1 棚卸し複製。STG フロントエンド(Vercel)"
}

resource "cloudflare_dns_record" "api_stg_backend" {
  zone_id = cloudflare_zone.noah_karte.id
  name    = "api.stg.${var.zone_name}"
  type    = "CNAME"
  content = "dcqico6azu5w2.cloudfront.net"
  ttl     = 300
  proxied = false
  comment = "P1-1 棚卸し複製。STG Backend API(CloudFront)。Phase 4 完了後に Worker 経由へ切替"
}

# ACM 証明書のDNS検証レコード。用途未確定だが削除すると証明書自動更新が失敗するリスクがあるため維持。
resource "cloudflare_dns_record" "acm_validation_stg" {
  zone_id = cloudflare_zone.noah_karte.id
  name    = "_a1ecec492bd7059488c176bca7348f1a.stg.${var.zone_name}"
  type    = "CNAME"
  content = "_75c08600d067c8a35ee081e0735fcefe.jkddzztszm.acm-validations.aws"
  ttl     = 300
  proxied = false
  comment = "P1-1 棚卸し複製。ACM証明書のDNS検証用(対象証明書は要人間確認)。削除禁止"
}

# 【試行7 棚卸し】Terraform 未管理レコード3件の方針決定(2026-07-05):
# 1. www.noah-karte.com A×2(216.150.1.129/216.150.1.193, proxied=true)→ 削除済み(import せず terraform 外で削除)。
#    `vercel dns ls noah-karte.com` で Vercel 側に www 専用レコードが存在しないことを確認済み。
#    実体は apex と同じ wildcard ALIAS(`*` → cname.vercel-dns-016.com)経由の動的解決の
#    スナップショット(Cloudflare Add a Site 時の自動スキャン取り込み)であり、apex で発覚した
#    「Anycast IPが変動する」問題と同種。固定IPのまま放置すると Vercel 側のIPローテーションで
#    将来的に www が到達不能になるリスクがあるため、既存の cloudflare_dns_record.wildcard
#    (CNAME flatten)にフォールバックさせる方が正しい。proxied=true だった点も
#    DNS onlyポリシーに反するため削除で解消。
# 2. _domainconnect.noah-karte.com CNAME(→ _domainconnect.vercel-dns.com)→ import して以下で管理。
#    Domain Connect プロトコルの discovery レコード(Vercel registrar が自動提供する第三者向け
#    ワンクリックDNS設定用エンドポイント)。用途未確定だが削除の安全性が確認できないため
#    ACM検証レコードと同様の慎重方針(維持)を採用し、DNS onlyポリシーに合わせて
#    proxied のみ true→false に正規化する。
resource "cloudflare_dns_record" "domainconnect" {
  zone_id = cloudflare_zone.noah_karte.id
  name    = "_domainconnect.${var.zone_name}"
  type    = "CNAME"
  content = "_domainconnect.vercel-dns.com"
  ttl     = 300
  proxied = false
  comment = "試行7 棚卸し import。Domain Connect discovery(Vercel registrar 自動提供)。用途未確定のため維持、proxied のみ正規化"
}

# apex CAA — カスタム登録分(amazon.com)+ Vercel 既定分(pki.goog / sectigo.com / letsencrypt.org)。
# いずれも "現行と異なる値を作らない" 原則により Vercel 側の実測値をそのまま複製する。
resource "cloudflare_dns_record" "caa_amazon" {
  zone_id = cloudflare_zone.noah_karte.id
  name    = var.zone_name
  type    = "CAA"
  ttl     = 60
  proxied = false
  data = {
    flags = 0
    tag   = "issue"
    value = "amazon.com"
  }
  comment = "P1-1 棚卸し複製。カスタム登録分(ACM等のCA許可)"
}

resource "cloudflare_dns_record" "caa_pki_goog" {
  zone_id = cloudflare_zone.noah_karte.id
  name    = var.zone_name
  type    = "CAA"
  ttl     = 60
  proxied = false
  data = {
    flags = 0
    tag   = "issue"
    value = "pki.goog"
  }
  comment = "P1-1 棚卸し複製。Vercel 既定CAA"
}

resource "cloudflare_dns_record" "caa_sectigo" {
  zone_id = cloudflare_zone.noah_karte.id
  name    = var.zone_name
  type    = "CAA"
  ttl     = 60
  proxied = false
  data = {
    flags = 0
    tag   = "issue"
    value = "sectigo.com"
  }
  comment = "P1-1 棚卸し複製。Vercel 既定CAA"
}

resource "cloudflare_dns_record" "caa_letsencrypt" {
  zone_id = cloudflare_zone.noah_karte.id
  name    = var.zone_name
  type    = "CAA"
  ttl     = 60
  proxied = false
  data = {
    flags = 0
    tag   = "issue"
    value = "letsencrypt.org"
  }
  comment = "P1-1 棚卸し複製。Vercel 既定CAA"
}

output "zone_name_servers" {
  description = "Cloudflare がこのゾーンに割り当てるネームサーバー(P1-2 で Vercel Domains 管理画面/APIに設定する値)"
  value       = cloudflare_zone.noah_karte.name_servers
}
