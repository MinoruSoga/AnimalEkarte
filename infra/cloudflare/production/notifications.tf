# production 専用の通知ポリシーは意図的に作成しない(このファイルにリソース定義はない)。
#
# 【理由】STG(infra/cloudflare/notifications.tf の cloudflare_notification_policy
# "worker_edge_error_rate")は alert_type = "http_alert_edge_error" で、フィルタ条件は
# `filters.zones = [<noah-karte.comのゾーンID>]` のみ。STGコード自身のコメントが明記する
# 通り、この alert_type にはホスト名単位(api.stg.noah-karte.com のみ、等)のフィルタが
# Cloudflare通知APIのスキーマ上そもそも存在しない(ゾーン全体の5xx率を見る指標)。
#
# production はSTGと**同一ゾーン**(noah-karte.com)を使うため、production用に同じ
# alert_typeで2本目のポリシーを作成しても、それは「同一ゾーンの同一指標」を2回監視する
# だけの二重管理になる。具体的な実害は2つ:
#   1. 1件の5xxイベントに対してSTG用・production用の両方が発火し、二重メールが飛ぶ
#   2. production用ポリシーの送信先を「production担当者のみ」に絞ったつもりでも、
#      STG起因の5xx(STGのContainerがクラッシュした場合等)でもproduction担当者に
#      通知が飛ぶ(ホスト名で分離できないため)。「production向けアラート」という
#      前提自体が成立しない
#
# したがって、STGの既存ポリシーが実質的にproductionのゾーンレベル5xxも既にカバーして
# いる(むしろカバーせざるを得ない)という事実をここに記録するに留める。
#
# 【ホスト名単位で分離したい場合の代替】`cloudflare_healthcheck`(能動的な/health監視。
# STGのnotifications.tfが検討済みの代替案と同じ)であればホスト名を指定できるが、
# ゾーンとは別課金のアドオンであり、STGでも「本セッションでは追加しない」と判断され
# 見送られている。production側で本当に必要になった場合は、STGとproduction両方の
# ヘルスチェックをこの粒度で新規に設計し直すこと(このファイル単独でのフォローアップは
# しない。§10相当のスコープ外として明記)。
#
# 【現状の運用】production起因の5xxはSTG側ポリシーの送信先(TF_VAR_notification_email、
# STG apply時に投入)に届く。production専用の受信者に分離したい場合は、まずSTG側ポリシーの
# 送信先を「STG/production共通の運用担当者」へ変更する運用判断が先決(このファイルの
# 変更だけでは解決しない)。
