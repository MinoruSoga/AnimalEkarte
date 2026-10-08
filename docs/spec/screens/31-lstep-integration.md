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
> 追記（2026-10-07 監査）: 経路別の失敗契約と二重 gate の契約本文は正本へ集約した。本節では契約の重複掲載をやめ、参照先のみを示す。

- **失敗契約の三分法**（request-local nonfatal secondary notification / single-owner propagation / scheduled multi-resource best effort — `BatchRunResult` と `processed_count`/`error_count` 監査による durable 計上を含む）の正本: [line/architecture.md](../line/architecture.md) §4。実装側の整理は [line/lstep-integration.md](../line/lstep-integration.md) §5。
- **Write 二重 gate**（deploy gate `LSTEP_WRITE_API_ENABLED` / clinic gate `is_sync_enabled` / API キー — `ErrWriteDisabled` とクライアント未構築の意図的スキップ `nil, nil` は別契約）: 契約本文は [line/architecture.md](../line/architecture.md) 冒頭注記。環境変数の評価規則（exact `true` のみ有効）と enable / stop / rollback 手順の正本は [`LSTEP_WRITE_API_PAUSE.md`](../../ops/deploy/LSTEP_WRITE_API_PAUSE.md)（本 spec に手順・環境実値を複製しない）。
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
