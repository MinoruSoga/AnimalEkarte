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
# したがって本ファイルでは production 用ポリシーを定義せず、「STG側にゾーンレベル
# ポリシーが定義されており、apply 済みかつ同一ゾーンをカバーしていれば production の
# ゾーンレベル5xxもそこに含まれる(分離できないため)」という設計判断を記録するに留める。
# STG側ポリシーの apply 状況・送信先・配送実績は外部検証が必要であり、production が
# 専用アカウントで構築される場合は共有アカウント側ポリシーのカバー範囲自体も要検証。
#
# 【ホスト名単位で分離したい場合の代替】`cloudflare_healthcheck`(能動的な/health監視。
# STGのnotifications.tfが検討済みの代替案と同じ)であればホスト名を指定できるが、
# ゾーンとは別課金のアドオンであり、STGでも「本セッションでは追加しない」と判断され
# 見送られている。production側で本当に必要になった場合は、STGとproduction両方の
# ヘルスチェックをこの粒度で新規に設計し直すこと(このファイル単独でのフォローアップは
# しない。§10相当のスコープ外として明記)。
#
# 【運用メモ】対象アカウント/ゾーンを確認し、STG側ポリシーが apply 済みで同一ゾーンを
# カバーしている場合に限り、production起因のゾーンレベル5xxもその送信先に含まれる設計。
# apply・送信先検証・配送は外部検証が必要であり、本ファイルは live の配送を保証しない。
# このカバー範囲が確認できた上で production 専用の受信者に分離したい場合は、STG側
# ポリシーの送信先を「STG/production共通の運用担当者」へ変更する運用判断が先決
# (このファイルの変更だけでは解決しない)。
