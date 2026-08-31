# Cloudflare 移行 外部連携棚卸し（P4-7・試行12）

> **目的**: AWS ECS/fck-nat 構成から Cloudflare Workers + Containers 構成への移行にあたり、
> 外部連携（LINE / Lステップ / SMTP / LIFF）が固定 egress IP 前提のallowlistに依存していないか、
> および現状の STG 実装状態を棚卸しする。
> **読者**: 開発者・PO・Phase 5 以降の実装担当者。
> **タイミング**: 本番カットオーバー判断時・外部連携の疑わしい失敗調査時。

---

## 背景（AWS → Cloudflare の egress 差分）

AWS STG は `fck-nat EC2 (t4g.nano) + Elastic IP ×1` により、Private Subnet（ECS Fargate タスク）
からの outbound 通信を **単一の固定 IP** に安定化していた（2026-07 調査 HTML。vault `evidence/2026-08-20-root-docs/research-cloudflare.html`）。

Cloudflare Workers + Containers には EIP に相当する概念がない。Container の outbound は
Cloudflare のエッジ網を経由するため、**送信元 IP は固定されない**（同調査）。
これが影響するのは、送信先サービスが「IP allowlist」方式のアクセス制御を採用している場合のみ。
本ドキュメントは LINE / Lステップ / SMTP / LIFF の各連携についてこの依存の有無を確認する。

---

## 現行棚卸し

この表は repository contract と dated evidence を分ける。deployed secret value、外部 console 設定、live send の成否は本更新では確認していない。

| 連携 | repository contract | 現行判定 |
|:---|:---|:---|
| LINE Messaging API | `backend/internal/infra/line/client.go` の bearer token。送信元 IP allowlist は外部 console の任意設定 | repository 上は対応可能。実 clinic の allowlist と live send は `UNVERIFIED` |
| Lステップ | `backend/internal/infra/lstep/client.go`, `tag.go`, `user.go`。deploy gate `LSTEP_WRITE_API_ENABLED` が exact `true` かつ clinic `is_sync_enabled=true` の二重 gate を通ると4 write methodが実 HTTPを送る。repository defaultはOFF | deployed gate と vendor allowlist は `UNVERIFIED`。`[DISABLED]` grep を現行証跡に使わない |
| SMTP | `backend/internal/infra/smtp/sender.go`。SMTP auth/TLSを使う。必要な名前の正本は target `backend/wrangler.jsonc` の `secrets.required` | names/value/provider-side IP restriction と controlled send は `UNVERIFIED` |
| LIFF | ID token検証。`api.stg.noah-karte.com` は 2026-07-17 に Cloudflare へ cutover 済みという ops SSOT を前提とする | DNS prerequisiteは解消。test channel/accountを使う controlled STG rehearsal までは `UNVERIFIED` |

### Lステップ release gate

- repository 設定の既定OFFと deployed state を混同しない。deployed value は承認済み names-only channel で確認する。
- writeを許可する前に、deploy gate、clinic flag、対象test account、vendorのIP制限、停止手段、失敗通知を確認する。
- `AddTag` / `RemoveTag` / `AddTagBulk` / `SetProperty` はgate通過時に実HTTPを行う。旧 `[DISABLED]` コメントは現行事実ではない。
- 詳細は [LSTEP_WRITE_API_PAUSE.md](./LSTEP_WRITE_API_PAUSE.md)。

### SMTP secret evidence

過去の `wrangler secret list` 結果は dated evidence にすぎない。現在の必要名は target Wrangler file の `secrets.required` から導出し、値を取得・記録しない。SMTP実装は `backend/internal/infra/smtp/sender.go` を参照し、brittleな行番号や旧service pathを根拠にしない。

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

IP allowlist: LINEは任意設定を要確認。Lステップ/SMTPはprovider設定とdeployed stateが未検証。LIFFのDNS prerequisiteは解消済みだがcontrolled STG rehearsalは未実施。

---

## まとめ

| 連携 | 判定 |
|:---|:---|
| LINE | PASS（IP allowlist inventory）／inbound redelivery: RELEASE PENDING（code deploy後のConsole有効化・error監視・rehearsal待ち）／live send: BLOCKED（誤配信リスク回避のためinventory only） |
| Lステップ | UNVERIFIED（repository default OFF。二重gate通過時は実HTTP） |
| SMTP | UNVERIFIED（target Wrangler の必要名、deployed value、controlled sendを要確認） |
| LIFF | UNVERIFIED（DNS prerequisiteは解消。controlled STG rehearsal待ち） |
