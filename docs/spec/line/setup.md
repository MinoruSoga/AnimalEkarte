# LINE・Lステップ 連携セットアップガイド (External Integration Setup)

> **目的**: LINE Developers Console と Animal Ekarte の初期設定 gate を提供する。
> **読者**: クリニック導入担当・運用者。
> **タイミング**: 新規クリニックの LINE 連携初期設定時。
> **最新更新**: 2026-08-31

この手順は実値を記録する場所ではない。環境ごとの public host、clinic ID、credential owner、rollback owner は承認済みの環境台帳で確認する。

## 1. 事前準備と credential 取扱い

必要な権限:
- LINE Developers / LINE Official Account Manager の管理権限
- Lステップ管理権限
- Animal Ekarte clinic 管理権限
- deployment gate と ops-provisioned field を扱える運用担当

Channel Secret、access token、API key、LIFF の実値を ticket、chat、log、screenshot、本文へ貼らない。管理画面は保存済み secret を mask し、blank 入力では既存値を変更しない契約として扱う。rotation と rollback の owner を開始前に決める。

## 2. LINE Developers Console

### 2.1 Messaging API channel（bot / webhook）

1. provider と LINE Official Account の関係を確認する。
2. Messaging API channel を作成または選択する。
3. Channel ID、Channel Secret、Channel Access Token を secret store へ登録する。
4. Webhook URL を `<API_PUBLIC_HOST>/api/line/webhook` に設定し、まだ enable しない。

### 2.2 LINE Login channel（LIFF）

LIFF app は Messaging API channel 内ではなく **LINE Login channel** に作成する。

1. 同じサービスで使う LINE Login channel を作成または選択し、Messaging API channel / Official Account との必要な linkage を LINE console で確認する。
2. LIFF app を追加し、scope `openid`, `profile` を許可する。
3. Endpoint URL を `<PUBLIC_HOST>/line-reserve/<clinicId>/`（末尾 slash 付き）にする。bare host は clinic を解決できない。repository 内の route 契約は [owner flow](../screens/37-line-reserve-owner-flow.md)、検証入口は [LIFF verification](../../ops/testing/liff-verification.md)。public host の環境対応は repository 外の承認済み環境台帳を正とし、未確認 hostname を本書へ固定しない。
4. LIFF ID を secret/config store へ登録する。

## 3. Lステップ

1. Lステップ API key を発行し secret store へ登録する。
2. managed tag code / admin mapping と source constants を照合する。代表例は CPM / VISIT と `PREV_ワクチン期限`。この文書に tag 全件を複製しない。

## 4. Animal Ekarte の ordered setup gates

以下を順番に満たす。前段が未確認なら外部 write を enable しない。

1. **LINE 予約設定**: `/line-reservation/settings` で受付状態、LINE Channel ID、LIFF ID（非secret）を登録する。Channel Secret / Channel Access Token はこの画面では扱わず、手順2のLステップ設定で保存する。
2. **Lステップ設定**: `/settings/integrations/lstep` で API key、base URL、Channel Access Token、Channel Secret、LIFF ID を保存する。
3. **destination routing**: 運用担当が `line_reservation_settings.line_bot_user_id` を ops / migration 経路で一意に provision する。この field は通常 UI/API save の対象外で、空のままでは webhook `destination` から clinic を解決できない。実値や直接 DB 操作を本書へ記載しない。
4. **webhook verification**: `POST /api/line/webhook` に対し、対象 channel の署名付き test event で destination routing と署名検証を確認してから LINE console の webhook を enable にする。
5. **LIFF verification**: clinic path から LIFF login、clinic 解決、予約設定取得までを別に確認する。connection-test button はこの確認を代替しない。
6. **clinic gate**: clinic の `is_sync_enabled` は、上記と監視・rollback owner が揃うまで false にする。
7. **deploy gate**: `LSTEP_WRITE_API_ENABLED` は既定 OFF。enable / stop / rollback は [LSTEP Write API pause runbook](../../ops/deploy/LSTEP_WRITE_API_PAUSE.md) に従う。環境変数の実値は本書へ記載しない。
8. **release evidence**: scoped external scenario、monitoring、alert、stop/rollback を記録する。generic staging gate は [STG_PRE_DEPLOY_READINESS_CHECK](../../ops/deploy/runbooks/STG_PRE_DEPLOY_READINESS_CHECK.md) を参照する。

## 5. 現行 connection test の範囲と既知 gap

画面の「Lステップ接続テスト」と「LINE接続テスト」は、どちらも `POST .../lstep-settings/test-connection` を呼ぶ。現行 backend が probe するのは次のみ:

- Lステップ: API key + base URL
- LINE Messaging API: Channel Access Token で `/v2/bot/info`

**検証しないもの**: Channel Secret、webhook signature/destination routing、LIFF ID、LIFF login、clinic path、deploy/clinic write gate。

**既知の source/UI gap（本 docs-only 変更では未修正）**:
- probe helper は 401/403 以外を成功扱いし得るため、404 / 429 / 5xx を正しく失敗にできない。
- backend は `{lstep_ok:false}` / `{line_ok:false}` を含んでも HTTP 200 を返し得る。
- frontend は typed body を評価せず、2xx を toast success として扱う。

したがって button の success は設定全体の合格証明ではない。各 probe は 2xx のみを成功とし、frontend が component result を確認する source 修正が別途必要。現状は webhook signature test と LIFF login test を独立して実施し、結果を release evidence に残す。

## 6. 停止・rotation

異常時は clinic gate を停止し、必要に応じて deploy kill switch を OFF にする。credential 漏えいまたは期限切れでは owner が provider 側で rotate し、masked save、個別 probe、webhook/LIFF verification の順で再確認する。外部 write の再開は pause runbook の gate を満たした後だけ行う。
