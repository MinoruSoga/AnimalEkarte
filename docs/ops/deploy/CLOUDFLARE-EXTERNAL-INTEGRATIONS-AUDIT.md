# Cloudflare 移行 外部連携棚卸し（P4-7・試行12）

> **目的**: AWS ECS/fck-nat 構成から Cloudflare Workers + Containers 構成への移行にあたり、
> 外部連携（LINE / Lステップ / SMTP / LIFF）が固定 egress IP 前提のallowlistに依存していないか、
> および現状の STG 実装状態を棚卸しする。
> **読者**: 開発者・PO・Phase 5 以降の実装担当者。
> **タイミング**: 本番カットオーバー判断時・外部連携の疑わしい失敗調査時。

---

## 背景（AWS → Cloudflare の egress 差分）

AWS STG は `fck-nat EC2 (t4g.nano) + Elastic IP ×1` により、Private Subnet（ECS Fargate タスク）
からの outbound 通信を **単一の固定 IP** に安定化していた（`research-cloudflare.html` L165-167）。

Cloudflare Workers + Containers には EIP に相当する概念がない。Container の outbound は
Cloudflare のエッジ網を経由するため、**送信元 IP は固定されない**（`research-cloudflare.html` L143, L539）。
これが影響するのは、送信先サービスが「IP allowlist」方式のアクセス制御を採用している場合のみ。
本ドキュメントは LINE / Lステップ / SMTP / LIFF の各連携についてこの依存の有無を確認する。

---

## 棚卸し表

| 連携 | 認証方式 | IP allowlist 要否 | STG 現状（試行9〜12時点） | Cloudflare 移行影響 | 試行12 結果 |
|:---|:---|:---|:---|:---|:---|
| **LINE Messaging API**（`api.line.me`） | Bearer Channel Access Token（`backend/internal/infra/line/client.go` L51） | **既定では不要**。LINE公式ガイドラインは「IPアドレスでのアクセス制御禁止（webhook受信側）」を明言。push送信側（outbound）は long-lived channel access token 使用時に**オプションでIP許可リストを設定可能**（LINE Developers Console > Security タブ）だが、既定は無効。 | Channel Access Token / Secret はクリニックごと `clinic_integrations` テーブルでDB管理（`config.go`にはグローバル環境変数経路なし）。`wrangler.jsonc` の `LINE_CHANNEL_ACCESS_TOKEN`/`LINE_CHANNEL_SECRET` はグローバルfallback用の予約枠のみで試行9時点で空投入。 | 対象クリニックがLINE Developers ConsoleでオプションのIP許可リストを**有効化していない限り無影響**。有効化しているクリニックがある場合はCloudflare移行後に送信失敗するため、本番カットオーバー前に全クリニックのLINE Developers Console > Security設定を確認する運用チェックが必要（P7-3で確認事項として引き継ぎ）。 | **PASS**（doc結論）。live push は BLOCKED — 理由: STG `clinic_integrations` に実クリニックのトークン/LINE User IDが登録されており、誤配信リスク回避のためテスト専用データが確認できない本試行では送信しない（inventory onlyに留める）。 |
| **Lステップ**（`api.lstep.jp`） | Bearer API Key（`lstep_settings_connection.go` L63, `tag.go`/`user.go`） | 不明（Lステップ公式に許可リスト機能の記載なし、要ベンダー確認）。現状は **Write系4メソッドが `[DISABLED]`** のため影響評価不要。 | `LSTEP_WRITE_API_PAUSE.md` に記載の通り、`tag.go`（`AddTag`/`RemoveTag`/`AddTagBulk`）と `user.go`（`SetProperty`）がHTTP送信を抑止し入力検証のみ実施。読み取り系（`TestConnection`の`testLstepAPI`によるGET `/api/v1/tags`疎通確認）は継続。 | Write再有効化時は、対象クリニックのLステップ管理画面でIP許可リスト機能の有無を確認する項目を`LSTEP_WRITE_API_PAUSE.md`の再有効化前提条件に追加すべき（本ドキュメントで指摘、実際のmdファイル更新はスコープ外）。 | **PASS**（DISABLED状態の確認+doc記載）。grep結果は下記参照。 |
| **SMTP**（587/465 egress） | `SMTP_HOST`/`SMTP_USER`/`SMTP_PASS`（`config.go` L60-64, `net/smtp.SendMail`使用） | IP allowlist方式ではなくSMTP認証（PLAIN AUTH）+ TLS/STARTTLS前提のため、送信元IPには非依存と想定されるが、送信先SMTPサーバー側の設定次第で送信元IPレンジ制限がある可能性は残る。 | `wrangler.jsonc`の`secrets.required`に`SMTP_HOST`/`SMTP_USER`/`SMTP_PASS`が登録されているが、試行9記録によれば空文字投入の可能性がある。試行12で`wrangler secret list`を実施し名前の存在のみ確認（値は非取得）。 | 値が空の場合、`appointment_notification_service.go` L301の`if s.cfg.SMTPHost == ""`ガードにより送信自体がスキップされるため、Cloudflare移行によるリグレッションは発生しない（無効化されたまま）。 | 下記「SMTP secret 確認結果」参照。 |
| **LIFF / コールバックURL** | LINE Login ID Token検証（`liff_auth.go`、`verifyLiffIDToken`） | 該当なし（IDトークン検証方式、IPには非依存） | `wrangler.jsonc` vars の `FRONTEND_URL=https://stg.noah-karte.com`固定。`frontend/vercel.json`のrewriteは`https://api.stg.noah-karte.com`（AWS側）向けで、workers.dev他ドメインへの動的切替機構はない。LIFF ID（`{channelID}-{appID}`形式）はクリニックごとDB管理（`liff_auth.go` L105）。 | P1-2（NS切替）完了前は、LIFFアプリ本番導線（`frontend/liff`, `frontend/line-reserve`）がworkers.dev経由のAPIに到達する経路が存在しない（vercel.jsonのrewrite先固定のため）。 | **BLOCKED**（P7-3へdefer）。理由: LIFF本番導線の検証にはDNS切替(P1-2)後の`api.stg.noah-karte.com`→Cloudflare Route解決が前提となり、workers.dev段階のPhase 4では検証不可。 |

---

## LSTEP `[DISABLED]` grep結果（AC-P47-2）

```
backend/internal/infra/lstep/tag.go:22:	// [DISABLED] HTTP call to POST /contacts/{id}/tags is suppressed.
backend/internal/infra/lstep/tag.go:35:	// [DISABLED] HTTP call to DELETE /contacts/{id}/tags is suppressed.
backend/internal/infra/lstep/tag.go:76:	// [DISABLED] HTTP call to POST /contacts/tags/bulk is suppressed.
backend/internal/infra/lstep/user.go:69:	// [DISABLED] HTTP call to POST /contacts/{id}/properties is suppressed.
```

4箇所すべて`[DISABLED]`のまま。`LSTEP_WRITE_API_PAUSE.md`の再有効化前提条件（5項目）も未達のため、本試行ではコード変更なし（スコープ外）。

---

## SMTP secret 確認結果（AC-P47-4）

`wrangler secret list --name animalekarte-stg-api` を試行12で実施（値は取得しない。名前のみ）。
結果、`SMTP_HOST` / `SMTP_USER` / `SMTP_PASS` の3つとも `secret_text` として登録済みであることを確認した
（`AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY`/`DB_HOST`/`DB_PASSWORD`/`DB_USER`/`INTEGRATION_ENCRYPTION_KEY`/
`JWT_SECRET`/`LINE_CHANNEL_ACCESS_TOKEN`/`LINE_CHANNEL_SECRET`/`LSTEP_API_KEY`/`MIGRATE_RUN_SECRET`も同時に
登録済みであることを確認。`wrangler.jsonc`の`secrets.required`一覧と一致）。

**結論: BLOCKED（inventory only）**。理由:
- `wrangler secret list` は名前の存在のみを返し、値（空文字か実値か）は取得できない。試行9記録では「SMTP/LINE/LSTEP（未使用分は空文字）を投入済み」とあり、空文字投入の可能性が残る。
- `config.go` L93 の `Validate()` は `SMTPHost != ""` の場合のみ `SMTP_PORT`妥当性を検査し、`appointment_notification_service.go` L301 の送信処理も `SMTPHost == ""` なら即スキップする設計のため、空文字であればCloudflare移行によるリグレッションは発生しない。
- 実際に空文字か実値かを判定するには、アプリ経由の送信トリガー（予約通知・パスワードリセットメール等）を実行する必要があるが、実値が設定されていた場合に実際のメールアドレスへ送信してしまうリスクがあるため、本試行では見合わせる。
- SMTP認証（PLAIN AUTH + STARTTLS/TLS）はIP allowlist方式ではないため、値が実際に設定されていた場合でもCloudflareの非固定egress IPによる新規の疎通リスクは低いと想定されるが、送信先SMTPサーバー側のIPレンジ制限方針は未確認のため実測が望ましい（P7以降でsecret値確認済みの担当者が実施することを推奨）。

---

## LINE Developers Console IP許可リスト機能（参考・公式ガイドライン根拠）

- 「Don't restrict access by IP address」（webhook受信側）: LINE Platform公式は送信元IPを開示しないため、webhook受信をIPで絞ることを明示的に禁止している（署名検証を使うこと）。
- 「Restrict who can call the API when using a long-lived channel access token (optional)」: push送信などのAPI呼び出し（outbound）側は、long-lived channel access token使用時に限り、LINE Developers Console > Security タブでIPアドレス/CIDRを登録してAPI呼び出し元を制限する**オプション機能**が存在する。
- 結論: 本システムがLINEへ送信する際（`api.line.me`へのoutbound）にこのオプション機能を有効化しているクリニックがあれば、Cloudflareの非固定egress IPでは送信が拒否される。デフォルトは無効なため、無効なクリニックには影響しない。カットオーバー前に対象クリニックのLINE Developers Console設定を確認する運用手順を追加することを推奨（本ドキュメントの提言、実施はP7-3）。

参考: LINE Developers公式ドキュメント（`developers.line.biz/en/docs/messaging-api/development-guidelines/`, `developers.line.biz/en/docs/messaging-api/building-bot/`）2026-07-05時点の内容。

---

## LINE webhook redelivery release gate（2026-07-24 follow-up）

[LINE公式のwebhook受信ガイド](https://developers.line.biz/en/docs/messaging-api/receiving-messages/)では、Webhook redeliveryは既定OFFであり、有効化済みかつbot serverが2xxを返さなかった場合に一定期間再送される。再送回数・間隔は非公開で、redelivery自体も確実な配送を保証しない。同一`webhookEventId`のeventが複数回届く場合があり、redeliveryにより受信順が発生順と異なる場合はevent `timestamp`で文脈を確認する必要がある。

現行codeはfollow/unfollowを、一意に署名一致したclinic、owner ID、expected LINE user ID、event timestampによるCASとして処理する。stale・duplicate・out-of-order・再連携前IDは`RowsAffected == 0`のno-op、同一timestampはunfollow優先であり、真のlookup/DB errorはnon-2xxへ伝播する。このcodeを先にdeployした後、次をrelease operationとして実施する。

1. test channel/accountでcode hashとwebhook URLを固定し、LINE Developers ConsoleのMessaging API tabで`Use webhook`を確認してから、既定OFFの`Webhook redelivery`を有効化する。
2. 同じConsoleで既定OFFの`Error statistics aggregation`を有効化する。[LINE公式のerror statisticsガイド](https://developers.line.biz/en/docs/messaging-api/check-webhook-error-statistics/)に従い、`Webhook errors` tabでnon-2xx（`error_status_code`）、connection failure、request timeout、unclassified errorを監視する。aggregation有効化前の期間は遡及表示されないため、有効化時刻を記録する。
3. test channel/accountだけでcontrolled non-2xxを発生させ、redeliveryを確認する。同一eventのduplicate、follow/unfollowのout-of-order、同一timestampの両順序を再現し、duplicate/staleがno-op、同一timestampはunfollowへ収束することを確認する。公式仕様上、回数・間隔は非公開であり、固定待ち時間をrunbook contractにしない。
4. rehearsal記録にはchannel、code hash、Console変更時刻、`webhookEventId`、event timestamp、初回/non-2xx/redeliveryのstatus、最終owner watermark、Webhook errors表示、rollback結果を残す。production userへの送信・状態変更をrehearsalに使用しない。

**Known LOW residual**: 同じLINE User IDへの再紐付け直後でownerの`line_followed_at` / `line_blocked_at`が両方nilの場合、そのIDに対する非常に古い正規署名済みredeliveryはtimestamp CASの初期比較を通り得る。expected LINE user ID CASにより別ID・別clinicへ波及せず、現時点ではこのwatermarkを業務判断に使うruntime decision consumerもないため、直ちにcode gateを再開するseverityではない。上記rehearsalの観測対象に含め、consumer追加または実害観測時にlink時watermark初期化やevent ID/last-event persistenceを再評価する。

**現在の判定**: code側は実装済みだが、Webhook redeliveryとError statistics aggregationのConsole有効化、non-2xx/error monitoring、duplicate/out-of-order/LOW residual rehearsalはrelease pending。この文書更新ではLINE Developers Consoleその他の外部設定を変更していない。

---

## リスクレジスタへの反映（AC-P47-6）

`../infra/_archive/migration-cloudflare.md` §9 のリスク登録簿「IP allowlist」行について、試行12の結論（LINE既定非依存・オプション機能のみ要確認、SMTP/LIFFはBLOCKED理由付き）を追記する。詳細は `../infra/_archive/migration-cloudflare.md` 試行12セクションを参照。

---

## まとめ

| 連携 | 判定 |
|:---|:---|
| LINE | PASS（IP allowlist inventory）／inbound redelivery: RELEASE PENDING（code deploy後のConsole有効化・error監視・rehearsal待ち）／live send: BLOCKED（誤配信リスク回避のためinventory only） |
| Lステップ | PASS（DISABLED状態確認） |
| SMTP | BLOCKED（secret名確認済み・値非取得・誤配信リスク回避のためlive smoke見合わせ） |
| LIFF | BLOCKED（P7-3 defer、DNS切替前提） |
