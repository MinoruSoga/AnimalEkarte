# LINE連携・MA アーキテクチャ (LINE & Marketing Automation)

> **目的**: LINE連携の同期フロー・認証設計を定義する。
> **読者**: バックエンド実装者。
> **タイミング**: 同期フローの実装・確認時。

> **Animal Ekarte**: LINE プラットフォームを活用した顧客体験の最大化
> **最新更新**: 2026-09-06（repo `7c6592f9f`）

---

> **注記 (2026-07-10 / 更新 2026-07-31)**: Lステップへの Write API（タグ付与・タグ解除・プロパティ更新）は **deploy gate `LSTEP_WRITE_API_ENABLED`（既定 OFF）+ clinic `is_sync_enabled` の二重 gate** で抑止される。gate OFF 時は外部 HTTP write を送らず `ErrWriteDisabled` を返す（silent nil 成功ではない）。その後のローカル DB / cache / audit / logging は呼び出し元ごとの失敗契約に従い、継続を一律には保証しない。詳細: [`docs/ops/deploy/runbooks/LSTEP_WRITE_API_PAUSE.md`](../../ops/deploy/runbooks/LSTEP_WRITE_API_PAUSE.md)

---

## 1. 二大機能の棲み分け

本システムにおける LINE 連携は、目的の異なる 2 つのサブシステムで構成されます。

### 1.1 LINE 予約システム (Inbound)
- **目的**: 飼主からの「予約」を自動で受け付ける。
- **技術**: LIFF (LINE Front-end Framework) + React 19。
- **フロー**: LIFF 経由で空き枠を検索し、カルテ側の `appointments` テーブルへ直接書き込み。

### 1.2 Lステップ連携 (Outbound / CRM)
- **目的**: 病院側から「最適な情報」を自動で届ける。
- **技術**: (a) イベント起因の同一リクエスト内タグ同期、(b) scheduled/cron の multi-resource バッチ、(c) 手動送信後の副次タグ付与、および L-Step API / Messaging API。
- **フロー**: 会計・診察等のイベントや定時バッチでタグ／配信トリガーを更新し、Lステップ側シナリオを起動する。経路ごとに失敗契約が異なる（§3・§4）。

---

## 2. 認証・紐付けロジック (Identity Mapping)

顧客の特定は次の経路で行われる（**予約時の氏名+電話による自動オーナー紐付けは行わない** — SEC-CS2-F02）。

1.  **LIFF ログイン**: LINE アプリからのアクセス時に ID トークン検証で `line_user_id` / `line_customers` を解決（未登録なら顧客行を作成）。
2.  **スタッフ発行 link token**: `POST /api/v1/owners/:id/line/link-token` で raw token を一度だけ返し、DB には unpadded base64url SHA-256 digest のみ保存（`line_link_tokens`）。
3.  **LIFF 紐付け**: `POST /api/liff/:clinicId/link` が `link_token` + `line_id_token` で自己認証し、飼主へ LINE User ID を紐付け（既存 LINE ID の上書き拒否、token 単回 CAS 消費、監査同一 TX）。
4.  **スタッフ手動リンク**: `PATCH .../line-customers/:id/link-owner` による院内オペレーション。

### Webhook 契約（要約）

Webhook（`POST /api/line/webhook`）の **executable SoT は code/tests**（`backend/internal/lstep/line_link_service.go` の `HandleWebhook` と `line_link_service_test.go` / signature routing tests）。本書は境界の要約と正本リンクのみを持つ（PO FINAL R-01: option B。詳細 event 表は二重管理しない）。

| 項目 | 現行境界 |
|------|----------|
| 処理する event type | `follow` / `unfollow` のみ（business side effect あり） |
| その他 event type | business side effect なしで skip（成功扱い） |
| 署名 | event 処理より先。`destination`（bot user ID）で clinic を一意解決し、当該 clinic の secret で最大 1 回 HMAC。全 clinic 走査・複数 secret 試行・destination 欠落時の推測は禁止 |
| 署名不成立 / setting 欠落 / DB failure | fail-closed（invalid input / 処理拒否） |
| 新規 event / message reply | 本要約から発明しない。set 変更時は code・contract test・本要約を同時更新 |

Provisioning・疎通・停止の gate checklist は [setup.md](./setup.md) を参照。そこでは Messaging API / LINE Login (LIFF) / webhook / destination routing / deploy・clinic gate / rollback を個別に確認する。

---

## 3. イベント起因タグ同期のライフサイクル（request-local）

会計確定などの **本処理** に対し、CPM/VISIT 等のタグ同期は同一リクエスト内で同期呼び出しされる（別プロセス/goroutine のワーカーではない）。これは **request-local nonfatal secondary notification** 契約である。

- タグ同期失敗はログするだけで、会計など本処理の成功契約を反転させない。
- 記録先は `slog` とタグキャッシュ（必要時は API 失敗カウンタ）。**`lstep_delivery_trigger_log` には書かない**（当該テーブルは自動配信トリガー専用。`audit_logs` もこの ordinary-sync 経路では書かない）。
- Write dual-gate 無効時は外部 HTTP write を送らない（`ErrWriteDisabled`）。後続のローカル DB / cache / audit / logging は呼び出し元の失敗契約に従う。たとえば手動タグ追加は外部 AddTag 成功後に cache / audit を更新するため、`ErrWriteDisabled` 時には継続しない。

```mermaid
sequenceDiagram
    participant Billing as Accounting Service (Go)
    participant Sync as Tag Sync (同一プロセス内・同期呼び出し)
    participant LS as L-Step API

    Billing->>Sync: 会計確定後の CPM 同期 (同期呼び出し, nonfatal)
    Sync->>Sync: 顧客状態再計算 (CPM/VISIT)
    Sync->>LS: タグ付与/解除リクエスト（deploy/clinic gate OFF 時は HTTP 0 + ErrWriteDisabled）
    LS-->>Sync: HTTP 200 OK（両 gate 有効時のみ）
    Sync->>Sync: slog / tag cache / 失敗カウンタ（trigger log ではない）
```

---

## 4. 失敗契約の三分法（LSTEP）

LSTEP 周辺の「best-effort」は一語で混同しない。経路ごとに契約を固定する（正本: `backend/CODING_RULES.md` の LSTEP バッチ条項）。

| 契約 | 代表経路 | 失敗時の振る舞い | 耐久計上 |
|:---|:---|:---|:---|
| **1. scheduled multi-resource best effort** | 定時バッチ（delivery / dormant / LTV / health-prevention 等）。multi-clinic・multi-owner・multi-trigger | 必須 dependency 欠落・clinic 一覧取得失敗は fail-closed。1 件失敗後も他 resource は続行 | 必須。`(processed, errs)` と durable 向け `BatchRunResult`（`Processed = Succeeded + Failed`）。error ログ、`processed_count`/`error_count` 監査、Partial/Failed 時の手動 fallback・再実行/idempotency |
| **2. single-owner propagation** | 1 飼主分のタグ同期本体（`SyncCPMStageTag` / VISIT / food 等）。バッチ内の 1 owner 処理も含む | 望ましい Add/Remove の失敗は **呼び出し元へ error 伝播**し、上位の Failed 計上に載せる。stale 除去の部分失敗は desired add を継続しつつ失敗計上を残す | 呼び出し元が batch なら `BatchRunResult`/監査へ。単独呼び出しなら error 返却で観測可能にする（silent `return 0, nil` は新規禁止） |
| **3. request-local nonfatal secondary notification** | 会計確定後の CPM 同期、LINE 手動送信後の purpose タグ付与、次回予定（tentative intent）の設定・解除・昇格・owner 変更に伴う post-commit の Lステップ同期など、本処理に付随する副次 side-effect | 本処理は成功のまま。副次失敗はログ必須で本処理を rollback しない | 本処理の成功契約を反転させない。副次失敗の収束/補償を短く明示（次回イベント再同期・失敗カウンタ等）。**trigger log は使わない** |

### 4.1 自動配信トリガーバッチと durable 計上

- 10:00 JST の配信トリガー等は契約 1。owner ループは continue-on-error、1 owner の失敗は契約 2 で伝播し上位が Failed に加算する。
- 配信実行・除外・優先度抑制・API 失敗は `lstep_delivery_trigger_log` に残し、配信監視画面の観測源とする。
- 候補 owner に対する owner / 当日 claim / 抑制 / tag-cache 読みは通常、clinic スコープの bulk-read を使う。production には bulk failure 時の per-owner day-log / owner / tag-cache read fallback が残る。この degraded mode は owner 数に比例する既知の性能 gap であり、実行上限・metrics を備えた bounded fallback への整理または source からの除去が必要。opt-out・suppression・daily-claim の意味論と bounded memory は維持する。

### 4.2 checked-in scheduler と運用証跡

[`backend/worker/scheduled-jobs.ts`](../../../backend/worker/scheduled-jobs.ts) の cron plan は、10:00 JST に `no_show` → `delivery`、15:00/20:00 JST に `no_show`、02:00 JST に `dormant`。[`backend/wrangler.jsonc`](../../../backend/wrangler.jsonc) に cron 定義があり、Go 側は [`batch_scheduler.go`](../../../backend/cmd/api/batch_scheduler.go) の内部認証付き route へ接続する。これはリポジトリ上の配線であり、実環境の設定値・cron 発火・外部送信を検証した記録ではない。

[#259 の再開条件](README.md#3-再開条件259) は Write と cron を同一リリースで扱い、少数実送信と停止確認を要求する。HTTP 成功だけで終了せず、scheduled outcome の件数整合と外部結果を照合する。

### 4.3 停止手段

- clinic 単位の `is_sync_enabled=false` で同期・配信バッチ対象から外す。
- Write API は deploy kill switch `LSTEP_WRITE_API_ENABLED` と clinic flag の二重 gate（[`LSTEP_WRITE_API_PAUSE.md`](../../ops/deploy/runbooks/LSTEP_WRITE_API_PAUSE.md)）。再開後も `is_sync_enabled=false` の clinic はサービス層で抑止する。

---

## 5. データ分離とセキュリティ

- **マルチテナント**: 各医院の API キーは AES-256-GCM で暗号化され、`clinic_integrations` テーブルに隔離保存されます。
- **流量**: Messaging API / Lステップ API のレート制限はクライアント実装の固定方針と運用監視で扱う。**バッチとリアルタイムを動的に切り替える rate adjustment は持たない**（未実装の動的流量調整を前提にしない）。

---

### 次回来院タグの再試行

`SyncNextVisitTag` は外部タグ操作前に、医院・飼主単位の `lstep_next_visit_sync_retries` を独立 commit する。owner lock 内で現在の予約・LINE連携先・同意を読み直し、全タグ操作とcache更新が成功した場合だけ、自身が enqueue した revision をcacheと同じtransactionで削除する。cache保存・commit・取消・外部操作の失敗は再試行記録を保持し、別の同期が追加した新しいrevisionは消さない。日次 tentative reconciliation は期限切れcacheと再試行記録をUNIONし、医院ごとに重複排除したowner IDのkeyset paginationで処理する。以前のcacheがない飼主も対象となる。

連携解除・opt-out・同期停止中は既存の再試行記録を保持し、再開時に現在のsourceと連携先を再評価する。通常の再試行記録は過去のタグやLINE IDのsnapshotを持たないため、LINE ID変更前の旧アカウントに残ったタグの除去は保証しない。例外は飼主削除で、削除transaction内で確定したLINE宛先が `cleanup_line_user_id` として不変に記録され、削除commit後の `SyncNextVisitTag` は外部操作をremove-onlyの `next_visit_*` 除去に限定する（予約参照・AddTagは行わない）。宛先が別のactive飼主へ再連携済みなら外部操作を行わず記録を保持する。新規migration `033` / `034` は通常のmigration手順で適用してからこの同期コードを稼働させる。臨床記録ではない再試行用の技術テーブルだけを成功時にhard deleteする。

<a id="followup-confirmation-adapter"></a>
## 初診・健診後フォローの確認adapter（2026-10-15目標）

[D15の承認済みFR-crm-followup-confirmation](lstep-integration.md#requirement-fr-crm-followup-confirmation)を実現する境界。既存の作成後welcome／checkup followupと定時初診候補は現行経路であり、記録作成だけで送らない目標への追従を完了扱いしない。コード／DB／配信設定は今回変更していない。

| 接続点 | 目標と現行差 | D15 AC suffix（AC-crm-followup-） |
|:---|:---|:---|
| D03／D09保存後→候補 | 保存済み元事象とスタッフの実施済み・送信対象確認を別に識別。Path Bは段階保存で全体atomicに変えず、未確定段階は送信へ進めない | record-only, partial-record |
| 案内目的／医師指定 | 一般お礼と医療的再検査・受診案内を区別。同一対象への医師要フォロー指定を医療的案内で検証し、結果数値や担当医を代用しない。現checkup callのpredicate不在は適合差 | general, medical |
| 確認保存／認可 | 元clinic・owner・pet・record／checkup・案内目的・保存結果に確認者／時点を結び、必要監査と同じ保存結果。既存grant・認証世代・依存照合／receiptを継承。DTO／route／格納形式は未実装・要適合確認 | confirmation-failure |
| claim前／外部作用直前 | 既存source再読へ確認・目的・医師指定の現状態を接続し、欠損／不明／取消で停止。直接sendだけでなく該当配信を起動する外部タグ／シナリオにも適用。opt-out・死亡由来・LINE・既存gate・優先度を緩和しない。現生死遷移との瞬間競合gapは保持 | source-changed |
| 再実行／結果unknown | 元事象と確認の反復を送信重複へ変換せず、現claimと送信結果照合へ接続。外部試行後blockingを保持し、臨床保存と外部HTTPを一つのtransactionにしたとは扱わない | duplicate-unknown |

送信に必要な確認・監査／結果の対応が読み取れない場合は保留する。旧ログのscheduled／firedだけではスタッフ確認や医師指定の成立証明にならない。新しい同意条件・臨床閾値・送信時刻をこのadapterで決定しない。


<a id="future-phase-followup-adapter-20261019"></a>
## 対象Phase開始時の内部adapter候補（未実行）

IR-07/IR-D01/IR-D02はLSTEP/DM対象Phase開始時だけの残工程で、今期必須gateへ加算しない。現PhaseのLINE予約・問診を対象外へ広げない。

将来の初診/健診後followup入力候補はsource owner/record ref・source revision・実施済みstaff確認ref・医療目的の医師指定ref・binding revision・operation key。D03/D09の保存済み事象をreadするadapterであり、単なるrecord作成やtag付与を実施確認・送信許可としない。source取消/訂正・確認版変更ならclaim前に再照合、旧版/結果不明を発火可へ変換しない。送信caller表に事象/目的/DM現在値/独立停止/binding/claim/外部結果を列別に持ち、未分類callerを予約通知例外へ落とさない。DM nullable、lstep_opt_out、delivery_excludedは独立に維持する。

将来のenable判定は先方確認・実送受信・cron停止・件数の同一版証跡で行う。HTTP成功やrepo配線だけで稼働済みにしない。今回は実送信・外部enable・cron操作なし。採用済みの実施確認/医師目的の業務を再質問せず、具体DTO/保存ref/全17trigger対応と実機証跡を対象Phaseの技術作業へ残す。
