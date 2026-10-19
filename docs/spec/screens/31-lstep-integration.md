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

> 追記（2026-10-07 再監査）: link token / LIFF 紐付けの契約本文（digest-only 保存・LIFF endpoint 制約・ID token 検証・transaction lock・単回 CAS 消費・legacy 64 桁 hex 互換）は [line/architecture.md](../line/architecture.md) §2 が正本。本節は業務概要のみを示す。

スタッフが飼主単位の 24 時間有効な link token を発行し、飼主が LIFF で本人認証（`link_token` + `line_id_token`）して LINE アカウントを紐付ける。既存 LINE ID の上書きは拒否され、紐付け操作は監査ログに記録される。

2026-10-15追加合意5の[主飼主変更](../domains/owners-pets/lifecycle.md#requirement-fr-owners-pets-primary-transfer)では、今後の予約とLINE通知先への影響をスタッフが一覧で確認してから切り替える。この確認はLINE本人紐付けの発行／確認／上書きの代わりにならず、電話・メールの重複候補から本人を自動判定しない。未リンクmanual予約の自動新規連携、配信除外・停止gate・結果不明保留の解除、通知の自動再送を加えない。条件付きの予約customer付替え／解除と現行差は[D04技術契約](../../design/domains/owners-pets/technical-spec.md#owner-transfer-target)へ。過去請求・預り金は元ownerに残る。

---

## 2. タグ管理と自動同期

### 2.1 自動管理タグ
システムが自動的に付与・剥がしを行うタグ群。
- **CPM ステージ (V2)**: `CPM_01_出会い`, `CPM_02_これから`, `CPM_03_いいかんじ`, `CPM_04_ファミリー`, `CPM_05_ノア`（V1 選択時は `cpm_encounter` / `cpm_growing` / `cpm_core` / `cpm_spot` / `cpm_noah` / `cpm_dormant` の英数字タグ）。
- **来院間隔**: `VISIT_120日超`, `VISIT_180日超` 等。
- **属性**: `has_dog`, `has_cat`（犬猫両方飼育は `has_both`）, `LTV_上位20`, `LTV_フード購入あり`。

### 2.2 健診対象者一括同期 (`CheckupSyncPage`)
- **抽出ロジック**: 健診種別・動物種・最終来院日・年齢・慢性疾患有無・CPM ステージ・累計診療費・年間来院回数・最終健診実施日を任意に組み合わせて絞り込むフォーム方式（固定条件ではない）。
- **アクション**: 抽出リストから送信可能な対象者を選択し、任意のタグ名（例: `checkup_annual_2026`）を指定して一括付与、Lステップ側のセグメント配信へ繋げます。案内・販促の配信を起動する場合は[DM目的別条件](../line/lstep-integration.md#crm-dm-purpose-contract)を満たし、DM未設定／不要を抽出・タグ付与経由で送信対象にしません。ordinaryな状態同期とは区別します。

---

## 3. 技術仕様

### 3.1 同期エンジンと失敗契約
> 追記（2026-10-07 監査）: 経路別の失敗契約と二重 gate の契約本文は正本へ集約した。本節では契約の重複掲載をやめ、参照先のみを示す。

- **失敗契約の三分法**（request-local nonfatal secondary notification / single-owner propagation / scheduled multi-resource best effort — `BatchRunResult` と `processed_count`/`error_count` 監査による durable 計上を含む）の正本: [line/architecture.md](../line/architecture.md) §4。実装側の整理は [line/lstep-integration.md](../line/lstep-integration.md) §5。
- **Write 二重 gate**（deploy gate `LSTEP_WRITE_API_ENABLED` / clinic gate `is_sync_enabled` / API キー — `ErrWriteDisabled` とクライアント未構築の意図的スキップ `nil, nil` は別契約）: 契約本文は [line/architecture.md](../line/architecture.md) 冒頭注記。環境変数の評価規則（exact `true` のみ有効）と enable / stop / rollback 手順の正本は [`LSTEP_WRITE_API_PAUSE.md`](../../ops/deploy/runbooks/LSTEP_WRITE_API_PAUSE.md)（本 spec に手順・環境実値を複製しない）。
- **バッチ同期（scheduler/cron）**: cron plan（`dormant` / `no_show`→`delivery` の時刻・durable coordinator・重複防止・pause/resume・missing-slot catch-up・失敗通知を含む）の正本は [line/architecture.md](../line/architecture.md) §4.2。**配線済みであることと、対象環境（STG/production）での自然発火・実送信・運用 rehearsal が完了していることは別事実である。** 後者は release gate として未実測（[Scheduler Operations](../../ops/deploy/runbooks/SCHEDULER_OPERATIONS.md)）。
- **配信トリガー候補の読み取り**: 読み取り要件（clinic スコープ bulk-read・owner ループ内 N+1 読み禁止・opt-out / suppression / daily-claim 意味論・bounded memory）の契約正本は [line/architecture.md](../line/architecture.md) §4.1。現行実装に残る bulk 失敗後の per-owner fallback と是正契約は [line/lstep-integration.md §5.2](../line/lstep-integration.md) を参照する。この劣化経路を含めた無条件の no-N+1 は未達である。
- **流量**: 固定クライアント方針と運用監視で扱い、バッチとリアルタイムを動的に切り替える rate adjustment は持たない旨の正本は [line/architecture.md](../line/architecture.md) §5。

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
| POST | `/api/v1/lstep-tag-config/auto-managed-prefixes` | 自動管理プレフィックス追加 | システム管理者のみ（RBAC 非対象・`requireSystemAdmin()`） | — |
| DELETE | `/api/v1/lstep-tag-config/auto-managed-prefixes/:id` | 自動管理プレフィックス削除 | システム管理者のみ（同上） | — |
| GET | `/api/v1/lstep-tag-config/condition-tag-mappings` | 条件別タグマッピング一覧取得 | `hospital-settings` | `view` |
| POST | `/api/v1/lstep-tag-config/condition-tag-mappings` | 条件別タグマッピング追加。慢性疾患コード重複は `localizeAlreadyExistsMessage` がコード値を含む日本語にする（例: 「慢性疾患コード『ckd』は既に使用されています」） | システム管理者のみ（同上） | — |
| DELETE | `/api/v1/lstep-tag-config/condition-tag-mappings/:id` | 条件別タグマッピング削除 | システム管理者のみ（同上） | — |
| GET | `/api/v1/lstep-tag-config/send-purpose-tag-prefixes` | 送信目的別タグプレフィックス一覧取得 | `hospital-settings` | `view` |
| POST | `/api/v1/lstep-tag-config/send-purpose-tag-prefixes` | 送信目的別タグプレフィックス追加 | システム管理者のみ（同上） | — |
| DELETE | `/api/v1/lstep-tag-config/send-purpose-tag-prefixes/:id` | 送信目的別タグプレフィックス削除 | システム管理者のみ（同上） | — |
| POST | `/api/v1/owners/:id/line/link-token` | 飼主LINE連携用の単回token発行 | `owners` | `edit` |
| POST | `/api/liff/:clinicId/link` | LIFFでLINE ID tokenを検証し飼主へ原子的に紐付け | 公開token検証 | — |
| GET | `/api/v1/clinics/:clinic_id/lstep/analytics/delivery-stats` | 月次配信統計の取得 | `lstep-analytics` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/analytics/visit-conversion` | 来院転換データの集計 | `lstep-analytics` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/csv-imports` | 友だち属性 CSV インポート履歴の取得 | `lstep-csv-import` | `view` |
| POST | `/api/v1/clinics/:clinic_id/lstep/csv-imports/friend-attributes` | 友だち属性 CSV のアップロード | `lstep-csv-import` | `edit` |

> 2026-09-29 訂正: `lstep-tag-config` 配下の POST/DELETE は `hospital-settings` の create/delete 権限 → **システム管理者専用**（`requireSystemAdmin()`、`is_system_admin` 判定）。GET は `hospital-settings:view` のまま。全医院共通の自動タグ設定を単一医院権限で変更できないよう SOLO-09 / LSA-04 / DEC-30 で意図的に閉じた（`backend/internal/lstep/routes.go`）。

---

<a id="crm-domain-owner-reference"></a>
## D15・D16の責務境界（2026-10-15）

タグ・健診同期・Webhook・停止／再開・外部副作用のFR/ACは[CRM正本のD15整理](../line/lstep-integration.md#crm-document-contract)を参照する。分析レポートの計数はD16、同画面からのCSVアップロード／インポートはD15のwriteであり、分析viewだけをアップロード権限へ拡張しない。CSVの列仕様・重複行／部分失敗・再取込、設定/タグ/健診/共有file/送信の詳細とACは[管理操作詳細](../line/lstep-integration.md#crm-operation-coverage)へ集約する。現行想定aliasと実export互換未検証を区別する。新規外部write・接続テスト・CSV取込は本作業で行わない。

[DM＝LINEの共同確認](../line/lstep-integration.md#crm-dm-purpose-contract)に従い、案内・販促は未設定／不要を抑止する。予約確認・変更・予約リマインドの例外でも既存アカウント・ブロック・本人／宛先・独立停止条件を維持する。会員区分変更で自動的に配信可否を変えず、退亡者を死亡と扱わない。希望欄の保存はD04が所有し、本画面に新たな設定writerを設けない。

## 分析の読取契約（D16）

/lstep/analytics の月次配信統計・配信後来院率は[経営・顧客集計の正本](../domains/analytics/business-rules.md#requirement-fr-analytics-lstep-observation)へ。現行統計はscheduled_atの月、来院率はfired_atの月に属する非抑止firedログが母数で、UIは30日窓。ユニーク飼主数・最終到達／既読・因果効果と同義にしない。同じ画面の友だち属性CSV取込はD15のwrite契約と独立権限であり、分析閲覧権限だけでは許可しない。

共有ファイルの手動削除／143時間cleanupと業務原本の保持は[D15の選定契約](../line/lstep-integration.md#crm-shared-file-retention)へ従う。管理delete権限があっても原本・必要履歴・分類不明を通常操作で物理削除できる意味ではない。現行storage先行deleteとの差は実装適合条件へ残す。

### 初診・健診後の確認と来院指標の境界

[D15の確認付きフォロー](../line/lstep-integration.md#crm-followup-confirmation)に従い、記録作成だけで送らず、スタッフが実施済み・送信対象を確認後に扱う。一般案内と医療的再検査／受診案内は分け、後者は医師の要フォロー指定が必要。健診対象preview・タグ同期・画面閲覧権限だけではこの確認／指定や送信許可を代行しない。現管理画面に新しい確認操作が実装済みとはしない。

累計・年間の来院回数と分析上の来院観測は[D16の実来院日基準](../domains/analytics/business-rules.md#requirement-fr-analytics-owner-counts)を参照する。同日複数ペットは飼主1回、ペットは各1回。旧カルテ日数による値は現行差として明示し、実来院の根拠不明を0回や「来院なし」へ自動補完しない。既存の分類閾値・観測窓・ログ母数と配信の認可／確認は別契約を維持する。
