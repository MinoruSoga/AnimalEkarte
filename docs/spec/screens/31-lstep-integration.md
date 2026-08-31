# Lステップ連携・管理 仕様書 (L-Step Integration)

## 概要
- **画面の目的**: Lステップ（LINE公式アカウント拡張ツール）との高度な連携設定、自動配信タグの管理、およびマーケティング施策の効果分析。
- **URLパターン**: 
  - 連携設定: `/settings/integrations/lstep`
  - タグ管理: `/settings/lstep/tags`
  - 健診対象者抽出: `/lstep/checkup-sync`
  - 分析レポート: `/lstep/analytics`
- **アクセス権限**（FE）:
  - 連携設定 `/settings/integrations/lstep`: `ResourceHospitalSettings`
  - タグ管理 `/settings/lstep/tags`: **`ResourceLstepAnalytics`**
  - 健診同期 `/lstep/checkup-sync`: `ResourceHospitalSettings`
  - 分析 `/lstep/analytics`: `ResourceLstepAnalytics` のみ
  - 親 `/lstep` は意図的に権限ガードを持たず、各子ルートが独立したガードを持つ。FE のルートガードとは別に、各 API は下表の BE 権限で認可する。

---

## 1. 連携設定 (Integration Settings)

### 1.1 API 設定
同一ページ（`/settings/integrations/lstep`）に接続情報・同期トグル・CPM/予防閾値・Messaging 接続テスト・設定削除・トリガー優先度・タグコード対応・自動管理プレフィックスがある。配信監視は [34-lstep-delivery-monitor.md](./34-lstep-delivery-monitor.md)。
- **Channel Access Token**: Messaging API 通信用の長期トークン。
- **LステップベースURL**: Lステップ側 API の接続先ベースURL。プレースホルダーは `https://api.lstep.jp`（許可ホスト）。`https://app.lstep.jp` は 400 で拒否されるため例示しない。

### 1.2 判定閾値設定 (CPM/LTV)
医院の運営方針に合わせ、以下の判定基準をカスタマイズ可能です。
- **CPM バージョン**: V1（金額＋回数）または V2（回数重視）を選択。
- **ステージ境界値**: 「コア」「ノア」等と判定するための累計売上額や来院回数の閾値。

### 1.3 飼主LINEアカウント連携

- スタッフが飼主単位の24時間link tokenを発行する。raw tokenは発行responseで一度だけ返し、DBにはunpadded base64url SHA-256 digestだけを保存する。
- LIFF公開endpointは16KiB以下の単一strict JSON objectだけを受け付け、`link_token`と`line_id_token`以外のfield（再紐付けを強制する`force`等）を拒否する。
- LINE ID token検証はrequest context、5秒timeout、64KiB response上限で行い、upstream response本文をclient errorへ転記しない。
- link token、対象clinic、飼主をtransaction内でlockし、既存LINE IDがある飼主への上書きを拒否する。飼主更新、tokenの単回CAS消費、成功監査は同一transactionでcommitし、いずれかが失敗した場合は全てrollbackする。
- 旧versionが発行済みの64桁hex raw tokenだけは期限内互換のため限定検索する。新規tokenを平文保存するfallbackは持たない。

---

## 2. タグ管理と自動同期

### 2.1 自動管理タグ
システムが自動的に付与・剥がしを行うタグ群。
- **CPM ステージ (V2)**: `CPM_01_出会い`, `CPM_02_これから`, `CPM_03_いいかんじ`, `CPM_04_ファミリー`, `CPM_05_ノア`（V1 選択時は `cpm_encounter` / `cpm_growing` / `cpm_core` / `cpm_spot` / `cpm_noah` / `cpm_dormant` の英数字タグ）。
- **来院間隔**: `VISIT_120日超`, `VISIT_180日超` 等。
- **属性**: `has_dog`, `has_cat`（犬猫両方飼育は `has_both`）, `LTV_上位20`, `LTV_フード購入あり`。

### 2.2 健診対象者一括同期 (`CheckupSyncPage`)
- **抽出ロジック**: 健診種別・動物種・最終来院日・年齢・慢性疾患有無・CPM ステージ・累計診療費・年間来院回数・最終健診実施日を任意に組み合わせて絞り込むフォーム方式（固定条件ではない）。
- **アクション**: 抽出リストから送信可能な対象者を選択し、任意のタグ名（例: `checkup_annual_2026`）を指定して一括付与、Lステップ側のセグメント配信へ繋げます。

---

## 3. 技術仕様

### 3.1 同期エンジンと失敗契約
経路ごとに失敗契約を混同しない（詳細: [line/architecture.md](../line/architecture.md) §4、[line/lstep-integration.md](../line/lstep-integration.md) §5）。

| 経路 | 契約 | 画面・運用への影響 |
|:---|:---|:---|
| 会計完了・ペット登録・死亡記録などイベント直後のタグ更新 | **request-local nonfatal secondary notification** | 本処理（会計等）は成功のまま。タグ同期失敗はログのみで本処理を反転させない。**配信トリガーログには書かない** |
| 1 飼主分のタグ同期本体（画面操作やバッチ内 1 owner） | **single-owner propagation** | 望ましいタグ Add/Remove 失敗は error 伝播。呼び出し元が失敗を観測・計上する |
| 定時バッチ（dormant / no_show / delivery / LTV 等 multi-resource） | **scheduled multi-resource best effort** | 1 件失敗後も続行。`BatchRunResult` と `processed_count`/`error_count` 監査による **durable 部分結果計上**が必須。必須 dependency 欠落は fail-closed |

- **Deploy gate（`LSTEP_WRITE_API_ENABLED`）**: Write 系（AddTag / RemoveTag / AddTagBulk / SetProperty）は exact `"true"` のときだけ HTTP を送る。未設定・空・`"false"`・未知値は無効。無効時は **`ErrWriteDisabled` を返し HTTP を送らない（`nil` 成功にしない）**。内部タグキャッシュ・判定・監査のアプリ内更新は経路により継続し得るが、Lステップ側実タグは変わらない。enable / stop / rollback の手順正本は [`LSTEP_WRITE_API_PAUSE.md`](../../ops/deploy/LSTEP_WRITE_API_PAUSE.md)（本 spec に手順・環境実値を複製しない）。
- **Clinic gate（`is_sync_enabled` / API キー）**: `is_sync_enabled=false` または API キー未設定の clinic は `buildClient` がクライアントを構築せず `nil, nil` を返す（意図的スキップ）。deploy gate の `ErrWriteDisabled` とは**別契約**である。
- **バッチ同期（scheduler/cron）**: Cloudflare scheduled event の式と job 対応は code/config へ配線済み（毎日 02:00 JST に `dormant`、10:00 JST に `no_show`→`delivery`、15:00/20:00 JST に `no_show`。durable coordinator、重複防止、pause/resume、missing-slot catch-up、失敗通知を含む）。**配線済みであることと、対象環境（STG/production）での自然発火・実送信・運用 rehearsal が完了していることは別事実である。** 後者は release gate として未実測（[Scheduler Operations](../../ops/deploy/runbooks/SCHEDULER_OPERATIONS.md)）。
- **配信トリガー候補の読み取り**: clinic スコープ bulk-read を必須とし、owner ループ内の N+1 読み（owner / 当日 claim / 抑制 / tag-cache）を置かない。opt-out・suppression・daily-claim 意味論と bounded memory は維持する。
- **流量**: Messaging API / Lステップ API のレート制限は固定のクライアント方針と監視で扱う。バッチとリアルタイムを動的に切り替える rate adjustment は持たない。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/clinics/:clinic_id/lstep-settings` | 連携設定と判定閾値の取得 | `hospital-settings` | `view` |
| PATCH | `/api/v1/clinics/:clinic_id/lstep-settings` | 連携設定の更新 | `hospital-settings` | `edit` |
| DELETE | `/api/v1/clinics/:clinic_id/lstep-settings` | 連携設定の削除 | `hospital-settings` | `delete` |
| POST | `/api/v1/clinics/:clinic_id/lstep-settings/test-connection` | 接続テスト | `hospital-settings` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/trigger-priorities` | 配信トリガー優先順位の参照 | `hospital-settings` | `view` |
| PATCH | `/api/v1/clinics/:clinic_id/lstep/trigger-priorities` | 配信トリガー優先順位の更新 | `hospital-settings` | `edit` |
| GET | `/api/v1/clinics/:clinic_id/lstep/tag-summary` | 現在のタグ保有者数の統計取得 | `lstep-analytics` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/owners` | 指定タグの飼主一覧取得（タグ管理） | `lstep-analytics` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/checkup-sync/preview` | 健診対象者抽出条件のプレビュー取得 | `owners` | `view` |
| POST | `/api/v1/clinics/:clinic_id/lstep/checkup-sync` | 指定条件の対象者へのタグ一括付与 | `owners` | `edit` |
| GET | `/api/v1/clinics/:clinic_id/lstep-tag-code-mappings` | タグ名ごとの外部コード紐付け取得 | `hospital-settings` | `view` |
| PUT | `/api/v1/clinics/:clinic_id/lstep-tag-code-mappings/:tag_name` | タグ別コード紐付け更新 | `hospital-settings` | `edit` |
| GET | `/api/v1/lstep-tag-config/auto-managed-prefixes` | 自動管理プレフィックス一覧取得 | `hospital-settings` | `view` |
| POST | `/api/v1/lstep-tag-config/auto-managed-prefixes` | 自動管理プレフィックス追加 | `hospital-settings` | `create` |
| DELETE | `/api/v1/lstep-tag-config/auto-managed-prefixes/:id` | 自動管理プレフィックス削除 | `hospital-settings` | `delete` |
| GET | `/api/v1/lstep-tag-config/condition-tag-mappings` | 条件別タグマッピング一覧取得 | `hospital-settings` | `view` |
| POST | `/api/v1/lstep-tag-config/condition-tag-mappings` | 条件別タグマッピング追加。慢性疾患コード重複は `localizeAlreadyExistsMessage` がコード値を含む日本語にする（例: 「慢性疾患コード『ckd』は既に使用されています」） | `hospital-settings` | `create` |
| DELETE | `/api/v1/lstep-tag-config/condition-tag-mappings/:id` | 条件別タグマッピング削除 | `hospital-settings` | `delete` |
| GET | `/api/v1/lstep-tag-config/send-purpose-tag-prefixes` | 送信目的別タグプレフィックス一覧取得 | `hospital-settings` | `view` |
| POST | `/api/v1/lstep-tag-config/send-purpose-tag-prefixes` | 送信目的別タグプレフィックス追加 | `hospital-settings` | `create` |
| DELETE | `/api/v1/lstep-tag-config/send-purpose-tag-prefixes/:id` | 送信目的別タグプレフィックス削除 | `hospital-settings` | `delete` |
| POST | `/api/v1/owners/:id/line/link-token` | 飼主LINE連携用の単回token発行 | `owners` | `edit` |
| POST | `/api/liff/:clinicId/link` | LIFFでLINE ID tokenを検証し飼主へ原子的に紐付け | 公開token検証 | — |
| GET | `/api/v1/clinics/:clinic_id/lstep/analytics/delivery-stats` | 月次配信統計の取得 | `lstep-analytics` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/analytics/visit-conversion` | 来院転換データの集計 | `lstep-analytics` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/csv-imports` | 友だち属性 CSV インポート履歴の取得 | `lstep-csv-import` | `view` |
| POST | `/api/v1/clinics/:clinic_id/lstep/csv-imports/friend-attributes` | 友だち属性 CSV のアップロード | `lstep-csv-import` | `edit` |

---
