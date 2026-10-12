# P6-3: Worker/Containers 異常検知の通知ポリシー
#
# 【2026-07-06 調査結果・重要】`terraform providers schema -json` で cloudflare provider
# (~> 5.21, 実バージョン 5.21.1)の `cloudflare_notification_policy` スキーマを直接抽出した結果、
# `alert_type` の受理値(52種)の中に Workers スクリプトエラーや Containers(コンテナ/Durable Object)
# のクラッシュ・OOM・再起動を直接指す alert_type は **存在しない**。近い候補は以下の2つ:
#   - http_alert_edge_error   … ゾーン全体でCloudflareのエッジが観測した5xx率のアラート
#   - http_alert_origin_error … ゾーンが「オリジン」へプロキシした際の5xx率のアラート
#     (Cloudflareの伝統的なリバースプロキシ構成 — proxied DNSレコードで外部オリジンへフォワード
#     する構成 — を前提にした指標。本構成はWorkerがリクエストを直接処理しContainerを
#     呼び出す構成であり、"外部オリジンへのプロキシ"に該当しないため、この指標が意味のある値を
#     返すか未検証。zoneのHTTPトラフィック全体を見るという性質上、Workerが返す5xxも
#     エッジ観測としては捕捉されると判断し、http_alert_origin_error ではなく
#     http_alert_edge_error を採用する)
#
# したがって本ファイルは http_alert_edge_error を「Workerエラー率の最も近い代替指標」として
# 実装する。ただし以下の制約を明記する:
#   1. ゾーンレベルの集計であり、`api.stg.noah-karte.com` サブドメインの5xxだけを狙い撃つ
#      フィルタは無い(zones フィルタのみサポート。hostname単位のフィルタは同スキーマに存在しない)。
#   2. シグナルが発生するには、対象ゾーンがCloudflare authoritativeで、実トラフィックが
#      エッジ(proxied)経由であることが前提。定義上は zone.tf の STG frontend / api.stg
#      レコードが proxied=true(ACM検証・DomainConnect・CAAはfalseのまま)だが、
#      実際の観測・通知配送は日付付き運用記録による外部検証が必要であり、本ファイルは
#      live の配送状態を主張しない。
#   3. Containers(Durable Object/コンテナインスタンス)のクラッシュ・OOM・再起動を専用に検知する
#      alert_type はCloudflare通知APIに存在しなかった(2026-07-06時点のスキーマ調査記録。
#      最新状況は再検証が必要)。最も近い代替は
#      `cloudflare_healthcheck`(能動的な/health監視 + health_check_status_notification)だが、
#      Health Checks はゾーンとは別の課金対象アドオンであり(当時のSTG月額試算には未計上
#      — 価格は2026-07時点の記録で再検証未実施)、
#      本来は外部オリジンIP監視を想定した機能でworkers.dev ホスト名への適用実績が確認できないため、
#      本構成では追加していない。現在の運用方針は docs/ops/infra/iac-guidelines.md /
#      docs/ops/infra/staging/runbook.md を参照。
#   4. `billing_usage_alert` という alert_type は実在するが、【2026-07-07 P6-4調査記録・
#      当時の結論で再検証未実施】Cloudflare公式ドキュメント確認の結果、これはArgo Smart Routing
#      (トラフィックのバイト数)やLoad Balancing(DNSクエリ数)のような特定の従量課金プロダクト専用の
#      使用量閾値通知であり、Workers/Containersは対象外だった。かつ当時Cloudflareには
#      AWS Budget Alert相当の「アカウント全体のドル建て月額支出」を監視する機構自体が
#      存在しないと判断されたため、P6-4(Budget Alert)はTerraform/CLIでは実装不可能と記録した。
#      代替策(Cloudflare GraphQL Analytics APIによる使用量の定期手動確認)の現在の運用方針は
#      docs/ops/infra/iac-guidelines.md / docs/ops/infra/staging/runbook.md を参照。
#
# 送信先メールアドレスは var.notification_email(必須・default無し)。値が無い環境では
# `terraform plan` がこの変数の未設定エラーで失敗する — これは意図した genuine BLOCKED
# であり、運用担当者のメールアドレスを人間が決定するまで推測で埋めない。
#
# 【code-reviewer指摘・2026-07-06追記・未検証】Cloudflareの通知メール送信先は、
# ダッシュボードの Notification Settings で確認リンククリックによる事前検証(verification)が
# 必要な可能性がある(Cloudflare通知APIの一般的な仕様。本プロバイダバージョンでこの制約が
# 実際に発現するかは本セッションでは未検証 — `terraform plan` は検証状態を問い合わせる
# APIを呼ばないため plan/validate PASS では確認できない)。そのため
# `TF_VAR_notification_email` に実アドレスを供給しても、`terraform apply` が
# 「未検証アドレス」エラーで失敗する可能性がある。apply 前に、当該メールアドレスが
# 既に Cloudflare アカウントの Notification 送信先として検証済みであることを
# 運用担当者に確認すること(未検証なら先にダッシュボードで確認メールを承認する必要がある
# ——運用原則の「例外的ダッシュボード操作」に該当する可能性が高い)。

resource "cloudflare_notification_policy" "worker_edge_error_rate" {
  account_id  = var.account_id
  name        = "animalekarte-${var.environment}-worker-edge-error-rate"
  description = "STG Cloudflare Workers/Containers(animalekarte-stg-api)のエッジ観測5xx率アラート。P1-2 NS切替完了までは無信号(dormant)。"
  alert_type  = "http_alert_edge_error"
  enabled     = true

  mechanisms = {
    email = [
      { id = var.notification_email },
    ]
  }

  filters = {
    zones = [cloudflare_zone.noah_karte.id]
  }
}
