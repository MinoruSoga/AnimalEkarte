# STG チャット履歴 未保存事象の調査（2026-10-05）

## 症状

STG でユーザーが当日チャット会話を行ったと報告されたが、共有一覧
`/settings/chat-history`（`GET /api/v1/support/chat/exchanges`）に記録がない。
10/2 に先方へチャットボットを通知済み。

## 結論

**10/2 以降、バックエンドに届いた `POST /api/v1/support/chat` は全件 DB に保存されている。
「今日の会話」に相当する POST は Cloudflare ゾーンのリクエストログに存在しない
— リクエスト自体が backend に到達していない。**

## 証拠: Cloudflare ゾーンリクエストログ × DB 突合

CF GraphQL `httpRequestsAdaptiveGroups`（zone `noah-karte.com`）で
`/api/v1/support/chat` POST を 10/01〜10/06 全件抽出:

| 日時 (UTC) | ホスト | 結果 | DB 対応 |
|---|---|---|---|
| 10/01 08:xx | stg.noah-karte.com | 200 ×2 | 10/01 17:54 JST の2件（wakt / ワクチン削除） |
| 10/02 05:xx | stg.noah-karte.com | 200 ×2 | 10/02 14:22 JST の2件（主訴/問診テンプレート） |
| 10/02 09:xx | stg.noah-karte.com | 200 ×1 | 10/02 18:25 JST の1件（診察なし薬のみ） |
| 10/05 11:xx | api.stg.noah-karte.com | 200 ×2 | 本調査の疎通テスト2件 |

**POST 合計 7件 = exchanges 増分 7件（9/30 分を除く窓内）。欠落ゼロ。**

10/5 の同日 STG 利用実績（ゾーンAPI、00:00–15:00 UTC = 09:00–24:00 JST）:
GET /pets ×185、GET /reservations ×103、POST /login ×11、POST /reservations ×7、
GET /support/chat/exchanges ×7、GET /support/chat/status ×2(+401 ×4)
→ 業務利用はあったが POST /support/chat は上記テスト2件のみ。

## 「回答は返るが保存されない」退化経路（今回は非該当だが構造的に存在）

`backend/internal/support/chat.go`:

1. `loadChatHistory` 失敗 → WARN のみ・履歴なしで継続
2. `RecordChatExchange` 失敗 → **WARN のみで回答は正常返却**（best-effort）
3. 回答が `chatHistoryMaxLength`(4000) 超 → WARN で保存スキップ

※今回はゾーンログに POST 自体が無いため、この経路は今回の事象の原因ではない。
（残る可能性: `*.workers.dev` 直アクセスはゾーン統計に出ないが、同一 Worker/DB
へ書き込むため、成功していれば DB に現れる → 該当なし）

## 考えられる実際の経路（要ヒアリング）

1. **マニュアル検索結果を会話と認識** — HelpChat は検索ヒットを sources として
   表示する。LLM 回答ではなく検索結果だけを見た可能性
2. **エラーターン表示** — POST 失敗時はフロントに明示的エラーが出る
   （サイレントフォールバックなし）
3. **別環境での利用** — prod は無トラフィック。local/他URL なら別 DB
4. 送信前に離脱・通信断

## exchanges の全件スナップショット（2026-10-05 21:00 JST 時点・8件）

| created_at (JST) | clinic | staff | question |
|---|---|---|---|
| 10/05 20:55 | 八王子病院 | 林 文明 | 疎通テスト2（本調査で投入） |
| 10/05 20:49 | 八王子病院 | 林 文明 | 疎通テスト（同） |
| 10/02 18:25 | 城東センター病院 | 林 文明 | 診察なし薬のみ受け取りに当日来る場合の使い方 |
| 10/02 14:22 | 城東センター病院 | 林 文明 | 主訴テンプレート |
| 10/02 14:22 | 城東センター病院 | 林 文明 | 問診テンプレート |
| 10/01 17:54 | 八王子病院 | 林 文明 | ワクチン間違えて入力してしまった 削除できない？ |
| 10/01 17:54 | 八王子病院 | 林 文明 | wakt |
| 09/30 19:22 | 八王子病院 | 林 文明 | カルテの作成方法は？ |

## ログから会話本文を抽出できない理由（構造的制約）

- 保存失敗 WARN には `clinic_id`/`staff_id`/`error` のみ。本文は PII 対策で非記録
- リクエストボディは Worker/Container どのログ経路にも出力されない
- 本文が残存する唯一の場所: 話者のブラウザ React state（タブが生きていれば
  リロード前にコピーで回収可能）/ LLM プロバイダ（Z.AI）コンソール側

## ログアクセス手段の棚卸し

| 経路 | 状態 |
|---|---|
| CF GraphQL Analytics（ゾーンHTTP・Worker集計） | **可**（wrangler OAuth で動作） |
| `wrangler tail`（ライブ） | 可。Worker 側ログのみ。Container 内 Go slog は実測で流れない |
| Workers Observability 過去ログ API | OAuth スコープ不足(403)。`.env*` の API token は失効(401) |
| PlanetScale 直読 | `pscale` token 失効。DB 認証は wrangler secret で読戻し不可 |

## 推奨アクション

1. 話者へ確認: AI の文章回答が返ったか / エラー表示だったか / どの URL で使ったか
2. 再現テスト: 該当スタッフに1件送信してもらい `wrangler tail` で観測
3. （恒久対策・別件提案）`RecordChatExchange` 失敗を WARN のみでなく
   audit_log 等へ記録する改修は検討余地あり — ただし本文を残すかは PII 判断要
